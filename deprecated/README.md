# Deprecated Components

⚠️ **DEPRECATED**: The components in this directory are deprecated and maintained for backward compatibility only.

## What's Deprecated

- **cmd/**: CLI and server applications
  - `cmd/cli/`: Command-line interface (use SDK directly instead)
  - `cmd/server/`: REST and GraphQL API server (integrate SDK in your own service)
- **api/**: REST and GraphQL API handlers (use SDK package directly)

## Migration Path

**innominatus-graph** is now an SDK library, not a standalone application.

### If you were using the CLI:
```go
// OLD: Running CLI commands
// ./innominatus-ctl graph export --app demo --format svg

// NEW: Use SDK directly in your Go code
import "github.com/innominatus/innominatus-graph/pkg/export"

exporter := export.NewExporter()
svgBytes, _ := exporter.ExportGraph(graph, export.FormatSVG)
```

### If you were using the REST API:
```go
// OLD: Making HTTP requests to standalone server
// GET http://localhost:8080/api/v1/graph?app=demo

// NEW: Import SDK in your application
import (
    "github.com/innominatus/innominatus-graph/pkg/graph"
    "github.com/innominatus/innominatus-graph/pkg/storage"
)

repo := storage.NewRepository(db)
g, _ := repo.LoadGraph("demo")
```

### If you were using GraphQL:
```go
// OLD: GraphQL queries to standalone server

// NEW: Use SDK directly for graph operations
import "github.com/innominatus/innominatus-graph/pkg/graph"

g := graph.NewGraph("my-app")
nodes := g.GetNodesByType(graph.NodeTypeWorkflow)
```

## Why Deprecated?

innominatus-graph has been refactored into a clean SDK library designed for integration into larger IDP platforms (like the innominatus orchestrator). The standalone CLI and server are no longer the recommended usage pattern.

## Support Timeline

- **Current**: Deprecated components remain in this directory for reference
- **Future**: These components may be removed in a future major version

## Recommended Approach

See the main README.md and `examples/demo/` for SDK integration patterns.
