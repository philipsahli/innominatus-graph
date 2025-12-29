package export_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/philipsahli/innominatus-graph/pkg/export"
	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

var _ = Describe("DOT Export", func() {
	var exporter *export.Exporter

	BeforeEach(func() {
		exporter = export.NewExporter()
	})

	AfterEach(func() {
		exporter.Close()
	})

	Describe("ExportGraph", func() {
		var g *graph.Graph

		BeforeEach(func() {
			g = graph.NewGraph("test-app")
		})

		Context("with DOT format", func() {
			It("should generate valid DOT output", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("digraph"))
				Expect(string(output)).To(ContainSubstring("test-app"))
			})

			It("should include node labels", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "My Workflow"})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("My Workflow"))
			})

			It("should include edge labels", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step", Type: graph.EdgeTypeContains})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("contains"))
			})
		})

		Context("with SVG format", func() {
			It("should generate SVG output", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				output, err := exporter.ExportGraph(g, export.FormatSVG)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("svg"))
			})
		})

		Context("with unsupported format", func() {
			It("should return error", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				_, err := exporter.ExportGraph(g, "unknown")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported format"))
			})
		})

		Context("with different node states", func() {
			It("should include state in label for running nodes", func() {
				g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step", State: graph.NodeStateRunning})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("running"))
			})

			It("should include state in label for failed nodes", func() {
				g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step", State: graph.NodeStateFailed})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("failed"))
			})

			It("should include state in label for succeeded nodes", func() {
				g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step", State: graph.NodeStateSucceeded})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("succeeded"))
			})

			It("should not include state for waiting nodes", func() {
				g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Step", State: graph.NodeStateWaiting})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).NotTo(ContainSubstring("[waiting]"))
			})
		})

		Context("with different node types", func() {
			It("should style spec nodes", func() {
				g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#E3F2FD"))
			})

			It("should style workflow nodes", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#FFF9C4"))
			})

			It("should style step nodes", func() {
				g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#FFE0B2"))
			})

			It("should style resource nodes", func() {
				g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#C8E6C9"))
			})
		})

		Context("with different edge types", func() {
			BeforeEach(func() {
				g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeWorkflow, Name: "Node 1"})
				g.AddNode(&graph.Node{ID: "n2", Type: graph.NodeTypeResource, Name: "Node 2"})
			})

			It("should style depends-on edges", func() {
				g.AddNode(&graph.Node{ID: "s1", Type: graph.NodeTypeStep, Name: "Step 1"})
				g.AddNode(&graph.Node{ID: "s2", Type: graph.NodeTypeStep, Name: "Step 2"})
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "s1", ToNodeID: "s2", Type: graph.EdgeTypeDependsOn})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#1976D2"))
			})

			It("should style provisions edges", func() {
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "n1", ToNodeID: "n2", Type: graph.EdgeTypeProvisions})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#388E3C"))
			})

			It("should style creates edges", func() {
				g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "n1", ToNodeID: "n2", Type: graph.EdgeTypeCreates})

				output, err := exporter.ExportGraph(g, export.FormatDOT)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("#F57C00"))
			})
		})
	})

	Describe("CreateSubgraph", func() {
		It("should create subgraph with specified nodes", func() {
			g := graph.NewGraph("main")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step1", Type: graph.NodeTypeStep, Name: "Step 1"})
			g.AddNode(&graph.Node{ID: "step2", Type: graph.NodeTypeStep, Name: "Step 2"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step1", Type: graph.EdgeTypeContains})
			g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "wf", ToNodeID: "step2", Type: graph.EdgeTypeContains})

			subgraph, err := exporter.CreateSubgraph(g, []string{"wf", "step1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(subgraph.Nodes).To(HaveLen(2))
			Expect(subgraph.Edges).To(HaveLen(1))
		})

		It("should only include edges between specified nodes", func() {
			g := graph.NewGraph("main")
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
			g.AddNode(&graph.Node{ID: "step1", Type: graph.NodeTypeStep, Name: "Step 1"})
			g.AddNode(&graph.Node{ID: "step2", Type: graph.NodeTypeStep, Name: "Step 2"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "wf", ToNodeID: "step1", Type: graph.EdgeTypeContains})
			g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "wf", ToNodeID: "step2", Type: graph.EdgeTypeContains})

			subgraph, err := exporter.CreateSubgraph(g, []string{"step1", "step2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(subgraph.Edges).To(BeEmpty())
		})
	})
})

var _ = Describe("Mermaid Export", func() {
	var g *graph.Graph

	BeforeEach(func() {
		g = graph.NewGraph("test-app")
	})

	Describe("ExportGraphMermaid", func() {
		Context("with flowchart type", func() {
			It("should generate flowchart with default options", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				output, err := export.ExportGraphMermaid(g, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("flowchart TB"))
				Expect(output).To(ContainSubstring("Workflow"))
			})

			It("should include state when IncludeState is true", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow", State: graph.NodeStateRunning})

				options := export.DefaultMermaidOptions()
				options.IncludeState = true

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("[running]"))
			})

			It("should use custom direction", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				options := export.DefaultMermaidOptions()
				options.Direction = "LR"

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("flowchart LR"))
			})

			It("should apply custom theme", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				options := export.DefaultMermaidOptions()
				options.Theme = "forest"

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("theme"))
				Expect(output).To(ContainSubstring("forest"))
			})

			It("should include timing when IncludeTiming is true", func() {
				now := time.Now()
				duration := 5 * time.Second
				g.AddNode(&graph.Node{
					ID:       "wf",
					Type:     graph.NodeTypeWorkflow,
					Name:     "Workflow",
					Duration: &duration,
				})

				options := export.DefaultMermaidOptions()
				options.IncludeTiming = true
				_ = now

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("5s"))
			})
		})

		Context("with state diagram type", func() {
			It("should generate state diagram", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow", State: graph.NodeStateRunning})

				options := export.DefaultMermaidOptions()
				options.DiagramType = export.MermaidStateDiagram

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("stateDiagram-v2"))
			})
		})

		Context("with gantt type", func() {
			It("should generate gantt chart", func() {
				now := time.Now()
				later := now.Add(5 * time.Minute)
				g.AddNode(&graph.Node{
					ID:          "wf",
					Type:        graph.NodeTypeWorkflow,
					Name:        "Workflow",
					State:       graph.NodeStateSucceeded,
					StartedAt:   &now,
					CompletedAt: &later,
				})

				options := export.DefaultMermaidOptions()
				options.DiagramType = export.MermaidGantt

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("gantt"))
				Expect(output).To(ContainSubstring("Workflow"))
			})

			It("should include status based on node state", func() {
				now := time.Now()
				g.AddNode(&graph.Node{
					ID:        "running-node",
					Type:      graph.NodeTypeStep,
					Name:      "Running Step",
					State:     graph.NodeStateRunning,
					StartedAt: &now,
				})
				g.AddNode(&graph.Node{
					ID:        "done-node",
					Type:      graph.NodeTypeStep,
					Name:      "Done Step",
					State:     graph.NodeStateSucceeded,
					StartedAt: &now,
				})
				g.AddNode(&graph.Node{
					ID:        "failed-node",
					Type:      graph.NodeTypeStep,
					Name:      "Failed Step",
					State:     graph.NodeStateFailed,
					StartedAt: &now,
				})

				options := export.DefaultMermaidOptions()
				options.DiagramType = export.MermaidGantt

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("active"))
				Expect(output).To(ContainSubstring("done"))
				Expect(output).To(ContainSubstring("crit"))
			})

			It("should skip nodes without start time", func() {
				g.AddNode(&graph.Node{
					ID:    "no-timing",
					Type:  graph.NodeTypeStep,
					Name:  "No Timing",
					State: graph.NodeStateWaiting,
				})

				options := export.DefaultMermaidOptions()
				options.DiagramType = export.MermaidGantt

				output, err := export.ExportGraphMermaid(g, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(ContainSubstring("No Timing"))
			})
		})

		Context("with unsupported diagram type", func() {
			It("should return error", func() {
				g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

				options := &export.MermaidExportOptions{
					DiagramType: "unknown",
				}

				_, err := export.ExportGraphMermaid(g, options)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported diagram type"))
			})
		})
	})

	Describe("Node shapes", func() {
		It("should use rounded shape for workflow", func() {
			g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("wf("))
		})

		It("should use stadium shape for step", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("step(["))
		})

		It("should use circle shape for resource", func() {
			g.AddNode(&graph.Node{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("res(("))
		})

		It("should use rectangle shape for spec", func() {
			g.AddNode(&graph.Node{ID: "spec", Type: graph.NodeTypeSpec, Name: "Spec"})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("spec["))
		})
	})

	Describe("Arrow styles", func() {
		BeforeEach(func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeWorkflow, Name: "N1"})
			g.AddNode(&graph.Node{ID: "n2", Type: graph.NodeTypeResource, Name: "N2"})
		})

		It("should use solid arrow for depends-on", func() {
			g.AddNode(&graph.Node{ID: "s1", Type: graph.NodeTypeStep, Name: "S1"})
			g.AddNode(&graph.Node{ID: "s2", Type: graph.NodeTypeStep, Name: "S2"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "s1", ToNodeID: "s2", Type: graph.EdgeTypeDependsOn})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("-->"))
		})

		It("should use thick arrow for provisions", func() {
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "n1", ToNodeID: "n2", Type: graph.EdgeTypeProvisions})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("==>"))
		})

		It("should use dotted arrow for binds-to", func() {
			g.AddNode(&graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"})
			g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "step", ToNodeID: "n2", Type: graph.EdgeTypeBindsTo})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("-.-"))
		})
	})

	Describe("Node classes", func() {
		It("should apply running class", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Running", State: graph.NodeStateRunning})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("class n1 running"))
		})

		It("should apply succeeded class", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Succeeded", State: graph.NodeStateSucceeded})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("class n1 succeeded"))
		})

		It("should apply failed class", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Failed", State: graph.NodeStateFailed})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("class n1 failed"))
		})

		It("should apply pending class", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Pending", State: graph.NodeStatePending})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("class n1 pending"))
		})

		It("should not apply class for waiting state", func() {
			g.AddNode(&graph.Node{ID: "n1", Type: graph.NodeTypeStep, Name: "Waiting", State: graph.NodeStateWaiting})

			output, err := export.ExportGraphMermaid(g, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(ContainSubstring("class n1 "))
		})
	})
})

var _ = Describe("Large Graph Handling", func() {
	var exporter *export.Exporter

	BeforeEach(func() {
		exporter = export.NewExporter()
	})

	AfterEach(func() {
		exporter.Close()
	})

	It("should handle graphs with many nodes", func() {
		g := graph.NewGraph("large-graph")

		for i := 0; i < 50; i++ {
			g.AddNode(&graph.Node{
				ID:    nodeID(i),
				Type:  graph.NodeTypeStep,
				Name:  "Step",
				State: graph.NodeStateWaiting,
			})
		}

		output, err := exporter.ExportGraph(g, export.FormatDOT)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("digraph"))

		lines := strings.Split(string(output), "\n")
		Expect(len(lines)).To(BeNumerically(">", 50))
	})

	It("should handle graphs with many edges", func() {
		g := graph.NewGraph("edge-heavy")

		g.AddNode(&graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"})
		for i := 0; i < 30; i++ {
			stepID := nodeID(i)
			g.AddNode(&graph.Node{ID: stepID, Type: graph.NodeTypeStep, Name: "Step"})
			g.AddEdge(&graph.Edge{
				ID:         "e-" + stepID,
				FromNodeID: "wf",
				ToNodeID:   stepID,
				Type:       graph.EdgeTypeContains,
			})
		}

		output, err := exporter.ExportGraph(g, export.FormatDOT)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Count(string(output), "->")).To(Equal(30))
	})
})

func nodeID(i int) string {
	return "node-" + string(rune('a'+i%26)) + string(rune('0'+i/26))
}
