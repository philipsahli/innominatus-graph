package storage

import (
	"os"
	"testing"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_SaveAndLoadGraph(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create test graph
	g := graph.NewGraph("test-app")
	
	workflow := &graph.Node{
		ID:          "wf-1",
		Type:        graph.NodeTypeWorkflow,
		Name:        "Test Workflow",
		Description: "Test workflow description",
		State:       graph.NodeStateWaiting,
	}
	err = g.AddNode(workflow)
	require.NoError(t, err)

	step := &graph.Node{
		ID:          "step-1",
		Type:        graph.NodeTypeStep,
		Name:        "Test Step",
		Description: "Test step description",
		State:       graph.NodeStateWaiting,
	}
	err = g.AddNode(step)
	require.NoError(t, err)

	edge := &graph.Edge{
		ID:          "edge-1",
		FromNodeID:  "wf-1",
		ToNodeID:    "step-1",
		Type:        graph.EdgeTypeContains,
		Description: "Workflow contains step",
	}
	err = g.AddEdge(edge)
	require.NoError(t, err)

	// Save graph
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	// Load graph
	loaded, err := repo.LoadGraph("test-app")
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify nodes
	assert.Len(t, loaded.Nodes, 2)
	
	loadedWorkflow, exists := loaded.GetNode("wf-1")
	assert.True(t, exists)
	assert.Equal(t, "Test Workflow", loadedWorkflow.Name)
	assert.Equal(t, graph.NodeTypeWorkflow, loadedWorkflow.Type)

	loadedStep, exists := loaded.GetNode("step-1")
	assert.True(t, exists)
	assert.Equal(t, "Test Step", loadedStep.Name)
	assert.Equal(t, graph.NodeTypeStep, loadedStep.Type)

	// Verify edges
	assert.Len(t, loaded.Edges, 1)
	loadedEdge, exists := loaded.Edges["edge-1"]
	assert.True(t, exists)
	assert.Equal(t, "wf-1", loadedEdge.FromNodeID)
	assert.Equal(t, "step-1", loadedEdge.ToNodeID)
	assert.Equal(t, graph.EdgeTypeContains, loadedEdge.Type)
}

func TestRepository_SaveGraph_UpdatesExisting(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Save initial graph
	g1 := graph.NewGraph("test-app")
	node1 := &graph.Node{
		ID:   "n1",
		Type: graph.NodeTypeWorkflow,
		Name: "Original",
	}
	g1.AddNode(node1)
	
	err = repo.SaveGraph("test-app", g1)
	require.NoError(t, err)

	// Save updated graph
	g2 := graph.NewGraph("test-app")
	node2 := &graph.Node{
		ID:   "n2",
		Type: graph.NodeTypeStep,
		Name: "Updated",
	}
	g2.AddNode(node2)
	
	err = repo.SaveGraph("test-app", g2)
	require.NoError(t, err)

	// Load and verify only updated graph exists
	loaded, err := repo.LoadGraph("test-app")
	require.NoError(t, err)
	
	assert.Len(t, loaded.Nodes, 1)
	_, exists := loaded.GetNode("n2")
	assert.True(t, exists)
	_, exists = loaded.GetNode("n1")
	assert.False(t, exists)
}

func TestRepository_LoadGraph_NotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Try to load non-existent app
	_, err = repo.LoadGraph("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRepository_UpdateNodeState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create and save graph
	g := graph.NewGraph("test-app")
	node := &graph.Node{
		ID:    "n1",
		Type:  graph.NodeTypeStep,
		Name:  "Test Node",
		State: graph.NodeStateWaiting,
	}
	g.AddNode(node)
	
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	// Update state
	err = repo.UpdateNodeState("test-app", "n1", graph.NodeStateRunning)
	require.NoError(t, err)

	// Load and verify state changed
	loaded, err := repo.LoadGraph("test-app")
	require.NoError(t, err)
	
	loadedNode, exists := loaded.GetNode("n1")
	assert.True(t, exists)
	assert.Equal(t, graph.NodeStateRunning, loadedNode.State)
}

func TestRepository_UpdateNodeState_NodeNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create empty app
	g := graph.NewGraph("test-app")
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	// Try to update non-existent node
	err = repo.UpdateNodeState("test-app", "non-existent", graph.NodeStateRunning)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRepository_CreateGraphRun(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create app first
	g := graph.NewGraph("test-app")
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	// Create graph run
	run, err := repo.CreateGraphRun("test-app", 1)
	require.NoError(t, err)
	require.NotNil(t, run)
	
	assert.Equal(t, 1, run.Version)
	assert.Equal(t, "pending", run.Status)
}

func TestRepository_GetGraphRuns(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create app
	g := graph.NewGraph("test-app")
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	// Create multiple runs
	_, err = repo.CreateGraphRun("test-app", 1)
	require.NoError(t, err)
	
	_, err = repo.CreateGraphRun("test-app", 2)
	require.NoError(t, err)

	// Get runs
	runs, err := repo.GetGraphRuns("test-app")
	require.NoError(t, err)
	assert.Len(t, runs, 2)
}

func TestRepository_UpdateGraphRun(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create app and run
	g := graph.NewGraph("test-app")
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	run, err := repo.CreateGraphRun("test-app", 1)
	require.NoError(t, err)

	// Update run status
	err = repo.UpdateGraphRun(run.ID, "completed", nil)
	require.NoError(t, err)

	// Verify update
	runs, err := repo.GetGraphRuns("test-app")
	require.NoError(t, err)
	assert.Equal(t, "completed", runs[0].Status)
}

func TestRepository_NodeToModel_WithProperties(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create graph with node properties
	g := graph.NewGraph("test-app")
	node := &graph.Node{
		ID:   "n1",
		Type: graph.NodeTypeStep,
		Name: "Node with properties",
		Properties: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}
	g.AddNode(node)

	// Save and load
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	loaded, err := repo.LoadGraph("test-app")
	require.NoError(t, err)

	loadedNode, exists := loaded.GetNode("n1")
	assert.True(t, exists)
	assert.NotNil(t, loadedNode.Properties)
	assert.Equal(t, "value1", loadedNode.Properties["key1"])
	// JSON unmarshaling converts numbers to float64
	assert.Equal(t, float64(42), loadedNode.Properties["key2"])
}

func TestRepository_EdgeToModel_WithProperties(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewSQLiteConnection(tmpFile.Name())
	require.NoError(t, err)
	
	err = AutoMigrate(db)
	require.NoError(t, err)

	repo := NewRepository(db)

	// Create graph with edge properties
	g := graph.NewGraph("test-app")
	
	n1 := &graph.Node{ID: "n1", Type: graph.NodeTypeWorkflow, Name: "Node 1"}
	n2 := &graph.Node{ID: "n2", Type: graph.NodeTypeStep, Name: "Node 2"}
	g.AddNode(n1)
	g.AddNode(n2)

	edge := &graph.Edge{
		ID:         "e1",
		FromNodeID: "n1",
		ToNodeID:   "n2",
		Type:       graph.EdgeTypeContains,
		Properties: map[string]interface{}{
			"weight": 1.5,
			"label":  "test",
		},
	}
	g.AddEdge(edge)

	// Save and load
	err = repo.SaveGraph("test-app", g)
	require.NoError(t, err)

	loaded, err := repo.LoadGraph("test-app")
	require.NoError(t, err)

	loadedEdge, exists := loaded.Edges["e1"]
	assert.True(t, exists)
	assert.NotNil(t, loadedEdge.Properties)
	assert.Equal(t, 1.5, loadedEdge.Properties["weight"])
	assert.Equal(t, "test", loadedEdge.Properties["label"])
}
