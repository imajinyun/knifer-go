package vhash

import hashimpl "github.com/imajinyun/knifer-go/internal/hash"

// ConsistentHashOption customizes a consistent hash ring.
type ConsistentHashOption = hashimpl.ConsistentHashOption

// ConsistentHash maps keys to nodes with bounded movement when nodes change.
type ConsistentHash = hashimpl.ConsistentHash

// Error represents an error produced by hash helpers.
type Error = hashimpl.Error

// WithVirtualNodes sets the number of virtual nodes per real node.
func WithVirtualNodes(n int) ConsistentHashOption { return hashimpl.WithVirtualNodes(n) }

// WithReplicaCount sets the number of virtual nodes per real node.
func WithReplicaCount(n int) ConsistentHashOption { return hashimpl.WithReplicaCount(n) }

// WithHashFunc sets the hash function used by the ring.
func WithHashFunc(hashFunc func([]byte) uint64) ConsistentHashOption {
	return hashimpl.WithHashFunc(hashFunc)
}

// NewConsistentHash creates an empty consistent hash ring.
func NewConsistentHash(opts ...ConsistentHashOption) *ConsistentHash {
	return hashimpl.NewConsistentHash(opts...)
}
