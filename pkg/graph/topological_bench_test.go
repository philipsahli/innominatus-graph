package graph

import (
	"fmt"
	"testing"
)

func createBenchGraphWithNodes(nodeCount int) *Graph {
	g := NewGraph("bench")
	
	// Add nodes
	for i := 0; i < nodeCount; i++ {
		node := &Node{
			ID:   fmt.Sprintf("node-%d", i),
			Type: NodeTypeStep,
			Name: fmt.Sprintf("Step %d", i),
		}
		g.AddNode(node)
	}
	
	// Add edges (linear chain)
	for i := 1; i < nodeCount; i++ {
		edge := &Edge{
			ID:         fmt.Sprintf("edge-%d", i),
			FromNodeID: fmt.Sprintf("node-%d", i),
			ToNodeID:   fmt.Sprintf("node-%d", i-1),
			Type:       EdgeTypeDependsOn,
		}
		g.AddEdge(edge)
	}
	
	return g
}

func BenchmarkTopologicalSort_10Nodes(b *testing.B) {
	g := createBenchGraphWithNodes(10)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.TopologicalSort()
	}
}

func BenchmarkTopologicalSort_100Nodes(b *testing.B) {
	g := createBenchGraphWithNodes(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.TopologicalSort()
	}
}

func BenchmarkTopologicalSort_1000Nodes(b *testing.B) {
	g := createBenchGraphWithNodes(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.TopologicalSort()
	}
}

func BenchmarkGetDependencies(b *testing.B) {
	g := createBenchGraphWithNodes(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetDependencies("node-50")
	}
}

func BenchmarkGetChildSteps(b *testing.B) {
	g := NewGraph("bench")
	
	// Setup: workflow with 100 steps
	workflow := &Node{ID: "workflow", Type: NodeTypeWorkflow, Name: "Workflow"}
	g.AddNode(workflow)
	
	for i := 0; i < 100; i++ {
		step := &Node{
			ID:   fmt.Sprintf("step-%d", i),
			Type: NodeTypeStep,
			Name: fmt.Sprintf("Step %d", i),
		}
		g.AddNode(step)
		
		edge := &Edge{
			ID:         fmt.Sprintf("edge-%d", i),
			FromNodeID: "workflow",
			ToNodeID:   fmt.Sprintf("step-%d", i),
			Type:       EdgeTypeContains,
		}
		g.AddEdge(edge)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetChildSteps("workflow")
	}
}
