package graph

import "fmt"

func (g *Graph) TopologicalSort() ([]*Node, error) {
	inDegree := make(map[string]int)

	for nodeID := range g.Nodes {
		inDegree[nodeID] = 0
	}

	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeDependsOn {
			inDegree[edge.FromNodeID]++
		} else {
			inDegree[edge.ToNodeID]++
		}
	}

	queue := make([]*Node, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, g.Nodes[nodeID])
		}
	}

	result := make([]*Node, 0, len(g.Nodes))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, edge := range g.Edges {
			var nextNodeID string
			if edge.Type == EdgeTypeDependsOn && edge.ToNodeID == current.ID {
				nextNodeID = edge.FromNodeID
			} else if edge.Type != EdgeTypeDependsOn && edge.FromNodeID == current.ID {
				nextNodeID = edge.ToNodeID
			} else {
				continue
			}

			inDegree[nextNodeID]--
			if inDegree[nextNodeID] == 0 {
				queue = append(queue, g.Nodes[nextNodeID])
			}
		}
	}

	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("graph contains cycles, cannot perform topological sort")
	}

	return result, nil
}

func (g *Graph) GetDependencies(nodeID string) ([]*Node, error) {
	_, exists := g.GetNode(nodeID)
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	dependencies := make([]*Node, 0)

	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeDependsOn && edge.FromNodeID == nodeID {
			if depNode, exists := g.GetNode(edge.ToNodeID); exists {
				dependencies = append(dependencies, depNode)
			}
		}
	}

	return dependencies, nil
}

func (g *Graph) GetDependents(nodeID string) ([]*Node, error) {
	_, exists := g.GetNode(nodeID)
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	dependents := make([]*Node, 0)

	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeDependsOn && edge.ToNodeID == nodeID {
			if depNode, exists := g.GetNode(edge.FromNodeID); exists {
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