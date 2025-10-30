# Getting Started

## Install
```bash
go get github.com/philipsahli/innominatus-graph
```

## Quick Start
```go
package main

import (
    "github.com/philipsahli/innominatus-graph/pkg/graph"
    "github.com/philipsahli/innominatus-graph/pkg/storage"
)

func main() {
    // Create graph
    g := graph.NewGraph("my-app")

    // Add workflow
    g.AddNode(&graph.Node{
        ID:   "wf-1",
        Type: graph.NodeTypeWorkflow,
        Name: "Deploy Workflow",
    })

    // Persist
    db, _ := storage.NewSQLiteConnection("app.db")
    storage.AutoMigrate(db)
    repo := storage.NewRepository(db)
    repo.SaveGraph("my-app", g)
}
```

## Full Example
See `examples/demo/main.go` for complete working example.
