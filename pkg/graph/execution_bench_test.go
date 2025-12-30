package graph

import (
	"fmt"
	"testing"
)

// buildTestGraph creates a test graph with the specified number of nodes and layers
func buildTestGraph(nodes, layers int) *Graph {
	g := NewGraph("benchmark")

	nodesPerLayer := nodes / layers
	if nodesPerLayer < 1 {
		nodesPerLayer = 1
	}

	// Create nodes in layers
	nodeID := 0
	for layer := 0; layer < layers; layer++ {
		for i := 0; i < nodesPerLayer && nodeID < nodes; i++ {
			id := fmt.Sprintf("node-%d", nodeID)
			g.AddNode(&Node{
				ID:    id,
				Type:  NodeTypeStep,
				Name:  fmt.Sprintf("Step %d", nodeID),
				State: NodeStateWaiting,
			})
			nodeID++
		}
	}

	// Create edges between layers
	nodeID = 0
	for layer := 0; layer < layers-1; layer++ {
		layerStart := layer * nodesPerLayer
		nextLayerStart := (layer + 1) * nodesPerLayer

		for i := 0; i < nodesPerLayer && layerStart+i < nodes; i++ {
			fromID := fmt.Sprintf("node-%d", layerStart+i)

			// Connect to 1-3 nodes in next layer
			connections := min(3, nodesPerLayer)
			for j := 0; j < connections && nextLayerStart+j < nodes; j++ {
				toID := fmt.Sprintf("node-%d", nextLayerStart+j)
				edgeID := fmt.Sprintf("edge-%s-%s", fromID, toID)
				g.AddEdge(&Edge{
					ID:         edgeID,
					FromNodeID: fromID,
					ToNodeID:   toID,
					Type:       EdgeTypeDependsOn,
				})
			}
		}
	}

	return g
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BenchmarkGetParallelLayers_10Nodes benchmarks parallel layers on small graphs
func BenchmarkGetParallelLayers_10Nodes(b *testing.B) {
	g := buildTestGraph(10, 3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = g.GetParallelLayers()
	}
}

// BenchmarkGetParallelLayers_100Nodes benchmarks parallel layers on medium graphs
func BenchmarkGetParallelLayers_100Nodes(b *testing.B) {
	g := buildTestGraph(100, 10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = g.GetParallelLayers()
	}
}

// BenchmarkGetParallelLayers_1000Nodes benchmarks parallel layers on large graphs
func BenchmarkGetParallelLayers_1000Nodes(b *testing.B) {
	g := buildTestGraph(1000, 20)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = g.GetParallelLayers()
	}
}

// BenchmarkGetParallelLayers_10000Nodes benchmarks parallel layers on very large graphs
func BenchmarkGetParallelLayers_10000Nodes(b *testing.B) {
	g := buildTestGraph(10000, 50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = g.GetParallelLayers()
	}
}

// BenchmarkGetReadyNodes_10Nodes benchmarks ready node detection on small graphs
func BenchmarkGetReadyNodes_10Nodes(b *testing.B) {
	g := buildTestGraph(10, 3)
	waitingStates := []NodeState{NodeStateWaiting, NodeStatePending}
	completedStates := []NodeState{NodeStateSucceeded, NodeStateSkipped}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = g.GetReadyNodes(waitingStates, completedStates)
	}
}

// BenchmarkGetReadyNodes_100Nodes benchmarks ready node detection on medium graphs
func BenchmarkGetReadyNodes_100Nodes(b *testing.B) {
	g := buildTestGraph(100, 10)
	waitingStates := []NodeState{NodeStateWaiting, NodeStatePending}
	completedStates := []NodeState{NodeStateSucceeded, NodeStateSkipped}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = g.GetReadyNodes(waitingStates, completedStates)
	}
}

// BenchmarkGetReadyNodes_1000Nodes benchmarks ready node detection on large graphs
func BenchmarkGetReadyNodes_1000Nodes(b *testing.B) {
	g := buildTestGraph(1000, 20)
	waitingStates := []NodeState{NodeStateWaiting, NodeStatePending}
	completedStates := []NodeState{NodeStateSucceeded, NodeStateSkipped}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = g.GetReadyNodes(waitingStates, completedStates)
	}
}

// BenchmarkGetReadyNodes_10000Nodes benchmarks ready node detection on very large graphs
func BenchmarkGetReadyNodes_10000Nodes(b *testing.B) {
	g := buildTestGraph(10000, 50)
	waitingStates := []NodeState{NodeStateWaiting, NodeStatePending}
	completedStates := []NodeState{NodeStateSucceeded, NodeStateSkipped}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = g.GetReadyNodes(waitingStates, completedStates)
	}
}

// BenchmarkPropagateState_10Nodes benchmarks state propagation on small graphs
func BenchmarkPropagateState_10Nodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		g := buildTestGraph(10, 3)
		affectedStates := []NodeState{NodeStateWaiting, NodeStatePending}
		b.StartTimer()

		_ = g.PropagateState("node-0", NodeStateFailed, affectedStates)
	}
}

// BenchmarkPropagateState_100Nodes benchmarks state propagation on medium graphs
func BenchmarkPropagateState_100Nodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		g := buildTestGraph(100, 10)
		affectedStates := []NodeState{NodeStateWaiting, NodeStatePending}
		b.StartTimer()

		_ = g.PropagateState("node-0", NodeStateFailed, affectedStates)
	}
}

// BenchmarkPropagateState_1000Nodes benchmarks state propagation on large graphs
func BenchmarkPropagateState_1000Nodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		g := buildTestGraph(1000, 20)
		affectedStates := []NodeState{NodeStateWaiting, NodeStatePending}
		b.StartTimer()

		_ = g.PropagateState("node-0", NodeStateFailed, affectedStates)
	}
}

// BenchmarkPropagateState_10000Nodes benchmarks state propagation on very large graphs
func BenchmarkPropagateState_10000Nodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		g := buildTestGraph(10000, 50)
		affectedStates := []NodeState{NodeStateWaiting, NodeStatePending}
		b.StartTimer()

		_ = g.PropagateState("node-0", NodeStateFailed, affectedStates)
	}
}

// BenchmarkGetDependencies_SmallDegree tests dependency queries with few edges
func BenchmarkGetDependencies_SmallDegree(b *testing.B) {
	g := buildTestGraph(1000, 100) // Many layers = small degree per node
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = g.GetDependencies("node-500")
	}
}

// BenchmarkGetDependencies_LargeDegree tests dependency queries with many edges
func BenchmarkGetDependencies_LargeDegree(b *testing.B) {
	g := buildTestGraph(1000, 10) // Few layers = large degree per node
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = g.GetDependencies("node-500")
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent graph access
func BenchmarkConcurrentOperations(b *testing.B) {
	g := buildTestGraph(1000, 20)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				_, _ = g.TopologicalSort()
			case 1:
				_, _ = g.GetParallelLayers()
			case 2:
				_ = g.GetReadyNodes(
					[]NodeState{NodeStateWaiting},
					[]NodeState{NodeStateSucceeded},
				)
			case 3:
				nodeID := fmt.Sprintf("node-%d", i%1000)
				_ = g.UpdateNodeState(nodeID, NodeStateRunning)
			}
			i++
		}
	})
}

// BenchmarkGraphConstruction measures the cost of building a graph with adjacency lists
func BenchmarkGraphConstruction_1000Nodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildTestGraph(1000, 20)
	}
}

// BenchmarkGraphConstruction_10000Nodes measures the cost of building a large graph
func BenchmarkGraphConstruction_10000Nodes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildTestGraph(10000, 50)
	}
}
