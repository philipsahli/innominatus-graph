# Innominatus Graph SDK

A Go SDK library for representing Internal Developer Platform (IDP) workflows as directed acyclic graphs (DAGs) with state management, persistence, and visualization.

**Designed for integration** into IDP orchestrators like [innominatus](https://github.com/innominatus/innominatus), this SDK provides the core graph model for tracking multi-step workflows, their execution state, and relationships to infrastructure resources.

## Features

### Core Graph Model
- **Node Types**: `spec`, `workflow`, `step`, `resource`
- **Edge Types**: `depends-on`, `provisions`, `creates`, `binds-to`, `contains`, `configures`
- **State Management**: Node states (`waiting`, `pending`, `running`, `failed`, `succeeded`)
- **State Propagation**: Automatic upward propagation (step failure → workflow failure)
- **Cycle Detection**: Topological sorting with comprehensive validation

### Persistence
- **PostgreSQL Backend**: GORM-based repository with transaction support
- **Repository Interface**: Pluggable backend support for future extensions
- **State Tracking**: Persistent node state with timestamp tracking

### Visualization
- **Export Formats**: DOT, SVG, PNG via GraphViz integration
- **State-Based Styling**:
  - Node colors by type (spec: blue, workflow: yellow, step: orange, resource: green)
  - Border colors by state (failed: red, running: blue, succeeded: green)
  - Edge styles by relationship type

### Execution
- **Topological Traversal**: Dependency-aware execution planning
- **Observer Pattern**: Real-time state change notifications via `ExecutionObserver`
- **Workflow Engine**: Extensible execution with custom runners

## Installation

```bash
go get github.com/innominatus/innominatus-graph
```

## Quick Start

### 1. Creating a Graph

```go
import "github.com/innominatus/innominatus-graph/pkg/graph"

// Create a new graph
g := graph.NewGraph("my-app")

// Add a workflow node
workflow := &graph.Node{
    ID:   "deploy-workflow",
    Type: graph.NodeTypeWorkflow,
    Name: "Deploy Application",
}
g.AddNode(workflow)

// Add step nodes
step1 := &graph.Node{
    ID:   "provision-infra",
    Type: graph.NodeTypeStep,
    Name: "Provision Infrastructure",
}
g.AddNode(step1)

step2 := &graph.Node{
    ID:   "deploy-app",
    Type: graph.NodeTypeStep,
    Name: "Deploy Application",
}
g.AddNode(step2)

// Add resource node
db := &graph.Node{
    ID:   "postgres-db",
    Type: graph.NodeTypeResource,
    Name: "PostgreSQL Database",
}
g.AddNode(db)

// Connect workflow → steps (contains)
g.AddEdge(&graph.Edge{
    ID:         "workflow-step1",
    FromNodeID: "deploy-workflow",
    ToNodeID:   "provision-infra",
    Type:       graph.EdgeTypeContains,
})

g.AddEdge(&graph.Edge{
    ID:         "workflow-step2",
    FromNodeID: "deploy-workflow",
    ToNodeID:   "deploy-app",
    Type:       graph.EdgeTypeContains,
})

// Connect step → resource (configures)
g.AddEdge(&graph.Edge{
    ID:         "step1-db",
    FromNodeID: "provision-infra",
    ToNodeID:   "postgres-db",
    Type:       graph.EdgeTypeConfigures,
})

// Add step dependency
g.AddEdge(&graph.Edge{
    ID:         "step2-depends-step1",
    FromNodeID: "deploy-app",
    ToNodeID:   "provision-infra",
    Type:       graph.EdgeTypeDependsOn,
})
```

### 2. State Management

```go
// Update node state (with automatic propagation)
err := g.UpdateNodeState("provision-infra", graph.NodeStateRunning)
err = g.UpdateNodeState("provision-infra", graph.NodeStateFailed)

// When a step fails, parent workflow automatically transitions to failed
workflow, _ := g.GetNode("deploy-workflow")
fmt.Println(workflow.State) // Output: failed

// Query nodes by state
runningNodes := g.GetNodesByState(graph.NodeStateRunning)
failedNodes := g.GetNodesByState(graph.NodeStateFailed)

// Query nodes by type
steps := g.GetNodesByType(graph.NodeTypeStep)
workflows := g.GetNodesByType(graph.NodeTypeWorkflow)

// Get parent/child relationships
childSteps := g.GetChildSteps("deploy-workflow")
parentWorkflow, _ := g.GetParentWorkflow("provision-infra")
```

### 3. Persistence

```go
import (
    "github.com/innominatus/innominatus-graph/pkg/storage"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Connect to PostgreSQL
dsn := "host=localhost user=postgres password=secret dbname=idp_orchestrator port=5432"
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// Create repository
repo := storage.NewRepository(db)

// Save graph
err := repo.SaveGraph("my-app", g)

// Load graph
loadedGraph, err := repo.LoadGraph("my-app")

// Update node state in database
err = repo.UpdateNodeState("my-app", "provision-infra", graph.NodeStateSucceeded)
```

### 4. Graph Export

```go
import "github.com/innominatus/innominatus-graph/pkg/export"

// Create exporter
exporter := export.NewExporter()
defer exporter.Close()

// Export to DOT format
dotBytes, _ := exporter.ExportGraph(g, export.FormatDOT)
os.WriteFile("graph.dot", dotBytes, 0644)

// Export to SVG (with state-based colors)
svgBytes, _ := exporter.ExportGraph(g, export.FormatSVG)
os.WriteFile("graph.svg", svgBytes, 0644)

// Export to PNG
pngBytes, _ := exporter.ExportGraph(g, export.FormatPNG)
os.WriteFile("graph.png", pngBytes, 0644)
```

### 5. Execution with Observer

```go
import "github.com/innominatus/innominatus-graph/pkg/execution"

// Implement ExecutionObserver
type MyObserver struct{}

func (o *MyObserver) OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState) {
    fmt.Printf("Node %s: %s → %s\n", node.Name, oldState, newState)
}

// Create execution engine
runner := execution.NewMockWorkflowRunner() // or your custom runner
engine := execution.NewEngine(repo, runner)

// Register observer for real-time state tracking
observer := &MyObserver{}
engine.RegisterObserver(observer)

// Execute workflow (triggers state change notifications)
plan, err := engine.ExecuteGraph("my-app")
```

## Graph Model

### Node Types

| Type | Description | Use Case |
|------|-------------|----------|
| `spec` | Configuration specification | Score specs, manifests |
| `workflow` | Multi-step orchestration process | Deployment workflows |
| `step` | Individual workflow step | Terraform apply, kubectl deploy |
| `resource` | Infrastructure or application resource | Database, K8s deployment |

### Edge Types

| Type | From → To | Description |
|------|-----------|-------------|
| `depends-on` | Any → Any | Generic dependency |
| `provisions` | Workflow → Resource | Workflow provisions resource |
| `creates` | Workflow → Any | Workflow creates node |
| `binds-to` | Any → Resource | Bind to existing resource |
| `contains` | Workflow → Step | Workflow contains step |
| `configures` | Step → Resource | Step configures resource |

### Node States

| State | Description |
|-------|-------------|
| `waiting` | Initial state, not yet ready |
| `pending` | Ready to execute |
| `running` | Currently executing |
| `failed` | Execution failed |
| `succeeded` | Execution completed successfully |

**State Propagation Rules:**
- When a `step` transitions to `failed`, its parent `workflow` automatically transitions to `failed`
- When a `workflow` completes (`failed` or `succeeded`), all running child `steps` inherit the workflow state

## Integration with Innominatus Orchestrator

This SDK is designed for integration into the [innominatus](https://github.com/innominatus/innominatus) orchestrator:

```go
// In innominatus orchestrator's internal/workflow/executor.go
import (
    graph "github.com/innominatus/innominatus-graph/pkg/graph"
    storage "github.com/innominatus/innominatus-graph/pkg/storage"
)

type WorkflowExecutor struct {
    graphRepo storage.RepositoryInterface
}

func (e *WorkflowExecutor) ExecuteWorkflow(appName string, workflow *Workflow) error {
    // Build graph representation
    g := graph.NewGraph(appName)

    // Add workflow node
    workflowNode := &graph.Node{
        ID:   workflow.ID,
        Type: graph.NodeTypeWorkflow,
        Name: workflow.Name,
    }
    g.AddNode(workflowNode)

    // Add steps
    for _, step := range workflow.Steps {
        stepNode := &graph.Node{
            ID:   step.ID,
            Type: graph.NodeTypeStep,
            Name: step.Name,
        }
        g.AddNode(stepNode)

        // Link workflow → step
        g.AddEdge(&graph.Edge{
            ID:         fmt.Sprintf("%s-%s", workflow.ID, step.ID),
            FromNodeID: workflow.ID,
            ToNodeID:   step.ID,
            Type:       graph.EdgeTypeContains,
        })
    }

    // Persist graph
    return e.graphRepo.SaveGraph(appName, g)
}
```

## Example Application

See [`examples/demo/main.go`](examples/demo/main.go) for a complete working example demonstrating:
- Building a graph with workflow → steps → resources
- State management and propagation
- Parent-child relationships
- Graph export to DOT/SVG
- Database persistence
- Execution with observer

Run the demo:
```bash
cd examples/demo
go run main.go

# With database persistence
DB_PASSWORD=yourpassword go run main.go
```

## Development

### Running Tests

```bash
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./pkg/graph
go test ./pkg/export
```

### Database Schema

Required PostgreSQL tables:
- `apps`: Application metadata
- `nodes`: Graph nodes with type and state
- `edges`: Graph edges with relationship types
- `graph_runs`: Execution history

See `migrations/` for schema definitions.

## API Reference

### Core Packages

- **`pkg/graph`**: Core graph model, validation, state management
- **`pkg/storage`**: PostgreSQL persistence (repository pattern)
- **`pkg/export`**: DOT/SVG/PNG export via GraphViz
- **`pkg/execution`**: Execution engine with observer support

### Key Interfaces

```go
// GraphRepository interface for pluggable backends
type RepositoryInterface interface {
    SaveGraph(appName string, g *graph.Graph) error
    LoadGraph(appName string) (*graph.Graph, error)
    UpdateNodeState(appName, nodeID string, state graph.NodeState) error
}

// ExecutionObserver for state change notifications
type ExecutionObserver interface {
    OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState)
}
```

## Migration from Previous Versions

⚠️ **Breaking Changes in SDK Refactoring:**

Previous versions included standalone CLI and API server. These have been moved to `deprecated/` directory.

**Old usage:**
```bash
./idp-o-ctl graph export --app demo --format svg
```

**New usage:**
```go
import "github.com/innominatus/innominatus-graph/pkg/export"

exporter := export.NewExporter()
svgBytes, _ := exporter.ExportGraph(g, export.FormatSVG)
```

See [`deprecated/README.md`](deprecated/README.md) for full migration guide.

## Prerequisites

- Go 1.22+
- PostgreSQL 12+ (for persistence features)
- GraphViz (for SVG/PNG export)

## License

[Your License Here]

## Contributing

1. Follow Go idioms and conventions
2. Add tests for new functionality
3. Update documentation for API changes
4. Ensure all tests pass: `go test ./...`

---

**Built for IDP Orchestration** | Integrates with [innominatus](https://github.com/innominatus/innominatus)
