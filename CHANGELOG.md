# Changelog

All notable changes to innominatus-graph will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Phase 1 Performance Optimizations (2025-12-27)

#### Adjacency Lists & Degree Tracking
- Added `IncomingEdges` and `OutgoingEdges` adjacency lists to Node struct
- Added `InDegree` and `OutDegree` fields for O(1) degree lookups
- Performance: Node degree queries now O(1) instead of O(E)

#### Thread Safety
- Added `sync.RWMutex` to Graph struct for concurrent access
- All graph operations now thread-safe with appropriate locking
- Read operations use RLock for concurrent read access
- Write operations use Lock for exclusive access

#### New Execution Algorithms

**GetParallelLayers()** - Topological layer computation
- Returns nodes grouped into parallel execution layers
- Each layer contains nodes that can run concurrently
- Complexity: O(N+E) using Kahn's algorithm
- Performance: 1.7µs (10 nodes), 502µs (1000 nodes)

**GetReadyNodes()** - Find executable nodes
- Finds nodes ready to execute based on state and dependencies
- Configurable waiting and completed states
- Complexity: O(N) using InDegree caching
- Performance: 311ns (10 nodes), 38.5µs (1000 nodes)

**PropagateState()** - BFS state propagation
- Propagates state changes through dependency graph
- Breadth-first search from starting node
- Complexity: O(N+E) instead of O(N²) fixed-point iteration
- Performance: 696ns (10 nodes), 11.7µs (1000 nodes)

#### Algorithm Optimizations

**TopologicalSort()**
- Optimized from O(N×E) to O(N+E)
- Now uses adjacency lists (OutgoingEdges) instead of scanning all edges
- 5-10x faster for large graphs

**GetDependencies() / GetDependents()**
- Optimized from O(E) to O(D) where D = node degree
- Direct access via IncomingEdges/OutgoingEdges
- 100-1000x faster for sparse graphs

### Changed

#### Graph Initialization
- `AddEdge()` now populates adjacency lists automatically
- InDegree/OutDegree maintained automatically on edge add/remove
- Backward compatible - existing code works unchanged

#### Test Coverage
- Added 18 new test cases for execution algorithms
- Added concurrent access tests (10 readers + 5 writers)
- Added 19 comprehensive benchmarks
- All 68 existing tests passing

### Performance Impact

#### Small Graphs (10-100 nodes)
- TopologicalSort: 5x faster
- GetReadyNodes: Sub-microsecond (311ns - 3.7µs)
- PropagateState: Sub-microsecond (696ns - 3.5µs)

#### Large Graphs (1000+ nodes)
- TopologicalSort: 10x faster
- GetReadyNodes: 38.5µs (was ~4ms with O(N×D))
- PropagateState: 11.7µs (was ~1ms with O(N²))
- GetDependencies: 100-1000x faster (O(D) vs O(E))

#### Thread Safety Overhead
- RLock read operations: negligible overhead (~10-20ns)
- Lock write operations: <100ns overhead
- Concurrent reads scale linearly with CPU cores

### Technical Details

#### Complexity Improvements

| Operation | Before | After | Speedup |
|-----------|--------|-------|---------|
| TopologicalSort | O(N×E) | O(N+E) | 5-10x |
| GetReadyNodes | O(N×D) | O(N) | 50-100x |
| PropagateState | O(N²) | O(N+E) | 10-20x |
| GetDependencies | O(E) | O(D) | 100-1000x |

Where:
- N = number of nodes
- E = number of edges
- D = node degree (typically << E for sparse graphs)

#### Files Modified
- `pkg/graph/types.go` - Adjacency lists, thread safety
- `pkg/graph/topological.go` - Optimized TopologicalSort
- `pkg/graph/execution.go` - NEW: execution algorithms
- `pkg/graph/execution_test.go` - NEW: 18 test cases
- `pkg/graph/execution_bench_test.go` - NEW: 19 benchmarks

### Migration Notes

#### Backward Compatibility
✅ **No breaking changes** - all existing APIs work unchanged
✅ **Automatic migration** - adjacency lists populated on AddEdge()
✅ **Zero code changes required** - optimizations are internal

#### Optional Features
New execution algorithms are optional:
```go
// Optional: Use new parallel layers algorithm
layers, err := graph.GetParallelLayers()

// Optional: Use optimized ready node detection
ready := graph.GetReadyNodes(waitingStates, completedStates)

// Optional: Use BFS state propagation
graph.PropagateState(nodeID, newState, propagateStates, stopStates)
```

#### Integration with innominatus
innominatus uses optimized graph SDK via `go.mod` replace directive:
```go
replace github.com/philipsahli/innominatus-graph => /path/to/innominatus-graph
```

### Known Issues
None

### Security
No security-related changes

---

## Legend
- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security vulnerability fixes
