package export

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"idp-orchestrator/pkg/graph"

	"github.com/goccy/go-graphviz"
)

type Format string

const (
	FormatDOT Format = "dot"
	FormatSVG Format = "svg"
	FormatPNG Format = "png"
)

type Exporter struct {
	graphviz *graphviz.Graphviz
}

func NewExporter() *Exporter {
	g, _ := graphviz.New(context.Background())
	return &Exporter{
		graphviz: g,
	}
}

func (e *Exporter) Close() error {
	return e.graphviz.Close()
}

func (e *Exporter) ExportGraph(g *graph.Graph, format Format) ([]byte, error) {
	dotContent, err := e.generateDOT(g)
	if err != nil {
		return nil, fmt.Errorf("failed to generate DOT: %w", err)
	}

	if format == FormatDOT {
		return []byte(dotContent), nil
	}

	gvGraph, err := graphviz.ParseBytes([]byte(dotContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DOT content: %w", err)
	}
	defer gvGraph.Close()

	var buf bytes.Buffer
	ctx := context.Background()
	switch format {
	case FormatSVG:
		err = e.graphviz.Render(ctx, gvGraph, graphviz.SVG, &buf)
	case FormatPNG:
		err = e.graphviz.Render(ctx, gvGraph, graphviz.PNG, &buf)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to render graph: %w", err)
	}

	return buf.Bytes(), nil
}

func (e *Exporter) generateDOT(g *graph.Graph) (string, error) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("digraph \"%s\" {\n", g.AppName))
	buf.WriteString("  rankdir=TB;\n")
	buf.WriteString("  node [shape=box, style=rounded];\n")
	buf.WriteString("  edge [fontsize=10];\n\n")

	for _, node := range g.Nodes {
		nodeColor := e.getNodeColor(node.Type)
		nodeLabel := e.escapeLabel(fmt.Sprintf("%s\\n(%s)", node.Name, node.Type))

		buf.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\", fillcolor=\"%s\", style=\"filled,rounded\"];\n",
			node.ID, nodeLabel, nodeColor))
	}

	buf.WriteString("\n")

	for _, edge := range g.Edges {
		edgeLabel := string(edge.Type)
		if edge.Description != "" {
			edgeLabel = fmt.Sprintf("%s\\n%s", edgeLabel, e.escapeLabel(edge.Description))
		}

		edgeColor := e.getEdgeColor(edge.Type)
		edgeStyle := e.getEdgeStyle(edge.Type)

		buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\", color=\"%s\", style=\"%s\"];\n",
			edge.FromNodeID, edge.ToNodeID, edgeLabel, edgeColor, edgeStyle))
	}

	buf.WriteString("}\n")

	return buf.String(), nil
}

func (e *Exporter) getNodeColor(nodeType graph.NodeType) string {
	switch nodeType {
	case graph.NodeTypeSpec:
		return "#E3F2FD" // Light blue
	case graph.NodeTypeWorkflow:
		return "#E8F5E8" // Light green
	case graph.NodeTypeResource:
		return "#FFF3E0" // Light orange
	default:
		return "#F5F5F5" // Light gray
	}
}

func (e *Exporter) getEdgeColor(edgeType graph.EdgeType) string {
	switch edgeType {
	case graph.EdgeTypeDependsOn:
		return "#1976D2" // Blue
	case graph.EdgeTypeProvisions:
		return "#388E3C" // Green
	case graph.EdgeTypeCreates:
		return "#F57C00" // Orange
	case graph.EdgeTypeBindsTo:
		return "#7B1FA2" // Purple
	default:
		return "#757575" // Gray
	}
}

func (e *Exporter) getEdgeStyle(edgeType graph.EdgeType) string {
	switch edgeType {
	case graph.EdgeTypeDependsOn:
		return "solid"
	case graph.EdgeTypeProvisions:
		return "bold"
	case graph.EdgeTypeCreates:
		return "dashed"
	case graph.EdgeTypeBindsTo:
		return "dotted"
	default:
		return "solid"
	}
}

func (e *Exporter) escapeLabel(label string) string {
	label = strings.ReplaceAll(label, "\"", "\\\"")
	label = strings.ReplaceAll(label, "\n", "\\n")
	return label
}

func (e *Exporter) CreateSubgraph(g *graph.Graph, nodeIDs []string) (*graph.Graph, error) {
	subgraph := graph.NewGraph(g.AppName + "-subgraph")

	nodeMap := make(map[string]bool)
	for _, id := range nodeIDs {
		nodeMap[id] = true
	}

	for _, nodeID := range nodeIDs {
		if node, exists := g.GetNode(nodeID); exists {
			if err := subgraph.AddNode(node); err != nil {
				return nil, fmt.Errorf("failed to add node %s to subgraph: %w", nodeID, err)
			}
		}
	}

	for _, edge := range g.Edges {
		if nodeMap[edge.FromNodeID] && nodeMap[edge.ToNodeID] {
			if err := subgraph.AddEdge(edge); err != nil {
				return nil, fmt.Errorf("failed to add edge %s to subgraph: %w", edge.ID, err)
			}
		}
	}

	return subgraph, nil
}