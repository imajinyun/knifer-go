package hash

import (
	"encoding/binary"
	"hash/fnv"
	"slices"
	"sort"
	"strconv"
)

const (
	defaultVirtualNodes       = 100
	maxCollisionProbeAttempts = 1024
)

type consistentHashConfig struct {
	virtualNodes int
	hashFunc     func([]byte) uint64
}

// ConsistentHashOption customizes a consistent hash ring.
type ConsistentHashOption func(*consistentHashConfig)

// WithVirtualNodes sets the number of virtual nodes per real node.
func WithVirtualNodes(n int) ConsistentHashOption {
	return func(c *consistentHashConfig) {
		if n > 0 {
			c.virtualNodes = n
		}
	}
}

// WithReplicaCount sets the number of virtual nodes per real node.
func WithReplicaCount(n int) ConsistentHashOption { return WithVirtualNodes(n) }

// WithHashFunc sets the hash function used by the ring.
func WithHashFunc(hashFunc func([]byte) uint64) ConsistentHashOption {
	return func(c *consistentHashConfig) {
		if hashFunc != nil {
			c.hashFunc = hashFunc
		}
	}
}

// ConsistentHash maps keys to nodes with bounded movement when nodes change.
type ConsistentHash struct {
	cfg    consistentHashConfig
	ring   map[uint64]string
	hashes []uint64
	nodes  map[string]struct{}
}

// NewConsistentHash creates an empty consistent hash ring.
func NewConsistentHash(opts ...ConsistentHashOption) *ConsistentHash {
	cfg := consistentHashConfig{
		virtualNodes: defaultVirtualNodes,
		hashFunc:     fnv64a,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.virtualNodes <= 0 {
		cfg.virtualNodes = defaultVirtualNodes
	}
	if cfg.hashFunc == nil {
		cfg.hashFunc = fnv64a
	}
	return &ConsistentHash{
		cfg:   cfg,
		ring:  map[uint64]string{},
		nodes: map[string]struct{}{},
	}
}

// Add inserts a node into the ring.
func (h *ConsistentHash) Add(node string) {
	if h == nil || node == "" {
		return
	}
	if _, ok := h.nodes[node]; ok {
		return
	}
	h.nodes[node] = struct{}{}
	added := 0
	for i := 0; i < h.cfg.virtualNodes; i++ {
		key := node + "#" + strconv.Itoa(i)
		sum := h.cfg.hashFunc([]byte(key))
		for attempt := 0; attempt < maxCollisionProbeAttempts; attempt++ {
			if _, exists := h.ring[sum]; !exists {
				break
			}
			buf := make([]byte, 16)
			binary.BigEndian.PutUint64(buf[:8], sum)
			binary.BigEndian.PutUint64(buf[8:], uint64(attempt+1))
			sum = h.cfg.hashFunc(append([]byte(key+"#"), buf...))
		}
		if _, exists := h.ring[sum]; exists {
			continue
		}
		h.ring[sum] = node
		h.hashes = append(h.hashes, sum)
		added++
	}
	if added == 0 {
		delete(h.nodes, node)
		return
	}
	sort.Slice(h.hashes, func(i, j int) bool { return h.hashes[i] < h.hashes[j] })
}

// Remove deletes a node from the ring.
func (h *ConsistentHash) Remove(node string) {
	if h == nil {
		return
	}
	if _, ok := h.nodes[node]; !ok {
		return
	}
	delete(h.nodes, node)
	out := h.hashes[:0]
	for _, sum := range h.hashes {
		if h.ring[sum] == node {
			delete(h.ring, sum)
			continue
		}
		out = append(out, sum)
	}
	h.hashes = slices.Clip(out)
}

// Get returns the node responsible for key.
func (h *ConsistentHash) Get(key string) (string, error) {
	if h == nil || len(h.hashes) == 0 {
		return "", invalidInputf("consistent hash ring is empty")
	}
	sum := h.cfg.hashFunc([]byte(key))
	idx := sort.Search(len(h.hashes), func(i int) bool { return h.hashes[i] >= sum })
	if idx == len(h.hashes) {
		idx = 0
	}
	return h.ring[h.hashes[idx]], nil
}

// GetN returns up to n distinct nodes for key in ring order.
func (h *ConsistentHash) GetN(key string, n int) ([]string, error) {
	if h == nil || len(h.hashes) == 0 {
		return nil, invalidInputf("consistent hash ring is empty")
	}
	if n <= 0 {
		return []string{}, nil
	}
	if n > len(h.nodes) {
		n = len(h.nodes)
	}
	sum := h.cfg.hashFunc([]byte(key))
	idx := sort.Search(len(h.hashes), func(i int) bool { return h.hashes[i] >= sum })
	if idx == len(h.hashes) {
		idx = 0
	}
	out := make([]string, 0, n)
	seen := make(map[string]struct{}, n)
	for i := 0; len(out) < n && i < len(h.hashes); i++ {
		node := h.ring[h.hashes[(idx+i)%len(h.hashes)]]
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		out = append(out, node)
	}
	return out, nil
}

func fnv64a(data []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(data)
	return avalanche64(h.Sum64())
}

func avalanche64(x uint64) uint64 {
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	return x
}
