package export

import (
	"encoding/json"
	"fmt"

	"github.com/philipsahli/innominatus-graph/pkg/graph"
)

// JSONExportOptions configures JSON export behavior
type JSONExportOptions struct {
	// IncludeMetadata includes graph metadata (created_at, updated_at, version)
	IncludeMetadata bool
	// IncludeProperties includes custom node/edge properties
	IncludeProperties bool
	// IncludeTiming includes timing information (started_at, completed_at, duration)
	IncludeTiming bool
	// PrettyPrint formats JSON with indentation
	PrettyPrint bool
}

// DefaultJSONOptions returns default JSON export options
func DefaultJSONOptions() *JSONExportOptions {
	return &JSONExportOptions{
		IncludeMetadata:   true,
		IncludeProperties: true,
		IncludeTiming:     true,
		PrettyPrint:       false,
	}
}

// JSONNode represents a node in JSON export format
type JSONNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	State       string                 `json:"state"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	StartedAt   *string                `json:"started_at,omitempty"`
	CompletedAt *string                `json:"completed_at,omitempty"`
	Duration    *string                `json:"duration,omitempty"`
	CreatedAt   *string                `json:"created_at,omitempty"`
	UpdatedAt   *string                `json:"updated_at,omitempty"`
}

// JSONEdge represents an edge in JSON export format
type JSONEdge struct {
	ID          string                 `json:"id"`
	FromNodeID  string                 `json:"from_node_id"`
	ToNodeID    string                 `json:"to_node_id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   *string                `json:"created_at,omitempty"`
}

// JSONGraph represents a graph in JSON export format
type JSONGraph struct {
	ID        string      `json:"id"`
	AppName   string      `json:"app_name"`
	Version   *int        `json:"version,omitempty"`
	Nodes     []JSONNode  `json:"nodes"`
	Edges     []JSONEdge  `json:"edges"`
	CreatedAt *string     `json:"created_at,omitempty"`
	UpdatedAt *string     `json:"updated_at,omitempty"`
}

// ExportGraphJSON exports a graph to JSON format with configurable options
func ExportGraphJSON(g *graph.Graph, options *JSONExportOptions) ([]byte, error) {
	if options == nil {
		options = DefaultJSONOptions()
	}

	jsonGraph := &JSONGraph{
		ID:      g.ID,
		AppName: g.AppName,
		Nodes:   make([]JSONNode, 0, len(g.Nodes)),
		Edges:   make([]JSONEdge, 0, len(g.Edges)),
	}

	// Include graph metadata
	if options.IncludeMetadata {
		version := g.Version
		jsonGraph.Version = &version
		createdAt := g.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		updatedAt := g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		jsonGraph.CreatedAt = &createdAt
		jsonGraph.UpdatedAt = &updatedAt
	}

	// Convert nodes
	for _, node := range g.Nodes {
		jsonNode := JSONNode{
			ID:          node.ID,
			Type:        string(node.Type),
			Name:        node.Name,
			Description: node.Description,
			State:       string(node.State),
		}

		// Include properties
		if options.IncludeProperties && len(node.Properties) > 0 {
			jsonNode.Properties = node.Properties
		}

		// Include timing information
		if options.IncludeTiming {
			if node.StartedAt != nil {
				startedAt := node.StartedAt.Format("2006-01-02T15:04:05Z07:00")
				jsonNode.StartedAt = &startedAt
			}
			if node.CompletedAt != nil {
				completedAt := node.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
				jsonNode.CompletedAt = &completedAt
			}
			if node.Duration != nil {
				duration := node.Duration.String()
				jsonNode.Duration = &duration
			}
		}

		// Include metadata
		if options.IncludeMetadata {
			createdAt := node.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
			updatedAt := node.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
			jsonNode.CreatedAt = &createdAt
			jsonNode.UpdatedAt = &updatedAt
		}

		jsonGraph.Nodes = append(jsonGraph.Nodes, jsonNode)
	}

	// Convert edges
	for _, edge := range g.Edges {
		jsonEdge := JSONEdge{
			ID:          edge.ID,
			FromNodeID:  edge.FromNodeID,
			ToNodeID:    edge.ToNodeID,
			Type:        string(edge.Type),
			Description: edge.Description,
		}

		// Include properties
		if options.IncludeProperties && len(edge.Properties) > 0 {
			jsonEdge.Properties = edge.Properties
		}

		// Include metadata
		if options.IncludeMetadata {
			createdAt := edge.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
			jsonEdge.CreatedAt = &createdAt
		}

		jsonGraph.Edges = append(jsonGraph.Edges, jsonEdge)
	}

	// Marshal to JSON
	var data []byte
	var err error
	if options.PrettyPrint {
		data, err = json.MarshalIndent(jsonGraph, "", "  ")
	} else {
		data, err = json.Marshal(jsonGraph)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal graph to JSON: %w", err)
	}

	return data, nil
}

// ExportGraphJSONCompact exports a graph to compact JSON (minimal metadata)
func ExportGraphJSONCompact(g *graph.Graph) ([]byte, error) {
	options := &JSONExportOptions{
		IncludeMetadata:   false,
		IncludeProperties: true,
		IncludeTiming:     true,
		PrettyPrint:       false,
	}
	return ExportGraphJSON(g, options)
}

// ExportGraphJSONPretty exports a graph to pretty-printed JSON
func ExportGraphJSONPretty(g *graph.Graph) ([]byte, error) {
	options := &JSONExportOptions{
		IncludeMetadata:   true,
		IncludeProperties: true,
		IncludeTiming:     true,
		PrettyPrint:       true,
	}
	return ExportGraphJSON(g, options)
}
