package vhash

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestConsistentHashFacade(t *testing.T) {
	ring := NewConsistentHash(WithVirtualNodes(10))
	ring.Add("a")
	ring.Add("b")
	ring.Add("c")

	node, err := ring.Get("user:1")
	if err != nil || node == "" {
		t.Fatalf("Get = %q, %v", node, err)
	}
	nodes, err := ring.GetN("user:1", 2)
	if err != nil || len(nodes) != 2 || nodes[0] == nodes[1] {
		t.Fatalf("GetN = %v, %v", nodes, err)
	}
	ring.Remove(node)
	next, err := ring.Get("user:1")
	if err != nil || next == "" || next == node {
		t.Fatalf("Get after Remove = %q, %v", next, err)
	}
}

func TestConsistentHashEmpty(t *testing.T) {
	_, err := NewConsistentHash().Get("key")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("empty ring err = %v", err)
	}
}
