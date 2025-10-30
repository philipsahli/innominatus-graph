# API Reference

> Detailed API documentation. For quick start, see `README.md`.

## Core Packages

### `pkg/graph` - Graph Model

**Types:**
- `Graph`: Directed graph with nodes and edges
- `Node`: Graph node (spec/workflow/step/resource)
- `Edge`: Graph edge (depends-on/provisions/creates/binds-to/contains/configures)
- `NodeType`: Type of node
- `EdgeType`: Type of edge
- `NodeState`: Node state (waiting/pending/running/failed/succeeded)

**Key Methods:**
```go
func NewGraph(appName string) *Graph
func (g *Graph) AddNode(node *Node) error
func (g *Graph) AddEdge(edge *Edge) error
func (g *Graph) UpdateNodeState(nodeID string, newState NodeState) error
func (g *Graph) GetNode(id string) (*Node, bool)
func (g *Graph) GetNodesByType(nodeType NodeType) []*Node
func (g *Graph) GetNodesByState(state NodeState) []*Node
func (g *Graph) GetChildSteps(workflowID string) []*Node
func (g *Graph) GetParentWorkflow(stepID string) (*Node, error)
```

### `pkg/storage` - Persistence

**Interfaces:**
```go
type RepositoryInterface interface {
    SaveGraph(appName string, g *graph.Graph) error
    LoadGraph(appName string) (*graph.Graph, error)
    UpdateNodeState(appName, nodeID string, state graph.NodeState) error
}
```

**Functions:**
```go
func NewSQLiteConnection(dbPath string) (*gorm.DB, error)
func NewPostgresConnection(host, user, password, dbname, sslmode string, port int) (*gorm.DB, error)
func NewConnection(config Config) (*gorm.DB, error)
func NewRepository(db *gorm.DB) *Repository
func AutoMigrate(db *gorm.DB) error
```

### `pkg/export` - Visualization

**Types:**
```go
type Format string
const (
    FormatDOT Format = "dot"
    FormatSVG Format = "svg"
    FormatPNG Format = "png"
)
```

**Functions:**
```go
func NewExporter() *Exporter
func (e *Exporter) ExportGraph(g *graph.Graph, format Format) ([]byte, error)
func (e *Exporter) Close() error
```

### `pkg/execution` - Execution Engine

**Interfaces:**
```go
type ExecutionObserver interface {
    OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState)
}

type WorkflowRunner interface {
    RunWorkflow(workflow *graph.Node) error
}
```

**Functions:**
```go
func NewEngine(repo storage.RepositoryInterface, runner WorkflowRunner) *Engine
func (e *Engine) RegisterObserver(observer ExecutionObserver)
func (e *Engine) ExecuteGraph(appName string) ([]*graph.Node, error)
```

## Node Types

| Type | Description | Use Case |
|------|-------------|----------|
| `spec` | Configuration specification | Score specs, manifests |
| `workflow` | Multi-step orchestration process | Deployment workflows |
| `step` | Individual workflow step | Terraform apply, kubectl deploy |
| `resource` | Infrastructure or application resource | Database, K8s deployment |

## Edge Types

| Type | From → To | Description |
|------|-----------|-------------|
| `depends-on` | Any → Any | Generic dependency |
| `provisions` | Workflow → Resource | Workflow provisions resource |
| `creates` | Workflow → Any | Workflow creates node |
| `binds-to` | Any → Resource | Bind to existing resource |
| `contains` | Workflow → Step | Workflow contains step |
| `configures` | Step → Resource | Step configures resource |

## Node States

| State | Description |
|-------|-------------|
| `waiting` | Initial state, not yet ready |
| `pending` | Ready to execute |
| `running` | Currently executing |
| `failed` | Execution failed |
| `succeeded` | Execution completed successfully |

## State Propagation

**Upward Propagation:**
- When a `step` transitions to `failed`, its parent `workflow` automatically transitions to `failed`

**Downward Propagation:**
- When a `workflow` completes (`failed` or `succeeded`), all running child `steps` inherit the workflow state

## Usage Examples

See original README.md content for detailed examples (moved here to keep README minimal).

### Creating a Graph
```go
g := graph.NewGraph("my-app")

workflow := &graph.Node{
    ID:   "deploy-workflow",
    Type: graph.NodeTypeWorkflow,
    Name: "Deploy Application",
}
g.AddNode(workflow)

step1 := &graph.Node{
    ID:   "provision-infra",
    Type: graph.NodeTypeStep,
    Name: "Provision Infrastructure",
}
g.AddNode(step1)

g.AddEdge(&graph.Edge{
    ID:         "workflow-step1",
    FromNodeID: "deploy-workflow",
    ToNodeID:   "provision-infra",
    Type:       graph.EdgeTypeContains,
})
```

### State Management
```go
err := g.UpdateNodeState("provision-infra", graph.NodeStateRunning)
err = g.UpdateNodeState("provision-infra", graph.NodeStateFailed)

// Check parent workflow state
workflow, _ := g.GetNode("deploy-workflow")
fmt.Println(workflow.State) // Output: failed
```

### Persistence - SQLite
```go
db, _ := storage.NewSQLiteConnection("graph.db")
storage.AutoMigrate(db)
repo := storage.NewRepository(db)

// Save
repo.SaveGraph("my-app", g)

// Load
loadedGraph, _ := repo.LoadGraph("my-app")
```

### Persistence - PostgreSQL
```go
db, _ := storage.NewPostgresConnection(
    "localhost", "postgres", "secret", "idp_orchestrator", "disable", 5432,
)
repo := storage.NewRepository(db)
repo.SaveGraph("my-app", g)
```

### Graph Export
```go
exporter := export.NewExporter()
defer exporter.Close()

// Export to SVG
svgBytes, _ := exporter.ExportGraph(g, export.FormatSVG)
os.WriteFile("graph.svg", svgBytes, 0644)
```

### Execution with Observer
```go
type MyObserver struct{}

func (o *MyObserver) OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState) {
    fmt.Printf("Node %s: %s → %s\n", node.Name, oldState, newState)
}

runner := execution.NewMockWorkflowRunner()
engine := execution.NewEngine(repo, runner)
engine.RegisterObserver(&MyObserver{})

plan, _ := engine.ExecuteGraph("my-app")
```

## Integration with Innominatus Orchestrator

See original README.md for integration examples (moved here to keep README minimal).

## Development

### Running Tests
```bash
go test ./...
go test -cover ./...
go test ./pkg/graph
```

### Database Schema
Auto-migration creates:
- `apps`: Application metadata
- `nodes`: Graph nodes with type and state
- `edges`: Graph edges with relationship types
- `graph_runs`: Execution history

## Prerequisites
- Go 1.24+
- GraphViz (for SVG/PNG export)
- PostgreSQL 12+ (optional, SQLite works for dev)
