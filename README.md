# Innominatus Graph SDK

Go SDK for IDP workflows as directed acyclic graphs with state management and persistence.

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

See [examples/demo/main.go](examples/demo/main.go) • [docs/](docs/) • Built for [innominatus](https://github.com/innominatus/innominatus)
