package layout

import (
	"fmt"
	"testing"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

func createBenchGraph(nodeCount int) *graph.Graph {
	g := graph.NewGraph("bench")

	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("n%d", i)
		g.AddNode(&graph.Node{ID: nodeID, Type: graph.NodeTypeStep, Name: nodeID})
	}

	// Create tree structure
	for i := 1; i < nodeCount; i++ {
		parentID := fmt.Sprintf("n%d", i/3)
		childID := fmt.Sprintf("n%d", i)
		edgeID := fmt.Sprintf("e%d", i)
		g.AddEdge(&graph.Edge{ID: edgeID, FromNodeID: parentID, ToNodeID: childID, Type: graph.EdgeTypeContains})
	}

	return g
}

func BenchmarkHierarchicalLayout_100(b *testing.B) {
	g := createBenchGraph(100)
	options := DefaultLayoutOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkHierarchicalLayout_500(b *testing.B) {
	g := createBenchGraph(500)
	options := DefaultLayoutOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkHierarchicalLayout_1000(b *testing.B) {
	g := createBenchGraph(1000)
	options := DefaultLayoutOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkRadialLayout_100(b *testing.B) {
	g := createBenchGraph(100)
	options := &LayoutOptions{
		Type:         LayoutRadial,
		NodeSpacing:  50.0,
		LevelSpacing: 75.0,
		Width:        2000.0,
		Height:       2000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkRadialLayout_500(b *testing.B) {
	g := createBenchGraph(500)
	options := &LayoutOptions{
		Type:         LayoutRadial,
		NodeSpacing:  50.0,
		LevelSpacing: 75.0,
		Width:        2000.0,
		Height:       2000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkGridLayout_100(b *testing.B) {
	g := createBenchGraph(100)
	options := &LayoutOptions{
		Type:         LayoutGrid,
		NodeSpacing:  50.0,
		LevelSpacing: 75.0,
		Width:        2000.0,
		Height:       2000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkGridLayout_500(b *testing.B) {
	g := createBenchGraph(500)
	options := &LayoutOptions{
		Type:         LayoutGrid,
		NodeSpacing:  50.0,
		LevelSpacing: 75.0,
		Width:        2000.0,
		Height:       2000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkForceLayout_50(b *testing.B) {
	g := createBenchGraph(50)
	options := &LayoutOptions{
		Type:         LayoutForce,
		NodeSpacing:  50.0,
		LevelSpacing: 75.0,
		Width:        2000.0,
		Height:       2000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}

func BenchmarkForceLayout_100(b *testing.B) {
	g := createBenchGraph(100)
	options := &LayoutOptions{
		Type:         LayoutForce,
		NodeSpacing:  50.0,
		LevelSpacing: 75.0,
		Width:        2000.0,
		Height:       2000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(g, options)
	}
}
