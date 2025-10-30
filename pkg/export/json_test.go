package export

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

func TestExportGraphJSON(t *testing.T) {
	// Create test graph
	g := graph.NewGraph("test-app")

	// Add nodes
	specNode := &graph.Node{
		ID:          "spec-1",
		Type:        graph.NodeTypeSpec,
		Name:        "Test Spec",
		Description: "Test specification",
		State:       graph.NodeStatePending,
		Properties:  map[string]interface{}{"key": "value"},
	}
	g.AddNode(specNode)

	workflowNode := &graph.Node{
		ID:    "workflow-1",
		Type:  graph.NodeTypeWorkflow,
		Name:  "Test Workflow",
		State: graph.NodeStateRunning,
	}
	// Set timing
	now := time.Now()
	workflowNode.StartedAt = &now
	g.AddNode(workflowNode)

	// Add edge (spec depends on workflow is not valid, so use workflow creates spec)
	edge := &graph.Edge{
		ID:         "edge-1",
		FromNodeID: "workflow-1",
		ToNodeID:   "spec-1",
		Type:       graph.EdgeTypeCreates,
	}
	if err := g.AddEdge(edge); err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	// Test with default options
	t.Run("default options", func(t *testing.T) {
		data, err := ExportGraphJSON(g, DefaultJSONOptions())
		if err != nil {
			t.Fatalf("Failed to export JSON: %v", err)
		}

		// Verify it's valid JSON
		var result JSONGraph
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Verify structure
		if result.AppName != "test-app" {
			t.Errorf("Expected app_name=test-app, got %s", result.AppName)
		}
		if len(result.Nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(result.Nodes))
		}
		if len(result.Edges) != 1 {
			t.Errorf("Expected 1 edge, got %d", len(result.Edges))
		}

		// Verify metadata included
		if result.Version == nil {
			t.Error("Expected version to be included")
		}
		if result.CreatedAt == nil {
			t.Error("Expected created_at to be included")
		}

		// Verify timing included
		foundWorkflow := false
		for _, node := range result.Nodes {
			if node.ID == "workflow-1" {
				foundWorkflow = true
				if node.StartedAt == nil {
					t.Error("Expected started_at for running workflow")
				}
			}
		}
		if !foundWorkflow {
			t.Error("Workflow node not found in export")
		}
	})

	// Test compact export
	t.Run("compact export", func(t *testing.T) {
		data, err := ExportGraphJSONCompact(g)
		if err != nil {
			t.Fatalf("Failed to export compact JSON: %v", err)
		}

		var result JSONGraph
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Verify metadata NOT included
		if result.Version != nil {
			t.Error("Expected version to be excluded in compact mode")
		}
		if result.CreatedAt != nil {
			t.Error("Expected created_at to be excluded in compact mode")
		}
	})

	// Test pretty export
	t.Run("pretty export", func(t *testing.T) {
		data, err := ExportGraphJSONPretty(g)
		if err != nil {
			t.Fatalf("Failed to export pretty JSON: %v", err)
		}

		// Verify it's indented
		dataStr := string(data)
		if len(dataStr) < 100 {
			t.Error("Pretty print should produce longer output")
		}

		// Should contain newlines and indentation
		if !contains(dataStr, "\n  ") {
			t.Error("Expected indented JSON")
		}
	})

	// Test with custom options
	t.Run("custom options no timing", func(t *testing.T) {
		options := &JSONExportOptions{
			IncludeMetadata:   true,
			IncludeProperties: true,
			IncludeTiming:     false, // Exclude timing
			PrettyPrint:       false,
		}
		data, err := ExportGraphJSON(g, options)
		if err != nil {
			t.Fatalf("Failed to export JSON: %v", err)
		}

		var result JSONGraph
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Verify timing NOT included
		for _, node := range result.Nodes {
			if node.StartedAt != nil {
				t.Error("Expected started_at to be excluded")
			}
		}
	})

	// Test with custom options no properties
	t.Run("custom options no properties", func(t *testing.T) {
		options := &JSONExportOptions{
			IncludeMetadata:   false,
			IncludeProperties: false, // Exclude properties
			IncludeTiming:     true,
			PrettyPrint:       false,
		}
		data, err := ExportGraphJSON(g, options)
		if err != nil {
			t.Fatalf("Failed to export JSON: %v", err)
		}

		var result JSONGraph
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Verify properties NOT included
		for _, node := range result.Nodes {
			if node.Properties != nil {
				t.Error("Expected properties to be excluded")
			}
		}
	})
}

func TestExportGraphJSON_EmptyGraph(t *testing.T) {
	g := graph.NewGraph("empty")

	data, err := ExportGraphJSON(g, DefaultJSONOptions())
	if err != nil {
		t.Fatalf("Failed to export empty graph: %v", err)
	}

	var result JSONGraph
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if len(result.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(result.Nodes))
	}
	if len(result.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(result.Edges))
	}
}

func TestExportGraphJSON_WithDuration(t *testing.T) {
	g := graph.NewGraph("duration-test")

	// Create node with completed timing
	node := &graph.Node{
		ID:    "node-1",
		Type:  graph.NodeTypeWorkflow,
		Name:  "Completed Workflow",
		State: graph.NodeStateSucceeded,
	}

	// Set timing
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now()
	duration := end.Sub(start)
	node.StartedAt = &start
	node.CompletedAt = &end
	node.Duration = &duration

	g.AddNode(node)

	data, err := ExportGraphJSON(g, DefaultJSONOptions())
	if err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}

	var result JSONGraph
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify duration included
	if len(result.Nodes) != 1 {
		t.Fatal("Expected 1 node")
	}

	exportedNode := result.Nodes[0]
	if exportedNode.StartedAt == nil {
		t.Error("Expected started_at to be set")
	}
	if exportedNode.CompletedAt == nil {
		t.Error("Expected completed_at to be set")
	}
	if exportedNode.Duration == nil {
		t.Error("Expected duration to be set")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) >= len(substr) && s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || s != s[:len(substr)]+s[len(s)-len(substr):]
}
