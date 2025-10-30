package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

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
		Feature:   "State Propagation",
		Checks:    0,
		Passed:    0,
		Failed:    0,
		Artifacts: []string{},
	}

	fmt.Println("=== Verification: State Propagation ===\n")

	// Setup: Create graph with workflow and steps
	g := graph.NewGraph("state-propagation-test")

	workflow := &graph.Node{
		ID:   "workflow-1",
		Type: graph.NodeTypeWorkflow,
		Name: "Test Workflow",
	}
	g.AddNode(workflow)

	step1 := &graph.Node{
		ID:   "step-1",
		Type: graph.NodeTypeStep,
		Name: "Step 1",
	}
	g.AddNode(step1)

	step2 := &graph.Node{
		ID:   "step-2",
		Type: graph.NodeTypeStep,
		Name: "Step 2",
	}
	g.AddNode(step2)

	// Connect workflow → steps
	g.AddEdge(&graph.Edge{
		ID:         "wf-s1",
		FromNodeID: "workflow-1",
		ToNodeID:   "step-1",
		Type:       graph.EdgeTypeContains,
	})

	g.AddEdge(&graph.Edge{
		ID:         "wf-s2",
		FromNodeID: "workflow-1",
		ToNodeID:   "step-2",
		Type:       graph.EdgeTypeContains,
	})

	// Test 1: Step failure propagates to workflow
	result.Checks++
	fmt.Println("▶️  Test 1: Step failure propagates to workflow")

	// Set step to running, then failed
	g.UpdateNodeState("step-1", graph.NodeStateRunning)
	g.UpdateNodeState("step-1", graph.NodeStateFailed)

	// Check that workflow is now failed
	workflowNode, _ := g.GetNode("workflow-1")
	if workflowNode.State != graph.NodeStateFailed {
		fmt.Printf("❌ Test 1 FAILED: Workflow state is %s, expected failed\n", workflowNode.State)
		result.Failed++
	} else {
		fmt.Println("✅ Test 1 PASSED: Workflow state correctly propagated to failed")
		result.Passed++
	}

	// Test 2: Verify step state is still failed
	result.Checks++
	fmt.Println("\n▶️  Test 2: Step state remains failed")

	step1Node, _ := g.GetNode("step-1")
	if step1Node.State != graph.NodeStateFailed {
		fmt.Printf("❌ Test 2 FAILED: Step state is %s, expected failed\n", step1Node.State)
		result.Failed++
	} else {
		fmt.Println("✅ Test 2 PASSED: Step state is failed")
		result.Passed++
	}

	// Test 3: Other step not affected by first step's failure
	result.Checks++
	fmt.Println("\n▶️  Test 3: Other step not affected by sibling failure")

	step2Node, _ := g.GetNode("step-2")
	if step2Node.State == graph.NodeStateFailed {
		fmt.Printf("❌ Test 3 FAILED: Step 2 incorrectly marked as failed\n")
		result.Failed++
	} else {
		fmt.Println("✅ Test 3 PASSED: Step 2 state is independent")
		result.Passed++
	}

	// Save artifacts
	fmt.Println("\n📁 Saving artifacts...")
	if err := os.MkdirAll("docs/verification", 0755); err != nil {
		fmt.Printf("⚠️  Failed to create directory: %v\n", err)
	}

	artifactPath := "docs/verification/state-propagation.json"
	if err := saveArtifact(result, artifactPath); err != nil {
		fmt.Printf("⚠️  Failed to save artifact: %v\n", err)
	} else {
		result.Artifacts = append(result.Artifacts, artifactPath)
		fmt.Printf("✅ Artifact saved: %s\n", artifactPath)
	}

	// Print final results
	fmt.Println("\n" + "==================================================")
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

func printJSON(data interface{}) {
	fmt.Println("\nVERIFICATION_RESULT:")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}
