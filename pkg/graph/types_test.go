package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGraph(t *testing.T) {
	appName := "test-app"
	g := NewGraph(appName)

	assert.Equal(t, "test-app-graph", g.ID)
	assert.Equal(t, appName, g.AppName)
	assert.Equal(t, 1, g.Version)
	assert.NotNil(t, g.Nodes)
	assert.NotNil(t, g.Edges)
	assert.Empty(t, g.Nodes)
	assert.Empty(t, g.Edges)
}

func TestGraph_AddNode(t *testing.T) {
	g := NewGraph("test")

	node := &Node{
		ID:   "node1",
		Type: NodeTypeSpec,
		Name: "Test Node",
	}

	err := g.AddNode(node)
	require.NoError(t, err)

	assert.Len(t, g.Nodes, 1)
	assert.Equal(t, node, g.Nodes["node1"])
	assert.False(t, node.CreatedAt.IsZero())
	assert.False(t, node.UpdatedAt.IsZero())
}

func TestGraph_AddNode_Validation(t *testing.T) {
	g := NewGraph("test")

	tests := []struct {
		name    string
		node    *Node
		wantErr string
	}{
		{
			name:    "nil node",
			node:    nil,
			wantErr: "node cannot be nil",
		},
		{
			name: "empty ID",
			node: &Node{
				Type: NodeTypeSpec,
				Name: "Test",
			},
			wantErr: "node ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.AddNode(tt.node)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGraph_AddNode_Duplicate(t *testing.T) {
	g := NewGraph("test")

	node1 := &Node{ID: "node1", Type: NodeTypeSpec, Name: "Test 1"}
	node2 := &Node{ID: "node1", Type: NodeTypeWorkflow, Name: "Test 2"}

	err := g.AddNode(node1)
	require.NoError(t, err)

	err = g.AddNode(node2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGraph_AddEdge(t *testing.T) {
	g := NewGraph("test")

	node1 := &Node{ID: "node1", Type: NodeTypeSpec, Name: "Spec"}
	node2 := &Node{ID: "node2", Type: NodeTypeWorkflow, Name: "Workflow"}

	require.NoError(t, g.AddNode(node1))
	require.NoError(t, g.AddNode(node2))

	edge := &Edge{
		ID:         "edge1",
		FromNodeID: "node1",
		ToNodeID:   "node2",
		Type:       EdgeTypeDependsOn,
	}

	err := g.AddEdge(edge)
	require.NoError(t, err)

	assert.Len(t, g.Edges, 1)
	assert.Equal(t, edge, g.Edges["edge1"])
	assert.False(t, edge.CreatedAt.IsZero())
}

func TestGraph_AddEdge_Validation(t *testing.T) {
	g := NewGraph("test")

	workflow := &Node{ID: "workflow1", Type: NodeTypeWorkflow, Name: "Workflow"}
	resource := &Node{ID: "resource1", Type: NodeTypeResource, Name: "Resource"}
	spec := &Node{ID: "spec1", Type: NodeTypeSpec, Name: "Spec"}

	require.NoError(t, g.AddNode(workflow))
	require.NoError(t, g.AddNode(resource))
	require.NoError(t, g.AddNode(spec))

	tests := []struct {
		name    string
		edge    *Edge
		wantErr string
	}{
		{
			name:    "nil edge",
			edge:    nil,
			wantErr: "edge cannot be nil",
		},
		{
			name: "empty ID",
			edge: &Edge{
				FromNodeID: "workflow1",
				ToNodeID:   "resource1",
				Type:       EdgeTypeProvisions,
			},
			wantErr: "edge ID cannot be empty",
		},
		{
			name: "non-existent from node",
			edge: &Edge{
				ID:         "edge1",
				FromNodeID: "missing",
				ToNodeID:   "resource1",
				Type:       EdgeTypeProvisions,
			},
			wantErr: "from node missing does not exist",
		},
		{
			name: "non-existent to node",
			edge: &Edge{
				ID:         "edge1",
				FromNodeID: "workflow1",
				ToNodeID:   "missing",
				Type:       EdgeTypeProvisions,
			},
			wantErr: "to node missing does not exist",
		},
		{
			name: "invalid provisions edge from spec",
			edge: &Edge{
				ID:         "edge1",
				FromNodeID: "spec1",
				ToNodeID:   "resource1",
				Type:       EdgeTypeProvisions,
			},
			wantErr: "provisions edge can only originate from workflow nodes",
		},
		{
			name: "invalid provisions edge to spec",
			edge: &Edge{
				ID:         "edge1",
				FromNodeID: "workflow1",
				ToNodeID:   "spec1",
				Type:       EdgeTypeProvisions,
			},
			wantErr: "provisions edge can only target resource nodes",
		},
		{
			name: "invalid creates edge from resource",
			edge: &Edge{
				ID:         "edge1",
				FromNodeID: "resource1",
				ToNodeID:   "spec1",
				Type:       EdgeTypeCreates,
			},
			wantErr: "creates edge can only originate from workflow nodes",
		},
		{
			name: "invalid binds-to edge to spec",
			edge: &Edge{
				ID:         "edge1",
				FromNodeID: "spec1",
				ToNodeID:   "spec1",
				Type:       EdgeTypeBindsTo,
			},
			wantErr: "binds-to edge can only target resource nodes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.AddEdge(tt.edge)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGraph_ValidEdges(t *testing.T) {
	g := NewGraph("test")

	workflow := &Node{ID: "workflow1", Type: NodeTypeWorkflow, Name: "Workflow"}
	resource := &Node{ID: "resource1", Type: NodeTypeResource, Name: "Resource"}
	spec := &Node{ID: "spec1", Type: NodeTypeSpec, Name: "Spec"}

	require.NoError(t, g.AddNode(workflow))
	require.NoError(t, g.AddNode(resource))
	require.NoError(t, g.AddNode(spec))

	validEdges := []*Edge{
		{ID: "edge1", FromNodeID: "spec1", ToNodeID: "workflow1", Type: EdgeTypeDependsOn},
		{ID: "edge2", FromNodeID: "workflow1", ToNodeID: "resource1", Type: EdgeTypeProvisions},
		{ID: "edge3", FromNodeID: "workflow1", ToNodeID: "spec1", Type: EdgeTypeCreates},
		{ID: "edge4", FromNodeID: "spec1", ToNodeID: "resource1", Type: EdgeTypeBindsTo},
	}

	for _, edge := range validEdges {
		err := g.AddEdge(edge)
		assert.NoError(t, err, "Edge %s should be valid", edge.ID)
	}

	assert.Len(t, g.Edges, 4)
}

func TestGraph_RemoveNode(t *testing.T) {
	g := NewGraph("test")

	node1 := &Node{ID: "node1", Type: NodeTypeSpec, Name: "Node 1"}
	node2 := &Node{ID: "node2", Type: NodeTypeWorkflow, Name: "Node 2"}

	require.NoError(t, g.AddNode(node1))
	require.NoError(t, g.AddNode(node2))

	edge := &Edge{
		ID:         "edge1",
		FromNodeID: "node1",
		ToNodeID:   "node2",
		Type:       EdgeTypeDependsOn,
	}
	require.NoError(t, g.AddEdge(edge))

	originalTime := g.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	err := g.RemoveNode("node1")
	require.NoError(t, err)

	assert.Len(t, g.Nodes, 1)
	assert.Len(t, g.Edges, 0) // Edge should be removed
	assert.True(t, g.UpdatedAt.After(originalTime))

	_, exists := g.GetNode("node1")
	assert.False(t, exists)

	_, exists = g.GetEdge("edge1")
	assert.False(t, exists)
}

func TestGraph_RemoveNode_NotFound(t *testing.T) {
	g := NewGraph("test")

	err := g.RemoveNode("missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGraph_RemoveEdge(t *testing.T) {
	g := NewGraph("test")

	node1 := &Node{ID: "node1", Type: NodeTypeSpec, Name: "Node 1"}
	node2 := &Node{ID: "node2", Type: NodeTypeWorkflow, Name: "Node 2"}

	require.NoError(t, g.AddNode(node1))
	require.NoError(t, g.AddNode(node2))

	edge := &Edge{
		ID:         "edge1",
		FromNodeID: "node1",
		ToNodeID:   "node2",
		Type:       EdgeTypeDependsOn,
	}
	require.NoError(t, g.AddEdge(edge))

	originalTime := g.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	err := g.RemoveEdge("edge1")
	require.NoError(t, err)

	assert.Len(t, g.Nodes, 2)
	assert.Len(t, g.Edges, 0)
	assert.True(t, g.UpdatedAt.After(originalTime))

	_, exists := g.GetEdge("edge1")
	assert.False(t, exists)
}

func TestGraph_RemoveEdge_NotFound(t *testing.T) {
	g := NewGraph("test")

	err := g.RemoveEdge("missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}