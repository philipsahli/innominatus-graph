package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGraph() *Graph {
	g := NewGraph("test")

	nodes := []*Node{
		{ID: "spec1", Type: NodeTypeSpec, Name: "Database Spec"},
		{ID: "spec2", Type: NodeTypeSpec, Name: "API Spec"},
		{ID: "workflow1", Type: NodeTypeWorkflow, Name: "Deploy Database"},
		{ID: "workflow2", Type: NodeTypeWorkflow, Name: "Deploy API"},
		{ID: "resource1", Type: NodeTypeResource, Name: "Database"},
		{ID: "resource2", Type: NodeTypeResource, Name: "API Service"},
	}

	for _, node := range nodes {
		require.NoError(nil, g.AddNode(node))
	}

	edges := []*Edge{
		{ID: "e1", FromNodeID: "workflow1", ToNodeID: "spec1", Type: EdgeTypeDependsOn},
		{ID: "e2", FromNodeID: "workflow2", ToNodeID: "spec2", Type: EdgeTypeDependsOn},
		{ID: "e3", FromNodeID: "workflow2", ToNodeID: "resource1", Type: EdgeTypeDependsOn},
		{ID: "e4", FromNodeID: "workflow1", ToNodeID: "resource1", Type: EdgeTypeProvisions},
		{ID: "e5", FromNodeID: "workflow2", ToNodeID: "resource2", Type: EdgeTypeProvisions},
	}

	for _, edge := range edges {
		require.NoError(nil, g.AddEdge(edge))
	}

	return g
}

func TestGraph_TopologicalSort(t *testing.T) {
	g := createTestGraph()

	sorted, err := g.TopologicalSort()
	require.NoError(t, err)

	assert.Len(t, sorted, 6)

	nodePositions := make(map[string]int)
	for i, node := range sorted {
		nodePositions[node.ID] = i
	}

	assert.True(t, nodePositions["spec1"] < nodePositions["workflow1"], "spec1 should come before workflow1")
	assert.True(t, nodePositions["spec2"] < nodePositions["workflow2"], "spec2 should come before workflow2")
	assert.True(t, nodePositions["resource1"] < nodePositions["workflow2"], "resource1 should come before workflow2")
}

func TestGraph_TopologicalSort_WithCycle(t *testing.T) {
	g := NewGraph("test")

	node1 := &Node{ID: "node1", Type: NodeTypeSpec, Name: "Node 1"}
	node2 := &Node{ID: "node2", Type: NodeTypeWorkflow, Name: "Node 2"}
	node3 := &Node{ID: "node3", Type: NodeTypeResource, Name: "Node 3"}

	require.NoError(t, g.AddNode(node1))
	require.NoError(t, g.AddNode(node2))
	require.NoError(t, g.AddNode(node3))

	edges := []*Edge{
		{ID: "e1", FromNodeID: "node1", ToNodeID: "node2", Type: EdgeTypeDependsOn},
		{ID: "e2", FromNodeID: "node2", ToNodeID: "node3", Type: EdgeTypeDependsOn},
		{ID: "e3", FromNodeID: "node3", ToNodeID: "node1", Type: EdgeTypeDependsOn},
	}

	for _, edge := range edges {
		require.NoError(t, g.AddEdge(edge))
	}

	_, err := g.TopologicalSort()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycles")
}

func TestGraph_TopologicalSort_EmptyGraph(t *testing.T) {
	g := NewGraph("test")

	sorted, err := g.TopologicalSort()
	require.NoError(t, err)
	assert.Empty(t, sorted)
}

func TestGraph_TopologicalSort_SingleNode(t *testing.T) {
	g := NewGraph("test")

	node := &Node{ID: "node1", Type: NodeTypeSpec, Name: "Single Node"}
	require.NoError(t, g.AddNode(node))

	sorted, err := g.TopologicalSort()
	require.NoError(t, err)
	assert.Len(t, sorted, 1)
	assert.Equal(t, "node1", sorted[0].ID)
}

func TestGraph_GetDependencies(t *testing.T) {
	g := createTestGraph()

	deps, err := g.GetDependencies("workflow1")
	require.NoError(t, err)
	assert.Len(t, deps, 1)
	assert.Equal(t, "spec1", deps[0].ID)

	deps, err = g.GetDependencies("workflow2")
	require.NoError(t, err)
	assert.Len(t, deps, 2)

	depIDs := make([]string, len(deps))
	for i, dep := range deps {
		depIDs[i] = dep.ID
	}
	assert.Contains(t, depIDs, "spec2")
	assert.Contains(t, depIDs, "resource1")
}

func TestGraph_GetDependencies_NotFound(t *testing.T) {
	g := createTestGraph()

	_, err := g.GetDependencies("missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGraph_GetDependencies_NoDeps(t *testing.T) {
	g := createTestGraph()

	deps, err := g.GetDependencies("spec1")
	require.NoError(t, err)
	assert.Empty(t, deps)
}

func TestGraph_GetDependents(t *testing.T) {
	g := createTestGraph()

	dependents, err := g.GetDependents("spec1")
	require.NoError(t, err)
	assert.Len(t, dependents, 1)
	assert.Equal(t, "workflow1", dependents[0].ID)

	dependents, err = g.GetDependents("resource1")
	require.NoError(t, err)
	assert.Len(t, dependents, 1)
	assert.Equal(t, "workflow2", dependents[0].ID)
}

func TestGraph_GetDependents_NotFound(t *testing.T) {
	g := createTestGraph()

	_, err := g.GetDependents("missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGraph_GetDependents_NoDependents(t *testing.T) {
	g := createTestGraph()

	dependents, err := g.GetDependents("resource2")
	require.NoError(t, err)
	assert.Empty(t, dependents)
}

func TestGraph_HasCycle(t *testing.T) {
	g := createTestGraph()
	assert.False(t, g.HasCycle())

	// Create a cycle: Add an edge that makes spec1 depend on workflow1
	// Since workflow1 already depends on spec1 (e1), this creates a cycle
	cycleEdge := &Edge{
		ID:         "cycle",
		FromNodeID: "spec1",
		ToNodeID:   "workflow1",
		Type:       EdgeTypeDependsOn,
	}
	require.NoError(t, g.AddEdge(cycleEdge))

	assert.True(t, g.HasCycle())
}