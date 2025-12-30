# Architecture

## Core Concepts

**Graph**: Directed acyclic graph (DAG) with nodes and edges.

**Nodes**: 4 types - `spec`, `workflow`, `step`, `resource`.

**Edges**: 6 types - `depends-on`, `provisions`, `creates`, `binds-to`, `contains`, `configures`.

**States**: `waiting`, `pending`, `running`, `failed`, `succeeded`.

## Domain Structure

```
pkg/
├── graph/       # Core graph model (nodes, edges, state)
│   ├── types.go          # Node/Edge structs with adjacency lists
│   ├── topological.go    # O(N+E) topological algorithms
│   ├── execution.go      # Parallel layers, ready nodes, state propagation
│   └── *_test.go         # 68 tests + 19 benchmarks
├── storage/     # Persistence (PostgreSQL, SQLite)
├── export/      # Visualization (DOT, SVG, PNG)
└── execution/   # Execution engine + observer pattern
```

## Performance Characteristics

### Graph Operations (Complexity)

| Operation | Complexity | Performance (1000 nodes) |
|-----------|------------|--------------------------|
| **TopologicalSort** | O(N+E) | <500µs |
| **GetReadyNodes** | O(N) | 38.5µs |
| **PropagateState** | O(N+E) | 11.7µs |
| **GetParallelLayers** | O(N+E) | 502µs |
| **GetDependencies** | O(D) | <1µs per node |
| **AddNode/AddEdge** | O(1) | <100ns |

Where: N = nodes, E = edges, D = node degree

### Optimization Techniques

**Adjacency Lists**: Each node maintains IncomingEdges and OutgoingEdges for O(D) traversal instead of O(E).

**Degree Caching**: InDegree and OutDegree fields enable O(1) degree lookups instead of counting edges.

**Thread Safety**: RWMutex allows concurrent reads while protecting writes. Read operations scale linearly with CPU cores.

**BFS Propagation**: State changes propagate via breadth-first search (O(N+E)) instead of fixed-point iteration (O(N²)).

## Key Patterns

**Repository Pattern**: `RepositoryInterface` for pluggable storage backends.

**Observer Pattern**: `ExecutionObserver` for state change notifications.

**State Propagation**: Step failure → workflow failure (automatic BFS upward propagation).

**Adjacency List Pattern**: Direct edge access for fast graph traversal.

## Design Principles

See `CLAUDE.md` for SOLID, KISS, YAGNI principles.

## Thread Safety

All graph operations are thread-safe:
- **Read operations** (GetNode, GetReadyNodes, TopologicalSort): Use `RLock` for concurrent access
- **Write operations** (AddNode, AddEdge, UpdateState): Use `Lock` for exclusive access
- **Performance**: Read lock overhead ~10-20ns, negligible for production workloads
