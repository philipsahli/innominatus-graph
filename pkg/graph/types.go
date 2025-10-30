package graph

import (
	"fmt"
	"time"
)

type NodeType string

const (
	NodeTypeSpec     NodeType = "spec"
	NodeTypeWorkflow NodeType = "workflow"
	NodeTypeStep     NodeType = "step"
	NodeTypeResource NodeType = "resource"
)

type EdgeType string

const (
	EdgeTypeDependsOn  EdgeType = "depends-on"
	EdgeTypeProvisions EdgeType = "provisions"
	EdgeTypeCreates    EdgeType = "creates"
	EdgeTypeBindsTo    EdgeType = "binds-to"
	EdgeTypeContains   EdgeType = "contains"   // workflow → step
	EdgeTypeConfigures EdgeType = "configures" // step → resource
)

type NodeState string

const (
	NodeStateWaiting   NodeState = "waiting"   // Initial state
	NodeStatePending   NodeState = "pending"   // Ready to execute
	NodeStateRunning   NodeState = "running"   // Currently executing
	NodeStateFailed    NodeState = "failed"    // Execution failed
	NodeStateSucceeded NodeState = "succeeded" // Execution succeeded
)

type Node struct {
	ID          string                 `json:"id"`
	Type        NodeType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	State       NodeState              `json:"state"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`   // When execution started
	CompletedAt *time.Time             `json:"completed_at,omitempty"` // When execution completed
	Duration    *time.Duration         `json:"duration,omitempty"`     // Execution duration
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
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

	// Initialize state if not set
	if node.State == "" {
		node.State = NodeStateWaiting
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
	case EdgeTypeContains:
		if fromNode.Type != NodeTypeWorkflow {
			return fmt.Errorf("contains edge can only originate from workflow nodes")
		}
		if toNode.Type != NodeTypeStep {
			return fmt.Errorf("contains edge can only target step nodes")
		}
	case EdgeTypeConfigures:
		if fromNode.Type != NodeTypeStep {
			return fmt.Errorf("configures edge can only originate from step nodes")
		}
		if toNode.Type != NodeTypeResource {
			return fmt.Errorf("configures edge can only target resource nodes")
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

// UpdateNodeState updates the state of a node and propagates state changes upward
func (g *Graph) UpdateNodeState(nodeID string, newState NodeState) error {
	node, exists := g.GetNode(nodeID)
	if !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}

	oldState := node.State
	node.State = newState
	now := time.Now()
	node.UpdatedAt = now
	g.UpdatedAt = now

	// Update timing fields based on state transitions
	if newState == NodeStateRunning && node.StartedAt == nil {
		node.StartedAt = &now
	}
	if (newState == NodeStateSucceeded || newState == NodeStateFailed) && node.CompletedAt == nil {
		node.CompletedAt = &now
		// Calculate duration if both start and completion times are set
		if node.StartedAt != nil {
			duration := node.CompletedAt.Sub(*node.StartedAt)
			node.Duration = &duration
		}
	}

	// Propagate state upward if step failed -> workflow failed
	if node.Type == NodeTypeStep && newState == NodeStateFailed {
		if err := g.propagateFailureToParent(nodeID); err != nil {
			return fmt.Errorf("failed to propagate state: %w", err)
		}
	}

	// If a workflow transitions to failed/succeeded, update all contained steps
	if node.Type == NodeTypeWorkflow && (newState == NodeStateFailed || newState == NodeStateSucceeded) {
		g.updateContainedSteps(nodeID, oldState, newState)
	}

	return nil
}

// propagateFailureToParent propagates step failure to parent workflow
func (g *Graph) propagateFailureToParent(stepID string) error {
	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeContains && edge.ToNodeID == stepID {
			// Found parent workflow
			parentNode, exists := g.GetNode(edge.FromNodeID)
			if exists && parentNode.State != NodeStateFailed {
				parentNode.State = NodeStateFailed
				parentNode.UpdatedAt = time.Now()
			}
			return nil
		}
	}
	return nil
}

// updateContainedSteps updates state of child steps when workflow completes
func (g *Graph) updateContainedSteps(workflowID string, oldState, newState NodeState) {
	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeContains && edge.FromNodeID == workflowID {
			stepNode, exists := g.GetNode(edge.ToNodeID)
			if exists && stepNode.State == NodeStateRunning {
				stepNode.State = newState
				stepNode.UpdatedAt = time.Now()
			}
		}
	}
}

// GetNodesByType returns all nodes of a specific type
func (g *Graph) GetNodesByType(nodeType NodeType) []*Node {
	nodes := make([]*Node, 0)
	for _, node := range g.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodesByState returns all nodes in a specific state
func (g *Graph) GetNodesByState(state NodeState) []*Node {
	nodes := make([]*Node, 0)
	for _, node := range g.Nodes {
		if node.State == state {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetChildSteps returns all step nodes contained by a workflow
func (g *Graph) GetChildSteps(workflowID string) []*Node {
	steps := make([]*Node, 0)
	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeContains && edge.FromNodeID == workflowID {
			if stepNode, exists := g.GetNode(edge.ToNodeID); exists {
				steps = append(steps, stepNode)
			}
		}
	}
	return steps
}

// GetParentWorkflow returns the parent workflow of a step node
func (g *Graph) GetParentWorkflow(stepID string) (*Node, error) {
	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeContains && edge.ToNodeID == stepID {
			if workflow, exists := g.GetNode(edge.FromNodeID); exists {
				return workflow, nil
			}
		}
	}
	return nil, fmt.Errorf("no parent workflow found for step %s", stepID)
}