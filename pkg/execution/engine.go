package execution

import (
	"fmt"
	"log"
	"time"

	"idp-orchestrator/pkg/graph"
	"idp-orchestrator/pkg/storage"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	StatusPending    ExecutionStatus = "pending"
	StatusRunning    ExecutionStatus = "running"
	StatusCompleted  ExecutionStatus = "completed"
	StatusFailed     ExecutionStatus = "failed"
	StatusSkipped    ExecutionStatus = "skipped"
)

type NodeExecution struct {
	NodeID    string          `json:"node_id"`
	Status    ExecutionStatus `json:"status"`
	StartTime *time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time      `json:"end_time,omitempty"`
	Error     string          `json:"error,omitempty"`
	Logs      []string        `json:"logs,omitempty"`
}

type ExecutionPlan struct {
	RunID      uuid.UUID                `json:"run_id"`
	AppName    string                   `json:"app_name"`
	Version    int                      `json:"version"`
	Status     ExecutionStatus          `json:"status"`
	StartTime  time.Time                `json:"start_time"`
	EndTime    *time.Time               `json:"end_time,omitempty"`
	Executions map[string]*NodeExecution `json:"executions"`
	Order      []*graph.Node            `json:"order"`
}

type Engine struct {
	repository storage.RepositoryInterface
	runner     WorkflowRunner
}

type WorkflowRunner interface {
	RunWorkflow(node *graph.Node) error
	ProvisionResource(workflow *graph.Node, resource *graph.Node) error
	CreateResource(workflow *graph.Node, target *graph.Node) error
}

func NewEngine(repository storage.RepositoryInterface, runner WorkflowRunner) *Engine {
	return &Engine{
		repository: repository,
		runner:     runner,
	}
}

func (e *Engine) ExecuteGraph(appName string) (*ExecutionPlan, error) {
	g, err := e.repository.LoadGraph(appName)
	if err != nil {
		return nil, fmt.Errorf("failed to load graph: %w", err)
	}

	sortedNodes, err := g.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to sort graph topologically: %w", err)
	}

	graphRun, err := e.repository.CreateGraphRun(appName, g.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph run: %w", err)
	}

	plan := &ExecutionPlan{
		RunID:      graphRun.ID,
		AppName:    appName,
		Version:    g.Version,
		Status:     StatusRunning,
		StartTime:  time.Now(),
		Executions: make(map[string]*NodeExecution),
		Order:      sortedNodes,
	}

	for _, node := range sortedNodes {
		plan.Executions[node.ID] = &NodeExecution{
			NodeID: node.ID,
			Status: StatusPending,
			Logs:   make([]string, 0),
		}
	}

	err = e.repository.UpdateGraphRun(graphRun.ID, string(StatusRunning), nil)
	if err != nil {
		log.Printf("Failed to update graph run status: %v", err)
	}

	executionSuccess := true
	for _, node := range sortedNodes {
		execution := plan.Executions[node.ID]

		if !e.shouldExecuteNode(node, plan, g) {
			execution.Status = StatusSkipped
			execution.Logs = append(execution.Logs, "Skipped due to failed dependencies")
			continue
		}

		if err := e.executeNode(node, execution, g); err != nil {
			execution.Status = StatusFailed
			execution.Error = err.Error()
			execution.Logs = append(execution.Logs, fmt.Sprintf("Execution failed: %v", err))
			executionSuccess = false
			log.Printf("Node %s failed: %v", node.ID, err)
		} else {
			execution.Status = StatusCompleted
			execution.Logs = append(execution.Logs, "Execution completed successfully")
		}

		if execution.EndTime == nil {
			now := time.Now()
			execution.EndTime = &now
		}
	}

	endTime := time.Now()
	plan.EndTime = &endTime

	if executionSuccess {
		plan.Status = StatusCompleted
		err = e.repository.UpdateGraphRun(graphRun.ID, string(StatusCompleted), nil)
	} else {
		plan.Status = StatusFailed
		errorMsg := "Some nodes failed to execute"
		err = e.repository.UpdateGraphRun(graphRun.ID, string(StatusFailed), &errorMsg)
	}

	if err != nil {
		log.Printf("Failed to update final graph run status: %v", err)
	}

	return plan, nil
}

func (e *Engine) shouldExecuteNode(node *graph.Node, plan *ExecutionPlan, g *graph.Graph) bool {
	dependencies, err := g.GetDependencies(node.ID)
	if err != nil {
		return false
	}

	for _, dep := range dependencies {
		if execution, exists := plan.Executions[dep.ID]; exists {
			if execution.Status == StatusFailed {
				return false
			}
		}
	}

	return true
}

func (e *Engine) executeNode(node *graph.Node, execution *NodeExecution, g *graph.Graph) error {
	startTime := time.Now()
	execution.StartTime = &startTime
	execution.Status = StatusRunning

	execution.Logs = append(execution.Logs, fmt.Sprintf("Starting execution of %s (%s)", node.Name, node.Type))

	switch node.Type {
	case graph.NodeTypeWorkflow:
		return e.executeWorkflow(node, execution, g)
	case graph.NodeTypeSpec:
		return e.executeSpec(node, execution)
	case graph.NodeTypeResource:
		return e.executeResource(node, execution, g)
	default:
		return fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func (e *Engine) executeWorkflow(node *graph.Node, execution *NodeExecution, g *graph.Graph) error {
	execution.Logs = append(execution.Logs, "Executing workflow...")

	if err := e.runner.RunWorkflow(node); err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	for _, edge := range g.Edges {
		if edge.FromNodeID == node.ID {
			targetNode, exists := g.GetNode(edge.ToNodeID)
			if !exists {
				continue
			}

			switch edge.Type {
			case graph.EdgeTypeProvisions:
				execution.Logs = append(execution.Logs, fmt.Sprintf("Provisioning resource: %s", targetNode.Name))
				if err := e.runner.ProvisionResource(node, targetNode); err != nil {
					return fmt.Errorf("resource provisioning failed: %w", err)
				}
			case graph.EdgeTypeCreates:
				execution.Logs = append(execution.Logs, fmt.Sprintf("Creating resource: %s", targetNode.Name))
				if err := e.runner.CreateResource(node, targetNode); err != nil {
					return fmt.Errorf("resource creation failed: %w", err)
				}
			}
		}
	}

	execution.Logs = append(execution.Logs, "Workflow execution completed")
	return nil
}

func (e *Engine) executeSpec(node *graph.Node, execution *NodeExecution) error {
	execution.Logs = append(execution.Logs, "Processing spec node...")
	execution.Logs = append(execution.Logs, "Spec validation completed")
	return nil
}

func (e *Engine) executeResource(node *graph.Node, execution *NodeExecution, g *graph.Graph) error {
	execution.Logs = append(execution.Logs, "Validating resource state...")

	provisioners := make([]*graph.Node, 0)
	for _, edge := range g.Edges {
		if edge.ToNodeID == node.ID && (edge.Type == graph.EdgeTypeProvisions || edge.Type == graph.EdgeTypeCreates) {
			if provisionerNode, exists := g.GetNode(edge.FromNodeID); exists {
				provisioners = append(provisioners, provisionerNode)
			}
		}
	}

	if len(provisioners) == 0 {
		execution.Logs = append(execution.Logs, "No provisioners found - resource may be external")
	} else {
		execution.Logs = append(execution.Logs, fmt.Sprintf("Resource provisioned by %d workflow(s)", len(provisioners)))
	}

	execution.Logs = append(execution.Logs, "Resource validation completed")
	return nil
}

type MockWorkflowRunner struct{}

func (r *MockWorkflowRunner) RunWorkflow(node *graph.Node) error {
	log.Printf("Mock: Running workflow %s (%s)", node.Name, node.ID)
	time.Sleep(100 * time.Millisecond)

	if node.Name == "failing-workflow" {
		return fmt.Errorf("mock workflow failure")
	}

	return nil
}

func (r *MockWorkflowRunner) ProvisionResource(workflow *graph.Node, resource *graph.Node) error {
	log.Printf("Mock: Workflow %s provisioning resource %s", workflow.Name, resource.Name)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (r *MockWorkflowRunner) CreateResource(workflow *graph.Node, target *graph.Node) error {
	log.Printf("Mock: Workflow %s creating resource %s", workflow.Name, target.Name)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func NewMockWorkflowRunner() WorkflowRunner {
	return &MockWorkflowRunner{}
}