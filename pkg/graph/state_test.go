package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph_UpdateNodeState(t *testing.T) {
	g := NewGraph("test-app")

	node := &Node{
		ID:   "node1",
		Type: NodeTypeWorkflow,
		Name: "Test Workflow",
	}
	g.AddNode(node)

	// Test state update
	err := g.UpdateNodeState("node1", NodeStateRunning)
	assert.NoError(t, err)

	updatedNode, _ := g.GetNode("node1")
	assert.Equal(t, NodeStateRunning, updatedNode.State)

	// Test non-existent node
	err = g.UpdateNodeState("non-existent", NodeStateRunning)
	assert.Error(t, err)
}

func TestGraph_StateProps_StepFailurePropagation(t *testing.T) {
	g := NewGraph("test-app")

	// Create workflow
	workflow := &Node{
		ID:   "workflow1",
		Type: NodeTypeWorkflow,
		Name: "Deploy Workflow",
	}
	g.AddNode(workflow)

	// Create steps
	step1 := &Node{
		ID:   "step1",
		Type: NodeTypeStep,
		Name: "Provision Step",
	}
	g.AddNode(step1)

	step2 := &Node{
		ID:   "step2",
		Type: NodeTypeStep,
		Name: "Deploy Step",
	}
	g.AddNode(step2)

	// Link workflow → steps
	g.AddEdge(&Edge{
		ID:         "wf-step1",
		FromNodeID: "workflow1",
		ToNodeID:   "step1",
		Type:       EdgeTypeContains,
	})

	g.AddEdge(&Edge{
		ID:         "wf-step2",
		FromNodeID: "workflow1",
		ToNodeID:   "step2",
		Type:       EdgeTypeContains,
	})

	// When step fails, workflow should fail
	err := g.UpdateNodeState("step1", NodeStateFailed)
	assert.NoError(t, err)

	workflowNode, _ := g.GetNode("workflow1")
	assert.Equal(t, NodeStateFailed, workflowNode.State, "Workflow should transition to failed when step fails")
}

func TestGraph_GetNodesByType(t *testing.T) {
	g := NewGraph("test-app")

	g.AddNode(&Node{ID: "spec1", Type: NodeTypeSpec, Name: "Spec 1"})
	g.AddNode(&Node{ID: "workflow1", Type: NodeTypeWorkflow, Name: "Workflow 1"})
	g.AddNode(&Node{ID: "step1", Type: NodeTypeStep, Name: "Step 1"})
	g.AddNode(&Node{ID: "step2", Type: NodeTypeStep, Name: "Step 2"})
	g.AddNode(&Node{ID: "resource1", Type: NodeTypeResource, Name: "Resource 1"})

	workflows := g.GetNodesByType(NodeTypeWorkflow)
	assert.Len(t, workflows, 1)

	steps := g.GetNodesByType(NodeTypeStep)
	assert.Len(t, steps, 2)

	specs := g.GetNodesByType(NodeTypeSpec)
	assert.Len(t, specs, 1)

	resources := g.GetNodesByType(NodeTypeResource)
	assert.Len(t, resources, 1)
}

func TestGraph_GetNodesByState(t *testing.T) {
	g := NewGraph("test-app")

	g.AddNode(&Node{ID: "node1", Type: NodeTypeWorkflow, Name: "Node 1", State: NodeStateWaiting})
	g.AddNode(&Node{ID: "node2", Type: NodeTypeStep, Name: "Node 2", State: NodeStateRunning})
	g.AddNode(&Node{ID: "node3", Type: NodeTypeStep, Name: "Node 3", State: NodeStateRunning})
	g.AddNode(&Node{ID: "node4", Type: NodeTypeResource, Name: "Node 4", State: NodeStateFailed})

	runningNodes := g.GetNodesByState(NodeStateRunning)
	assert.Len(t, runningNodes, 2)

	failedNodes := g.GetNodesByState(NodeStateFailed)
	assert.Len(t, failedNodes, 1)

	waitingNodes := g.GetNodesByState(NodeStateWaiting)
	assert.Len(t, waitingNodes, 1)
}

func TestGraph_GetChildSteps(t *testing.T) {
	g := NewGraph("test-app")

	workflow := &Node{ID: "workflow1", Type: NodeTypeWorkflow, Name: "Workflow"}
	step1 := &Node{ID: "step1", Type: NodeTypeStep, Name: "Step 1"}
	step2 := &Node{ID: "step2", Type: NodeTypeStep, Name: "Step 2"}

	g.AddNode(workflow)
	g.AddNode(step1)
	g.AddNode(step2)

	g.AddEdge(&Edge{
		ID:         "wf-step1",
		FromNodeID: "workflow1",
		ToNodeID:   "step1",
		Type:       EdgeTypeContains,
	})

	g.AddEdge(&Edge{
		ID:         "wf-step2",
		FromNodeID: "workflow1",
		ToNodeID:   "step2",
		Type:       EdgeTypeContains,
	})

	childSteps := g.GetChildSteps("workflow1")
	assert.Len(t, childSteps, 2)
	assert.Contains(t, []string{childSteps[0].ID, childSteps[1].ID}, "step1")
	assert.Contains(t, []string{childSteps[0].ID, childSteps[1].ID}, "step2")
}

func TestGraph_GetParentWorkflow(t *testing.T) {
	g := NewGraph("test-app")

	workflow := &Node{ID: "workflow1", Type: NodeTypeWorkflow, Name: "Workflow"}
	step := &Node{ID: "step1", Type: NodeTypeStep, Name: "Step 1"}

	g.AddNode(workflow)
	g.AddNode(step)

	g.AddEdge(&Edge{
		ID:         "wf-step1",
		FromNodeID: "workflow1",
		ToNodeID:   "step1",
		Type:       EdgeTypeContains,
	})

	parent, err := g.GetParentWorkflow("step1")
	assert.NoError(t, err)
	assert.Equal(t, "workflow1", parent.ID)

	// Test step with no parent
	orphanStep := &Node{ID: "orphan-step", Type: NodeTypeStep, Name: "Orphan Step"}
	g.AddNode(orphanStep)

	_, err = g.GetParentWorkflow("orphan-step")
	assert.Error(t, err)
}

func TestGraph_AddEdge_NewEdgeTypes(t *testing.T) {
	tests := []struct {
		name        string
		setupGraph  func() *Graph
		edge        *Edge
		expectError bool
	}{
		{
			name: "valid contains edge (workflow → step)",
			setupGraph: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "wf1", Type: NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&Node{ID: "step1", Type: NodeTypeStep, Name: "Step"})
				return g
			},
			edge: &Edge{
				ID:         "wf-step",
				FromNodeID: "wf1",
				ToNodeID:   "step1",
				Type:       EdgeTypeContains,
			},
			expectError: false,
		},
		{
			name: "invalid contains edge (spec → step)",
			setupGraph: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "spec1", Type: NodeTypeSpec, Name: "Spec"})
				g.AddNode(&Node{ID: "step1", Type: NodeTypeStep, Name: "Step"})
				return g
			},
			edge: &Edge{
				ID:         "spec-step",
				FromNodeID: "spec1",
				ToNodeID:   "step1",
				Type:       EdgeTypeContains,
			},
			expectError: true,
		},
		{
			name: "invalid contains edge (workflow → resource)",
			setupGraph: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "wf1", Type: NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&Node{ID: "res1", Type: NodeTypeResource, Name: "Resource"})
				return g
			},
			edge: &Edge{
				ID:         "wf-res",
				FromNodeID: "wf1",
				ToNodeID:   "res1",
				Type:       EdgeTypeContains,
			},
			expectError: true,
		},
		{
			name: "valid configures edge (step → resource)",
			setupGraph: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "step1", Type: NodeTypeStep, Name: "Step"})
				g.AddNode(&Node{ID: "res1", Type: NodeTypeResource, Name: "Resource"})
				return g
			},
			edge: &Edge{
				ID:         "step-res",
				FromNodeID: "step1",
				ToNodeID:   "res1",
				Type:       EdgeTypeConfigures,
			},
			expectError: false,
		},
		{
			name: "invalid configures edge (workflow → resource)",
			setupGraph: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "wf1", Type: NodeTypeWorkflow, Name: "Workflow"})
				g.AddNode(&Node{ID: "res1", Type: NodeTypeResource, Name: "Resource"})
				return g
			},
			edge: &Edge{
				ID:         "wf-res",
				FromNodeID: "wf1",
				ToNodeID:   "res1",
				Type:       EdgeTypeConfigures,
			},
			expectError: true,
		},
		{
			name: "invalid configures edge (step → spec)",
			setupGraph: func() *Graph {
				g := NewGraph("test")
				g.AddNode(&Node{ID: "step1", Type: NodeTypeStep, Name: "Step"})
				g.AddNode(&Node{ID: "spec1", Type: NodeTypeSpec, Name: "Spec"})
				return g
			},
			edge: &Edge{
				ID:         "step-spec",
				FromNodeID: "step1",
				ToNodeID:   "spec1",
				Type:       EdgeTypeConfigures,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setupGraph()
			err := g.AddEdge(tt.edge)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGraph_ComplexStatePropagation(t *testing.T) {
	// Test complex scenario: workflow with multiple steps
	g := NewGraph("test-app")

	// Create workflow
	workflow := &Node{
		ID:   "deploy-workflow",
		Type: NodeTypeWorkflow,
		Name: "Deploy Workflow",
	}
	g.AddNode(workflow)

	// Create 3 steps
	for i := 1; i <= 3; i++ {
		step := &Node{
			ID:   "step" + string(rune('0'+i)),
			Type: NodeTypeStep,
			Name: "Step " + string(rune('0'+i)),
		}
		g.AddNode(step)

		g.AddEdge(&Edge{
			ID:         "wf-step" + string(rune('0'+i)),
			FromNodeID: "deploy-workflow",
			ToNodeID:   step.ID,
			Type:       EdgeTypeContains,
		})
	}

	// All steps start as waiting
	steps := g.GetChildSteps("deploy-workflow")
	for _, step := range steps {
		assert.Equal(t, NodeStateWaiting, step.State)
	}

	// Simulate step 2 failing
	g.UpdateNodeState("step2", NodeStateRunning)
	g.UpdateNodeState("step2", NodeStateFailed)

	// Workflow should be failed
	workflowNode, _ := g.GetNode("deploy-workflow")
	assert.Equal(t, NodeStateFailed, workflowNode.State)

	// Other steps should still be waiting
	step1, _ := g.GetNode("step1")
	step3, _ := g.GetNode("step3")
	assert.Equal(t, NodeStateWaiting, step1.State)
	assert.Equal(t, NodeStateWaiting, step3.State)
}
