# Innominatus Graph SDK

Go SDK for IDP workflows as directed acyclic graphs with state management and persistence.

**Performance-optimized** with O(N+E) algorithms, thread-safe concurrent access, and sub-millisecond operations for 1000+ node graphs.

## Install
```bash
go get github.com/philipsahli/innominatus-graph
```

## Quick Start
```go
g := graph.NewGraph("my-app")
g.AddNode(&graph.Node{ID: "wf-1", Type: graph.NodeTypeWorkflow})

db, _ := storage.NewSQLiteConnection("app.db")
repo := storage.NewRepository(db)
repo.SaveGraph("my-app", g)
```

## Features

✅ **High Performance** - Sub-millisecond graph operations for 1000+ nodes
✅ **Thread-Safe** - Concurrent read/write access with RWMutex
✅ **Execution Algorithms** - GetReadyNodes, GetParallelLayers, PropagateState
✅ **Optimized Topology** - O(N+E) topological sort and cycle detection
✅ **State Management** - Workflow execution states with automatic propagation
✅ **Persistence** - PostgreSQL and SQLite storage backends
✅ **Visualization** - Export to Mermaid, DOT, SVG, PNG formats

## Performance

| Operation | Complexity | 10 nodes | 1000 nodes |
|-----------|------------|----------|------------|
| GetReadyNodes | O(N) | 311ns | 38.5µs |
| PropagateState | O(N+E) | 696ns | 11.7µs |
| GetParallelLayers | O(N+E) | 1.7µs | 502µs |
| TopologicalSort | O(N+E) | <1µs | <500µs |

*Benchmarked on Apple M4*

## Documentation

See [examples/demo/main.go](examples/demo/main.go) • [docs/](docs/) • [CHANGELOG.md](CHANGELOG.md) • Built for [innominatus](https://github.com/innominatus/innominatus)
