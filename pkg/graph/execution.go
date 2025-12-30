package graph

import "fmt"

// GetParallelLayers returns nodes grouped into layers for parallel execution.
// Each layer contains nodes that can be executed in parallel (no dependencies on each other).
// Layers are ordered such that layer N must complete before layer N+1 can start.
// This is the topological layers algorithm transferred from innominatus custom DAG.
//
// Complexity: O(N+E) using adjacency lists and in-degree tracking.
func (g *Graph) GetParallelLayers() ([][]*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.Nodes) == 0 {
		return [][]*Node{}, nil
	}

	// Use cached InDegree from nodes (already maintained by AddEdge/RemoveEdge)
	inDegree := make(map[string]int)
	for nodeID, node := range g.Nodes {
		inDegree[nodeID] = node.InDegree
	}

	layers := make([][]*Node, 0)
	processed := make(map[string]bool)

	for {
		// Find all nodes ready for this layer (in-degree = 0)
		currentLayer := make([]*Node, 0)
		for _, node := range g.Nodes {
			if !processed[node.ID] && inDegree[node.ID] == 0 {
				currentLayer = append(currentLayer, node)
			}
		}

		// If no nodes ready, check if we're done or have a cycle
		if len(currentLayer) == 0 {
			// Check if all nodes processed
			if len(processed) == len(g.Nodes) {
				break // All done!
			}
			// Otherwise, we have a cycle
			return nil, fmt.Errorf("cycle detected in graph - cannot compute parallel layers")
		}

		// Mark current layer as processed and decrease in-degree of dependents
		for _, node := range currentLayer {
			processed[node.ID] = true

			// Decrement InDegree for nodes that depend on current node
			// For DependsOn edges: FROM node's InDegree was incremented (FROM depends on current)
			// For other edges: TO node's InDegree was incremented (current points to TO)
			for _, edge := range node.IncomingEdges {
				if edge.Type == EdgeTypeDependsOn {
					inDegree[edge.FromNodeID]--
				}
			}
			for _, edge := range node.OutgoingEdges {
				if edge.Type != EdgeTypeDependsOn {
					inDegree[edge.ToNodeID]--
				}
			}
		}

		layers = append(layers, currentLayer)
	}

	return layers, nil
}

// GetReadyNodes returns all nodes that are ready to execute.
// A node is ready if all its dependencies have completed or been skipped.
// Transferred from innominatus custom DAG and optimized using in-degree tracking.
//
// Complexity: O(N) using cached in-degree (vs O(N×D) without caching).
//
// Parameters:
// - waitingStates: Node states considered "waiting" (e.g., Waiting, Pending)
// - completedStates: Node states considered "completed" (e.g., Succeeded, Skipped)
func (g *Graph) GetReadyNodes(waitingStates, completedStates []NodeState) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ready := make([]*Node, 0)

	for _, node := range g.Nodes {
		// Check if node is in a waiting state
		if !containsState(node.State, waitingStates) {
			continue
		}

		// Check if all dependencies are satisfied (O(1) with in-degree!)
		// A node is ready if it has no unsatisfied dependencies
		satisfiedCount := 0
		for _, edge := range node.IncomingEdges {
			sourceNode := g.Nodes[edge.FromNodeID]
			if containsState(sourceNode.State, completedStates) {
				satisfiedCount++
			}
		}

		// If all incoming edges come from completed nodes, it's ready
		if satisfiedCount == len(node.IncomingEdges) {
			ready = append(ready, node)
		}
	}

	return ready
}

// PropagateState propagates a state change to dependent nodes using BFS.
// This is an optimized version of MarkNodesWithFailedDependencies from custom DAG.
// Transferred and generalized for any state propagation scenario.
//
// Complexity: O(N+E) using BFS with adjacency lists (vs O(N²) with fixed-point iteration).
//
// Parameters:
// - startNodeID: The node whose state changed
// - targetState: The state to propagate to dependents
// - affectedStates: Only propagate to nodes in these states
//
// Example: Propagate failure from a failed step to all dependent waiting/pending steps.
func (g *Graph) PropagateState(startNodeID string, targetState NodeState, affectedStates []NodeState) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[startNodeID]; !exists {
		return fmt.Errorf("start node %s does not exist", startNodeID)
	}

	// BFS to propagate state to all dependent nodes
	queue := []string{startNodeID}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		currentNode := g.Nodes[currentID]

		// Propagate to all dependent nodes (outgoing edges)
		for _, edge := range currentNode.OutgoingEdges {
			targetNode := g.Nodes[edge.ToNodeID]

			// Only propagate if target node is in an affected state
			if containsState(targetNode.State, affectedStates) {
				targetNode.State = targetState
				queue = append(queue, edge.ToNodeID)
			}
		}
	}

	return nil
}

// containsState checks if a state is in a list of states.
func containsState(state NodeState, states []NodeState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}
