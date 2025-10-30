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
├── storage/     # Persistence (PostgreSQL, SQLite)
├── export/      # Visualization (DOT, SVG, PNG)
└── execution/   # Execution engine + observer pattern
```

## Key Patterns

**Repository Pattern**: `RepositoryInterface` for pluggable storage backends.

**Observer Pattern**: `ExecutionObserver` for state change notifications.

**State Propagation**: Step failure → workflow failure (automatic upward propagation).

## Design Principles

See `CLAUDE.md` for SOLID, KISS, YAGNI principles.
