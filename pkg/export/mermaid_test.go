package export

import (
	"strings"
	"testing"
	"time"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

func TestExportGraphMermaid_Flowchart(t *testing.T) {
	// Create test graph
	g := graph.NewGraph("test-workflow")

	// Add nodes
	specNode := &graph.Node{
		ID:    "spec-1",
		Type:  graph.NodeTypeSpec,
		Name:  "My App",
		State: graph.NodeStatePending,
	}
	g.AddNode(specNode)

	workflowNode := &graph.Node{
		ID:    "workflow-1",
		Type:  graph.NodeTypeWorkflow,
		Name:  "Deploy Workflow",
		State: graph.NodeStateRunning,
	}
	g.AddNode(workflowNode)

	stepNode := &graph.Node{
		ID:    "step-1",
		Type:  graph.NodeTypeStep,
		Name:  "Create Namespace",
		State: graph.NodeStateSucceeded,
	}
	g.AddNode(stepNode)

	// Add edges
	g.AddEdge(&graph.Edge{
		ID:         "edge-1",
		FromNodeID: "workflow-1",
		ToNodeID:   "spec-1",
		Type:       graph.EdgeTypeCreates,
	})

	g.AddEdge(&graph.Edge{
		ID:         "edge-2",
		FromNodeID: "workflow-1",
		ToNodeID:   "step-1",
		Type:       graph.EdgeTypeContains,
	})

	// Test flowchart export
	output, err := ExportGraphMermaid(g, DefaultMermaidOptions())
	if err != nil {
		t.Fatalf("Failed to export Mermaid: %v", err)
	}

	// Verify output structure
	if !strings.Contains(output, "flowchart TB") {
		t.Error("Expected flowchart TB header")
	}
	if !strings.Contains(output, "spec_1") {
		t.Error("Expected spec node ID")
	}
	if !strings.Contains(output, "My App") {
		t.Error("Expected spec node name")
	}
	if !strings.Contains(output, "Deploy Workflow") {
		t.Error("Expected workflow node name")
	}
	if !strings.Contains(output, "-->") {
		t.Error("Expected arrow connection")
	}
	if !strings.Contains(output, "classDef running") {
		t.Error("Expected running class definition")
	}
}

func TestExportGraphMermaid_StateDiagram(t *testing.T) {
	g := graph.NewGraph("state-test")

	// Add nodes
	node1 := &graph.Node{
		ID:    "node-1",
		Type:  graph.NodeTypeWorkflow,
		Name:  "Workflow",
		State: graph.NodeStateRunning,
	}
	g.AddNode(node1)

	node2 := &graph.Node{
		ID:    "node-2",
		Type:  graph.NodeTypeStep,
		Name:  "Step",
		State: graph.NodeStateSucceeded,
	}
	g.AddNode(node2)

	// Add edge
	g.AddEdge(&graph.Edge{
		ID:         "edge-1",
		FromNodeID: "node-1",
		ToNodeID:   "node-2",
		Type:       graph.EdgeTypeContains,
	})

	// Export as state diagram
	options := &MermaidExportOptions{
		DiagramType: MermaidStateDiagram,
	}
	output, err := ExportGraphMermaid(g, options)
	if err != nil {
		t.Fatalf("Failed to export state diagram: %v", err)
	}

	// Verify output
	if !strings.Contains(output, "stateDiagram-v2") {
		t.Error("Expected stateDiagram-v2 header")
	}
	if !strings.Contains(output, "Workflow: running") {
		t.Error("Expected state in label")
	}
}

func TestExportGraphMermaid_Gantt(t *testing.T) {
	g := graph.NewGraph("gantt-test")

	// Add nodes with timing
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now()

	node1 := &graph.Node{
		ID:          "node-1",
		Type:        graph.NodeTypeWorkflow,
		Name:        "Workflow Task",
		State:       graph.NodeStateSucceeded,
		StartedAt:   &start,
		CompletedAt: &end,
	}
	g.AddNode(node1)

	// Export as Gantt chart
	options := &MermaidExportOptions{
		DiagramType:   MermaidGantt,
		IncludeTiming: true,
	}
	output, err := ExportGraphMermaid(g, options)
	if err != nil {
		t.Fatalf("Failed to export Gantt chart: %v", err)
	}

	// Verify output
	if !strings.Contains(output, "gantt") {
		t.Error("Expected gantt header")
	}
	if !strings.Contains(output, "Workflow Task") {
		t.Error("Expected task name")
	}
	if !strings.Contains(output, "done") {
		t.Error("Expected done status for succeeded task")
	}
}

func TestExportGraphMermaid_WithTiming(t *testing.T) {
	g := graph.NewGraph("timing-test")

	// Create node with duration
	start := time.Now().Add(-2 * time.Minute)
	end := time.Now()
	duration := end.Sub(start)

	node := &graph.Node{
		ID:          "node-1",
		Type:        graph.NodeTypeWorkflow,
		Name:        "Task",
		State:       graph.NodeStateSucceeded,
		StartedAt:   &start,
		CompletedAt: &end,
		Duration:    &duration,
	}
	g.AddNode(node)

	// Export with timing
	options := &MermaidExportOptions{
		DiagramType:   MermaidFlowchart,
		IncludeTiming: true,
		IncludeState:  true,
	}
	output, err := ExportGraphMermaid(g, options)
	if err != nil {
		t.Fatalf("Failed to export with timing: %v", err)
	}

	// Verify duration in output
	if !strings.Contains(output, "m") { // Duration should contain "m" for minutes
		t.Error("Expected duration in label")
	}
	if !strings.Contains(output, "succeeded") {
		t.Error("Expected state in label")
	}
}

func TestExportGraphMermaid_EmptyGraph(t *testing.T) {
	g := graph.NewGraph("empty")

	output, err := ExportGraphMermaid(g, DefaultMermaidOptions())
	if err != nil {
		t.Fatalf("Failed to export empty graph: %v", err)
	}

	// Should still have header
	if !strings.Contains(output, "flowchart TB") {
		t.Error("Expected flowchart header even for empty graph")
	}
}

func TestMermaidSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"spec-1", "spec_1"},
		{"workflow.2", "workflow_2"},
		{"step 3", "step_3"},
		{"resource-foo.bar", "resource_foo_bar"},
	}

	for _, tt := range tests {
		result := sanitizeID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeID(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestMermaidEscapeLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`Hello "World"`, `Hello &quot;World&quot;`},
		{"Text #tag", "Text &num;tag"},
		{`Quote "and" #hash`, `Quote &quot;and&quot; &num;hash`},
	}

	for _, tt := range tests {
		result := escapeLabel(tt.input)
		if result != tt.expected {
			t.Errorf("escapeLabel(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
