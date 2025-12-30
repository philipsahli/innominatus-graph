package graph

import (
	"sync"
	"testing"
)

// TestGetParallelLayers tests the parallel layers algorithm
func TestGetParallelLayers(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *Graph
		expectedLayers int
		expectError    bool
	}{
		{
			name: "simple linear chain",
			setup: func() *Graph {
				g := NewGraph("test")
				// A → B → C (linear chain, 3 layers)
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-C", FromNodeID: "B", ToNodeID: "C", Type: EdgeTypeDependsOn})
				return g
			},
			expectedLayers: 3,
			expectError:    false,
		},
		{
			name: "diamond dependency",
			setup: func() *Graph {
				g := NewGraph("test")
				//     A
				//    / \
				//   B   C
				//    \ /
				//     D
				// Expected layers: [A], [B, C], [D]
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "D", Type: NodeTypeStep, Name: "Step D", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "A-C", FromNodeID: "A", ToNodeID: "C", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-D", FromNodeID: "B", ToNodeID: "D", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "C-D", FromNodeID: "C", ToNodeID: "D", Type: EdgeTypeDependsOn})
				return g
			},
			expectedLayers: 3,
			expectError:    false,
		},
		{
			name: "parallel independent nodes",
			setup: func() *Graph {
				g := NewGraph("test")
				// A, B, C (all independent, 1 layer)
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				return g
			},
			expectedLayers: 1,
			expectError:    false,
		},
		{
			name: "empty graph",
			setup: func() *Graph {
				return NewGraph("test")
			},
			expectedLayers: 0,
			expectError:    false,
		},
		{
			name: "graph with cycle",
			setup: func() *Graph {
				g := NewGraph("test")
				// A → B → C → A (cycle)
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-C", FromNodeID: "B", ToNodeID: "C", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "C-A", FromNodeID: "C", ToNodeID: "A", Type: EdgeTypeDependsOn})
				return g
			},
			expectedLayers: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup()
			layers, err := g.GetParallelLayers()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(layers) != tt.expectedLayers {
				t.Errorf("Expected %d layers, got %d", tt.expectedLayers, len(layers))
			}

			// Verify each layer contains nodes with satisfied dependencies
			processed := make(map[string]bool)
			for layerIdx, layer := range layers {
				for _, node := range layer {
					// Check all dependencies are in previous layers
					// For DependsOn edges, dependency is the ToNodeID of OutgoingEdges
					for _, edge := range node.OutgoingEdges {
						if edge.Type == EdgeTypeDependsOn && !processed[edge.ToNodeID] && layerIdx > 0 {
							t.Errorf("Node %s in layer %d has unprocessed dependency %s",
								node.ID, layerIdx, edge.ToNodeID)
						}
					}
					processed[node.ID] = true
				}
			}
		})
	}
}

// TestGetReadyNodes tests the ready node detection algorithm
func TestGetReadyNodes(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() *Graph
		waitingStates   []NodeState
		completedStates []NodeState
		expectedReady   []string
	}{
		{
			name: "simple ready detection",
			setup: func() *Graph {
				g := NewGraph("test")
				// A (completed) → B (waiting) → C (waiting)
				// Only B should be ready
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateSucceeded})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-C", FromNodeID: "B", ToNodeID: "C", Type: EdgeTypeDependsOn})
				return g
			},
			waitingStates:   []NodeState{NodeStateWaiting, NodeStatePending},
			completedStates: []NodeState{NodeStateSucceeded, NodeStateSkipped},
			expectedReady:   []string{"B"},
		},
		{
			name: "parallel ready nodes",
			setup: func() *Graph {
				g := NewGraph("test")
				//     A (completed)
				//    / \
				//   B   C (both waiting)
				//    \ /
				//     D (waiting)
				// B and C should both be ready
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateSucceeded})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "D", Type: NodeTypeStep, Name: "Step D", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "A-C", FromNodeID: "A", ToNodeID: "C", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-D", FromNodeID: "B", ToNodeID: "D", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "C-D", FromNodeID: "C", ToNodeID: "D", Type: EdgeTypeDependsOn})
				return g
			},
			waitingStates:   []NodeState{NodeStateWaiting},
			completedStates: []NodeState{NodeStateSucceeded},
			expectedReady:   []string{"B", "C"},
		},
		{
			name: "no dependencies ready immediately",
			setup: func() *Graph {
				g := NewGraph("test")
				// A, B, C (all independent, all waiting)
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				return g
			},
			waitingStates:   []NodeState{NodeStateWaiting},
			completedStates: []NodeState{NodeStateSucceeded},
			expectedReady:   []string{"A", "B", "C"},
		},
		{
			name: "skipped dependencies count as satisfied",
			setup: func() *Graph {
				g := NewGraph("test")
				// A (skipped) → B (waiting)
				// B should be ready because skipped counts as completed
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateSkipped})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				return g
			},
			waitingStates:   []NodeState{NodeStateWaiting},
			completedStates: []NodeState{NodeStateSucceeded, NodeStateSkipped},
			expectedReady:   []string{"B"},
		},
		{
			name: "no ready nodes when dependencies not satisfied",
			setup: func() *Graph {
				g := NewGraph("test")
				// A (waiting) → B (waiting)
				// No nodes ready (A is waiting, B depends on A)
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				return g
			},
			waitingStates:   []NodeState{NodeStateWaiting},
			completedStates: []NodeState{NodeStateSucceeded},
			expectedReady:   []string{"A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup()
			ready := g.GetReadyNodes(tt.waitingStates, tt.completedStates)

			if len(ready) != len(tt.expectedReady) {
				t.Errorf("Expected %d ready nodes, got %d", len(tt.expectedReady), len(ready))
			}

			// Convert to map for easy lookup
			readyMap := make(map[string]bool)
			for _, node := range ready {
				readyMap[node.ID] = true
			}

			for _, expectedID := range tt.expectedReady {
				if !readyMap[expectedID] {
					t.Errorf("Expected node %s to be ready, but it wasn't", expectedID)
				}
			}
		})
	}
}

// TestPropagateState tests the BFS state propagation algorithm
func TestPropagateState(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *Graph
		startNodeID    string
		targetState    NodeState
		affectedStates []NodeState
		expectedStates map[string]NodeState
		expectError    bool
	}{
		{
			name: "propagate failure down linear chain",
			setup: func() *Graph {
				g := NewGraph("test")
				// A → B → C
				// All waiting, propagate failure from A
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateFailed})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-C", FromNodeID: "B", ToNodeID: "C", Type: EdgeTypeDependsOn})
				return g
			},
			startNodeID:    "A",
			targetState:    NodeStateFailed,
			affectedStates: []NodeState{NodeStateWaiting, NodeStatePending},
			expectedStates: map[string]NodeState{
				"A": NodeStateFailed,
				"B": NodeStateFailed,
				"C": NodeStateFailed,
			},
			expectError: false,
		},
		{
			name: "propagate failure in diamond",
			setup: func() *Graph {
				g := NewGraph("test")
				//     A (failed)
				//    / \
				//   B   C (waiting)
				//    \ /
				//     D (waiting)
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateFailed})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddNode(&Node{ID: "D", Type: NodeTypeStep, Name: "Step D", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "A-C", FromNodeID: "A", ToNodeID: "C", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-D", FromNodeID: "B", ToNodeID: "D", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "C-D", FromNodeID: "C", ToNodeID: "D", Type: EdgeTypeDependsOn})
				return g
			},
			startNodeID:    "A",
			targetState:    NodeStateFailed,
			affectedStates: []NodeState{NodeStateWaiting},
			expectedStates: map[string]NodeState{
				"A": NodeStateFailed,
				"B": NodeStateFailed,
				"C": NodeStateFailed,
				"D": NodeStateFailed,
			},
			expectError: false,
		},
		{
			name: "only propagate to affected states",
			setup: func() *Graph {
				g := NewGraph("test")
				// A → B (running) → C (waiting)
				// B is running, should NOT be affected
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateFailed})
				g.AddNode(&Node{ID: "B", Type: NodeTypeStep, Name: "Step B", State: NodeStateRunning})
				g.AddNode(&Node{ID: "C", Type: NodeTypeStep, Name: "Step C", State: NodeStateWaiting})
				g.AddEdge(&Edge{ID: "A-B", FromNodeID: "A", ToNodeID: "B", Type: EdgeTypeDependsOn})
				g.AddEdge(&Edge{ID: "B-C", FromNodeID: "B", ToNodeID: "C", Type: EdgeTypeDependsOn})
				return g
			},
			startNodeID:    "A",
			targetState:    NodeStateFailed,
			affectedStates: []NodeState{NodeStateWaiting, NodeStatePending},
			expectedStates: map[string]NodeState{
				"A": NodeStateFailed,
				"B": NodeStateRunning, // Should NOT change
				"C": NodeStateWaiting, // Should NOT change (B blocks propagation)
			},
			expectError: false,
		},
		{
			name: "error on non-existent start node",
			setup: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
				return g
			},
			startNodeID:    "non-existent",
			targetState:    NodeStateFailed,
			affectedStates: []NodeState{NodeStateWaiting},
			expectedStates: map[string]NodeState{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup()
			err := g.PropagateState(tt.startNodeID, tt.targetState, tt.affectedStates)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify expected states
			for nodeID, expectedState := range tt.expectedStates {
				node, exists := g.GetNode(nodeID)
				if !exists {
					t.Errorf("Node %s not found", nodeID)
					continue
				}
				if node.State != expectedState {
					t.Errorf("Node %s: expected state %s, got %s",
						nodeID, expectedState, node.State)
				}
			}
		})
	}
}

// TestConcurrentGraphAccess tests thread safety of graph operations
func TestConcurrentGraphAccess(t *testing.T) {
	g := NewGraph("concurrent-test")

	// Pre-populate with some nodes
	for i := 0; i < 10; i++ {
		nodeID := string(rune('A' + i))
		g.AddNode(&Node{
			ID:    nodeID,
			Type:  NodeTypeStep,
			Name:  "Step " + nodeID,
			State: NodeStateWaiting,
		})
	}

	// Add edges
	for i := 0; i < 9; i++ {
		fromID := string(rune('A' + i))
		toID := string(rune('A' + i + 1))
		g.AddEdge(&Edge{
			ID:         fromID + "-" + toID,
			FromNodeID: fromID,
			ToNodeID:   toID,
			Type:       EdgeTypeDependsOn,
		})
	}

	// Concurrent operations
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = g.TopologicalSort()
				_ = g.GetReadyNodes(
					[]NodeState{NodeStateWaiting},
					[]NodeState{NodeStateSucceeded},
				)
				_, _ = g.GetParallelLayers()
			}
		}(i)
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				nodeID := string(rune('A' + (j % 10)))
				currentState, exists := g.GetNodeState(nodeID)
				if exists {
					// Cycle through states
					newState := NodeStateWaiting
					switch currentState {
					case NodeStateWaiting:
						newState = NodeStateRunning
					case NodeStateRunning:
						newState = NodeStateSucceeded
					case NodeStateSucceeded:
						newState = NodeStateWaiting
					}
					_ = g.UpdateNodeState(nodeID, newState)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify graph is still valid
	if len(g.Nodes) != 10 {
		t.Errorf("Expected 10 nodes, got %d", len(g.Nodes))
	}

	// Verify all adjacency lists are intact
	for _, node := range g.Nodes {
		if len(node.IncomingEdges)+len(node.OutgoingEdges) == 0 && node.ID != "A" && node.ID != "J" {
			t.Errorf("Node %s has no edges", node.ID)
		}
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("GetParallelLayers on single node", func(t *testing.T) {
		g := NewGraph("test")
		g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateWaiting})
		layers, err := g.GetParallelLayers()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(layers) != 1 || len(layers[0]) != 1 {
			t.Errorf("Expected 1 layer with 1 node")
		}
	})

	t.Run("GetReadyNodes with no nodes", func(t *testing.T) {
		g := NewGraph("test")
		ready := g.GetReadyNodes(
			[]NodeState{NodeStateWaiting},
			[]NodeState{NodeStateSucceeded},
		)
		if len(ready) != 0 {
			t.Errorf("Expected 0 ready nodes on empty graph, got %d", len(ready))
		}
	})

	t.Run("PropagateState with no dependents", func(t *testing.T) {
		g := NewGraph("test")
		g.AddNode(&Node{ID: "A", Type: NodeTypeStep, Name: "Step A", State: NodeStateFailed})
		err := g.PropagateState("A", NodeStateFailed, []NodeState{NodeStateWaiting})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
