# Innominatus Graph SDK

[![CI](https://github.com/philipsahli/innominatus-graph/actions/workflows/ci.yml/badge.svg)](https://github.com/philipsahli/innominatus-graph/actions/workflows/ci.yml)

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

- **High Performance** - Sub-millisecond operations, O(N+E) algorithms
- **Thread-Safe** - Concurrent access with RWMutex protection
- **Execution** - GetReadyNodes, GetParallelLayers, PropagateState, TopologicalSort
- **State Management** - 6 node states with automatic failure propagation
- **Persistence** - PostgreSQL (production) and SQLite (development)
- **Export** - DOT, SVG, PNG, JSON, Mermaid (flowchart/state/gantt)
- **Layout** - Hierarchical, Radial, Grid, Force-directed positioning
- **Observable** - GraphObserver pattern for reactive state changes

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
