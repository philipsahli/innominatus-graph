package graph

import (
	"fmt"
	"sync"
	"time"
)

type NodeType string

const (
	NodeTypeTeam        NodeType = "team"
	NodeTypeApplication NodeType = "application"
	NodeTypeSpec        NodeType = "spec"
	NodeTypeWorkflow    NodeType = "workflow"
	NodeTypeStep        NodeType = "step"
	NodeTypeResource    NodeType = "resource"
	NodeTypeProvider    NodeType = "provider"
)

type EdgeType string

const (
	EdgeTypeDependsOn  EdgeType = "depends-on"
	EdgeTypeProvisions EdgeType = "provisions"
	EdgeTypeCreates    EdgeType = "creates"
	EdgeTypeBindsTo    EdgeType = "binds-to"
	EdgeTypeContains   EdgeType = "contains"   // workflow → step, spec → resource
	EdgeTypeConfigures EdgeType = "configures" // step → resource
	EdgeTypeRequires   EdgeType = "requires"   // resource → provider
	EdgeTypeExecutes   EdgeType = "executes"   // provider → workflow
	EdgeTypeTriggers   EdgeType = "triggers"   // spec → workflow
	EdgeTypeOwns       EdgeType = "owns"       // team → application
	EdgeTypeHasSpec    EdgeType = "has-spec"   // application → spec
)

type NodeState string

const (
	NodeStateWaiting   NodeState = "waiting"   // Initial state
	NodeStatePending   NodeState = "pending"   // Ready to execute
	NodeStateRunning   NodeState = "running"   // Currently executing
	NodeStateFailed    NodeState = "failed"    // Execution failed
	NodeStateSucceeded NodeState = "succeeded" // Execution succeeded
	NodeStateSkipped   NodeState = "skipped"   // Skipped (condition not met)
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

	// Performance optimization: Adjacency lists for O(D) instead of O(E) traversal
	IncomingEdges []*Edge `json:"-"` // Edges pointing TO this node (dependencies)
	OutgoingEdges []*Edge `json:"-"` // Edges pointing FROM this node (dependents)

	// Performance optimization: Cached degree counts for O(1) ready node detection
	InDegree  int `json:"in_degree"`  // Count of incoming dependency edges
	OutDegree int `json:"out_degree"` // Count of outgoing edges
}

type Edge struct {
	ID          string                 `json:"id"`
	FromNodeID  string                 `json:"from_node_id"`
	ToNodeID    string                 `json:"to_node_id"`
	Type        EdgeType               `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type Graph struct {
	mu sync.RWMutex // Thread safety for concurrent access

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

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[node.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", node.ID)
	}

	// Initialize state if not set
	if node.State == "" {
		node.State = NodeStateWaiting
	}

	// Initialize adjacency lists
	if node.IncomingEdges == nil {
		node.IncomingEdges = make([]*Edge, 0)
	}
	if node.OutgoingEdges == nil {
		node.OutgoingEdges = make([]*Edge, 0)
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

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Edges[edge.ID]; exists {
		return fmt.Errorf("edge with ID %s already exists", edge.ID)
	}

	fromNode, fromExists := g.Nodes[edge.FromNodeID]
	if !fromExists {
		return fmt.Errorf("from node %s does not exist", edge.FromNodeID)
	}
	toNode, toExists := g.Nodes[edge.ToNodeID]
	if !toExists {
		return fmt.Errorf("to node %s does not exist", edge.ToNodeID)
	}

	if err := g.validateEdge(edge); err != nil {
		return err
	}

	edge.CreatedAt = time.Now()
	g.Edges[edge.ID] = edge
	g.UpdatedAt = time.Now()

	// Maintain adjacency lists for O(D) traversal
	fromNode.OutgoingEdges = append(fromNode.OutgoingEdges, edge)
	toNode.IncomingEdges = append(toNode.IncomingEdges, edge)

	// Update degree counts for O(1) ready node detection
	fromNode.OutDegree++
	// For DependsOn edges, the FROM node is the dependent (has a dependency)
	// For other edges, track TO node's InDegree for general purposes
	if edge.Type == EdgeTypeDependsOn {
		fromNode.InDegree++
	} else {
		toNode.InDegree++
	}

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
		// Allow spec → resource AND workflow → step
		if fromNode.Type == NodeTypeSpec && toNode.Type == NodeTypeResource {
			return nil // spec contains resources
		}
		if fromNode.Type == NodeTypeWorkflow && toNode.Type == NodeTypeStep {
			return nil // workflow contains steps
		}
		return fmt.Errorf("contains edge requires (spec→resource) or (workflow→step), got (%s→%s)", fromNode.Type, toNode.Type)
	case EdgeTypeConfigures:
		if fromNode.Type != NodeTypeStep {
			return fmt.Errorf("configures edge can only originate from step nodes")
		}
		if toNode.Type != NodeTypeResource {
			return fmt.Errorf("configures edge can only target resource nodes")
		}
	case EdgeTypeRequires:
		// resource → provider (resource requires provider)
		if fromNode.Type != NodeTypeResource {
			return fmt.Errorf("requires edge can only originate from resource nodes, got %s", fromNode.Type)
		}
		if toNode.Type != NodeTypeProvider {
			return fmt.Errorf("requires edge can only target provider nodes, got %s", toNode.Type)
		}
	case EdgeTypeExecutes:
		// provider → workflow (provider executes workflow)
		if fromNode.Type != NodeTypeProvider {
			return fmt.Errorf("executes edge can only originate from provider nodes, got %s", fromNode.Type)
		}
		if toNode.Type != NodeTypeWorkflow {
			return fmt.Errorf("executes edge can only target workflow nodes, got %s", toNode.Type)
		}
	case EdgeTypeTriggers:
		// spec → workflow (spec triggers workflow)
		if fromNode.Type != NodeTypeSpec {
			return fmt.Errorf("triggers edge can only originate from spec nodes, got %s", fromNode.Type)
		}
		if toNode.Type != NodeTypeWorkflow {
			return fmt.Errorf("triggers edge can only target workflow nodes, got %s", toNode.Type)
		}
	case EdgeTypeOwns:
		// team → application (team owns application)
		if fromNode.Type != NodeTypeTeam {
			return fmt.Errorf("owns edge can only originate from team nodes, got %s", fromNode.Type)
		}
		if toNode.Type != NodeTypeApplication {
			return fmt.Errorf("owns edge can only target application nodes, got %s", toNode.Type)
		}
	case EdgeTypeHasSpec:
		// application → spec (application has spec version)
		if fromNode.Type != NodeTypeApplication {
			return fmt.Errorf("has-spec edge can only originate from application nodes, got %s", fromNode.Type)
		}
		if toNode.Type != NodeTypeSpec {
			return fmt.Errorf("has-spec edge can only target spec nodes, got %s", toNode.Type)
		}
	default:
		return fmt.Errorf("invalid edge type: %s", edge.Type)
	}

	return nil
}

func (g *Graph) GetNode(id string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	node, exists := g.Nodes[id]
	return node, exists
}

// GetNodeState returns the current state of a node (thread-safe)
func (g *Graph) GetNodeState(id string) (NodeState, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	node, exists := g.Nodes[id]
	if !exists {
		return NodeState(""), false
	}
	return node.State, true
}

func (g *Graph) GetEdge(id string) (*Edge, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	edge, exists := g.Edges[id]
	return edge, exists
}

func (g *Graph) RemoveNode(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, exists := g.Nodes[id]
	if !exists {
		return fmt.Errorf("node %s does not exist", id)
	}

	// Remove all edges connected to this node and update adjacency lists
	edgesToRemove := make([]string, 0)

	// Remove outgoing edges
	for _, edge := range node.OutgoingEdges {
		edgesToRemove = append(edgesToRemove, edge.ID)
		// Update target node's incoming edges
		if targetNode, ok := g.Nodes[edge.ToNodeID]; ok {
			// For non-DependsOn edges, decrement target's InDegree
			if edge.Type != EdgeTypeDependsOn {
				targetNode.InDegree--
			}
			targetNode.IncomingEdges = removeEdgeFromList(targetNode.IncomingEdges, edge)
		}
		// For DependsOn edges, the current node (being removed) had its InDegree incremented
		// This is handled implicitly since the node is being deleted
	}

	// Remove incoming edges
	for _, edge := range node.IncomingEdges {
		edgesToRemove = append(edgesToRemove, edge.ID)
		// Update source node's outgoing edges and out-degree
		if sourceNode, ok := g.Nodes[edge.FromNodeID]; ok {
			sourceNode.OutDegree--
			sourceNode.OutgoingEdges = removeEdgeFromList(sourceNode.OutgoingEdges, edge)
			// For DependsOn edges, FROM node's InDegree was incremented
			if edge.Type == EdgeTypeDependsOn {
				sourceNode.InDegree--
			}
		}
	}

	// Remove edges from map
	for _, edgeID := range edgesToRemove {
		delete(g.Edges, edgeID)
	}

	delete(g.Nodes, id)
	g.UpdatedAt = time.Now()

	return nil
}

func (g *Graph) RemoveEdge(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	edge, exists := g.Edges[id]
	if !exists {
		return fmt.Errorf("edge %s does not exist", id)
	}

	// Update adjacency lists and degree counts
	if fromNode, ok := g.Nodes[edge.FromNodeID]; ok {
		fromNode.OutDegree--
		fromNode.OutgoingEdges = removeEdgeFromList(fromNode.OutgoingEdges, edge)
		// For DependsOn edges, FROM node's InDegree was incremented
		if edge.Type == EdgeTypeDependsOn {
			fromNode.InDegree--
		}
	}
	if toNode, ok := g.Nodes[edge.ToNodeID]; ok {
		toNode.IncomingEdges = removeEdgeFromList(toNode.IncomingEdges, edge)
		// For non-DependsOn edges, TO node's InDegree was incremented
		if edge.Type != EdgeTypeDependsOn {
			toNode.InDegree--
		}
	}

	delete(g.Edges, id)
	g.UpdatedAt = time.Now()

	return nil
}

// removeEdgeFromList removes an edge from an adjacency list
func removeEdgeFromList(edges []*Edge, edgeToRemove *Edge) []*Edge {
	for i, edge := range edges {
		if edge.ID == edgeToRemove.ID {
			return append(edges[:i], edges[i+1:]...)
		}
	}
	return edges
}

// UpdateNodeState updates the state of a node and propagates state changes upward
func (g *Graph) UpdateNodeState(nodeID string, newState NodeState) error {
	_, err := g.UpdateNodeStateWithOldState(nodeID, newState)
	return err
}

// UpdateNodeStateWithOldState updates the state of a node and returns the old state.
// This method is thread-safe and returns the old state atomically.
func (g *Graph) UpdateNodeStateWithOldState(nodeID string, newState NodeState) (oldState NodeState, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, exists := g.Nodes[nodeID]
	if !exists {
		return NodeState(""), fmt.Errorf("node %s does not exist", nodeID)
	}

	oldState = node.State
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
			return oldState, fmt.Errorf("failed to propagate state: %w", err)
		}
	}

	// If a workflow transitions to failed/succeeded, update all contained steps
	if node.Type == NodeTypeWorkflow && (newState == NodeStateFailed || newState == NodeStateSucceeded) {
		g.updateContainedSteps(nodeID, oldState, newState)
	}

	return oldState, nil
}

// propagateFailureToParent propagates step failure to parent workflow
// Note: This is called while holding the write lock, so access nodes directly
func (g *Graph) propagateFailureToParent(stepID string) error {
	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeContains && edge.ToNodeID == stepID {
			// Found parent workflow - access directly without lock (already held)
			if parentNode, exists := g.Nodes[edge.FromNodeID]; exists {
				if parentNode.State != NodeStateFailed {
					parentNode.State = NodeStateFailed
					parentNode.UpdatedAt = time.Now()
				}
			}
			return nil
		}
	}
	return nil
}

// updateContainedSteps updates state of child steps when workflow completes
// Note: This is called while holding the write lock, so access nodes directly
func (g *Graph) updateContainedSteps(workflowID string, oldState, newState NodeState) {
	for _, edge := range g.Edges {
		if edge.Type == EdgeTypeContains && edge.FromNodeID == workflowID {
			// Access directly without lock (already held)
			if stepNode, exists := g.Nodes[edge.ToNodeID]; exists {
				if stepNode.State == NodeStateRunning {
					stepNode.State = newState
					stepNode.UpdatedAt = time.Now()
				}
			}
		}
	}
}

// GetNodesByType returns all nodes of a specific type
func (g *Graph) GetNodesByType(nodeType NodeType) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

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
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*Node, 0)
	for _, node := range g.Nodes {
		if node.State == state {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetChildSteps returns all step nodes contained by a workflow
// Uses adjacency lists for O(D) instead of O(E)
func (g *Graph) GetChildSteps(workflowID string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	workflow, exists := g.Nodes[workflowID]
	if !exists {
		return []*Node{}
	}

	steps := make([]*Node, 0)
	for _, edge := range workflow.OutgoingEdges {
		if edge.Type == EdgeTypeContains {
			if stepNode, ok := g.Nodes[edge.ToNodeID]; ok {
				steps = append(steps, stepNode)
			}
		}
	}
	return steps
}

// GetParentWorkflow returns the parent workflow of a step node
// Uses adjacency lists for O(D) instead of O(E)
func (g *Graph) GetParentWorkflow(stepID string) (*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	step, exists := g.Nodes[stepID]
	if !exists {
		return nil, fmt.Errorf("step %s does not exist", stepID)
	}

	for _, edge := range step.IncomingEdges {
		if edge.Type == EdgeTypeContains {
			if workflow, ok := g.Nodes[edge.FromNodeID]; ok {
				return workflow, nil
			}
		}
	}
	return nil, fmt.Errorf("no parent workflow found for step %s", stepID)
}
