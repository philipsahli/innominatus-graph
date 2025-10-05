package storage

import (
	"github.com/philipsahli/innominatus-graph/pkg/graph"

	"github.com/google/uuid"
)

type RepositoryInterface interface {
	SaveGraph(appName string, g *graph.Graph) error
	LoadGraph(appName string) (*graph.Graph, error)
	CreateGraphRun(appName string, version int) (*GraphRunModel, error)
	UpdateGraphRun(runID uuid.UUID, status string, errorMessage *string) error
	GetGraphRuns(appName string) ([]GraphRunModel, error)
	UpdateNodeState(appName string, nodeID string, state graph.NodeState) error
}
