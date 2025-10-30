package export

import (
	"fmt"
	"strings"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

// MermaidDiagramType specifies the type of Mermaid diagram to generate
type MermaidDiagramType string

const (
	// MermaidFlowchart generates a flowchart (graph TD)
	MermaidFlowchart MermaidDiagramType = "flowchart"
	// MermaidStateDiagram generates a state diagram
	MermaidStateDiagram MermaidDiagramType = "state"
	// MermaidGantt generates a Gantt chart (timeline view)
	MermaidGantt MermaidDiagramType = "gantt"
)

// MermaidExportOptions configures Mermaid export behavior
type MermaidExportOptions struct {
	// DiagramType specifies the Mermaid diagram type
	DiagramType MermaidDiagramType
	// Direction sets flowchart direction (TB, TD, BT, RL, LR)
	Direction string
	// IncludeState shows node states in labels
	IncludeState bool
	// IncludeTiming shows timing information
	IncludeTiming bool
	// Theme specifies Mermaid theme (default, forest, dark, neutral)
	Theme string
}

// DefaultMermaidOptions returns default Mermaid export options
func DefaultMermaidOptions() *MermaidExportOptions {
	return &MermaidExportOptions{
		DiagramType:   MermaidFlowchart,
		Direction:     "TB",
		IncludeState:  true,
		IncludeTiming: false,
		Theme:         "default",
	}
}

// ExportGraphMermaid exports a graph to Mermaid diagram format
func ExportGraphMermaid(g *graph.Graph, options *MermaidExportOptions) (string, error) {
	if options == nil {
		options = DefaultMermaidOptions()
	}

	switch options.DiagramType {
	case MermaidFlowchart:
		return exportMermaidFlowchart(g, options)
	case MermaidStateDiagram:
		return exportMermaidStateDiagram(g, options)
	case MermaidGantt:
		return exportMermaidGantt(g, options)
	default:
		return "", fmt.Errorf("unsupported diagram type: %s", options.DiagramType)
	}
}

// exportMermaidFlowchart generates a Mermaid flowchart
func exportMermaidFlowchart(g *graph.Graph, options *MermaidExportOptions) (string, error) {
	var buf strings.Builder

	// Header
	buf.WriteString(fmt.Sprintf("---\ntitle: %s\n---\n", g.AppName))
	buf.WriteString(fmt.Sprintf("flowchart %s\n", options.Direction))

	// Theme initialization (optional)
	if options.Theme != "default" {
		buf.WriteString(fmt.Sprintf("    %%{init: {'theme':'%s'}}%%\n", options.Theme))
	}

	// Define nodes
	for _, node := range g.Nodes {
		nodeID := sanitizeID(node.ID)
		label := node.Name

		// Add state to label if requested
		if options.IncludeState && node.State != "" {
			label = fmt.Sprintf("%s [%s]", label, node.State)
		}

		// Add timing if requested
		if options.IncludeTiming && node.Duration != nil {
			label = fmt.Sprintf("%s (%s)", label, node.Duration.String())
		}

		// Determine node shape based on type
		nodeShape := getNodeShape(node.Type)
		nodeClass := getNodeClass(node.State)

		switch nodeShape {
		case "rectangle":
			buf.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeID, escapeLabel(label)))
		case "rounded":
			buf.WriteString(fmt.Sprintf("    %s(\"%s\")\n", nodeID, escapeLabel(label)))
		case "stadium":
			buf.WriteString(fmt.Sprintf("    %s([%s])\n", nodeID, escapeLabel(label)))
		case "diamond":
			buf.WriteString(fmt.Sprintf("    %s{%s}\n", nodeID, escapeLabel(label)))
		case "circle":
			buf.WriteString(fmt.Sprintf("    %s((%s))\n", nodeID, escapeLabel(label)))
		default:
			buf.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeID, escapeLabel(label)))
		}

		// Apply class styling
		if nodeClass != "" {
			buf.WriteString(fmt.Sprintf("    class %s %s\n", nodeID, nodeClass))
		}
	}

	buf.WriteString("\n")

	// Define edges
	for _, edge := range g.Edges {
		fromID := sanitizeID(edge.FromNodeID)
		toID := sanitizeID(edge.ToNodeID)
		edgeLabel := string(edge.Type)

		// Different arrow styles based on edge type
		arrow := getArrowStyle(edge.Type)

		buf.WriteString(fmt.Sprintf("    %s %s|%s| %s\n", fromID, arrow, edgeLabel, toID))
	}

	buf.WriteString("\n")

	// Define class styles
	buf.WriteString("    classDef running fill:#bbdefb,stroke:#1976d2,stroke-width:3px\n")
	buf.WriteString("    classDef succeeded fill:#c8e6c9,stroke:#388e3c,stroke-width:2px\n")
	buf.WriteString("    classDef failed fill:#ffcdd2,stroke:#d32f2f,stroke-width:3px\n")
	buf.WriteString("    classDef pending fill:#fff9c4,stroke:#fbc02d,stroke-width:2px\n")

	return buf.String(), nil
}

// exportMermaidStateDiagram generates a Mermaid state diagram
func exportMermaidStateDiagram(g *graph.Graph, options *MermaidExportOptions) (string, error) {
	var buf strings.Builder

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("title: %s State Machine\n", g.AppName))
	buf.WriteString("---\n")
	buf.WriteString("stateDiagram-v2\n")

	// Define states
	for _, node := range g.Nodes {
		nodeID := sanitizeID(node.ID)
		label := fmt.Sprintf("%s: %s", node.Name, node.State)
		buf.WriteString(fmt.Sprintf("    %s: %s\n", nodeID, label))
	}

	buf.WriteString("\n")

	// Define transitions
	for _, edge := range g.Edges {
		fromID := sanitizeID(edge.FromNodeID)
		toID := sanitizeID(edge.ToNodeID)
		buf.WriteString(fmt.Sprintf("    %s --> %s: %s\n", fromID, toID, edge.Type))
	}

	return buf.String(), nil
}

// exportMermaidGantt generates a Mermaid Gantt chart (timeline)
func exportMermaidGantt(g *graph.Graph, options *MermaidExportOptions) (string, error) {
	var buf strings.Builder

	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("title: %s Timeline\n", g.AppName))
	buf.WriteString("---\n")
	buf.WriteString("gantt\n")
	buf.WriteString("    dateFormat YYYY-MM-DD HH:mm:ss\n")
	buf.WriteString("    axisFormat %H:%M:%S\n")

	// Group nodes by type
	sections := make(map[string][]*graph.Node)
	for _, node := range g.Nodes {
		sections[string(node.Type)] = append(sections[string(node.Type)], node)
	}

	// Generate sections for each node type
	for nodeType, nodes := range sections {
		buf.WriteString(fmt.Sprintf("\n    section %s\n", nodeType))

		for _, node := range nodes {
			// Only include nodes with timing information
			if node.StartedAt == nil {
				continue
			}

			status := getGanttStatus(node.State)
			taskName := node.Name

			if node.CompletedAt != nil {
				// Task with start and end
				buf.WriteString(fmt.Sprintf("    %s : %s, %s, %s\n",
					taskName,
					status,
					node.StartedAt.Format("2006-01-02 15:04:05"),
					node.CompletedAt.Format("2006-01-02 15:04:05")))
			} else {
				// Task with only start time
				buf.WriteString(fmt.Sprintf("    %s : %s, %s, 1m\n",
					taskName,
					status,
					node.StartedAt.Format("2006-01-02 15:04:05")))
			}
		}
	}

	return buf.String(), nil
}

// Helper functions

func getNodeShape(nodeType graph.NodeType) string {
	switch nodeType {
	case graph.NodeTypeSpec:
		return "rectangle"
	case graph.NodeTypeWorkflow:
		return "rounded"
	case graph.NodeTypeStep:
		return "stadium"
	case graph.NodeTypeResource:
		return "circle"
	default:
		return "rectangle"
	}
}

func getNodeClass(state graph.NodeState) string {
	switch state {
	case graph.NodeStateRunning:
		return "running"
	case graph.NodeStateSucceeded:
		return "succeeded"
	case graph.NodeStateFailed:
		return "failed"
	case graph.NodeStatePending:
		return "pending"
	default:
		return ""
	}
}

func getArrowStyle(edgeType graph.EdgeType) string {
	switch edgeType {
	case graph.EdgeTypeDependsOn:
		return "-->"
	case graph.EdgeTypeProvisions:
		return "==>"
	case graph.EdgeTypeCreates:
		return "-->"
	case graph.EdgeTypeBindsTo:
		return "-.->"
	case graph.EdgeTypeContains:
		return "-->"
	case graph.EdgeTypeConfigures:
		return "-.->"
	default:
		return "-->"
	}
}

func getGanttStatus(state graph.NodeState) string {
	switch state {
	case graph.NodeStateRunning:
		return "active"
	case graph.NodeStateSucceeded:
		return "done"
	case graph.NodeStateFailed:
		return "crit"
	default:
		return ""
	}
}

func sanitizeID(id string) string {
	// Replace invalid characters with underscores
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, ".", "_")
	id = strings.ReplaceAll(id, " ", "_")
	return id
}

func escapeLabel(label string) string {
	// Escape quotes and special characters
	label = strings.ReplaceAll(label, "\"", "&quot;")
	label = strings.ReplaceAll(label, "#", "&num;")
	return label
}
