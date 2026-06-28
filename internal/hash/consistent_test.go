package hash

import (
	"errors"
	"fmt"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestConsistentHashGetNReturnsDistinctNodes(t *testing.T) {
	ring := NewConsistentHash(WithVirtualNodes(8))
	ring.Add("cache-a")
	ring.Add("cache-b")
	ring.Add("cache-c")

	nodes, err := ring.GetN("user:42", 10)
	if err != nil {
		t.Fatalf("GetN error = %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("GetN len = %d, want 3", len(nodes))
	}
	seen := map[string]bool{}
	for _, node := range nodes {
		if seen[node] {
			t.Fatalf("GetN returned duplicate node %q in %v", node, nodes)
		}
		seen[node] = true
	}
}

func TestConsistentHashEmptyRingErrors(t *testing.T) {
	_, err := NewConsistentHash().Get("key")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Get empty ring error = %v, want ErrCodeInvalidInput", err)
	}

	nodes, err := NewConsistentHash().GetN("key", 1)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("GetN empty ring error = %v, want ErrCodeInvalidInput", err)
	}
	if nodes != nil {
		t.Fatalf("GetN empty ring nodes = %v, want nil", nodes)
	}
}

func TestConsistentHashMovementAfterAddingNodeIsBounded(t *testing.T) {
	const keyCount = 5000
	before := NewConsistentHash(WithVirtualNodes(512))
	for _, node := range []string{"cache-a", "cache-b", "cache-c"} {
		before.Add(node)
	}

	assignments := make([]string, 0, keyCount)
	for i := 0; i < keyCount; i++ {
		node, err := before.Get(fmt.Sprintf("key:%d", i))
		if err != nil {
			t.Fatalf("before.Get error = %v", err)
		}
		assignments = append(assignments, node)
	}

	after := NewConsistentHash(WithVirtualNodes(512))
	for _, node := range []string{"cache-a", "cache-b", "cache-c", "cache-d"} {
		after.Add(node)
	}

	moved := 0
	for i, previous := range assignments {
		node, err := after.Get(fmt.Sprintf("key:%d", i))
		if err != nil {
			t.Fatalf("after.Get error = %v", err)
		}
		if node != previous {
			moved++
		}
	}
	ratio := float64(moved) / keyCount
	if ratio > 0.35 {
		t.Fatalf("movement ratio = %.3f, want <= 0.35", ratio)
	}
}

func TestConsistentHashDistributionIsReasonablyUniform(t *testing.T) {
	const keyCount = 10000
	ring := NewConsistentHash(WithVirtualNodes(1024))
	nodes := []string{"cache-a", "cache-b", "cache-c", "cache-d"}
	for _, node := range nodes {
		ring.Add(node)
	}

	counts := make(map[string]int, len(nodes))
	for i := 0; i < keyCount; i++ {
		node, err := ring.Get(fmt.Sprintf("key:%d", i))
		if err != nil {
			t.Fatalf("Get error = %v", err)
		}
		counts[node]++
	}

	expected := keyCount / len(nodes)
	for _, node := range nodes {
		count := counts[node]
		if count < expected/2 || count > expected+expected/2 {
			t.Fatalf("node %s count = %d, want within 50%% of %d; counts=%v", node, count, expected, counts)
		}
	}
}

func TestConsistentHashConstantHashFunctionDoesNotHang(t *testing.T) {
	done := make(chan struct{})
	ring := NewConsistentHash(
		WithVirtualNodes(4),
		WithHashFunc(func([]byte) uint64 { return 1 }),
	)

	go func() {
		defer close(done)
		ring.Add("cache-a")
		ring.Add("cache-b")
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Add with constant hash function did not return")
	}

	node, err := ring.Get("key")
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if node != "cache-a" {
		t.Fatalf("Get = %q, want first inserted node", node)
	}
}
