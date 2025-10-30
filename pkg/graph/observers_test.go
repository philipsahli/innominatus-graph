package graph

import (
	"sync"
	"testing"
)

// MockObserver implements GraphObserver for testing
type MockObserver struct {
	stateChanges   []StateChange
	nodeUpdates    []string
	edgeAdditions  []string
	graphUpdates   int
	mu             sync.Mutex
}

type StateChange struct {
	NodeID   string
	OldState NodeState
	NewState NodeState
}

func NewMockObserver() *MockObserver {
	return &MockObserver{
		stateChanges:  make([]StateChange, 0),
		nodeUpdates:   make([]string, 0),
		edgeAdditions: make([]string, 0),
	}
}

func (mo *MockObserver) OnNodeStateChanged(g *Graph, nodeID string, oldState, newState NodeState) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.stateChanges = append(mo.stateChanges, StateChange{
		NodeID:   nodeID,
		OldState: oldState,
		NewState: newState,
	})
}

func (mo *MockObserver) OnNodeUpdated(g *Graph, nodeID string) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.nodeUpdates = append(mo.nodeUpdates, nodeID)
}

func (mo *MockObserver) OnEdgeAdded(g *Graph, edge *Edge) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.edgeAdditions = append(mo.edgeAdditions, edge.ID)
}

func (mo *MockObserver) OnGraphUpdated(g *Graph) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.graphUpdates++
}

func (mo *MockObserver) GetStateChanges() []StateChange {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	return append([]StateChange{}, mo.stateChanges...)
}

func (mo *MockObserver) GetNodeUpdates() []string {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	return append([]string{}, mo.nodeUpdates...)
}

func (mo *MockObserver) GetEdgeAdditions() []string {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	return append([]string{}, mo.edgeAdditions...)
}

func (mo *MockObserver) GetGraphUpdates() int {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	return mo.graphUpdates
}

func TestNewObservableGraph(t *testing.T) {
	og := NewObservableGraph("test")

	if og.Graph == nil {
		t.Fatal("Expected graph to be initialized")
	}
	if og.AppName != "test" {
		t.Errorf("Expected app name 'test', got %s", og.AppName)
	}
	if og.GetObserverCount() != 0 {
		t.Errorf("Expected 0 observers, got %d", og.GetObserverCount())
	}
}

func TestWrapGraphAsObservable(t *testing.T) {
	g := NewGraph("test")
	g.AddNode(&Node{ID: "node-1", Type: NodeTypeSpec, Name: "Test"})

	og := WrapGraphAsObservable(g)

	if og.Graph != g {
		t.Error("Expected wrapped graph to be the same instance")
	}
	if len(og.Nodes) != 1 {
		t.Error("Expected wrapped graph to have existing nodes")
	}
}

func TestAddRemoveObserver(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()

	// Add observer
	og.AddObserver(observer)
	if og.GetObserverCount() != 1 {
		t.Errorf("Expected 1 observer, got %d", og.GetObserverCount())
	}

	// Remove observer
	og.RemoveObserver(observer)
	if og.GetObserverCount() != 0 {
		t.Errorf("Expected 0 observers after removal, got %d", og.GetObserverCount())
	}
}

func TestObserverNotifications_StateChange(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()
	og.AddObserver(observer)

	// Add node
	node := &Node{
		ID:    "node-1",
		Type:  NodeTypeWorkflow,
		Name:  "Test Workflow",
		State: NodeStateWaiting,
	}
	og.AddNode(node)

	// Update state
	og.UpdateNodeState("node-1", NodeStateRunning)

	// Verify notifications
	stateChanges := observer.GetStateChanges()
	if len(stateChanges) != 1 {
		t.Fatalf("Expected 1 state change notification, got %d", len(stateChanges))
	}

	change := stateChanges[0]
	if change.NodeID != "node-1" {
		t.Errorf("Expected node-1, got %s", change.NodeID)
	}
	if change.OldState != NodeStateWaiting {
		t.Errorf("Expected old state waiting, got %s", change.OldState)
	}
	if change.NewState != NodeStateRunning {
		t.Errorf("Expected new state running, got %s", change.NewState)
	}

	// Verify node update notification
	nodeUpdates := observer.GetNodeUpdates()
	if len(nodeUpdates) != 1 {
		t.Errorf("Expected 1 node update notification, got %d", len(nodeUpdates))
	}
}

func TestObserverNotifications_EdgeAdded(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()
	og.AddObserver(observer)

	// Add nodes (workflow can create spec)
	og.AddNode(&Node{ID: "node-1", Type: NodeTypeWorkflow, Name: "Node 1"})
	og.AddNode(&Node{ID: "node-2", Type: NodeTypeSpec, Name: "Node 2"})

	// Add edge (workflow creates spec - valid)
	edge := &Edge{
		ID:         "edge-1",
		FromNodeID: "node-1",
		ToNodeID:   "node-2",
		Type:       EdgeTypeCreates,
	}
	err := og.AddEdge(edge)
	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	// Verify edge addition notification
	edgeAdditions := observer.GetEdgeAdditions()
	if len(edgeAdditions) != 1 {
		t.Fatalf("Expected 1 edge addition notification, got %d", len(edgeAdditions))
	}
	if edgeAdditions[0] != "edge-1" {
		t.Errorf("Expected edge-1, got %s", edgeAdditions[0])
	}
}

func TestObserverNotifications_GraphUpdated(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()
	og.AddObserver(observer)

	// Operations that trigger graph updates
	og.AddNode(&Node{ID: "node-1", Type: NodeTypeWorkflow, Name: "Node 1"})
	og.AddNode(&Node{ID: "node-2", Type: NodeTypeSpec, Name: "Node 2"})

	edge := &Edge{
		ID:         "edge-1",
		FromNodeID: "node-1",  // Workflow creates spec
		ToNodeID:   "node-2",
		Type:       EdgeTypeCreates,
	}
	og.AddEdge(edge)

	og.RemoveEdge("edge-1")
	og.RemoveNode("node-2")

	// Verify graph update count
	graphUpdates := observer.GetGraphUpdates()
	// AddNode x2, AddEdge, RemoveEdge, RemoveNode = 5 updates
	if graphUpdates != 5 {
		t.Errorf("Expected 5 graph updates, got %d", graphUpdates)
	}
}

func TestMultipleObservers(t *testing.T) {
	og := NewObservableGraph("test")

	observer1 := NewMockObserver()
	observer2 := NewMockObserver()

	og.AddObserver(observer1)
	og.AddObserver(observer2)

	// Add node
	og.AddNode(&Node{ID: "node-1", Type: NodeTypeSpec, Name: "Test"})

	// Both observers should be notified
	if observer1.GetGraphUpdates() != 1 {
		t.Errorf("Observer 1: expected 1 graph update, got %d", observer1.GetGraphUpdates())
	}
	if observer2.GetGraphUpdates() != 1 {
		t.Errorf("Observer 2: expected 1 graph update, got %d", observer2.GetGraphUpdates())
	}
}

func TestObserver_NoNotificationAfterRemoval(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()

	og.AddObserver(observer)

	// First operation - should notify
	og.AddNode(&Node{ID: "node-1", Type: NodeTypeSpec, Name: "Test"})

	// Remove observer
	og.RemoveObserver(observer)

	// Second operation - should NOT notify
	og.AddNode(&Node{ID: "node-2", Type: NodeTypeWorkflow, Name: "Test 2"})

	// Should only have 1 notification (from before removal)
	graphUpdates := observer.GetGraphUpdates()
	if graphUpdates != 1 {
		t.Errorf("Expected 1 graph update (before removal), got %d", graphUpdates)
	}
}

func TestObserver_StateChangePropagation(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()
	og.AddObserver(observer)

	// Add workflow and step
	og.AddNode(&Node{ID: "workflow-1", Type: NodeTypeWorkflow, Name: "Workflow", State: NodeStateWaiting})
	og.AddNode(&Node{ID: "step-1", Type: NodeTypeStep, Name: "Step", State: NodeStateWaiting})

	// Add contains edge
	og.AddEdge(&Edge{
		ID:         "edge-1",
		FromNodeID: "workflow-1",
		ToNodeID:   "step-1",
		Type:       EdgeTypeContains,
	})

	// Update step to failed (propagates to workflow internally)
	og.UpdateNodeState("step-1", NodeStateFailed)

	// Verify step state change was notified
	stateChanges := observer.GetStateChanges()
	if len(stateChanges) < 1 {
		t.Fatalf("Expected at least 1 state change (step), got %d", len(stateChanges))
	}

	// Find step failure
	foundStepFailure := false
	for _, change := range stateChanges {
		if change.NodeID == "step-1" && change.NewState == NodeStateFailed {
			foundStepFailure = true
		}
	}

	if !foundStepFailure {
		t.Error("Expected step-1 failure notification")
	}

	// Verify workflow was actually updated (even if not notified separately)
	workflow, _ := og.GetNode("workflow-1")
	if workflow.State != NodeStateFailed {
		t.Error("Expected workflow-1 to have failed state (propagated internally)")
	}
}

func TestObserver_ConcurrentAccess(t *testing.T) {
	og := NewObservableGraph("test")
	observer := NewMockObserver()
	og.AddObserver(observer)

	var wg sync.WaitGroup

	// Concurrent state updates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			nodeID := "node-1"
			if index == 0 {
				og.AddNode(&Node{ID: nodeID, Type: NodeTypeWorkflow, Name: "Test", State: NodeStateWaiting})
			}
			og.UpdateNodeState(nodeID, NodeStateRunning)
		}(i)
	}

	wg.Wait()

	// Should have notifications (exact count may vary due to concurrency)
	stateChanges := observer.GetStateChanges()
	if len(stateChanges) == 0 {
		t.Error("Expected state change notifications from concurrent updates")
	}
}
