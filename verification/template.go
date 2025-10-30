package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

// VerificationResult holds structured verification output
type VerificationResult struct {
	Feature   string   `json:"feature"`
	Status    string   `json:"status"`
	Checks    int      `json:"checks"`
	Passed    int      `json:"passed"`
	Failed    int      `json:"failed"`
	Artifacts []string `json:"artifacts"`
}

func main() {
	result := VerificationResult{
		Feature:   "Feature Name",
		Checks:    0,
		Passed:    0,
		Failed:    0,
		Artifacts: []string{},
	}

	fmt.Println("=== Verification: Feature Name ===\n")

	// Setup
	g := graph.NewGraph("verify-test")

	// Test 1: [Description of what this test verifies]
	result.Checks++
	fmt.Println("▶️  Test 1: [Description]")
	if err := test1(g); err != nil {
		fmt.Printf("❌ Test 1 FAILED: %v\n", err)
		result.Failed++
	} else {
		fmt.Println("✅ Test 1 PASSED")
		result.Passed++
	}

	// Test 2: [Description of what this test verifies]
	result.Checks++
	fmt.Println("\n▶️  Test 2: [Description]")
	if err := test2(g); err != nil {
		fmt.Printf("❌ Test 2 FAILED: %v\n", err)
		result.Failed++
	} else {
		fmt.Println("✅ Test 2 PASSED")
		result.Passed++
	}

	// Save artifacts
	fmt.Println("\n📁 Saving artifacts...")
	artifactPath := "docs/verification/feature-name.json"
	if err := os.MkdirAll("docs/verification", 0755); err != nil {
		fmt.Printf("⚠️  Failed to create directory: %v\n", err)
	}

	if err := saveArtifact(result, artifactPath); err != nil {
		fmt.Printf("⚠️  Failed to save artifact: %v\n", err)
	} else {
		result.Artifacts = append(result.Artifacts, artifactPath)
		fmt.Printf("✅ Artifact saved: %s\n", artifactPath)
	}

	// Print final results
	fmt.Println("\n" + "=".repeat(50))
	if result.Failed > 0 {
		result.Status = "failed"
		fmt.Printf("❌ VERIFICATION FAILED: %d/%d checks passed\n", result.Passed, result.Checks)
		printJSON(result)
		os.Exit(1)
	}

	result.Status = "passed"
	fmt.Printf("✅ VERIFICATION PASSED: %d/%d checks passed\n", result.Passed, result.Checks)
	printJSON(result)
}

// test1 implements the first verification test
func test1(g *graph.Graph) error {
	// TODO: Implement test logic
	// Example:
	node := &graph.Node{
		ID:   "test-node-1",
		Type: graph.NodeTypeWorkflow,
		Name: "Test Workflow",
	}

	if err := g.AddNode(node); err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	retrievedNode, exists := g.GetNode("test-node-1")
	if !exists {
		return fmt.Errorf("node not found after adding")
	}

	if retrievedNode.Name != "Test Workflow" {
		return fmt.Errorf("node name mismatch: got %s, want Test Workflow", retrievedNode.Name)
	}

	return nil
}

// test2 implements the second verification test
func test2(g *graph.Graph) error {
	// TODO: Implement test logic
	// Example:
	nodes := g.GetNodesByType(graph.NodeTypeWorkflow)
	if len(nodes) != 1 {
		return fmt.Errorf("expected 1 workflow node, got %d", len(nodes))
	}

	return nil
}

// saveArtifact saves verification result to JSON file
func saveArtifact(data interface{}, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// printJSON prints structured output for AI evaluation
func printJSON(data interface{}) {
	fmt.Println("\nVERIFICATION_RESULT:")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}
