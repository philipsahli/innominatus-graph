# Innominatus Graph SDK - Claude Code Configuration

## Project Overview

**Innominatus Graph SDK** is a Go library for representing Internal Developer Platform (IDP) workflows as directed acyclic graphs (DAGs) with state management, persistence, and visualization. Designed for integration into IDP orchestrators like [innominatus](https://github.com/innominatus/innominatus).

### Core Purpose
Provide a robust graph model for tracking multi-step workflows, their execution state, and relationships to infrastructure resources.

## Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL (production) + SQLite (development)
  - ORM: GORM with auto-migration
- **Visualization**: GraphViz (DOT/SVG/PNG export)
- **Testing**: `testing` package + `testify/assert`
- **CLI Framework**: Cobra (deprecated, in `deprecated/`)
- **GraphQL**: gqlgen (deprecated, in `deprecated/`)
- **HTTP**: Gin framework (deprecated, in `deprecated/`)

## Key Commands

```bash
# Install dependencies
go mod download

# Run tests (must maintain >80% coverage)
go test ./...
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run specific package tests
go test ./pkg/graph
go test ./pkg/storage
go test ./pkg/export
go test ./pkg/execution

# Format code (run before committing)
gofmt -w .

# Lint and vet
go vet ./...
golint ./...

# Run demo example
cd examples/demo
go run main.go                        # SQLite mode
DB_PASSWORD=secret go run main.go     # PostgreSQL mode
USE_SQLITE=1 go run main.go           # Explicit SQLite

# Build (if needed)
go build -o bin/demo examples/demo/main.go

# Dependency audit
go mod tidy
go mod verify
```

## Architecture Rules

### SOLID Principles (Non-Negotiable)

#### Single Responsibility Principle (SRP)
- One struct/interface = one reason to change
- `Graph` handles graph operations only
- `Repository` handles persistence only
- `Exporter` handles visualization only
- `Engine` handles execution orchestration only

#### Open/Closed Principle (OCP)
- Interfaces for extension points: `RepositoryInterface`, `ExecutionObserver`, `WorkflowRunner`
- New storage backends via `RepositoryInterface` implementation
- New execution logic via `WorkflowRunner` implementation
- Do NOT modify existing interfaces; extend with new ones

#### Liskov Substitution Principle (LSP)
- All `RepositoryInterface` implementations must be interchangeable
- SQLite and PostgreSQL repositories behave identically from consumer perspective
- Mock implementations for testing must honor real interface contracts

#### Interface Segregation Principle (ISP)
- Small, focused interfaces over large ones
- `ExecutionObserver` has single method: `OnNodeStateChange`
- `RepositoryInterface` has focused methods: Save, Load, UpdateState
- Clients depend only on methods they use

#### Dependency Inversion Principle (DIP)
- Depend on abstractions (`RepositoryInterface`), not concrete implementations
- `Engine` depends on `RepositoryInterface`, not `Repository`
- Inject dependencies via constructors (e.g., `NewEngine(repo, runner)`)

### KISS (Keep It Simple, Stupid)

- **Favor simple solutions over clever ones**
  - Direct field access over complex getters/setters
  - Simple loops over abstract iterators
  - Explicit error handling over panic recovery

- **Write code that's easy to understand**
  - Clear variable names: `workflowNode` not `wn`
  - Self-documenting function names: `GetChildSteps()` not `GCS()`
  - Prefer flat code over deep nesting

- **Avoid premature optimization**
  - Measure before optimizing
  - Readable code first, fast code second
  - Use profiling (`go test -bench`, `pprof`) to find real bottlenecks

- **If it's complex, it's probably wrong - simplify**
  - Max 3 levels of nesting
  - Functions under 50 lines (ideally under 30)
  - If function needs extensive comments, break it down

- **Prefer explicit over implicit**
  - Explicit error returns over hidden panics
  - Explicit state transitions over magic behavior
  - Explicit dependencies in function signatures

### YAGNI (You Aren't Gonna Need It)

- **Build only what is needed NOW**
  - No "future-proof" abstractions
  - No "just in case" features
  - No "might need later" configuration options

- **Don't add features not explicitly required**
  - Current scope: Graph model, persistence, visualization, execution
  - Do NOT add: REST API (deprecated), GraphQL server (deprecated), CLI tools (deprecated)

- **Avoid over-engineering**
  - Use map lookup before building index structures
  - Use slices before custom collections
  - Use standard library before external packages

- **Refactor when needed, not preemptively**
  - Wait for 3rd use case before abstracting (Rule of Three)
  - Extract interfaces only when multiple implementations exist
  - Generalize when pattern emerges, not before

### Minimal Documentation Philosophy

#### Code Comments
- **Explain "WHY", never "WHAT"**
  ```go
  // GOOD: Propagate failure upward to prevent orphaned running steps
  if node.Type == NodeTypeStep && newState == NodeStateFailed {
      g.propagateFailureToParent(nodeID)
  }

  // BAD: Update node state to failed
  node.State = NodeStateFailed
  ```

- **No redundant comments**
  ```go
  // BAD: Get node by ID
  func (g *Graph) GetNode(id string) (*Node, bool)

  // GOOD: No comment needed - function name is clear
  func (g *Graph) GetNode(id string) (*Node, bool)
  ```

- **Self-documenting code through naming**
  ```go
  // GOOD
  func (g *Graph) GetChildSteps(workflowID string) []*Node

  // BAD - needs comment to explain
  func (g *Graph) Get(id string) []*Node // Gets child steps of workflow
  ```

#### Documentation Files
- **README.md**: 15-20 lines max (what it is, install, quick start)
- **No verbose sections**: Contributing, Badges, Screenshots (unless critical)
- **"Telegram style"**: Ultra-short, concise, actionable
- **API docs**: Use Go doc comments for public APIs, not separate markdown

#### Go Doc Comments
- Public functions/types MUST have doc comments (enforced by `golint`)
- Doc comments start with the name of the element
  ```go
  // NewGraph creates a new graph instance for the given application name.
  func NewGraph(appName string) *Graph
  ```
- Keep doc comments to 1-2 sentences max

### Go-Specific Best Practices

#### Composition Over Inheritance
```go
// GOOD: Composition
type Engine struct {
    repo   RepositoryInterface
    runner WorkflowRunner
}

// BAD: No inheritance in Go - use interfaces
```

#### Interface-Based Design
- Define interfaces at consumer side (not provider side)
- Small, focused interfaces (1-3 methods ideal)
- Accept interfaces, return concrete types
  ```go
  // GOOD
  func NewEngine(repo RepositoryInterface, runner WorkflowRunner) *Engine

  // BAD
  func NewEngine(repo *Repository, runner *MockRunner) *Engine
  ```

#### Error Handling
- Always check errors explicitly
- Return errors, don't panic (except for programming errors)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use custom error types for domain errors

#### Zero Values
- Design structs so zero value is useful
- Use `NewX()` constructors when initialization logic needed
- Example: `Graph.Nodes` initialized in `NewGraph()`, not on first use

#### Pointer vs Value Receivers
- Use pointer receivers when:
  - Method modifies the receiver
  - Receiver is large struct
  - Consistency with other methods
- Use value receivers for small, immutable types

#### Package Structure
```
pkg/
  graph/         # Core graph model
  storage/       # Persistence layer
  export/        # Visualization
  execution/     # Execution engine
```
- Each package has single, clear responsibility
- Avoid circular dependencies
- Internal implementation details in `internal/`

#### Testing
- Test files alongside source: `types_test.go` next to `types.go`
- Table-driven tests for multiple cases
- Use `testify/assert` for readable assertions
- Mock external dependencies via interfaces

### Dependency Management

- **Only install necessary packages**
  - Current dependencies are minimal and justified
  - Question every new dependency: "Can standard library do this?"

- **Avoid large frameworks when simple libraries suffice**
  - GORM chosen for multi-DB abstraction (justified)
  - GraphViz wrapper for visualization (justified)
  - Cobra/Gin moved to `deprecated/` (no longer needed)

- **Regularly audit unused dependencies**
  ```bash
  go mod tidy
  go mod why <package>  # Understand why package is needed
  ```

### Code Quality Standards

- **Test Coverage**: Minimum 80% coverage
  - Run: `go test -cover ./...`
  - Critical paths (state propagation, validation) must have 100% coverage

- **Linting**: Zero warnings from `go vet` and `golint`
  ```bash
  go vet ./...
  golint ./...
  ```

- **Formatting**: All code must be `gofmt` formatted
  ```bash
  gofmt -w .
  ```

- **Naming Conventions**:
  - Exported: `PascalCase` (e.g., `NodeType`, `GetNode`)
  - Unexported: `camelCase` (e.g., `validateEdge`, `propagateFailure`)
  - Acronyms: `ID`, `URL`, `HTTP` (not `Id`, `Url`, `Http`)

- **Function Length**: Max 50 lines (ideally under 30)
- **Cyclomatic Complexity**: Max 10 per function
- **Nesting Depth**: Max 3 levels

## Domain Structure

### Core Domain: Graph Model (`pkg/graph`)
- **Entities**: `Node`, `Edge`, `Graph`
- **Value Objects**: `NodeType`, `EdgeType`, `NodeState`
- **Operations**: Add, Remove, Query, State Management
- **Validation**: Edge type validation, cycle detection

### Persistence Domain (`pkg/storage`)
- **Interfaces**: `RepositoryInterface`
- **Implementations**: PostgreSQL (`Repository`), SQLite (via `Repository`)
- **Models**: GORM models for persistence
- **Connection Management**: Database connection factories

### Visualization Domain (`pkg/export`)
- **Exporter**: DOT/SVG/PNG export
- **Rendering**: State-based colors, edge styles

### Execution Domain (`pkg/execution`)
- **Engine**: Workflow execution orchestration
- **Observer Pattern**: `ExecutionObserver` for state change notifications
- **Runner Interface**: `WorkflowRunner` for custom execution logic

## Current Focus

### SDK Library Maturity
- **Stability**: Core graph model is stable
- **Documentation**: API docs via Go doc comments
- **Examples**: `examples/demo/` demonstrates all features
- **Testing**: Maintain >80% coverage

### Recent Completions
- ✅ Refactored from monolith to SDK library
- ✅ SQLite support for development
- ✅ State propagation (step → workflow)
- ✅ Observer pattern for execution monitoring
- ✅ Demo example with full workflow

### Next Steps
- Integration into innominatus orchestrator
- Performance benchmarking for large graphs (1000+ nodes)
- Additional storage backends if needed (Redis, etcd)

## Verification Protocol

Every feature MUST include a verification script that:

1. **Runs actual code** (not just unit tests)
2. **Captures real outputs** (e.g., graph exports, database records)
3. **Prints clear pass/fail results**
4. **Saves artifacts** to `docs/verification/`

### Verification Template (Go)

```go
package main

import (
    "fmt"
    "os"
    "github.com/philipsahli/innominatus-graph/pkg/graph"
)

func main() {
    fmt.Println("=== Feature Verification: [Feature Name] ===")

    // Setup
    g := graph.NewGraph("verify-test")

    // Execute feature
    node := &graph.Node{
        ID:   "test-node",
        Type: graph.NodeTypeWorkflow,
        Name: "Test Workflow",
    }
    err := g.AddNode(node)

    // Verify results
    if err != nil {
        fmt.Printf("❌ FAILED: %v\n", err)
        os.Exit(1)
    }

    retrievedNode, exists := g.GetNode("test-node")
    if !exists || retrievedNode.Name != "Test Workflow" {
        fmt.Println("❌ FAILED: Node not found or incorrect")
        os.Exit(1)
    }

    // Save artifacts
    // (e.g., export graph, save logs, etc.)

    fmt.Println("✅ PASSED: All checks succeeded")

    // Output structured data for AI evaluation
    fmt.Printf(`
VERIFICATION_RESULT:
  feature: [Feature Name]
  status: passed
  checks: 2
  artifacts:
    - docs/verification/[feature-name].json
`)
}
```

### Running Verifications
```bash
go run verification/verify-[feature-name].go
```

## Integration Examples

### Using in Orchestrator

```go
import (
    "github.com/philipsahli/innominatus-graph/pkg/graph"
    "github.com/philipsahli/innominatus-graph/pkg/storage"
)

// In your orchestrator
db, _ := storage.NewSQLiteConnection("orchestrator.db")
storage.AutoMigrate(db)
repo := storage.NewRepository(db)

// Build graph
g := graph.NewGraph("my-app")
// ... add nodes and edges ...

// Persist
repo.SaveGraph("my-app", g)

// Execute and track state
g.UpdateNodeState("step-1", graph.NodeStateRunning)
```

## Common Patterns

### Adding New Node Types
1. Add constant to `NodeType` in `pkg/graph/types.go`
2. Update validation logic in `validateEdge()`
3. Add export styling in `pkg/export/dot.go`
4. Write verification script
5. Update tests

### Adding New Edge Types
1. Add constant to `EdgeType` in `pkg/graph/types.go`
2. Add validation case in `validateEdge()`
3. Add export styling in `pkg/export/dot.go`
4. Write verification script
5. Update tests

### Adding New Storage Backend
1. Implement `RepositoryInterface` in `pkg/storage/`
2. Add connection factory function
3. Update `NewConnection()` to support new type
4. Write integration tests
5. Write verification script

## Anti-Patterns to Avoid

### ❌ Don't Do This
```go
// Swallowing errors
node, _ := g.GetNode(id) // Ignoring existence check

// Magic numbers
if len(steps) > 5 { } // What does 5 mean?

// Mutating input parameters
func ProcessGraph(g *Graph) {
    g.Version++ // Unexpected side effect
}

// Generic variable names
for _, n := range nodes { // What is 'n'?
    ...
}
```

### ✅ Do This Instead
```go
// Explicit error handling
node, exists := g.GetNode(id)
if !exists {
    return fmt.Errorf("node %s not found", id)
}

// Named constants
const maxConcurrentSteps = 5
if len(steps) > maxConcurrentSteps { }

// Pure functions or explicit mutation
func IncrementVersion(g *Graph) *Graph {
    // Clear from name that mutation occurs
    g.Version++
    return g
}

// Descriptive variable names
for _, workflowNode := range workflowNodes {
    ...
}
```

## Performance Considerations

- Graph operations are O(n) for most queries
- Use `GetNodesByType()` and cache results if querying multiple times
- State updates trigger propagation - consider batch updates if needed
- Database persistence: Use transactions for atomic multi-graph updates

## Questions to Ask Before Coding

1. **Is this needed NOW?** (YAGNI)
2. **Can I make this simpler?** (KISS)
3. **Does this violate single responsibility?** (SOLID)
4. **Am I adding a dependency I don't need?** (Minimal Dependencies)
5. **Does this need a comment, or can I make the code clearer?** (Self-Documenting Code)
6. **Have I written a verification script?** (Verification-First)

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [SOLID Principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design)
