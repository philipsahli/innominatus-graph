package graph_test

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

var _ = Describe("Edge Type Validation", func() {
	var g *graph.Graph

	BeforeEach(func() {
		g = graph.NewGraph("validation-test")
	})

	Describe("DependsOn edge type", func() {
		It("should allow any node type combination", func() {
			g.AddNode(&graph.Node{ID: "step-1", Type: graph.NodeTypeStep, Name: "Step 1"})
			g.AddNode(&graph.Node{ID: "step-2", Type: graph.NodeTypeStep, Name: "Step 2"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step-1",
				ToNodeID:   "step-2",
				Type:       graph.EdgeTypeDependsOn,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Provisions edge type", func() {
		It("should allow workflow -> resource", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "wf",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeProvisions,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-workflow source", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeProvisions,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from workflow"))
		})

		It("should reject non-resource target", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "wf",
				ToNodeID:   "step",
				Type:       graph.EdgeTypeProvisions,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target resource"))
		})
	})

	Describe("Creates edge type", func() {
		It("should allow workflow -> any", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "wf",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeCreates,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-workflow source", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeCreates,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from workflow"))
		})
	})

	Describe("BindsTo edge type", func() {
		It("should allow any -> resource", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeBindsTo,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-resource target", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeBindsTo,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target resource"))
		})
	})

	Describe("Contains edge type", func() {
		Context("workflow -> step", func() {
			It("should allow workflow containing step", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

				err := g.AddEdge(&graph.Edge{
					ID:         "e1",
					FromNodeID: "wf",
					ToNodeID:   "step",
					Type:       graph.EdgeTypeContains,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("spec -> resource", func() {
			It("should allow spec containing resource", func() {
				g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

				err := g.AddEdge(&graph.Edge{
					ID:         "e1",
					FromNodeID: "spec",
					ToNodeID:   "res",
					Type:       graph.EdgeTypeContains,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should reject invalid combinations", func() {
			g.AddNode(&graph.Node{ID: "step1", Type: graph.NodeTypeStep, Name: "Step 1"})
			g.AddNode(&graph.Node{ID: "step2", Type: graph.NodeTypeStep, Name: "Step 2"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step1",
				ToNodeID:   "step2",
				Type:       graph.EdgeTypeContains,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("requires (spec→resource) or (workflow→step)"))
		})
	})

	Describe("Configures edge type", func() {
		It("should allow step -> resource", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeConfigures,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-step source", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "wf",
				ToNodeID:   "res",
				Type:       graph.EdgeTypeConfigures,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from step"))
		})

		It("should reject non-resource target", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeConfigures,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target resource"))
		})
	})

	Describe("Requires edge type", func() {
		It("should allow resource -> provider", func() {
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
			g.AddNode(&graph.Node{ID: "prov", Type: graph.NodeTypeProvider, Name: "Provider"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "res",
				ToNodeID:   "prov",
				Type:       graph.EdgeTypeRequires,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-resource source", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "prov", Type: graph.NodeTypeProvider, Name: "Provider"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "prov",
				Type:       graph.EdgeTypeRequires,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from resource"))
		})

		It("should reject non-provider target", func() {
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "res",
				ToNodeID:   "step",
				Type:       graph.EdgeTypeRequires,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target provider"))
		})
	})

	Describe("Executes edge type", func() {
		It("should allow provider -> workflow", func() {
			g.AddNode(&graph.Node{ID: "prov", Type: graph.NodeTypeProvider, Name: "Provider"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "prov",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeExecutes,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-provider source", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeExecutes,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from provider"))
		})

		It("should reject non-workflow target", func() {
			g.AddNode(&graph.Node{ID: "prov", Type: graph.NodeTypeProvider, Name: "Provider"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "prov",
				ToNodeID:   "step",
				Type:       graph.EdgeTypeExecutes,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target workflow"))
		})
	})

	Describe("Triggers edge type", func() {
		It("should allow spec -> workflow", func() {
			g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "spec",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeTriggers,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-spec source", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeTriggers,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from spec"))
		})

		It("should reject non-workflow target", func() {
			g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "spec",
				ToNodeID:   "step",
				Type:       graph.EdgeTypeTriggers,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target workflow"))
		})
	})

	Describe("Owns edge type", func() {
		It("should allow team -> application", func() {
			g.AddNode(&graph.Node{ID: "team", Type: graph.NodeTypeTeam, Name: "Team"})
			g.AddNode(&graph.Node{ID: "app", Type: graph.NodeTypeApplication, Name: "Application"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "team",
				ToNodeID:   "app",
				Type:       graph.EdgeTypeOwns,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-team source", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddNode(&graph.Node{ID: "app", Type: graph.NodeTypeApplication, Name: "Application"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "step",
				ToNodeID:   "app",
				Type:       graph.EdgeTypeOwns,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from team"))
		})

		It("should reject non-application target", func() {
			g.AddNode(&graph.Node{ID: "team", Type: graph.NodeTypeTeam, Name: "Team"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "team",
				ToNodeID:   "step",
				Type:       graph.EdgeTypeOwns,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target application"))
		})
	})

	Describe("HasSpec edge type", func() {
		It("should allow application -> spec", func() {
			g.AddNode(&graph.Node{ID: "app", Type: graph.NodeTypeApplication, Name: "Application"})
			g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "app",
				ToNodeID:   "spec",
				Type:       graph.EdgeTypeHasSpec,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject non-application source", func() {
			g.AddNode(&graph.Node{ID: "team", Type: graph.NodeTypeTeam, Name: "Team"})
			g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "team",
				ToNodeID:   "spec",
				Type:       graph.EdgeTypeHasSpec,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only originate from application"))
		})

		It("should reject non-spec target", func() {
			g.AddNode(&graph.Node{ID: "app", Type: graph.NodeTypeApplication, Name: "Application"})
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "app",
				ToNodeID:   "wf",
				Type:       graph.EdgeTypeHasSpec,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("can only target spec"))
		})
	})

	Describe("Invalid edge type", func() {
		It("should return error for unknown edge type", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Node 1"})
			g.AddNode(&graph.Node{ID: "n2", Type: graph.NodeTypeStep, Name: "Node 2"})

			err := g.AddEdge(&graph.Edge{
				ID:         "e1",
				FromNodeID: "n1",
				ToNodeID:   "n2",
				Type:       "unknown-edge-type",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid edge type"))
		})
	})
})

var _ = Describe("State Propagation", func() {
	var g *graph.Graph

	BeforeEach(func() {
		g = graph.NewGraph("state-test")
	})

	Describe("Step failure propagation to parent workflow", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow", State: graph.NodeStateRunning})
			g.AddNode(&graph.Node{ID: "step-1", Type: graph.NodeTypeStep, Name: "Step 1", State: graph.NodeStateRunning})
			g.AddNode(&graph.Node{ID: "step-2", Type: graph.NodeTypeStep, Name: "Step 2", State: graph.NodeStateWaiting})

			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step-1", Type: graph.EdgeTypeContains})
			g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "wf", ToNodeID: "step-2", Type: graph.EdgeTypeContains})
		})

		It("should propagate step failure to parent workflow", func() {
			err := g.UpdateNodeState("step-1", graph.NodeStateFailed)
			Expect(err).NotTo(HaveOccurred())

			wf, _ := g.GetNode("wf")
			Expect(wf.State).To(Equal(graph.NodeStateFailed))
		})

		It("should not affect sibling steps when one step fails", func() {
			g.UpdateNodeState("step-1", graph.NodeStateFailed)

			step2, _ := g.GetNode("step-2")
			Expect(step2.State).To(Equal(graph.NodeStateWaiting))
		})
	})

	Describe("Workflow completion propagation to contained steps", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow", State: graph.NodeStateRunning})
			g.AddNode(&graph.Node{ID: "step-1", Type: graph.NodeTypeStep, Name: "Step 1", State: graph.NodeStateRunning})
			g.AddNode(&graph.Node{ID: "step-2", Type: graph.NodeTypeStep, Name: "Step 2", State: graph.NodeStateWaiting})
			g.AddNode(&graph.Node{ID: "step-3", Type: graph.NodeTypeStep, Name: "Step 3", State: graph.NodeStateSucceeded})

			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step-1", Type: graph.EdgeTypeContains})
			g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "wf", ToNodeID: "step-2", Type: graph.EdgeTypeContains})
			g.AddEdge(&graph.Edge{ID: "e3", FromNodeID: "wf", ToNodeID: "step-3", Type: graph.EdgeTypeContains})
		})

		It("should update running steps to succeeded when workflow succeeds", func() {
			g.UpdateNodeState("wf", graph.NodeStateSucceeded)

			step1, _ := g.GetNode("step-1")
			Expect(step1.State).To(Equal(graph.NodeStateSucceeded))
		})

		It("should not affect waiting steps when workflow completes", func() {
			g.UpdateNodeState("wf", graph.NodeStateSucceeded)

			step2, _ := g.GetNode("step-2")
			Expect(step2.State).To(Equal(graph.NodeStateWaiting))
		})

		It("should not affect already succeeded steps", func() {
			g.UpdateNodeState("wf", graph.NodeStateSucceeded)

			step3, _ := g.GetNode("step-3")
			Expect(step3.State).To(Equal(graph.NodeStateSucceeded))
		})

		It("should update running steps to failed when workflow fails", func() {
			g.UpdateNodeState("wf", graph.NodeStateFailed)

			step1, _ := g.GetNode("step-1")
			Expect(step1.State).To(Equal(graph.NodeStateFailed))
		})
	})

	Describe("State timing updates", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step", State: graph.NodeStateWaiting})
		})

		It("should set StartedAt when transitioning to running", func() {
			g.UpdateNodeState("step", graph.NodeStateRunning)

			step, _ := g.GetNode("step")
			Expect(step.StartedAt).NotTo(BeNil())
		})

		It("should set CompletedAt when transitioning to succeeded", func() {
			g.UpdateNodeState("step", graph.NodeStateRunning)
			g.UpdateNodeState("step", graph.NodeStateSucceeded)

			step, _ := g.GetNode("step")
			Expect(step.CompletedAt).NotTo(BeNil())
		})

		It("should calculate Duration when completed", func() {
			g.UpdateNodeState("step", graph.NodeStateRunning)
			g.UpdateNodeState("step", graph.NodeStateSucceeded)

			step, _ := g.GetNode("step")
			Expect(step.Duration).NotTo(BeNil())
		})
	})
})

var _ = Describe("Thread Safety", func() {
	var g *graph.Graph

	BeforeEach(func() {
		g = graph.NewGraph("concurrent-test")
		for i := 0; i < 10; i++ {
			g.AddNode(&graph.Node{
				ID:    nodeID(i),
				Type:  graph.NodeTypeStep,
				Name:  "Step",
				State: graph.NodeStateWaiting,
			})
		}
	})

	It("should safely add nodes concurrently", func() {
		g2 := graph.NewGraph("concurrent-add")
		var wg sync.WaitGroup
		errors := make(chan error, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				err := g2.AddNode(&graph.Node{
					ID:   nodeID(idx),
					Type: graph.NodeTypeStep,
					Name: "Step",
				})
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		Expect(errors).To(BeEmpty())
		Expect(g2.Nodes).To(HaveLen(100))
	})

	It("should safely update node states concurrently", func() {
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				g.UpdateNodeState(nodeID(idx), graph.NodeStateRunning)
			}(i)
		}

		wg.Wait()

		for i := 0; i < 10; i++ {
			node, exists := g.GetNode(nodeID(i))
			Expect(exists).To(BeTrue())
			Expect(node.State).To(Equal(graph.NodeStateRunning))
		}
	})

	It("should safely read nodes while writing", func() {
		var wg sync.WaitGroup
		reads := make(chan *graph.Node, 100)

		// Start readers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					node, _ := g.GetNode(nodeID(j % 10))
					if node != nil {
						reads <- node
					}
				}
			}()
		}

		// Start writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				g.UpdateNodeState(nodeID(idx), graph.NodeStateRunning)
			}(i)
		}

		wg.Wait()
		close(reads)

		readCount := 0
		for range reads {
			readCount++
		}
		Expect(readCount).To(BeNumerically(">", 0))
	})
})

var _ = Describe("Node and Edge Operations", func() {
	var g *graph.Graph

	BeforeEach(func() {
		g = graph.NewGraph("ops-test")
	})

	Describe("AddNode", func() {
		It("should reject nil node", func() {
			err := g.AddNode(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node cannot be nil"))
		})

		It("should reject empty ID", func() {
			err := g.AddNode(&graph.Node{ID: "", Type: graph.NodeTypeStep, Name: "Step"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ID cannot be empty"))
		})

		It("should reject duplicate node ID", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step"})
			err := g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already exists"))
		})

		It("should initialize state to waiting if not set", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step"})
			node, _ := g.GetNode("n1")
			Expect(node.State).To(Equal(graph.NodeStateWaiting))
		})
	})

	Describe("AddEdge", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
		})

		It("should reject nil edge", func() {
			err := g.AddEdge(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("edge cannot be nil"))
		})

		It("should reject empty ID", func() {
			err := g.AddEdge(&graph.Edge{ID: "", FromNodeID: "wf", ToNodeID: "step", Type: graph.EdgeTypeContains})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ID cannot be empty"))
		})

		It("should reject non-existent from node", func() {
			err := g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "nonexistent", ToNodeID: "step", Type: graph.EdgeTypeContains})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})

		It("should reject non-existent to node", func() {
			err := g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "nonexistent", Type: graph.EdgeTypeContains})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})
	})

	Describe("RemoveNode", func() {
		It("should return error for non-existent node", func() {
			err := g.RemoveNode("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})

		It("should remove connected edges when removing node", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step", Type: graph.EdgeTypeContains})

			g.RemoveNode("step")

			_, exists := g.GetEdge("e1")
			Expect(exists).To(BeFalse())
		})
	})

	Describe("RemoveEdge", func() {
		It("should return error for non-existent edge", func() {
			err := g.RemoveEdge("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})

		It("should successfully remove existing edge", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step", Type: graph.EdgeTypeContains})

			err := g.RemoveEdge("e1")
			Expect(err).NotTo(HaveOccurred())

			_, exists := g.GetEdge("e1")
			Expect(exists).To(BeFalse())
		})
	})

	Describe("GetNodesByType", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "wf1", Type: graph.NodeTypeWorkflow, Name: "Workflow 1"})
			g.AddNode(&graph.Node{ID: "wf2", Type: graph.NodeTypeWorkflow, Name: "Workflow 2"})
			g.AddNode(&graph.Node{ID: "step1", Type: graph.NodeTypeStep, Name: "Step 1"})
		})

		It("should return all nodes of specified type", func() {
			workflows := g.GetNodesByType(graph.NodeTypeWorkflow)
			Expect(workflows).To(HaveLen(2))
		})

		It("should return empty slice for type with no nodes", func() {
			resources := g.GetNodesByType(graph.NodeTypeResource)
			Expect(resources).To(BeEmpty())
		})
	})

	Describe("GetNodesByState", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step 1", State: graph.NodeStateRunning})
			g.AddNode(&graph.Node{ID: "n2", Type: graph.NodeTypeStep, Name: "Step 2", State: graph.NodeStateRunning})
			g.AddNode(&graph.Node{ID: "n3", Type: graph.NodeTypeStep, Name: "Step 3", State: graph.NodeStateWaiting})
		})

		It("should return all nodes in specified state", func() {
			running := g.GetNodesByState(graph.NodeStateRunning)
			Expect(running).To(HaveLen(2))
		})

		It("should return empty slice for state with no nodes", func() {
			failed := g.GetNodesByState(graph.NodeStateFailed)
			Expect(failed).To(BeEmpty())
		})
	})

	Describe("GetChildSteps", func() {
		It("should return all steps contained by workflow", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step1", Type: graph.NodeTypeStep, Name: "Step 1"})
			g.AddNode(&graph.Node{ID: "step2", Type: graph.NodeTypeStep, Name: "Step 2"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step1", Type: graph.EdgeTypeContains})
			g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "wf", ToNodeID: "step2", Type: graph.EdgeTypeContains})

			steps := g.GetChildSteps("wf")
			Expect(steps).To(HaveLen(2))
		})

		It("should return empty slice for non-existent workflow", func() {
			steps := g.GetChildSteps("nonexistent")
			Expect(steps).To(BeEmpty())
		})
	})

	Describe("GetParentWorkflow", func() {
		It("should return parent workflow of step", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step", Type: graph.EdgeTypeContains})

			parent, err := g.GetParentWorkflow("step")
			Expect(err).NotTo(HaveOccurred())
			Expect(parent.ID).To(Equal("wf"))
		})

		It("should return error for step without parent", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			_, err := g.GetParentWorkflow("step")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no parent workflow"))
		})

		It("should return error for non-existent step", func() {
			_, err := g.GetParentWorkflow("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})
	})
})

func nodeID(i int) string {
	return "node-" + string(rune('0'+i))
}
