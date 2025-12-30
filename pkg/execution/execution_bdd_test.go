package execution_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/philipsahli/innominatus-graph/pkg/execution"
	"github.com/philipsahli/innominatus-graph/pkg/graph"
	"github.com/philipsahli/innominatus-graph/pkg/storage"
)

// MockRunner implements WorkflowRunner for testing
type MockRunner struct {
	WorkflowCalls   []string
	ProvisionCalls  []string
	CreateCalls     []string
	ShouldFail      bool
	FailOn          string
	ProvisionFail   bool
	CreateFail      bool
	StepCalls       []string
	StepImplemented bool
}

func NewMockRunner() *MockRunner {
	return &MockRunner{
		WorkflowCalls:  make([]string, 0),
		ProvisionCalls: make([]string, 0),
		CreateCalls:    make([]string, 0),
		StepCalls:      make([]string, 0),
	}
}

func (m *MockRunner) RunWorkflow(node *graph.Node) error {
	m.WorkflowCalls = append(m.WorkflowCalls, node.ID)
	if m.ShouldFail && (m.FailOn == "" || m.FailOn == node.ID) {
		return fmt.Errorf("workflow %s failed", node.ID)
	}
	return nil
}

func (m *MockRunner) ProvisionResource(workflow *graph.Node, resource *graph.Node) error {
	m.ProvisionCalls = append(m.ProvisionCalls, resource.ID)
	if m.ProvisionFail {
		return fmt.Errorf("provision failed for %s", resource.ID)
	}
	return nil
}

func (m *MockRunner) CreateResource(workflow *graph.Node, target *graph.Node) error {
	m.CreateCalls = append(m.CreateCalls, target.ID)
	if m.CreateFail {
		return fmt.Errorf("create failed for %s", target.ID)
	}
	return nil
}

func (m *MockRunner) RunStep(node *graph.Node) error {
	m.StepCalls = append(m.StepCalls, node.ID)
	return nil
}

// MockObserver tracks state changes
type MockObserver struct {
	Changes []StateChange
}

type StateChange struct {
	NodeID   string
	OldState graph.NodeState
	NewState graph.NodeState
}

func NewMockObserver() *MockObserver {
	return &MockObserver{Changes: make([]StateChange, 0)}
}

func (m *MockObserver) OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState) {
	m.Changes = append(m.Changes, StateChange{
		NodeID:   node.ID,
		OldState: oldState,
		NewState: newState,
	})
}

var _ = Describe("Execution Engine", func() {
	var repo *storage.Repository
	var runner *MockRunner
	var engine *execution.Engine
	var cleanup func()

	BeforeEach(func() {
		tmpFile, err := os.CreateTemp("", "bdd-exec-*.db")
		Expect(err).NotTo(HaveOccurred())

		db, err := storage.NewSQLiteConnection(tmpFile.Name())
		Expect(err).NotTo(HaveOccurred())

		err = storage.AutoMigrate(db)
		Expect(err).NotTo(HaveOccurred())

		repo = storage.NewRepository(db)
		runner = NewMockRunner()
		engine = execution.NewEngine(repo, runner)
		cleanup = func() { os.Remove(tmpFile.Name()) }
	})

	AfterEach(func() {
		if cleanup != nil {
			cleanup()
		}
	})

	Describe("ExecuteGraph", func() {
		Context("with successful execution", func() {
			It("should execute workflow node successfully", func() {
				g := graph.NewGraph("test-app")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				repo.SaveGraph("test-app", g)

				plan, err := engine.ExecuteGraph("test-app")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Status).To(Equal(execution.StatusCompleted))
				Expect(runner.WorkflowCalls).To(ContainElement("wf"))
			})

			It("should execute multiple nodes in topological order", func() {
				g := graph.NewGraph("multi-node")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step", Type: graph.EdgeTypeContains})
				repo.SaveGraph("multi-node", g)

				plan, err := engine.ExecuteGraph("multi-node")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Status).To(Equal(execution.StatusCompleted))
				Expect(len(plan.Executions)).To(Equal(2))
			})

			It("should record execution times", func() {
				g := graph.NewGraph("timing-test")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				repo.SaveGraph("timing-test", g)

				plan, err := engine.ExecuteGraph("timing-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.EndTime).NotTo(BeNil())
				Expect(plan.Executions["wf"].StartTime).NotTo(BeNil())
				Expect(plan.Executions["wf"].EndTime).NotTo(BeNil())
			})
		})

		Context("when workflow fails", func() {
			It("should mark execution as failed", func() {
				g := graph.NewGraph("fail-test")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				repo.SaveGraph("fail-test", g)

				runner.ShouldFail = true
				plan, err := engine.ExecuteGraph("fail-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Status).To(Equal(execution.StatusFailed))
				Expect(plan.Executions["wf"].Status).To(Equal(execution.StatusFailed))
			})

			It("should record error message", func() {
				g := graph.NewGraph("error-test")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				repo.SaveGraph("error-test", g)

				runner.ShouldFail = true
				plan, _ := engine.ExecuteGraph("error-test")
				Expect(plan.Executions["wf"].Error).To(ContainSubstring("failed"))
			})

			It("should skip dependent nodes when parent fails", func() {
				g := graph.NewGraph("skip-test")
				g.AddNode(&graph.Node{ID: "wf1", Type: graph.NodeTypeWorkflow, Name: "Workflow 1"})
				g.AddNode(&graph.Node{ID: "wf2", Type: graph.NodeTypeWorkflow, Name: "Workflow 2"})
				// wf2 depends on wf1, so wf1 must run first
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf2", ToNodeID: "wf1", Type: graph.EdgeTypeDependsOn})
				repo.SaveGraph("skip-test", g)

				runner.ShouldFail = true
				runner.FailOn = "wf1"
				plan, _ := engine.ExecuteGraph("skip-test")

				Expect(plan.Executions["wf1"].Status).To(Equal(execution.StatusFailed))
				Expect(plan.Executions["wf2"].Status).To(Equal(execution.StatusSkipped))
			})
		})

		Context("when graph loading fails", func() {
			It("should return error for non-existent app", func() {
				_, err := engine.ExecuteGraph("non-existent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to load graph"))
			})
		})

		Context("with provisions edge", func() {
			It("should provision resources", func() {
				g := graph.NewGraph("provision-test")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "res", Type: graph.EdgeTypeProvisions})
				repo.SaveGraph("provision-test", g)

				_, err := engine.ExecuteGraph("provision-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.ProvisionCalls).To(ContainElement("res"))
			})

			It("should fail when provision fails", func() {
				g := graph.NewGraph("provision-fail")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "res", Type: graph.EdgeTypeProvisions})
				repo.SaveGraph("provision-fail", g)

				runner.ProvisionFail = true
				plan, _ := engine.ExecuteGraph("provision-fail")
				Expect(plan.Status).To(Equal(execution.StatusFailed))
				Expect(plan.Executions["wf"].Error).To(ContainSubstring("provisioning failed"))
			})
		})

		Context("with creates edge", func() {
			It("should create resources", func() {
				g := graph.NewGraph("create-test")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "res", Type: graph.EdgeTypeCreates})
				repo.SaveGraph("create-test", g)

				_, err := engine.ExecuteGraph("create-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(runner.CreateCalls).To(ContainElement("res"))
			})

			It("should fail when create fails", func() {
				g := graph.NewGraph("create-fail")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "res", Type: graph.EdgeTypeCreates})
				repo.SaveGraph("create-fail", g)

				runner.CreateFail = true
				plan, _ := engine.ExecuteGraph("create-fail")
				Expect(plan.Status).To(Equal(execution.StatusFailed))
				Expect(plan.Executions["wf"].Error).To(ContainSubstring("creation failed"))
			})
		})

		Context("with step execution", func() {
			It("should execute step nodes", func() {
				g := graph.NewGraph("step-test")
				g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
				repo.SaveGraph("step-test", g)

				plan, err := engine.ExecuteGraph("step-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Executions["step"].Status).To(Equal(execution.StatusCompleted))
			})

			It("should process configures edges", func() {
				g := graph.NewGraph("config-test")
				g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "step", ToNodeID: "res", Type: graph.EdgeTypeConfigures})
				repo.SaveGraph("config-test", g)

				plan, err := engine.ExecuteGraph("config-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Executions["step"].Logs).To(ContainElement(ContainSubstring("Configuring resource")))
			})
		})

		Context("with spec execution", func() {
			It("should execute spec nodes", func() {
				g := graph.NewGraph("spec-test")
				g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})
				repo.SaveGraph("spec-test", g)

				plan, err := engine.ExecuteGraph("spec-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Executions["spec"].Status).To(Equal(execution.StatusCompleted))
				Expect(plan.Executions["spec"].Logs).To(ContainElement(ContainSubstring("validation completed")))
			})
		})

		Context("with resource execution", func() {
			It("should validate resource nodes", func() {
				g := graph.NewGraph("resource-test")
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				repo.SaveGraph("resource-test", g)

				plan, err := engine.ExecuteGraph("resource-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Executions["res"].Status).To(Equal(execution.StatusCompleted))
				Expect(plan.Executions["res"].Logs).To(ContainElement(ContainSubstring("validation completed")))
			})

			It("should identify provisioners", func() {
				g := graph.NewGraph("provisioner-test")
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "res", Type: graph.EdgeTypeProvisions})
				repo.SaveGraph("provisioner-test", g)

				plan, err := engine.ExecuteGraph("provisioner-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Executions["res"].Logs).To(ContainElement(ContainSubstring("provisioned by")))
			})

			It("should handle external resources", func() {
				g := graph.NewGraph("external-test")
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "External Resource"})
				repo.SaveGraph("external-test", g)

				plan, err := engine.ExecuteGraph("external-test")
				Expect(err).NotTo(HaveOccurred())
				Expect(plan.Executions["res"].Logs).To(ContainElement(ContainSubstring("may be external")))
			})
		})
	})

	Describe("Observer Pattern", func() {
		It("should notify observers on state changes", func() {
			observer := NewMockObserver()
			engine.RegisterObserver(observer)

			g := graph.NewGraph("observer-test")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			repo.SaveGraph("observer-test", g)

			_, err := engine.ExecuteGraph("observer-test")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(observer.Changes)).To(BeNumerically(">=", 2))

			// Should have transition to running
			runningChange := findChange(observer.Changes, "wf", graph.NodeStateRunning)
			Expect(runningChange).NotTo(BeNil())

			// Should have transition to succeeded
			succeededChange := findChange(observer.Changes, "wf", graph.NodeStateSucceeded)
			Expect(succeededChange).NotTo(BeNil())
		})

		It("should notify on failure", func() {
			observer := NewMockObserver()
			engine.RegisterObserver(observer)

			g := graph.NewGraph("fail-observer")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			repo.SaveGraph("fail-observer", g)

			runner.ShouldFail = true
			engine.ExecuteGraph("fail-observer")

			failedChange := findChange(observer.Changes, "wf", graph.NodeStateFailed)
			Expect(failedChange).NotTo(BeNil())
		})

		It("should support multiple observers", func() {
			observer1 := NewMockObserver()
			observer2 := NewMockObserver()
			engine.RegisterObserver(observer1)
			engine.RegisterObserver(observer2)

			g := graph.NewGraph("multi-observer")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			repo.SaveGraph("multi-observer", g)

			engine.ExecuteGraph("multi-observer")

			Expect(len(observer1.Changes)).To(Equal(len(observer2.Changes)))
		})
	})

	Describe("Execution Logging", func() {
		It("should log execution steps", func() {
			g := graph.NewGraph("log-test")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			repo.SaveGraph("log-test", g)

			plan, _ := engine.ExecuteGraph("log-test")
			Expect(plan.Executions["wf"].Logs).To(ContainElement(ContainSubstring("Starting execution")))
			Expect(plan.Executions["wf"].Logs).To(ContainElement(ContainSubstring("completed")))
		})

		It("should log failure messages", func() {
			g := graph.NewGraph("log-fail")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			repo.SaveGraph("log-fail", g)

			runner.ShouldFail = true
			plan, _ := engine.ExecuteGraph("log-fail")
			Expect(plan.Executions["wf"].Logs).To(ContainElement(ContainSubstring("failed")))
		})

		It("should log skipped nodes", func() {
			g := graph.NewGraph("log-skip")
			g.AddNode(&graph.Node{ID: "wf1", Type: graph.NodeTypeWorkflow, Name: "Workflow 1"})
			g.AddNode(&graph.Node{ID: "wf2", Type: graph.NodeTypeWorkflow, Name: "Workflow 2"})
			// wf2 depends on wf1, so wf1 runs first
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf2", ToNodeID: "wf1", Type: graph.EdgeTypeDependsOn})
			repo.SaveGraph("log-skip", g)

			runner.ShouldFail = true
			runner.FailOn = "wf1"
			plan, _ := engine.ExecuteGraph("log-skip")
			Expect(plan.Executions["wf2"].Logs).To(ContainElement(ContainSubstring("Skipped")))
		})
	})
})

func findChange(changes []StateChange, nodeID string, state graph.NodeState) *StateChange {
	for _, c := range changes {
		if c.NodeID == nodeID && c.NewState == state {
			return &c
		}
	}
	return nil
}
