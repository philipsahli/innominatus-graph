package layout

import (
	"fmt"
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

func TestLayoutSingleNode(t *testing.T) {
	g := graph.NewGraph("single")
	g.AddNode(&graph.Node{ID: "only", Type: graph.NodeTypeWorkflow, Name: "Only Node"})

	layouts := []LayoutType{LayoutHierarchical, LayoutRadial, LayoutGrid, LayoutForce}
	for _, layoutType := range layouts {
		t.Run(string(layoutType), func(t *testing.T) {
			options := &LayoutOptions{
				Type:         layoutType,
				NodeSpacing:  100.0,
				LevelSpacing: 150.0,
				Width:        1200.0,
				Height:       800.0,
			}

			layout, err := ComputeLayout(g, options)
			if err != nil {
				t.Fatalf("Failed to compute %s layout: %v", layoutType, err)
			}

			if len(layout.Nodes) != 1 {
				t.Errorf("Expected 1 node in layout, got %d", len(layout.Nodes))
			}

			_, exists := layout.GetNodePosition("only")
			if !exists {
				t.Error("Expected node 'only' to have a position")
			}
		})
	}
}

func TestLayoutDisconnectedGraph(t *testing.T) {
	g := graph.NewGraph("disconnected")

	// Create two disconnected subgraphs
	g.AddNode(&graph.Node{ID: "A1", Type: graph.NodeTypeWorkflow, Name: "A1"})
	g.AddNode(&graph.Node{ID: "A2", Type: graph.NodeTypeStep, Name: "A2"})
	g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "A1", ToNodeID: "A2", Type: graph.EdgeTypeContains})

	g.AddNode(&graph.Node{ID: "B1", Type: graph.NodeTypeWorkflow, Name: "B1"})
	g.AddNode(&graph.Node{ID: "B2", Type: graph.NodeTypeStep, Name: "B2"})
	g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "B1", ToNodeID: "B2", Type: graph.EdgeTypeContains})

	options := &LayoutOptions{
		Type:         LayoutHierarchical,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute layout for disconnected graph: %v", err)
	}

	if len(layout.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in layout, got %d", len(layout.Nodes))
	}

	// Verify both roots are at level 0
	level0Nodes := layout.GetNodesByLevel(0)
	if len(level0Nodes) != 2 {
		t.Errorf("Expected 2 nodes at level 0, got %d", len(level0Nodes))
	}
}

func TestLayoutDeepHierarchy(t *testing.T) {
	g := graph.NewGraph("deep")

	// Create a deep chain: N0 -> N1 -> N2 -> ... -> N9
	depth := 10
	for i := 0; i < depth; i++ {
		nodeID := fmt.Sprintf("N%d", i)
		g.AddNode(&graph.Node{ID: nodeID, Type: graph.NodeTypeStep, Name: nodeID})

		if i > 0 {
			edgeID := fmt.Sprintf("e%d", i)
			fromID := fmt.Sprintf("N%d", i-1)
			g.AddEdge(&graph.Edge{ID: edgeID, FromNodeID: fromID, ToNodeID: nodeID, Type: graph.EdgeTypeDependsOn})
		}
	}

	options := &LayoutOptions{
		Type:         LayoutHierarchical,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute layout for deep hierarchy: %v", err)
	}

	// Verify each node is at its correct level
	for i := 0; i < depth; i++ {
		nodeID := fmt.Sprintf("N%d", i)
		nodeLayout := layout.Nodes[nodeID]
		if nodeLayout.Level != i {
			t.Errorf("Expected node %s at level %d, got %d", nodeID, i, nodeLayout.Level)
		}
	}

	// Verify Y coordinates increase with level
	for i := 1; i < depth; i++ {
		prevID := fmt.Sprintf("N%d", i-1)
		currID := fmt.Sprintf("N%d", i)
		if layout.Nodes[currID].Position.Y <= layout.Nodes[prevID].Position.Y {
			t.Errorf("Expected node %s to be below node %s", currID, prevID)
		}
	}
}

func TestLayoutWideGraph(t *testing.T) {
	g := graph.NewGraph("wide")

	// Create a wide graph: root with 20 children
	g.AddNode(&graph.Node{ID: "root", Type: graph.NodeTypeWorkflow, Name: "Root"})

	width := 20
	for i := 0; i < width; i++ {
		childID := fmt.Sprintf("child%d", i)
		g.AddNode(&graph.Node{ID: childID, Type: graph.NodeTypeStep, Name: childID})
		g.AddEdge(&graph.Edge{
			ID:         fmt.Sprintf("e%d", i),
			FromNodeID: "root",
			ToNodeID:   childID,
			Type:       graph.EdgeTypeContains,
		})
	}

	options := &LayoutOptions{
		Type:         LayoutHierarchical,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute layout for wide graph: %v", err)
	}

	// Verify all children are at level 1
	level1Nodes := layout.GetNodesByLevel(1)
	if len(level1Nodes) != width {
		t.Errorf("Expected %d nodes at level 1, got %d", width, len(level1Nodes))
	}

	// Verify all children have unique X positions
	xPositions := make(map[float64]bool)
	for i := 0; i < width; i++ {
		childID := fmt.Sprintf("child%d", i)
		x := layout.Nodes[childID].Position.X
		if xPositions[x] {
			t.Errorf("Duplicate X position found for node %s", childID)
		}
		xPositions[x] = true
	}
}

func TestLayoutUnknownType(t *testing.T) {
	g := createTestGraph()

	options := &LayoutOptions{
		Type:         LayoutType("unknown"),
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	_, err := ComputeLayout(g, options)
	if err == nil {
		t.Error("Expected error for unknown layout type")
	}
}

func TestLayoutNilOptions(t *testing.T) {
	g := createTestGraph()

	layout, err := ComputeLayout(g, nil)
	if err != nil {
		t.Fatalf("Failed to compute layout with nil options: %v", err)
	}

	// Should use default options (hierarchical)
	if len(layout.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in layout, got %d", len(layout.Nodes))
	}

	if layout.Options.Type != LayoutHierarchical {
		t.Errorf("Expected default layout type hierarchical, got %s", layout.Options.Type)
	}
}

func TestRadialLayoutNoRoots(t *testing.T) {
	g := graph.NewGraph("cycle")

	// Create a cycle: A -> B -> C -> A (no true roots)
	g.AddNode(&graph.Node{ID: "A", Type: graph.NodeTypeStep, Name: "A"})
	g.AddNode(&graph.Node{ID: "B", Type: graph.NodeTypeStep, Name: "B"})
	g.AddNode(&graph.Node{ID: "C", Type: graph.NodeTypeStep, Name: "C"})

	g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "A", ToNodeID: "B", Type: graph.EdgeTypeDependsOn})
	g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "B", ToNodeID: "C", Type: graph.EdgeTypeDependsOn})
	g.AddEdge(&graph.Edge{ID: "e3", FromNodeID: "C", ToNodeID: "A", Type: graph.EdgeTypeDependsOn})

	options := &LayoutOptions{
		Type:         LayoutRadial,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}

	layout, err := ComputeLayout(g, options)
	if err != nil {
		t.Fatalf("Failed to compute radial layout for cycle: %v", err)
	}

	if len(layout.Nodes) != 3 {
		t.Errorf("Expected 3 nodes in layout, got %d", len(layout.Nodes))
	}
}

func TestLayoutLargeGraph(t *testing.T) {
	g := graph.NewGraph("large")

	// Create graph with 100 nodes
	nodeCount := 100
	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("n%d", i)
		g.AddNode(&graph.Node{ID: nodeID, Type: graph.NodeTypeStep, Name: nodeID})
	}

	// Add edges to create a tree structure
	for i := 1; i < nodeCount; i++ {
		parentID := fmt.Sprintf("n%d", i/3)
		childID := fmt.Sprintf("n%d", i)
		edgeID := fmt.Sprintf("e%d", i)
		g.AddEdge(&graph.Edge{ID: edgeID, FromNodeID: parentID, ToNodeID: childID, Type: graph.EdgeTypeContains})
	}

	layouts := []LayoutType{LayoutHierarchical, LayoutRadial, LayoutGrid, LayoutForce}
	for _, layoutType := range layouts {
		t.Run(string(layoutType), func(t *testing.T) {
			options := &LayoutOptions{
				Type:         layoutType,
				NodeSpacing:  50.0,
				LevelSpacing: 75.0,
				Width:        2000.0,
				Height:       2000.0,
			}

			layout, err := ComputeLayout(g, options)
			if err != nil {
				t.Fatalf("Failed to compute %s layout for large graph: %v", layoutType, err)
			}

			if len(layout.Nodes) != nodeCount {
				t.Errorf("Expected %d nodes in layout, got %d", nodeCount, len(layout.Nodes))
			}
		})
	}
}

func TestGetNodesByLevel_NoMatchingLevel(t *testing.T) {
	g := createTestGraph()
	layout, err := ComputeLayout(g, DefaultLayoutOptions())
	if err != nil {
		t.Fatalf("Failed to compute layout: %v", err)
	}

	// Test level that doesn't exist
	level99Nodes := layout.GetNodesByLevel(99)
	if len(level99Nodes) != 0 {
		t.Errorf("Expected 0 nodes at level 99, got %d", len(level99Nodes))
	}
}

func TestFindRootNodes_AllRoots(t *testing.T) {
	g := graph.NewGraph("all-roots")

	// Create graph with no edges (all nodes are roots)
	g.AddNode(&graph.Node{ID: "A", Type: graph.NodeTypeStep, Name: "A"})
	g.AddNode(&graph.Node{ID: "B", Type: graph.NodeTypeStep, Name: "B"})
	g.AddNode(&graph.Node{ID: "C", Type: graph.NodeTypeStep, Name: "C"})

	roots := findRootNodes(g)

	if len(roots) != 3 {
		t.Errorf("Expected 3 root nodes, got %d", len(roots))
	}
}

func TestFindRootNodes_NoRoots(t *testing.T) {
	g := graph.NewGraph("no-roots")

	// Create a full cycle (no true roots)
	g.AddNode(&graph.Node{ID: "A", Type: graph.NodeTypeStep, Name: "A"})
	g.AddNode(&graph.Node{ID: "B", Type: graph.NodeTypeStep, Name: "B"})
	g.AddEdge(&graph.Edge{ID: "e1", FromNodeID: "A", ToNodeID: "B", Type: graph.EdgeTypeDependsOn})
	g.AddEdge(&graph.Edge{ID: "e2", FromNodeID: "B", ToNodeID: "A", Type: graph.EdgeTypeDependsOn})

	roots := findRootNodes(g)

	if len(roots) != 0 {
		t.Errorf("Expected 0 root nodes for cycle, got %d", len(roots))
	}
}
