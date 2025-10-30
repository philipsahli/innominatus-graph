package execution

import (
	"testing"

	"github.com/philipsahli/innominatus-graph/pkg/storage"

	"github.com/philipsahli/innominatus-graph/pkg/graph"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) LoadGraph(appName string) (*graph.Graph, error) {
	args := m.Called(appName)
	return args.Get(0).(*graph.Graph), args.Error(1)
}

func (m *MockRepository) CreateGraphRun(appName string, version int) (*storage.GraphRunModel, error) {
	args := m.Called(appName, version)
	return args.Get(0).(*storage.GraphRunModel), args.Error(1)
}

func (m *MockRepository) UpdateGraphRun(runID uuid.UUID, status string, errorMessage *string) error {
	args := m.Called(runID, status, errorMessage)
	return args.Error(0)
}

func (m *MockRepository) SaveGraph(appName string, g *graph.Graph) error {
	args := m.Called(appName, g)
	return args.Error(0)
}

func (m *MockRepository) GetGraphRuns(appName string) ([]storage.GraphRunModel, error) {
	args := m.Called(appName)
	return args.Get(0).([]storage.GraphRunModel), args.Error(1)
}

func (m *MockRepository) UpdateNodeState(appName string, nodeID string, state graph.NodeState) error {
	args := m.Called(appName, nodeID, state)
	return args.Error(0)
}

// Mock WorkflowRunner
type MockWorkflowRunnerTest struct {
	mock.Mock
}

func (m *MockWorkflowRunnerTest) RunWorkflow(node *graph.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockWorkflowRunnerTest) ProvisionResource(workflow *graph.Node, resource *graph.Node) error {
	args := m.Called(workflow, resource)
	return args.Error(0)
}

func (m *MockWorkflowRunnerTest) CreateResource(workflow *graph.Node, target *graph.Node) error {
	args := m.Called(workflow, target)
	return args.Error(0)
}

func createTestGraphForExecution() *graph.Graph {
	g := graph.NewGraph("test-app")

	nodes := []*graph.Node{
		{ID: "spec1", Type: graph.NodeTypeSpec, Name: "Database Spec"},
		{ID: "workflow1", Type: graph.NodeTypeWorkflow, Name: "Deploy Database"},
		{ID: "resource1", Type: graph.NodeTypeResource, Name: "Database"},
		{ID: "workflow2", Type: graph.NodeTypeWorkflow, Name: "Deploy API"},
		{ID: "resource2", Type: graph.NodeTypeResource, Name: "API Service"},
	}

	for _, node := range nodes {
		require.NoError(nil, g.AddNode(node))
	}

	edges := []*graph.Edge{
		{ID: "e1", FromNodeID: "workflow1", ToNodeID: "spec1", Type: graph.EdgeTypeDependsOn},
		{ID: "e2", FromNodeID: "workflow1", ToNodeID: "resource1", Type: graph.EdgeTypeProvisions},
		{ID: "e3", FromNodeID: "workflow2", ToNodeID: "workflow1", Type: graph.EdgeTypeDependsOn}, // workflow2 depends on workflow1
		{ID: "e4", FromNodeID: "workflow2", ToNodeID: "resource2", Type: graph.EdgeTypeProvisions},
	}

	for _, edge := range edges {
		require.NoError(nil, g.AddEdge(edge))
	}

	return g
}

func TestEngine_ExecuteGraph_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRunner := &MockWorkflowRunnerTest{}

	g := createTestGraphForExecution()
	mockRepo.On("LoadGraph", "test-app").Return(g, nil)

	runModel := &storage.GraphRunModel{ID: uuid.New()}
	mockRepo.On("CreateGraphRun", "test-app", 1).Return(runModel, nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "running", (*string)(nil)).Return(nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "completed", (*string)(nil)).Return(nil)

	// Expect workflow executions
	mockRunner.On("RunWorkflow", mock.AnythingOfType("*graph.Node")).Return(nil)
	mockRunner.On("ProvisionResource", mock.AnythingOfType("*graph.Node"), mock.AnythingOfType("*graph.Node")).Return(nil)

	engine := NewEngine(mockRepo, mockRunner)

	plan, err := engine.ExecuteGraph("test-app")
	require.NoError(t, err)

	assert.Equal(t, "test-app", plan.AppName)
	assert.Equal(t, StatusCompleted, plan.Status)
	assert.Len(t, plan.Executions, 5)

	// Check that spec1 was executed first
	spec1Exec := plan.Executions["spec1"]
	assert.Equal(t, StatusCompleted, spec1Exec.Status)

	// Check that workflow1 was executed after spec1
	workflow1Exec := plan.Executions["workflow1"]
	assert.Equal(t, StatusCompleted, workflow1Exec.Status)

	mockRepo.AssertExpectations(t)
	mockRunner.AssertExpectations(t)
}

func TestEngine_ExecuteGraph_WorkflowFailure(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRunner := &MockWorkflowRunnerTest{}

	g := createTestGraphForExecution()
	mockRepo.On("LoadGraph", "test-app").Return(g, nil)

	runModel := &storage.GraphRunModel{ID: uuid.New()}
	mockRepo.On("CreateGraphRun", "test-app", 1).Return(runModel, nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "running", (*string)(nil)).Return(nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "failed", mock.AnythingOfType("*string")).Return(nil)

	// Make workflow1 fail
	mockRunner.On("RunWorkflow", mock.MatchedBy(func(node *graph.Node) bool {
		return node.ID == "workflow1"
	})).Return(assert.AnError)

	// workflow2 should not be executed at all due to dependency failure

	engine := NewEngine(mockRepo, mockRunner)

	plan, err := engine.ExecuteGraph("test-app")
	require.NoError(t, err)

	assert.Equal(t, StatusFailed, plan.Status)

	workflow1Exec := plan.Executions["workflow1"]
	assert.Equal(t, StatusFailed, workflow1Exec.Status)
	assert.NotEmpty(t, workflow1Exec.Error)

	// workflow2 should be skipped due to failed dependency
	workflow2Exec := plan.Executions["workflow2"]
	assert.Equal(t, StatusSkipped, workflow2Exec.Status)

	mockRepo.AssertExpectations(t)
	mockRunner.AssertExpectations(t)
}

func TestEngine_shouldExecuteNode(t *testing.T) {
	g := createTestGraphForExecution()
	engine := NewEngine(nil, nil)

	plan := &ExecutionPlan{
		Executions: map[string]*NodeExecution{
			"spec1":     {NodeID: "spec1", Status: StatusCompleted},
			"workflow1": {NodeID: "workflow1", Status: StatusFailed},
			"resource1": {NodeID: "resource1", Status: StatusPending},
		},
	}

	// spec1 has no dependencies - should execute
	shouldExecute := engine.shouldExecuteNode(g.Nodes["spec1"], plan, g)
	assert.True(t, shouldExecute)

	// workflow1 depends on spec1 which completed - should execute
	shouldExecute = engine.shouldExecuteNode(g.Nodes["workflow1"], plan, g)
	assert.True(t, shouldExecute)

	// workflow2 depends on workflow1 which failed - should not execute
	shouldExecute = engine.shouldExecuteNode(g.Nodes["workflow2"], plan, g)
	assert.False(t, shouldExecute)
}

func TestMockWorkflowRunner(t *testing.T) {
	runner := NewMockWorkflowRunner()

	node := &graph.Node{ID: "test", Type: graph.NodeTypeWorkflow, Name: "test-workflow"}
	err := runner.RunWorkflow(node)
	assert.NoError(t, err)

	// Test failing workflow
	failingNode := &graph.Node{ID: "fail", Type: graph.NodeTypeWorkflow, Name: "failing-workflow"}
	err = runner.RunWorkflow(failingNode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock workflow failure")
}

func TestMockWorkflowRunner_ProvisionResource(t *testing.T) {
	runner := NewMockWorkflowRunner()

	workflow := &graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "workflow"}
	resource := &graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "resource"}

	err := runner.ProvisionResource(workflow, resource)
	assert.NoError(t, err)
}

func TestMockWorkflowRunner_CreateResource(t *testing.T) {
	runner := NewMockWorkflowRunner()

	workflow := &graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "workflow"}
	target := &graph.Node{ID: "tgt", Type: graph.NodeTypeResource, Name: "target"}

	err := runner.CreateResource(workflow, target)
	assert.NoError(t, err)
}

func TestEngine_RegisterObserver(t *testing.T) {
	engine := NewEngine(nil, nil)

	observer1 := &MockObserver{}
	observer2 := &MockObserver{}

	engine.RegisterObserver(observer1)
	engine.RegisterObserver(observer2)

	assert.Len(t, engine.observers, 2)
}

func TestEngine_NotifyStateChange(t *testing.T) {
	engine := NewEngine(nil, nil)

	observer := &MockObserver{}
	engine.RegisterObserver(observer)

	node := &graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Test"}

	observer.On("OnNodeStateChange", node, graph.NodeStateWaiting, graph.NodeStateRunning).Return()

	engine.notifyStateChange(node, graph.NodeStateWaiting, graph.NodeStateRunning)

	observer.AssertExpectations(t)
}

func TestEngine_ExecuteWorkflow_WithCreatesEdge(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRunner := &MockWorkflowRunnerTest{}

	g := graph.NewGraph("test-app")
	workflow := &graph.Node{ID: "wf1", Type: graph.NodeTypeWorkflow, Name: "Deploy"}
	resource := &graph.Node{ID: "res1", Type: graph.NodeTypeResource, Name: "Database"}

	g.AddNode(workflow)
	g.AddNode(resource)

	// Creates edge instead of Provisions
	edge := &graph.Edge{
		ID:         "e1",
		FromNodeID: "wf1",
		ToNodeID:   "res1",
		Type:       graph.EdgeTypeCreates,
	}
	g.AddEdge(edge)

	g.Version = 1 // Set version to 1
	mockRepo.On("LoadGraph", "test-app").Return(g, nil)

	runModel := &storage.GraphRunModel{ID: uuid.New()}
	mockRepo.On("CreateGraphRun", "test-app", 1).Return(runModel, nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "running", (*string)(nil)).Return(nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "completed", (*string)(nil)).Return(nil)

	mockRunner.On("RunWorkflow", mock.AnythingOfType("*graph.Node")).Return(nil)
	mockRunner.On("CreateResource", mock.AnythingOfType("*graph.Node"), mock.AnythingOfType("*graph.Node")).Return(nil)

	engine := NewEngine(mockRepo, mockRunner)

	plan, err := engine.ExecuteGraph("test-app")
	require.NoError(t, err)

	assert.Equal(t, StatusCompleted, plan.Status)

	mockRepo.AssertExpectations(t)
	mockRunner.AssertExpectations(t)
}

func TestEngine_ExecuteStep_WithConfiguresEdge(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRunner := &MockWorkflowRunnerTest{}

	g := graph.NewGraph("test-app")
	workflow := &graph.Node{ID: "wf1", Type: graph.NodeTypeWorkflow, Name: "Deploy"}
	step := &graph.Node{ID: "step1", Type: graph.NodeTypeStep, Name: "Configure"}
	resource := &graph.Node{ID: "res1", Type: graph.NodeTypeResource, Name: "Database"}

	g.AddNode(workflow)
	g.AddNode(step)
	g.AddNode(resource)

	g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf1", ToNodeID: "step1", Type: graph.EdgeTypeContains})
	g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "step1", ToNodeID: "res1", Type: graph.EdgeTypeConfigures})

	g.Version = 1 // Set version to 1
	mockRepo.On("LoadGraph", "test-app").Return(g, nil)

	runModel := &storage.GraphRunModel{ID: uuid.New()}
	mockRepo.On("CreateGraphRun", "test-app", 1).Return(runModel, nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "running", (*string)(nil)).Return(nil)
	mockRepo.On("UpdateGraphRun", runModel.ID, "completed", (*string)(nil)).Return(nil)

	mockRunner.On("RunWorkflow", mock.AnythingOfType("*graph.Node")).Return(nil)

	engine := NewEngine(mockRepo, mockRunner)

	plan, err := engine.ExecuteGraph("test-app")
	require.NoError(t, err)

	assert.Equal(t, StatusCompleted, plan.Status)

	stepExec := plan.Executions["step1"]
	assert.Equal(t, StatusCompleted, stepExec.Status)

	// Check that logs contain resource configuration
	found := false
	for _, log := range stepExec.Logs {
		if log == "Configuring resource: Database" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected log about configuring resource")

	mockRepo.AssertExpectations(t)
	mockRunner.AssertExpectations(t)
}

// MockObserver for testing observer pattern
type MockObserver struct {
	mock.Mock
}

func (m *MockObserver) OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState) {
	m.Called(node, oldState, newState)
}
