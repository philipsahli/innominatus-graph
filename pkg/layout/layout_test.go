package layout

import (
	"testing"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

func createTestGraph() *graph.Graph {
	g := graph.NewGraph("test")

	// Create simple DAG: A -> B -> C
	//                      A -> D
	g.AddNode(&graph.Node{ID: "A", Type: graph.NodeTypeSpec, Name: "A"})
	g.AddNode(&graph.Node{ID: "B", Type: graph.NodeTypeWorkflow, Name: "B"})
	g.AddNode(&graph.Node{ID: "C", Type: graph.NodeTypeStep, Name: "C"})
	g.AddNode(&graph.Node{ID: "D", Type: graph.NodeTypeResource, Name: "D"})

	g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "A", ToNodeID: "B", Type: graph.EdgeTypeDependsOn})
	g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "B", ToNodeID: "C", Type: graph.EdgeTypeDependsOn})
	g.AddEdge(&graph.Edge{ID: "e3", FromNodeID: "A", ToNodeID: "D", Type: graph.EdgeTypeDependsOn})

	return g
}

func TestComputeHierarchicalLayout(t *testing.T) {
	g := createTestGraph()

	options := &LayoutOptions{
		Type:         LayoutHierarchical,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute layout: %v", err)
	}

	// Verify all nodes have positions
	if len(layout.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in layout, got %d", len(layout.Nodes))
	}

	// Debug: print all levels
	t.Logf("Levels: A=%d, B=%d, C=%d, D=%d",
		layout.Nodes["A"].Level,
		layout.Nodes["B"].Level,
		layout.Nodes["C"].Level,
		layout.Nodes["D"].Level)

	// Verify node A is at level 0
	if layout.Nodes["A"].Level != 0 {
		t.Errorf("Expected node A at level 0, got %d", layout.Nodes["A"].Level)
	}

	// Verify node B is at level 1 (depends on A)
	if layout.Nodes["B"].Level != 1 {
		t.Errorf("Expected node B at level 1, got %d", layout.Nodes["B"].Level)
	}

	// Verify node C is at level 2 (depends on B)
	if layout.Nodes["C"].Level != 2 {
		t.Errorf("Expected node C at level 2, got %d", layout.Nodes["C"].Level)
	}

	// Verify positions are set
	posA := layout.Nodes["A"].Position
	if posA.X == 0 && posA.Y == 0 {
		t.Error("Expected non-zero position for node A")
	}
}

func TestComputeRadialLayout(t *testing.T) {
	g := createTestGraph()

	options := &LayoutOptions{
		Type:         LayoutRadial,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute radial layout: %v", err)
	}

	// Verify all nodes have positions
	if len(layout.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in layout, got %d", len(layout.Nodes))
	}

	// Verify root node is at center (level 0)
	rootLevel := layout.Nodes["A"].Level
	if rootLevel != 0 {
		t.Errorf("Expected root at level 0, got %d", rootLevel)
	}
}

func TestComputeGridLayout(t *testing.T) {
	g := createTestGraph()

	options := &LayoutOptions{
		Type:         LayoutGrid,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute grid layout: %v", err)
	}

	// Verify all nodes have positions
	if len(layout.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in layout, got %d", len(layout.Nodes))
	}

	// Verify positions are distributed
	positions := make(map[Position]bool)
	for _, nodeLayout := range layout.Nodes {
		positions[nodeLayout.Position] = true
	}

	// All positions should be unique
	if len(positions) != 4 {
		t.Error("Expected unique positions for all nodes in grid layout")
	}
}

func TestComputeForceLayout(t *testing.T) {
	g := createTestGraph()

	options := &LayoutOptions{
		Type:         LayoutForce,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute force layout: %v", err)
	}

	// Verify all nodes have positions
	if len(layout.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in layout, got %d", len(layout.Nodes))
	}

	// Verify positions changed from initialization
	hasMovement := false
	for _, nodeLayout := range layout.Nodes {
		if nodeLayout.Position.X != 0 || nodeLayout.Position.Y != 0 {
			hasMovement = true
			break
		}
	}

	if !hasMovement {
		t.Error("Expected nodes to have moved during force simulation")
	}
}

func TestLayoutEmptyGraph(t *testing.T) {
	g := graph.NewGraph("empty")

	layout, err := ComputeLayout(g, DefaultLayoutOptions())
	if err != nil {
		t.Fatalf("Failed to compute layout for empty graph: %v", err)
	}

	if len(layout.Nodes) != 0 {
		t.Errorf("Expected 0 nodes in layout, got %d", len(layout.Nodes))
	}
}

func TestGetNodePosition(t *testing.T) {
	g := createTestGraph()
	layout, err := ComputeLayout(g, DefaultLayoutOptions())
	if err != nil {
		t.Fatalf("Failed to compute layout: %v", err)
	}

	// Test existing node
	pos, exists := layout.GetNodePosition("A")
	if !exists {
		t.Error("Expected node A to exist")
	}
	if pos.X == 0 && pos.Y == 0 {
		t.Error("Expected non-zero position")
	}

	// Test non-existing node
	_, exists = layout.GetNodePosition("Z")
	if exists {
		t.Error("Expected node Z to not exist")
	}
}

func TestGetNodesByLevel(t *testing.T) {
	g := createTestGraph()

	options := &LayoutOptions{
		Type:         LayoutHierarchical,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute layout: %v", err)
	}

	// Get nodes at level 0 (should be node A)
	level0Nodes := layout.GetNodesByLevel(0)
	if len(level0Nodes) != 1 {
		t.Errorf("Expected 1 node at level 0, got %d", len(level0Nodes))
	}
	if level0Nodes[0] != "A" {
		t.Errorf("Expected node A at level 0, got %s", level0Nodes[0])
	}

	// Get nodes at level 1
	level1Nodes := layout.GetNodesByLevel(1)
	if len(level1Nodes) != 2 {
		t.Errorf("Expected 2 nodes at level 1, got %d", len(level1Nodes))
	}
}

func TestDefaultLayoutOptions(t *testing.T) {
	options := DefaultLayoutOptions()

	if options.Type != LayoutHierarchical {
		t.Errorf("Expected default type=hierarchical, got %s", options.Type)
	}
	if options.NodeSpacing <= 0 {
		t.Error("Expected positive node spacing")
	}
	if options.Width <= 0 {
		t.Error("Expected positive width")
	}
}

func TestFindRootNodes(t *testing.T) {
	g := createTestGraph()

	roots := findRootNodes(g)

	// Node A should be the only root
	if len(roots) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(roots))
	}
	if len(roots) > 0 && roots[0] != "A" {
		t.Errorf("Expected root to be A, got %s", roots[0])
	}
}
