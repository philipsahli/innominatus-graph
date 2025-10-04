# Innominatus-Graph SDK API Reference

## Core Types

### Node Types
```go
type NodeType string

const (
    NodeTypeSpec     NodeType = "spec"     // Configuration specification
    NodeTypeWorkflow NodeType = "workflow" // Multi-step orchestration
    NodeTypeStep     NodeType = "step"     // Individual workflow step
    NodeTypeResource NodeType = "resource" // Infrastructure/app resource
)
```

### Edge Types
```go
type EdgeType string

const (
    EdgeTypeDependsOn  EdgeType = "depends-on"  // Generic dependency
    EdgeTypeProvisions EdgeType = "provisions"  // Workflow → Resource
    EdgeTypeCreates    EdgeType = "creates"     // Workflow → Any
    EdgeTypeBindsTo    EdgeType = "binds-to"    // Any → Resource
    EdgeTypeContains   EdgeType = "contains"    // Workflow → Step
    EdgeTypeConfigures EdgeType = "configures"  // Step → Resource
)
```

### Node States
```go
type NodeState string

const (
    NodeStateWaiting   NodeState = "waiting"   // Initial state
    NodeStatePending   NodeState = "pending"   // Ready to execute
    NodeStateRunning   NodeState = "running"   // Currently executing
    NodeStateFailed    NodeState = "failed"    // Execution failed
    NodeStateSucceeded NodeState = "succeeded" // Execution succeeded
)
```

## Core Structs

### Node
```go
type Node struct {
    ID          string                 `json:"id"`
    Type        NodeType               `json:"type"`
    Name        string                 `json:"name"`
    Description string                 `json:"description,omitempty"`
    State       NodeState              `json:"state"`
    Properties  map[string]interface{} `json:"properties,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

### Edge
```go
type Edge struct {
    ID          string                 `json:"id"`
    FromNodeID  string                 `json:"from_node_id"`
    ToNodeID    string                 `json:"to_node_id"`
    Type        EdgeType               `json:"type"`
    Description string                 `json:"description,omitempty"`
    Properties  map[string]interface{} `json:"properties,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
}
```

### Graph
```go
type Graph struct {
    ID        string           `json:"id"`
    AppName   string           `json:"app_name"`
    Version   int              `json:"version"`
    Nodes     map[string]*Node `json:"nodes"`
    Edges     map[string]*Edge `json:"edges"`
    CreatedAt time.Time        `json:"created_at"`
    UpdatedAt time.Time        `json:"updated_at"`
}
```

## Core Graph Methods

### Graph Construction
```go
// NewGraph creates a new graph
func NewGraph(appName string) *Graph

// AddNode adds a node to the graph
func (g *Graph) AddNode(node *Node) error

// AddEdge adds an edge to the graph (with validation)
func (g *Graph) AddEdge(edge *Edge) error

// RemoveNode removes a node and its edges
func (g *Graph) RemoveNode(id string) error

// RemoveEdge removes an edge
func (g *Graph) RemoveEdge(id string) error
```

### Graph Queries
```go
// GetNode retrieves a node by ID
func (g *Graph) GetNode(id string) (*Node, bool)

// GetEdge retrieves an edge by ID
func (g *Graph) GetEdge(id string) (*Edge, bool)

// GetNodesByType returns all nodes of a specific type
func (g *Graph) GetNodesByType(nodeType NodeType) []*Node

// GetNodesByState returns all nodes in a specific state
func (g *Graph) GetNodesByState(state NodeState) []*Node

// GetChildSteps returns all step nodes contained by a workflow
func (g *Graph) GetChildSteps(workflowID string) []*Node

// GetParentWorkflow returns the parent workflow of a step node
func (g *Graph) GetParentWorkflow(stepID string) (*Node, error)

// GetDependencies returns nodes that a node depends on
func (g *Graph) GetDependencies(nodeID string) ([]*Node, error)

// GetDependents returns nodes that depend on a node
func (g *Graph) GetDependents(nodeID string) ([]*Node, error)
```

### State Management
```go
// UpdateNodeState updates the state of a node (with propagation)
func (g *Graph) UpdateNodeState(nodeID string, newState NodeState) error
```

**State Propagation Rules:**
- When a `step` transitions to `failed` → parent `workflow` transitions to `failed`
- When a `workflow` transitions to `failed` or `succeeded` → running child `steps` inherit the state

### Topological Sort
```go
// TopologicalSort returns nodes in dependency-aware execution order
func (g *Graph) TopologicalSort() ([]*Node, error)

// HasCycle checks if the graph contains cycles
func (g *Graph) HasCycle() bool
```

## Storage Package (pkg/storage)

### Repository Interface
```go
type RepositoryInterface interface {
    SaveGraph(appName string, g *graph.Graph) error
    LoadGraph(appName string) (*graph.Graph, error)
    CreateGraphRun(appName string, version int) (*GraphRunModel, error)
    UpdateGraphRun(runID uuid.UUID, status string, errorMessage *string) error
    GetGraphRuns(appName string) ([]GraphRunModel, error)
    UpdateNodeState(appName string, nodeID string, state graph.NodeState) error
}
```

### Repository Implementation
```go
// NewRepository creates a PostgreSQL-backed repository
func NewRepository(db *gorm.DB) *Repository

// Example usage:
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
repo := storage.NewRepository(db)
```

## Export Package (pkg/export)

### Exporter
```go
type Format string

const (
    FormatDOT Format = "dot"
    FormatSVG Format = "svg"
    FormatPNG Format = "png"
)

// NewExporter creates a new graph exporter
func NewExporter() *Exporter

// Close closes the exporter (frees GraphViz resources)
func (e *Exporter) Close() error

// ExportGraph exports a graph to the specified format
func (e *Exporter) ExportGraph(g *graph.Graph, format Format) ([]byte, error)

// CreateSubgraph creates a subgraph containing only specified nodes
func (e *Exporter) CreateSubgraph(g *graph.Graph, nodeIDs []string) (*graph.Graph, error)
```

### Visual Styling

**Node Colors (by type):**
- Spec: `#E3F2FD` (Light blue)
- Workflow: `#FFF9C4` (Light yellow)
- Step: `#FFE0B2` (Light orange)
- Resource: `#C8E6C9` (Light green)

**Node Border Colors (by state):**
- Failed: `red`
- Running: `#1976D2` (Blue, bold)
- Succeeded: `#388E3C` (Green)
- Default: `black`

**Edge Colors:**
- depends-on: `#1976D2` (Blue)
- provisions: `#388E3C` (Green)
- creates: `#F57C00` (Orange)
- binds-to: `#7B1FA2` (Purple)
- contains: `#FBC02D` (Yellow)
- configures: `#E64A19` (Deep orange)

**Edge Styles:**
- depends-on: solid
- provisions: bold
- creates: dashed
- binds-to: dotted
- contains: bold
- configures: dashed

## Execution Package (pkg/execution)

### ExecutionObserver Interface
```go
type ExecutionObserver interface {
    OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState)
}
```

### WorkflowRunner Interface
```go
type WorkflowRunner interface {
    RunWorkflow(node *graph.Node) error
    ProvisionResource(workflow *graph.Node, resource *graph.Node) error
    CreateResource(workflow *graph.Node, target *graph.Node) error
}
```

### Engine
```go
// NewEngine creates a new execution engine
func NewEngine(repository storage.RepositoryInterface, runner WorkflowRunner) *Engine

// RegisterObserver registers an observer for state change notifications
func (e *Engine) RegisterObserver(observer ExecutionObserver)

// ExecuteGraph executes a graph topologically
func (e *Engine) ExecuteGraph(appName string) (*ExecutionPlan, error)
```

### Execution Types
```go
type ExecutionStatus string

const (
    StatusPending   ExecutionStatus = "pending"
    StatusRunning   ExecutionStatus = "running"
    StatusCompleted ExecutionStatus = "completed"
    StatusFailed    ExecutionStatus = "failed"
    StatusSkipped   ExecutionStatus = "skipped"
)

type NodeExecution struct {
    NodeID    string          `json:"node_id"`
    Status    ExecutionStatus `json:"status"`
    StartTime *time.Time      `json:"start_time,omitempty"`
    EndTime   *time.Time      `json:"end_time,omitempty"`
    Error     string          `json:"error,omitempty"`
    Logs      []string        `json:"logs,omitempty"`
}

type ExecutionPlan struct {
    RunID      uuid.UUID                `json:"run_id"`
    AppName    string                   `json:"app_name"`
    Version    int                      `json:"version"`
    Status     ExecutionStatus          `json:"status"`
    StartTime  time.Time                `json:"start_time"`
    EndTime    *time.Time               `json:"end_time,omitempty"`
    Executions map[string]*NodeExecution `json:"executions"`
    Order      []*graph.Node            `json:"order"`
}
```

### Mock Implementation
```go
// NewMockWorkflowRunner creates a mock runner for testing
func NewMockWorkflowRunner() WorkflowRunner
```

## Edge Validation Rules

| Edge Type | From Node Type | To Node Type | Description |
|-----------|----------------|--------------|-------------|
| depends-on | Any | Any | Generic dependency |
| provisions | Workflow | Resource | Workflow provisions resource |
| creates | Workflow | Any | Workflow creates node |
| binds-to | Any | Resource | Bind to existing resource |
| contains | Workflow | Step | Workflow contains step |
| configures | Step | Resource | Step configures resource |

**Validation Errors:**
- ❌ `provisions` from non-workflow node
- ❌ `provisions` to non-resource node
- ❌ `creates` from non-workflow node
- ❌ `binds-to` to non-resource node
- ❌ `contains` from non-workflow node
- ❌ `contains` to non-step node
- ❌ `configures` from non-step node
- ❌ `configures` to non-resource node

## Usage Examples

### 1. Basic Graph Creation
```go
import "github.com/innominatus/innominatus-graph/pkg/graph"

g := graph.NewGraph("my-app")

workflow := &graph.Node{
    ID:   "deploy",
    Type: graph.NodeTypeWorkflow,
    Name: "Deploy Application",
}
g.AddNode(workflow)

step := &graph.Node{
    ID:   "provision",
    Type: graph.NodeTypeStep,
    Name: "Provision Infrastructure",
}
g.AddNode(step)

g.AddEdge(&graph.Edge{
    ID:         "wf-step",
    FromNodeID: "deploy",
    ToNodeID:   "provision",
    Type:       graph.EdgeTypeContains,
})
```

### 2. State Management
```go
// Update state with automatic propagation
g.UpdateNodeState("provision", graph.NodeStateRunning)
g.UpdateNodeState("provision", graph.NodeStateFailed)

// Parent workflow automatically becomes failed
workflow, _ := g.GetNode("deploy")
fmt.Println(workflow.State) // Output: failed
```

### 3. Persistence
```go
import "github.com/innominatus/innominatus-graph/pkg/storage"

repo := storage.NewRepository(db)
repo.SaveGraph("my-app", g)

loadedGraph, _ := repo.LoadGraph("my-app")
repo.UpdateNodeState("my-app", "provision", graph.NodeStateSucceeded)
```

### 4. Export
```go
import "github.com/innominatus/innominatus-graph/pkg/export"

exporter := export.NewExporter()
defer exporter.Close()

svgBytes, _ := exporter.ExportGraph(g, export.FormatSVG)
os.WriteFile("graph.svg", svgBytes, 0644)
```

### 5. Execution with Observer
```go
import "github.com/innominatus/innominatus-graph/pkg/execution"

type MyObserver struct{}

func (o *MyObserver) OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState) {
    fmt.Printf("Node %s: %s → %s\n", node.Name, oldState, newState)
}

runner := execution.NewMockWorkflowRunner()
engine := execution.NewEngine(repo, runner)
engine.RegisterObserver(&MyObserver{})

plan, _ := engine.ExecuteGraph("my-app")
```

## Error Handling

All methods that can fail return `error` as the last return value:

```go
if err := g.AddNode(node); err != nil {
    // Handle error (e.g., duplicate ID, nil node, etc.)
}

if err := g.AddEdge(edge); err != nil {
    // Handle error (e.g., validation failure, missing nodes, etc.)
}

nodes, err := g.TopologicalSort()
if err != nil {
    // Handle error (e.g., cycle detected)
}
```

## Thread Safety

**Note:** The graph implementation is **not thread-safe**. If you need concurrent access, use external synchronization (e.g., `sync.RWMutex`).

---

For complete examples, see:
- `examples/demo/main.go`: Full integration example
- `README.md`: Quick start guide
- Test files in `pkg/*/`: Usage patterns
