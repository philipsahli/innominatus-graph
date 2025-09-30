package graph

import (
	"fmt"
	"time"
)

type NodeType string

const (
	NodeTypeSpec     NodeType = "spec"
	NodeTypeWorkflow NodeType = "workflow"
	NodeTypeResource NodeType = "resource"
)

type EdgeType string

const (
	EdgeTypeDependsOn  EdgeType = "depends-on"
	EdgeTypeProvisions EdgeType = "provisions"
	EdgeTypeCreates    EdgeType = "creates"
	EdgeTypeBindsTo    EdgeType = "binds-to"
)

type Node struct {
	ID          string            `json:"id"`
	Type        NodeType          `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Edge struct {
	ID          string            `json:"id"`
	FromNodeID  string            `json:"from_node_id"`
	ToNodeID    string            `json:"to_node_id"`
	Type        EdgeType          `json:"type"`
	Description string            `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

type Graph struct {
	ID        string           `json:"id"`
	AppName   string           `json:"app_name"`
	Version   int              `json:"version"`
	Nodes     map[string]*Node `json:"nodes"`
	Edges     map[string]*Edge `json:"edges"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func NewGraph(appName string) *Graph {
	return &Graph{
		ID:        fmt.Sprintf("%s-graph", appName),
		AppName:   appName,
		Version:   1,
		Nodes:     make(map[string]*Node),
		Edges:     make(map[string]*Edge),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (g *Graph) AddNode(node *Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}
	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if _, exists := g.Nodes[node.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", node.ID)
	}

	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	g.Nodes[node.ID] = node
	g.UpdatedAt = time.Now()

	return nil
}

func (g *Graph) AddEdge(edge *Edge) error {
	if edge == nil {
		return fmt.Errorf("edge cannot be nil")
	}
	if edge.ID == "" {
		return fmt.Errorf("edge ID cannot be empty")
	}
	if _, exists := g.Edges[edge.ID]; exists {
		return fmt.Errorf("edge with ID %s already exists", edge.ID)
	}

	if _, exists := g.Nodes[edge.FromNodeID]; !exists {
		return fmt.Errorf("from node %s does not exist", edge.FromNodeID)
	}
	if _, exists := g.Nodes[edge.ToNodeID]; !exists {
		return fmt.Errorf("to node %s does not exist", edge.ToNodeID)
	}

	if err := g.validateEdge(edge); err != nil {
		return err
	}

	edge.CreatedAt = time.Now()
	g.Edges[edge.ID] = edge
	g.UpdatedAt = time.Now()

	return nil
}

func (g *Graph) validateEdge(edge *Edge) error {
	fromNode := g.Nodes[edge.FromNodeID]
	toNode := g.Nodes[edge.ToNodeID]

	switch edge.Type {
	case EdgeTypeDependsOn:
		return nil
	case EdgeTypeProvisions:
		if fromNode.Type != NodeTypeWorkflow {
			return fmt.Errorf("provisions edge can only originate from workflow nodes")
		}
		if toNode.Type != NodeTypeResource {
			return fmt.Errorf("provisions edge can only target resource nodes")
		}
	case EdgeTypeCreates:
		if fromNode.Type != NodeTypeWorkflow {
			return fmt.Errorf("creates edge can only originate from workflow nodes")
		}
	case EdgeTypeBindsTo:
		if toNode.Type != NodeTypeResource {
			return fmt.Errorf("binds-to edge can only target resource nodes")
		}
	default:
		return fmt.Errorf("invalid edge type: %s", edge.Type)
	}

	return nil
}

func (g *Graph) GetNode(id string) (*Node, bool) {
	node, exists := g.Nodes[id]
	return node, exists
}

func (g *Graph) GetEdge(id string) (*Edge, bool) {
	edge, exists := g.Edges[id]
	return edge, exists
}

func (g *Graph) RemoveNode(id string) error {
	if _, exists := g.Nodes[id]; !exists {
		return fmt.Errorf("node %s does not exist", id)
	}

	edgesToRemove := []string{}
	for edgeID, edge := range g.Edges {
		if edge.FromNodeID == id || edge.ToNodeID == id {
			edgesToRemove = append(edgesToRemove, edgeID)
		}
	}

	for _, edgeID := range edgesToRemove {
		delete(g.Edges, edgeID)
	}

	delete(g.Nodes, id)
	g.UpdatedAt = time.Now()

	return nil
}

func (g *Graph) RemoveEdge(id string) error {
	if _, exists := g.Edges[id]; !exists {
		return fmt.Errorf("edge %s does not exist", id)
	}

	delete(g.Edges, id)
	g.UpdatedAt = time.Now()

	return nil
}