# IDP Orchestrator

A graph-based orchestrator for Internal Developer Platforms (IDPs) where everything (specs, workflows, resources) is represented as nodes in a graph, with relationships as edges. The graph serves as the execution, persistence, and visualization model.

## Features

### MVP Deliverables ✅

1. **Graph Model**
   - Nodes: spec, workflow, resource with validation
   - Edges: depends-on, provisions, creates, binds-to with canonical types
   - Methods: AddNode, AddEdge with comprehensive validation
   - Topological sorting and cycle detection

2. **Postgres Persistence**
   - Tables: apps, nodes, edges, graph_runs with proper relationships
   - Functions: SaveGraph, LoadGraph with transaction support
   - Graph versioning and execution history via graph_runs

3. **REST API**
   - `GET /api/v1/graph?app=demo` → returns latest graph JSON
   - `POST /api/v1/graph/export?app=demo&format=svg` → exports DOT/SVG/PNG
   - Graph run management endpoints

4. **GraphQL API**
   - Schema-first approach using gqlgen
   - Queries: `graph(app: String!)`, `node(id: ID!)`
   - GraphQL playground available at `/graphql`

5. **CLI Integration**
   - Command: `idp-o-ctl graph export --app demo --format svg --output demo.svg`
   - Default: DOT to stdout
   - Support for subgraph exports with `--nodes` filter

6. **Graph Export**
   - DOT generation with proper styling and colors
   - SVG/PNG rendering via GraphViz integration
   - Node colors by type, edge styles by relationship

7. **Execution Engine**
   - Topological traversal with dependency resolution
   - Mock workflow runner for MVP demonstration
   - Execution state tracking in database
   - Comprehensive logging and error handling

8. **Tests & Mocks**
   - Unit tests for all core components
   - Mock implementations for external dependencies
   - Test coverage for graph operations, persistence, and export

## Project Structure

```
idp-orchestrator/
├── cmd/
│   ├── cli/           # CLI tool (idp-o-ctl)
│   └── server/        # REST + GraphQL API server
├── pkg/
│   ├── api/           # REST and GraphQL handlers + generated code
│   ├── execution/     # Graph execution engine with mocks
│   ├── export/        # DOT/SVG/PNG export functionality
│   ├── graph/         # Core graph model and operations
│   └── storage/       # Postgres persistence layer
├── internal/
│   └── config/        # Configuration management
├── migrations/        # Database migration files
├── schema.graphql     # GraphQL schema definition
├── gqlgen.yml         # GraphQL code generation config
└── go.mod             # Go module dependencies
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- GraphViz (for SVG/PNG export)

### Database Setup

1. Create PostgreSQL database:
```sql
CREATE DATABASE idp_orchestrator;
```

2. Run migrations:
```bash
psql -d idp_orchestrator -f migrations/001_create_tables.sql
```

### Build & Run

1. **API Server**:
```bash
go run cmd/server/main.go --db-password=yourpassword
```

2. **CLI Tool**:
```bash
go run cmd/cli/main.go graph export --app demo --format svg --output graph.svg
```

### Example Usage

**Create a simple graph via GraphQL**:
```graphql
query GetGraph {
  graph(app: "demo") {
    id
    appName
    nodes {
      id
      type
      name
    }
    edges {
      id
      type
      fromNodeId
      toNodeId
    }
  }
}
```

**Export graph via REST API**:
```bash
curl "http://localhost:8080/api/v1/graph/export?app=demo&format=svg" > demo-graph.svg
```

**CLI Export**:
```bash
# Export to DOT (stdout)
idp-o-ctl graph export --app demo

# Export to SVG file
idp-o-ctl graph export --app demo --format svg --output demo.svg

# Export subgraph with specific nodes
idp-o-ctl graph export --app demo --nodes spec1,workflow1 --format png --output subgraph.png
```

## API Endpoints

### REST API
- `GET /api/v1/graph?app={name}` - Get latest graph
- `POST /api/v1/graph/export?app={name}&format={format}` - Export graph
- `GET /api/v1/apps/{app}/runs` - Get execution history
- `POST /api/v1/apps/{app}/runs` - Create execution run
- `PUT /api/v1/runs/{runId}` - Update execution status

### GraphQL
- Playground: `http://localhost:8080/graphql`
- Endpoint: `POST /graphql`

## Configuration

Environment variables:
- `POSTGRES_PASSWORD` - Database password
- Config file: `~/.idp-orchestrator.yaml`

CLI flags:
- `--db-host`, `--db-port`, `--db-user`, `--db-name`
- `--config` for custom config file

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/graph
go test ./pkg/export
```

## Graph Model

### Node Types
- **spec**: Configuration specifications
- **workflow**: Executable processes
- **resource**: Infrastructure or application resources

### Edge Types
- **depends-on**: Generic dependency relationship
- **provisions**: Workflow provisions a resource
- **creates**: Workflow creates another node/resource
- **binds-to**: Bind to an existing resource

### Validation Rules
- `provisions` edges: workflow → resource only
- `creates` edges: workflow → any node type
- `binds-to` edges: any → resource only
- `depends-on` edges: any → any (generic dependency)

## Development

### Generate GraphQL Code
```bash
go run github.com/99designs/gqlgen generate
```

### Add Dependencies
```bash
go get <package>
go mod tidy
```

## Architecture Notes

- **Modular Design**: Clear separation between graph operations, persistence, APIs, and execution
- **Schema-First GraphQL**: Uses gqlgen for type-safe GraphQL development
- **Transaction Safety**: All database operations use GORM transactions
- **Graph Validation**: Comprehensive validation of node types and edge relationships
- **Execution Safety**: Topological sorting prevents execution of cycles
- **Export Flexibility**: Support for multiple output formats with proper styling

## Contributing

1. Follow Go idioms and conventions
2. Add tests for new functionality
3. Update documentation for API changes
4. Use meaningful commit messages
5. Ensure all tests pass before submitting

## License

[Add your license here]