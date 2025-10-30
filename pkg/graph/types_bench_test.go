package graph

import (
	"fmt"
	"testing"
)

func BenchmarkAddNode(b *testing.B) {
	g := NewGraph("bench")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := &Node{
			ID:   fmt.Sprintf("node-%d", i),
			Type: NodeTypeStep,
			Name: "Benchmark Node",
		}
		g.AddNode(node)
	}
}

func BenchmarkGetNode(b *testing.B) {
	g := NewGraph("bench")
	
	// Setup: Add 1000 nodes
	for i := 0; i < 1000; i++ {
		node := &Node{
			ID:   fmt.Sprintf("node-%d", i),
			Type: NodeTypeStep,
			Name: "Test Node",
		}
		g.AddNode(node)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetNode(fmt.Sprintf("node-%d", i%1000))
	}
}

func BenchmarkAddEdge(b *testing.B) {
	g := NewGraph("bench")
	
	// Setup: Add nodes
	from := &Node{ID: "workflow", Type: NodeTypeWorkflow, Name: "Workflow"}
	to := &Node{ID: "step", Type: NodeTypeStep, Name: "Step"}
	g.AddNode(from)
	g.AddNode(to)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		edge := &Edge{
			ID:         fmt.Sprintf("edge-%d", i),
			FromNodeID: "workflow",
			ToNodeID:   "step",
			Type:       EdgeTypeContains,
		}
		g.AddEdge(edge)
	}
}

func BenchmarkUpdateNodeState(b *testing.B) {
	g := NewGraph("bench")
	
	// Setup: Add node
	node := &Node{
		ID:    "test-node",
		Type:  NodeTypeStep,
		Name:  "Test",
		State: NodeStateWaiting,
	}
	g.AddNode(node)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate states
		if i%2 == 0 {
			g.UpdateNodeState("test-node", NodeStateRunning)
		} else {
			g.UpdateNodeState("test-node", NodeStateWaiting)
		}
	}
}

func BenchmarkGetNodesByType(b *testing.B) {
	g := NewGraph("bench")
	
	// Setup: Add 100 nodes of mixed types
	for i := 0; i < 100; i++ {
		nodeType := NodeTypeStep
		if i%4 == 0 {
			nodeType = NodeTypeWorkflow
		} else if i%4 == 1 {
			nodeType = NodeTypeSpec
		} else if i%4 == 2 {
			nodeType = NodeTypeResource
		}
		
		node := &Node{
			ID:   fmt.Sprintf("node-%d", i),
			Type: nodeType,
			Name: "Test Node",
		}
		g.AddNode(node)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetNodesByType(NodeTypeWorkflow)
	}
}

func BenchmarkGetNodesByState(b *testing.B) {
	g := NewGraph("bench")
	
	// Setup: Add 100 nodes with mixed states
	for i := 0; i < 100; i++ {
		state := NodeStateWaiting
		if i%3 == 0 {
			state = NodeStateRunning
		} else if i%3 == 1 {
			state = NodeStateSucceeded
		}
		
		node := &Node{
			ID:    fmt.Sprintf("node-%d", i),
			Type:  NodeTypeStep,
			Name:  "Test Node",
			State: state,
		}
		g.AddNode(node)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetNodesByState(NodeStateRunning)
	}
}

func BenchmarkLargeGraph_100Nodes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := NewGraph("bench")
		
		// Add 100 nodes
		for j := 0; j < 100; j++ {
			node := &Node{
				ID:   fmt.Sprintf("node-%d", j),
				Type: NodeTypeStep,
				Name: "Test Node",
			}
			g.AddNode(node)
		}
		
		// Add 99 edges (linear chain)
		for j := 1; j < 100; j++ {
			edge := &Edge{
				ID:         fmt.Sprintf("edge-%d", j),
				FromNodeID: fmt.Sprintf("node-%d", j-1),
				ToNodeID:   fmt.Sprintf("node-%d", j),
				Type:       EdgeTypeDependsOn,
			}
			g.AddEdge(edge)
		}
	}
}

func BenchmarkLargeGraph_1000Nodes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := NewGraph("bench")
		
		// Add 1000 nodes
		for j := 0; j < 1000; j++ {
			node := &Node{
				ID:   fmt.Sprintf("node-%d", j),
				Type: NodeTypeStep,
				Name: "Test Node",
			}
			g.AddNode(node)
		}
		
		// Add 999 edges (linear chain)
		for j := 1; j < 1000; j++ {
			edge := &Edge{
				ID:         fmt.Sprintf("edge-%d", j),
				FromNodeID: fmt.Sprintf("node-%d", j-1),
				ToNodeID:   fmt.Sprintf("node-%d", j),
				Type:       EdgeTypeDependsOn,
			}
			g.AddEdge(edge)
		}
	}
}
