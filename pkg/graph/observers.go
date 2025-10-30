package graph

import "sync"

// GraphObserver is an interface for observing graph changes
type GraphObserver interface {
	// OnNodeStateChanged is called when a node's state changes
	OnNodeStateChanged(g *Graph, nodeID string, oldState, newState NodeState)

	// OnNodeUpdated is called when a node is updated (including timing changes)
	OnNodeUpdated(g *Graph, nodeID string)

	// OnEdgeAdded is called when an edge is added to the graph
	OnEdgeAdded(g *Graph, edge *Edge)

	// OnGraphUpdated is called when the graph structure changes
	OnGraphUpdated(g *Graph)
}

// ObservableGraph wraps a Graph with observer pattern support
type ObservableGraph struct {
	*Graph
	observers []GraphObserver
	mu        sync.RWMutex
}

// NewObservableGraph creates a new observable graph
func NewObservableGraph(appName string) *ObservableGraph {
	return &ObservableGraph{
		Graph:     NewGraph(appName),
		observers: make([]GraphObserver, 0),
	}
}

// WrapGraphAsObservable wraps an existing graph with observable functionality
func WrapGraphAsObservable(g *Graph) *ObservableGraph {
	return &ObservableGraph{
		Graph:     g,
		observers: make([]GraphObserver, 0),
	}
}

// AddObserver registers a new observer
func (og *ObservableGraph) AddObserver(observer GraphObserver) {
	og.mu.Lock()
	defer og.mu.Unlock()
	og.observers = append(og.observers, observer)
}

// RemoveObserver unregisters an observer
func (og *ObservableGraph) RemoveObserver(observer GraphObserver) {
	og.mu.Lock()
	defer og.mu.Unlock()

	for i, obs := range og.observers {
		if obs == observer {
			og.observers = append(og.observers[:i], og.observers[i+1:]...)
			break
		}
	}
}

// notifyNodeStateChanged notifies all observers of a state change
func (og *ObservableGraph) notifyNodeStateChanged(nodeID string, oldState, newState NodeState) {
	og.mu.RLock()
	defer og.mu.RUnlock()

	for _, observer := range og.observers {
		observer.OnNodeStateChanged(og.Graph, nodeID, oldState, newState)
	}
}

// notifyNodeUpdated notifies all observers of a node update
func (og *ObservableGraph) notifyNodeUpdated(nodeID string) {
	og.mu.RLock()
	defer og.mu.RUnlock()

	for _, observer := range og.observers {
		observer.OnNodeUpdated(og.Graph, nodeID)
	}
}

// notifyEdgeAdded notifies all observers of an edge addition
func (og *ObservableGraph) notifyEdgeAdded(edge *Edge) {
	og.mu.RLock()
	defer og.mu.RUnlock()

	for _, observer := range og.observers {
		observer.OnEdgeAdded(og.Graph, edge)
	}
}

// notifyGraphUpdated notifies all observers of a graph update
func (og *ObservableGraph) notifyGraphUpdated() {
	og.mu.RLock()
	defer og.mu.RUnlock()

	for _, observer := range og.observers {
		observer.OnGraphUpdated(og.Graph)
	}
}

// UpdateNodeState overrides the base implementation to add notifications
func (og *ObservableGraph) UpdateNodeState(nodeID string, newState NodeState) error {
	node, exists := og.GetNode(nodeID)
	if !exists {
		return og.Graph.UpdateNodeState(nodeID, newState)
	}

	oldState := node.State

	// Update state using base implementation
	if err := og.Graph.UpdateNodeState(nodeID, newState); err != nil {
		return err
	}

	// Notify observers if state actually changed
	if oldState != newState {
		og.notifyNodeStateChanged(nodeID, oldState, newState)
	}

	// Always notify of update (timing may have changed)
	og.notifyNodeUpdated(nodeID)

	return nil
}

// AddNode overrides to add notifications
func (og *ObservableGraph) AddNode(node *Node) error {
	if err := og.Graph.AddNode(node); err != nil {
		return err
	}

	og.notifyGraphUpdated()
	return nil
}

// AddEdge overrides to add notifications
func (og *ObservableGraph) AddEdge(edge *Edge) error {
	if err := og.Graph.AddEdge(edge); err != nil {
		return err
	}

	og.notifyEdgeAdded(edge)
	og.notifyGraphUpdated()
	return nil
}

// RemoveNode overrides to add notifications
func (og *ObservableGraph) RemoveNode(id string) error {
	if err := og.Graph.RemoveNode(id); err != nil {
		return err
	}

	og.notifyGraphUpdated()
	return nil
}

// RemoveEdge overrides to add notifications
func (og *ObservableGraph) RemoveEdge(id string) error {
	if err := og.Graph.RemoveEdge(id); err != nil {
		return err
	}

	og.notifyGraphUpdated()
	return nil
}

// GetObserverCount returns the number of registered observers (for testing)
func (og *ObservableGraph) GetObserverCount() int {
	og.mu.RLock()
	defer og.mu.RUnlock()
	return len(og.observers)
}
