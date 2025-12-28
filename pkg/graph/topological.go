package graph

import "fmt"

// TopologicalSort returns nodes in topological order using Kahn's algorithm.
// Optimized from O(N×E) to O(N+E) using adjacency lists.
// Returns error if graph contains cycles.
func (g *Graph) TopologicalSort() ([]*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Use cached InDegree from nodes (already maintained by AddEdge/RemoveEdge)
	inDegree := make(map[string]int)
	for nodeID, node := range g.Nodes {
		inDegree[nodeID] = node.InDegree
	}

	// Find all nodes with no dependencies (in-degree = 0)
	queue := make([]*Node, 0)
	for _, node := range g.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node)
		}
	}

	result := make([]*Node, 0, len(g.Nodes))

	// Process nodes in topological order
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Decrement InDegree for nodes that depend on current node
		// For DependsOn edges: FROM node's InDegree was incremented (FROM depends on current)
		// For other edges: TO node's InDegree was incremented (current points to TO)
		for _, edge := range current.IncomingEdges {
			if edge.Type == EdgeTypeDependsOn {
				dependentID := edge.FromNodeID
				inDegree[dependentID]--
				if inDegree[dependentID] == 0 {
					queue = append(queue, g.Nodes[dependentID])
				}
			}
		}
		for _, edge := range current.OutgoingEdges {
			if edge.Type != EdgeTypeDependsOn {
				targetID := edge.ToNodeID
				inDegree[targetID]--
				if inDegree[targetID] == 0 {
					queue = append(queue, g.Nodes[targetID])
				}
			}
		}
	}

	// Cycle detection: if we didn't process all nodes, there's a cycle
	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("graph contains cycles, cannot perform topological sort")
	}

	return result, nil
}

// GetDependencies returns nodes that this node depends on.
// Optimized from O(E) to O(D) using adjacency lists.
func (g *Graph) GetDependencies(nodeID string) ([]*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	dependencies := make([]*Node, 0)

	// Use adjacency list: iterate outgoing edges to find dependencies
	// Edge semantics: {FROM: A, TO: B, DependsOn} means "A depends on B"
	for _, edge := range node.OutgoingEdges {
		if edge.Type == EdgeTypeDependsOn {
			if depNode, ok := g.Nodes[edge.ToNodeID]; ok {
				dependencies = append(dependencies, depNode)
			}
		}
	}

	return dependencies, nil
}

// GetDependents returns nodes that depend on this node.
// Optimized from O(E) to O(D) using adjacency lists.
func (g *Graph) GetDependents(nodeID string) ([]*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	dependents := make([]*Node, 0)

	// Use adjacency list: iterate incoming edges to find dependents
	// Edge semantics: {FROM: A, TO: B, DependsOn} means "A depends on B"
	// So nodes that depend on current (B) have edges pointing TO current
	for _, edge := range node.IncomingEdges {
		if edge.Type == EdgeTypeDependsOn {
			if depNode, ok := g.Nodes[edge.FromNodeID]; ok {
				dependents = append(dependents, depNode)
			}
		}
	}

	return dependents, nil
}

func (g *Graph) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}