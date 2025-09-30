package storage

import (
	"encoding/json"
	"fmt"

	"idp-orchestrator/pkg/graph"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveGraph(appName string, g *graph.Graph) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var app App
		err := tx.Where("name = ?", appName).First(&app).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				app = App{Name: appName}
				if err := tx.Create(&app).Error; err != nil {
					return fmt.Errorf("failed to create app: %w", err)
				}
			} else {
				return fmt.Errorf("failed to find app: %w", err)
			}
		}

		if err := tx.Where("app_id = ?", app.ID).Delete(&EdgeModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing edges: %w", err)
		}
		if err := tx.Where("app_id = ?", app.ID).Delete(&NodeModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing nodes: %w", err)
		}

		for _, node := range g.Nodes {
			nodeModel, err := r.nodeToModel(node, app.ID)
			if err != nil {
				return fmt.Errorf("failed to convert node to model: %w", err)
			}
			if err := tx.Create(&nodeModel).Error; err != nil {
				return fmt.Errorf("failed to save node %s: %w", node.ID, err)
			}
		}

		for _, edge := range g.Edges {
			edgeModel, err := r.edgeToModel(edge, app.ID)
			if err != nil {
				return fmt.Errorf("failed to convert edge to model: %w", err)
			}
			if err := tx.Create(&edgeModel).Error; err != nil {
				return fmt.Errorf("failed to save edge %s: %w", edge.ID, err)
			}
		}

		return nil
	})
}

func (r *Repository) LoadGraph(appName string) (*graph.Graph, error) {
	var app App
	err := r.db.Where("name = ?", appName).First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("app %s not found", appName)
		}
		return nil, fmt.Errorf("failed to find app: %w", err)
	}

	var nodeModels []NodeModel
	if err := r.db.Where("app_id = ?", app.ID).Find(&nodeModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load nodes: %w", err)
	}

	var edgeModels []EdgeModel
	if err := r.db.Where("app_id = ?", app.ID).Find(&edgeModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	g := graph.NewGraph(appName)
	g.ID = fmt.Sprintf("%s-graph", app.ID)

	for _, nodeModel := range nodeModels {
		node, err := r.modelToNode(&nodeModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node model: %w", err)
		}
		if err := g.AddNode(node); err != nil {
			return nil, fmt.Errorf("failed to add node to graph: %w", err)
		}
	}

	for _, edgeModel := range edgeModels {
		edge, err := r.modelToEdge(&edgeModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert edge model: %w", err)
		}
		if err := g.AddEdge(edge); err != nil {
			return nil, fmt.Errorf("failed to add edge to graph: %w", err)
		}
	}

	return g, nil
}

func (r *Repository) CreateGraphRun(appName string, version int) (*GraphRunModel, error) {
	var app App
	err := r.db.Where("name = ?", appName).First(&app).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find app: %w", err)
	}

	graphRun := &GraphRunModel{
		AppID:   app.ID,
		Version: version,
		Status:  "pending",
	}

	if err := r.db.Create(graphRun).Error; err != nil {
		return nil, fmt.Errorf("failed to create graph run: %w", err)
	}

	return graphRun, nil
}

func (r *Repository) UpdateGraphRun(runID uuid.UUID, status string, errorMessage *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == "completed" || status == "failed" {
		updates["completed_at"] = "NOW()"
	}

	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	}

	return r.db.Model(&GraphRunModel{}).Where("id = ?", runID).Updates(updates).Error
}

func (r *Repository) GetGraphRuns(appName string) ([]GraphRunModel, error) {
	var app App
	err := r.db.Where("name = ?", appName).First(&app).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find app: %w", err)
	}

	var runs []GraphRunModel
	err = r.db.Where("app_id = ?", app.ID).Order("started_at DESC").Find(&runs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load graph runs: %w", err)
	}

	return runs, nil
}

func (r *Repository) nodeToModel(node *graph.Node, appID uuid.UUID) (*NodeModel, error) {
	propertiesJSON, err := json.Marshal(node.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node properties: %w", err)
	}

	return &NodeModel{
		ID:          node.ID,
		AppID:       appID,
		Type:        string(node.Type),
		Name:        node.Name,
		Description: node.Description,
		Properties:  string(propertiesJSON),
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
	}, nil
}

func (r *Repository) modelToNode(model *NodeModel) (*graph.Node, error) {
	var properties map[string]interface{}
	if model.Properties != "" {
		if err := json.Unmarshal([]byte(model.Properties), &properties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node properties: %w", err)
		}
	}

	return &graph.Node{
		ID:          model.ID,
		Type:        graph.NodeType(model.Type),
		Name:        model.Name,
		Description: model.Description,
		Properties:  properties,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (r *Repository) edgeToModel(edge *graph.Edge, appID uuid.UUID) (*EdgeModel, error) {
	propertiesJSON, err := json.Marshal(edge.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal edge properties: %w", err)
	}

	return &EdgeModel{
		ID:          edge.ID,
		AppID:       appID,
		FromNodeID:  edge.FromNodeID,
		ToNodeID:    edge.ToNodeID,
		Type:        string(edge.Type),
		Description: edge.Description,
		Properties:  string(propertiesJSON),
		CreatedAt:   edge.CreatedAt,
	}, nil
}

func (r *Repository) modelToEdge(model *EdgeModel) (*graph.Edge, error) {
	var properties map[string]interface{}
	if model.Properties != "" {
		if err := json.Unmarshal([]byte(model.Properties), &properties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal edge properties: %w", err)
		}
	}

	return &graph.Edge{
		ID:          model.ID,
		FromNodeID:  model.FromNodeID,
		ToNodeID:    model.ToNodeID,
		Type:        graph.EdgeType(model.Type),
		Description: model.Description,
		Properties:  properties,
		CreatedAt:   model.CreatedAt,
	}, nil
}