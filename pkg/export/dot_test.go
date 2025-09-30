package export

import (
	"testing"

	"idp-orchestrator/pkg/graph"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGraph() *graph.Graph {
	g := graph.NewGraph("test-app")

	nodes := []*graph.Node{
		{ID: "spec1", Type: graph.NodeTypeSpec, Name: "Database Spec"},
		{ID: "workflow1", Type: graph.NodeTypeWorkflow, Name: "Deploy Database"},
		{ID: "resource1", Type: graph.NodeTypeResource, Name: "Database"},
	}

	for _, node := range nodes {
		require.NoError(nil, g.AddNode(node))
	}

	edges := []*graph.Edge{
		{ID: "e1", FromNodeID: "workflow1", ToNodeID: "spec1", Type: graph.EdgeTypeDependsOn, Description: "needs spec"},
		{ID: "e2", FromNodeID: "workflow1", ToNodeID: "resource1", Type: graph.EdgeTypeProvisions, Description: "creates database"},
	}

	for _, edge := range edges {
		require.NoError(nil, g.AddEdge(edge))
	}

	return g
}

func TestExporter_generateDOT(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	dotContent, err := exporter.generateDOT(g)
	require.NoError(t, err)

	assert.Contains(t, dotContent, `digraph "test-app"`)
	assert.Contains(t, dotContent, `rankdir=TB`)

	assert.Contains(t, dotContent, `"spec1"`)
	assert.Contains(t, dotContent, `"workflow1"`)
	assert.Contains(t, dotContent, `"resource1"`)

	assert.Contains(t, dotContent, `"workflow1" -> "spec1"`)
	assert.Contains(t, dotContent, `"workflow1" -> "resource1"`)

	assert.Contains(t, dotContent, `[label="depends-on\nneeds spec"`)
	assert.Contains(t, dotContent, `[label="provisions\ncreates database"`)
}

func TestExporter_ExportGraph_DOT(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	data, err := exporter.ExportGraph(g, FormatDOT)
	require.NoError(t, err)

	dotContent := string(data)
	assert.Contains(t, dotContent, `digraph "test-app"`)
	assert.Contains(t, dotContent, `"spec1"`)
}

func TestExporter_ExportGraph_SVG(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	data, err := exporter.ExportGraph(g, FormatSVG)
	require.NoError(t, err)

	svgContent := string(data)
	assert.Contains(t, svgContent, `<svg`)
	assert.Contains(t, svgContent, `</svg>`)
}

func TestExporter_ExportGraph_PNG(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	data, err := exporter.ExportGraph(g, FormatPNG)
	require.NoError(t, err)

	assert.True(t, len(data) > 0, "PNG data should not be empty")
	assert.Equal(t, byte(0x89), data[0], "PNG should start with PNG signature")
	assert.Equal(t, "PNG", string(data[1:4]), "PNG signature should be correct")
}

func TestExporter_ExportGraph_UnsupportedFormat(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	_, err := exporter.ExportGraph(g, Format("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestExporter_getNodeColor(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	tests := []struct {
		nodeType graph.NodeType
		expected string
	}{
		{graph.NodeTypeSpec, "#E3F2FD"},
		{graph.NodeTypeWorkflow, "#E8F5E8"},
		{graph.NodeTypeResource, "#FFF3E0"},
		{graph.NodeType("unknown"), "#F5F5F5"},
	}

	for _, test := range tests {
		color := exporter.getNodeColor(test.nodeType)
		assert.Equal(t, test.expected, color)
	}
}

func TestExporter_getEdgeColor(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	tests := []struct {
		edgeType graph.EdgeType
		expected string
	}{
		{graph.EdgeTypeDependsOn, "#1976D2"},
		{graph.EdgeTypeProvisions, "#388E3C"},
		{graph.EdgeTypeCreates, "#F57C00"},
		{graph.EdgeTypeBindsTo, "#7B1FA2"},
		{graph.EdgeType("unknown"), "#757575"},
	}

	for _, test := range tests {
		color := exporter.getEdgeColor(test.edgeType)
		assert.Equal(t, test.expected, color)
	}
}

func TestExporter_getEdgeStyle(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	tests := []struct {
		edgeType graph.EdgeType
		expected string
	}{
		{graph.EdgeTypeDependsOn, "solid"},
		{graph.EdgeTypeProvisions, "bold"},
		{graph.EdgeTypeCreates, "dashed"},
		{graph.EdgeTypeBindsTo, "dotted"},
		{graph.EdgeType("unknown"), "solid"},
	}

	for _, test := range tests {
		style := exporter.getEdgeStyle(test.edgeType)
		assert.Equal(t, test.expected, style)
	}
}

func TestExporter_escapeLabel(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	tests := []struct {
		input    string
		expected string
	}{
		{`simple`, `simple`},
		{`with"quotes`, `with\"quotes`},
		{"with\nnewlines", `with\nnewlines`},
		{`with"quotes\nand\nnewlines`, `with\"quotes\nand\nnewlines`},
	}

	for _, test := range tests {
		result := exporter.escapeLabel(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestExporter_CreateSubgraph(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	nodeIDs := []string{"spec1", "workflow1"}

	subgraph, err := exporter.CreateSubgraph(g, nodeIDs)
	require.NoError(t, err)

	assert.Len(t, subgraph.Nodes, 2)
	assert.Len(t, subgraph.Edges, 1)

	_, exists := subgraph.GetNode("spec1")
	assert.True(t, exists)

	_, exists = subgraph.GetNode("workflow1")
	assert.True(t, exists)

	_, exists = subgraph.GetNode("resource1")
	assert.False(t, exists)

	_, exists = subgraph.GetEdge("e1")
	assert.True(t, exists)

	_, exists = subgraph.GetEdge("e2")
	assert.False(t, exists)
}

func TestExporter_CreateSubgraph_EmptyNodeList(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	nodeIDs := []string{}

	subgraph, err := exporter.CreateSubgraph(g, nodeIDs)
	require.NoError(t, err)

	assert.Empty(t, subgraph.Nodes)
	assert.Empty(t, subgraph.Edges)
}

func TestExporter_CreateSubgraph_NonExistentNode(t *testing.T) {
	exporter := NewExporter()
	defer exporter.Close()

	g := createTestGraph()
	nodeIDs := []string{"spec1", "missing"}

	subgraph, err := exporter.CreateSubgraph(g, nodeIDs)
	require.NoError(t, err)

	assert.Len(t, subgraph.Nodes, 1)
	_, exists := subgraph.GetNode("spec1")
	assert.True(t, exists)
	_, exists = subgraph.GetNode("missing")
	assert.False(t, exists)
}