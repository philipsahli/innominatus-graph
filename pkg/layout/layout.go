package layout

import (
	"fmt"
	"math"
	"sort"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

// Position represents a 2D coordinate
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// LayoutType specifies the layout algorithm to use
type LayoutType string

const (
	// LayoutHierarchical arranges nodes in layers (top-to-bottom)
	LayoutHierarchical LayoutType = "hierarchical"
	// LayoutRadial arranges nodes in concentric circles
	LayoutRadial LayoutType = "radial"
	// LayoutForce uses force-directed algorithm
	LayoutForce LayoutType = "force"
	// LayoutGrid arranges nodes in a grid
	LayoutGrid LayoutType = "grid"
)

// LayoutOptions configures layout computation
type LayoutOptions struct {
	// Type specifies the layout algorithm
	Type LayoutType
	// NodeSpacing controls spacing between nodes
	NodeSpacing float64
	// LevelSpacing controls spacing between levels (hierarchical)
	LevelSpacing float64
	// Width of the layout area
	Width float64
	// Height of the layout area
	Height float64
}

// DefaultLayoutOptions returns default layout options
func DefaultLayoutOptions() *LayoutOptions {
	return &LayoutOptions{
		Type:         LayoutHierarchical,
		NodeSpacing:  100.0,
		LevelSpacing: 150.0,
		Width:        1200.0,
		Height:       800.0,
	}
}

// NodeLayout contains positioning information for a node
type NodeLayout struct {
	NodeID   string   `json:"node_id"`
	Position Position `json:"position"`
	Level    int      `json:"level"` // For hierarchical layouts
}

// GraphLayout contains layout information for an entire graph
type GraphLayout struct {
	Nodes   map[string]*NodeLayout `json:"nodes"`
	Options *LayoutOptions         `json:"options"`
}

// ComputeLayout calculates positions for all nodes in a graph
func ComputeLayout(g *graph.Graph, options *LayoutOptions) (*GraphLayout, error) {
	if options == nil {
		options = DefaultLayoutOptions()
	}

	switch options.Type {
	case LayoutHierarchical:
		return computeHierarchicalLayout(g, options)
	case LayoutRadial:
		return computeRadialLayout(g, options)
	case LayoutGrid:
		return computeGridLayout(g, options)
	case LayoutForce:
		return computeForceLayout(g, options)
	default:
		return nil, fmt.Errorf("unsupported layout type: %s", options.Type)
	}
}

// computeHierarchicalLayout arranges nodes in layers based on dependencies
func computeHierarchicalLayout(g *graph.Graph, options *LayoutOptions) (*GraphLayout, error) {
	layout := &GraphLayout{
		Nodes:   make(map[string]*NodeLayout),
		Options: options,
	}

	// Calculate levels using BFS from root nodes
	// For hierarchical layout, we use VISUAL edge direction (A -> B means A is above B)
	// ignoring execution semantics

	levels := make(map[string]int)
	maxLevel := 0

	// Find root nodes (nodes with no incoming edges)
	roots := findRootNodes(g)

	// BFS to assign levels
	queue := make([]string, len(roots))
	for i, rootID := range roots {
		queue[i] = rootID
		levels[rootID] = 0
	}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		currentLevel := levels[currentID]

		// Find all nodes that this node points TO
		for _, edge := range g.Edges {
			if edge.FromNodeID == currentID {
				childID := edge.ToNodeID

				// Update child's level if we found a longer path
				newLevel := currentLevel + 1
				if existingLevel, exists := levels[childID]; !exists || newLevel > existingLevel {
					levels[childID] = newLevel
					queue = append(queue, childID)

					if newLevel > maxLevel {
						maxLevel = newLevel
					}
				}
			}
		}
	}

	// Ensure all nodes have a level (handle disconnected nodes)
	for nodeID := range g.Nodes {
		if _, exists := levels[nodeID]; !exists {
			levels[nodeID] = 0
		}
	}

	// Group nodes by level
	nodesPerLevel := make(map[int][]string)
	for nodeID, level := range levels {
		nodesPerLevel[level] = append(nodesPerLevel[level], nodeID)
	}

	// Position nodes
	for level := 0; level <= maxLevel; level++ {
		nodes := nodesPerLevel[level]
		numNodes := len(nodes)

		y := float64(level) * options.LevelSpacing

		for i, nodeID := range nodes {
			// Center nodes horizontally
			totalWidth := float64(numNodes-1) * options.NodeSpacing
			startX := (options.Width - totalWidth) / 2
			x := startX + float64(i)*options.NodeSpacing

			layout.Nodes[nodeID] = &NodeLayout{
				NodeID:   nodeID,
				Position: Position{X: x, Y: y},
				Level:    level,
			}
		}
	}

	return layout, nil
}

// computeRadialLayout arranges nodes in concentric circles
func computeRadialLayout(g *graph.Graph, options *LayoutOptions) (*GraphLayout, error) {
	layout := &GraphLayout{
		Nodes:   make(map[string]*NodeLayout),
		Options: options,
	}

	// Find root nodes (nodes with no incoming edges)
	roots := findRootNodes(g)
	if len(roots) == 0 {
		// If no roots, pick any node
		for nodeID := range g.Nodes {
			roots = append(roots, nodeID)
			break
		}
	}

	// Assign nodes to levels based on distance from root
	levels := make(map[string]int)
	visited := make(map[string]bool)
	maxLevel := 0

	var assignLevel func(nodeID string, level int)
	assignLevel = func(nodeID string, level int) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true
		levels[nodeID] = level
		if level > maxLevel {
			maxLevel = level
		}

		// Visit children
		for _, edge := range g.Edges {
			if edge.FromNodeID == nodeID {
				assignLevel(edge.ToNodeID, level+1)
			}
		}
	}

	for _, root := range roots {
		assignLevel(root, 0)
	}

	// Group nodes by level
	nodesPerLevel := make(map[int][]string)
	for nodeID, level := range levels {
		nodesPerLevel[level] = append(nodesPerLevel[level], nodeID)
	}

	// Position nodes in circles
	centerX := options.Width / 2
	centerY := options.Height / 2
	baseRadius := 50.0

	for level := 0; level <= maxLevel; level++ {
		nodes := nodesPerLevel[level]
		numNodes := len(nodes)
		radius := baseRadius + float64(level)*options.LevelSpacing

		for i, nodeID := range nodes {
			angle := 2 * math.Pi * float64(i) / float64(numNodes)
			x := centerX + radius*math.Cos(angle)
			y := centerY + radius*math.Sin(angle)

			layout.Nodes[nodeID] = &NodeLayout{
				NodeID:   nodeID,
				Position: Position{X: x, Y: y},
				Level:    level,
			}
		}
	}

	return layout, nil
}

// computeGridLayout arranges nodes in a simple grid
func computeGridLayout(g *graph.Graph, options *LayoutOptions) (*GraphLayout, error) {
	layout := &GraphLayout{
		Nodes:   make(map[string]*NodeLayout),
		Options: options,
	}

	// Collect all node IDs
	nodeIDs := make([]string, 0, len(g.Nodes))
	for nodeID := range g.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs) // Stable ordering

	// Calculate grid dimensions
	numNodes := len(nodeIDs)
	cols := int(math.Ceil(math.Sqrt(float64(numNodes))))

	// Position nodes in grid
	for i, nodeID := range nodeIDs {
		row := i / cols
		col := i % cols

		x := float64(col) * options.NodeSpacing
		y := float64(row) * options.LevelSpacing

		layout.Nodes[nodeID] = &NodeLayout{
			NodeID:   nodeID,
			Position: Position{X: x, Y: y},
			Level:    row,
		}
	}

	return layout, nil
}

// computeForceLayout uses a simple force-directed algorithm
func computeForceLayout(g *graph.Graph, options *LayoutOptions) (*GraphLayout, error) {
	layout := &GraphLayout{
		Nodes:   make(map[string]*NodeLayout),
		Options: options,
	}

	// Initialize random positions
	i := 0
	for nodeID := range g.Nodes {
		x := float64(i%10) * options.NodeSpacing
		y := float64(i/10) * options.LevelSpacing
		layout.Nodes[nodeID] = &NodeLayout{
			NodeID:   nodeID,
			Position: Position{X: x, Y: y},
			Level:    0,
		}
		i++
	}

	// Run force simulation (simplified version)
	iterations := 100
	repulsionStrength := 1000.0
	attractionStrength := 0.1
	damping := 0.9

	for iter := 0; iter < iterations; iter++ {
		forces := make(map[string]Position)

		// Calculate repulsion between all nodes
		for nodeID1 := range g.Nodes {
			force := Position{X: 0, Y: 0}

			for nodeID2 := range g.Nodes {
				if nodeID1 == nodeID2 {
					continue
				}

				pos1 := layout.Nodes[nodeID1].Position
				pos2 := layout.Nodes[nodeID2].Position

				dx := pos1.X - pos2.X
				dy := pos1.Y - pos2.Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance < 1.0 {
					distance = 1.0
				}

				repulsion := repulsionStrength / (distance * distance)
				force.X += (dx / distance) * repulsion
				force.Y += (dy / distance) * repulsion
			}

			forces[nodeID1] = force
		}

		// Calculate attraction along edges
		for _, edge := range g.Edges {
			pos1 := layout.Nodes[edge.FromNodeID].Position
			pos2 := layout.Nodes[edge.ToNodeID].Position

			dx := pos2.X - pos1.X
			dy := pos2.Y - pos1.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance > 0 {
				attraction := distance * attractionStrength

				force1 := forces[edge.FromNodeID]
				force1.X += (dx / distance) * attraction
				force1.Y += (dy / distance) * attraction
				forces[edge.FromNodeID] = force1

				force2 := forces[edge.ToNodeID]
				force2.X -= (dx / distance) * attraction
				force2.Y -= (dy / distance) * attraction
				forces[edge.ToNodeID] = force2
			}
		}

		// Apply forces
		for nodeID, force := range forces {
			pos := layout.Nodes[nodeID].Position
			pos.X += force.X * damping
			pos.Y += force.Y * damping
			layout.Nodes[nodeID].Position = pos
		}
	}

	return layout, nil
}

// Helper functions

func findRootNodes(g *graph.Graph) []string {
	hasIncoming := make(map[string]bool)

	// Mark nodes with incoming edges
	for _, edge := range g.Edges {
		hasIncoming[edge.ToNodeID] = true
	}

	// Find nodes without incoming edges
	roots := make([]string, 0)
	for nodeID := range g.Nodes {
		if !hasIncoming[nodeID] {
			roots = append(roots, nodeID)
		}
	}

	return roots
}

// GetNodePosition returns the position of a specific node
func (gl *GraphLayout) GetNodePosition(nodeID string) (Position, bool) {
	if nodeLayout, exists := gl.Nodes[nodeID]; exists {
		return nodeLayout.Position, true
	}
	return Position{}, false
}

// GetNodesByLevel returns all nodes at a specific level (for hierarchical layouts)
func (gl *GraphLayout) GetNodesByLevel(level int) []string {
	nodes := make([]string, 0)
	for _, nodeLayout := range gl.Nodes {
		if nodeLayout.Level == level {
			nodes = append(nodes, nodeLayout.NodeID)
		}
	}
	return nodes
}
