package storage_test

import (
	"fmt"
	"os"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/philipsahli/innominatus-graph/pkg/graph"
	"github.com/philipsahli/innominatus-graph/pkg/storage"
)

var _ = Describe("Database Connections", func() {
	Describe("NewConnection", func() {
		Context("with SQLite database type", func() {
			It("should create connection with valid file path", func() {
				tmpFile, err := os.CreateTemp("", "bdd-test-*.db")
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tmpFile.Name())

				config := storage.Config{
					Type:   storage.DatabaseTypeSQLite,
					DBName: tmpFile.Name(),
				}
				db, err := storage.NewConnection(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(db).NotTo(BeNil())
			})

			It("should create in-memory database with :memory:", func() {
				config := storage.Config{
					Type:   storage.DatabaseTypeSQLite,
					DBName: ":memory:",
				}
				db, err := storage.NewConnection(config)
				Expect(err).NotTo(HaveOccurred())
				Expect(db).NotTo(BeNil())
			})
		})

		Context("with PostgreSQL database type", func() {
			It("should return error with invalid connection parameters", func() {
				config := storage.Config{
					Type:     storage.DatabaseTypePostgres,
					Host:     "invalid-host-that-does-not-exist.local",
					Port:     5432,
					User:     "invalid",
					Password: "invalid",
					DBName:   "invalid",
					SSLMode:  "disable",
				}
				_, err := storage.NewConnection(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to connect to PostgreSQL"))
			})
		})

		Context("with unsupported database type", func() {
			It("should return descriptive error", func() {
				config := storage.Config{
					Type:   "mysql",
					DBName: "test",
				}
				_, err := storage.NewConnection(config)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported database type"))
			})
		})
	})

	Describe("NewSQLiteConnection", func() {
		It("should create connection with valid file path", func() {
			tmpFile, err := os.CreateTemp("", "bdd-sqlite-*.db")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(tmpFile.Name())

			db, err := storage.NewSQLiteConnection(tmpFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(db).NotTo(BeNil())
		})
	})

	Describe("NewPostgresConnection", func() {
		It("should return error with invalid host", func() {
			_, err := storage.NewPostgresConnection(
				"invalid-host.local",
				"user",
				"password",
				"dbname",
				"disable",
				5432,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to connect to PostgreSQL"))
		})
	})

	Describe("AutoMigrate", func() {
		It("should create all required tables", func() {
			db, err := storage.NewSQLiteConnection(":memory:")
			Expect(err).NotTo(HaveOccurred())

			err = storage.AutoMigrate(db)
			Expect(err).NotTo(HaveOccurred())

			// Verify tables exist by checking if we can query them
			var count int64
			Expect(db.Table("graph_apps").Count(&count).Error).NotTo(HaveOccurred())
			Expect(db.Table("graph_nodes").Count(&count).Error).NotTo(HaveOccurred())
			Expect(db.Table("graph_edges").Count(&count).Error).NotTo(HaveOccurred())
			Expect(db.Table("graph_runs").Count(&count).Error).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Graph Persistence", func() {
	var repo *storage.Repository
	var cleanup func()

	BeforeEach(func() {
		tmpFile, err := os.CreateTemp("", "bdd-repo-*.db")
		Expect(err).NotTo(HaveOccurred())

		db, err := storage.NewSQLiteConnection(tmpFile.Name())
		Expect(err).NotTo(HaveOccurred())

		err = storage.AutoMigrate(db)
		Expect(err).NotTo(HaveOccurred())

		repo = storage.NewRepository(db)
		cleanup = func() { os.Remove(tmpFile.Name()) }
	})

	AfterEach(func() {
		if cleanup != nil {
			cleanup()
		}
	})

	Describe("SaveGraph", func() {
		Context("when saving a new graph", func() {
			It("should create app record if not exists", func() {
				g := graph.NewGraph("new-app")
				err := repo.SaveGraph("new-app", g)
				Expect(err).NotTo(HaveOccurred())

				loaded, err := repo.LoadGraph("new-app")
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded.AppName).To(Equal("new-app"))
			})

			It("should save nodes with all properties", func() {
				g := graph.NewGraph("prop-test")
				node := &graph.Node{
					ID:          "node-1",
					Type:        graph.NodeTypeWorkflow,
					Name:        "Test Node",
					Description: "Test description",
					State:       graph.NodeStateRunning,
					Properties: map[string]interface{}{
						"key1": "value1",
						"key2": 123,
					},
				}
				err := g.AddNode(node)
				Expect(err).NotTo(HaveOccurred())

				err = repo.SaveGraph("prop-test", g)
				Expect(err).NotTo(HaveOccurred())

				loaded, err := repo.LoadGraph("prop-test")
				Expect(err).NotTo(HaveOccurred())

				loadedNode, exists := loaded.GetNode("node-1")
				Expect(exists).To(BeTrue())
				Expect(loadedNode.Name).To(Equal("Test Node"))
				Expect(loadedNode.Description).To(Equal("Test description"))
				Expect(loadedNode.State).To(Equal(graph.NodeStateRunning))
				Expect(loadedNode.Properties["key1"]).To(Equal("value1"))
			})

			It("should save edges with all properties", func() {
				g := graph.NewGraph("edge-test")

				wf := &graph.Node{ID: "wf-1", Type: graph.NodeTypeWorkflow, Name: "Workflow"}
				step := &graph.Node{ID: "step-1", Type: graph.NodeTypeStep, Name: "Step"}
				g.AddNode(wf)
				g.AddNode(step)

				edge := &graph.Edge{
					ID:          "edge-1",
					FromNodeID:  "wf-1",
					ToNodeID:    "step-1",
					Type:        graph.EdgeTypeContains,
					Description: "Contains step",
					Properties: map[string]interface{}{
						"priority": 1,
					},
				}
				err := g.AddEdge(edge)
				Expect(err).NotTo(HaveOccurred())

				err = repo.SaveGraph("edge-test", g)
				Expect(err).NotTo(HaveOccurred())

				loaded, err := repo.LoadGraph("edge-test")
				Expect(err).NotTo(HaveOccurred())

				Expect(loaded.Edges).To(HaveLen(1))
				loadedEdge := loaded.Edges["edge-1"]
				Expect(loadedEdge.Description).To(Equal("Contains step"))
				Expect(loadedEdge.Properties["priority"]).To(BeEquivalentTo(1))
			})
		})

		Context("when saving an empty graph", func() {
			It("should handle graph with no nodes or edges", func() {
				g := graph.NewGraph("empty-app")
				err := repo.SaveGraph("empty-app", g)
				Expect(err).NotTo(HaveOccurred())

				loaded, err := repo.LoadGraph("empty-app")
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded.Nodes).To(BeEmpty())
				Expect(loaded.Edges).To(BeEmpty())
			})
		})

		Context("when updating an existing graph", func() {
			It("should replace all nodes and edges", func() {
				g1 := graph.NewGraph("update-test")
				n1 := &graph.Node{ID: "old-node", Type: graph.NodeTypeWorkflow, Name: "Old"}
				g1.AddNode(n1)
				repo.SaveGraph("update-test", g1)

				g2 := graph.NewGraph("update-test")
				n2 := &graph.Node{ID: "new-node", Type: graph.NodeTypeStep, Name: "New"}
				g2.AddNode(n2)
				repo.SaveGraph("update-test", g2)

				loaded, _ := repo.LoadGraph("update-test")
				Expect(loaded.Nodes).To(HaveLen(1))
				_, oldExists := loaded.GetNode("old-node")
				_, newExists := loaded.GetNode("new-node")
				Expect(oldExists).To(BeFalse())
				Expect(newExists).To(BeTrue())
			})
		})

		Context("when saving large graphs", func() {
			It("should handle 100+ nodes efficiently", func() {
				g := graph.NewGraph("large-graph")

				for i := 0; i < 100; i++ {
					node := &graph.Node{
						ID:   fmt.Sprintf("step-%d", i),
						Type: graph.NodeTypeStep,
						Name: fmt.Sprintf("Step %d", i),
					}
					g.AddNode(node)
				}

				err := repo.SaveGraph("large-graph", g)
				Expect(err).NotTo(HaveOccurred())

				loaded, err := repo.LoadGraph("large-graph")
				Expect(err).NotTo(HaveOccurred())
				Expect(loaded.Nodes).To(HaveLen(100))
			})
		})
	})

	Describe("LoadGraph", func() {
		Context("when app does not exist", func() {
			It("should return descriptive error", func() {
				_, err := repo.LoadGraph("non-existent-app")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when loading a saved graph", func() {
			It("should restore node types correctly", func() {
				g := graph.NewGraph("type-test")
				nodes := []*graph.Node{
					{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"},
					{ID: "step", Type: graph.NodeTypeStep, Name: "Step"},
					{ID: "res", Type: graph.NodeTypeResource, Name: "Resource"},
					{ID: "prov", Type: graph.NodeTypeProvider, Name: "Provider"},
				}
				for _, n := range nodes {
					g.AddNode(n)
				}
				repo.SaveGraph("type-test", g)

				loaded, _ := repo.LoadGraph("type-test")
				wf, _ := loaded.GetNode("wf")
				step, _ := loaded.GetNode("step")
				res, _ := loaded.GetNode("res")
				prov, _ := loaded.GetNode("prov")

				Expect(wf.Type).To(Equal(graph.NodeTypeWorkflow))
				Expect(step.Type).To(Equal(graph.NodeTypeStep))
				Expect(res.Type).To(Equal(graph.NodeTypeResource))
				Expect(prov.Type).To(Equal(graph.NodeTypeProvider))
			})

			It("should restore edge types correctly", func() {
				g := graph.NewGraph("edge-type-test")
				wf := &graph.Node{ID: "wf", Type: graph.NodeTypeWorkflow, Name: "Workflow"}
				step := &graph.Node{ID: "step", Type: graph.NodeTypeStep, Name: "Step"}
				g.AddNode(wf)
				g.AddNode(step)
				g.AddEdge(&graph.Edge{
					ID:         "e1",
					FromNodeID: "wf",
					ToNodeID:   "step",
					Type:       graph.EdgeTypeContains,
				})
				repo.SaveGraph("edge-type-test", g)

				loaded, _ := repo.LoadGraph("edge-type-test")
				Expect(loaded.Edges["e1"].Type).To(Equal(graph.EdgeTypeContains))
			})
		})
	})

	Describe("UpdateNodeState", func() {
		Context("when node exists", func() {
			It("should update state successfully", func() {
				g := graph.NewGraph("state-test")
				node := &graph.Node{
					ID:    "n1",
					Type:  graph.NodeTypeStep,
					Name:  "Test",
					State: graph.NodeStateWaiting,
				}
				g.AddNode(node)
				repo.SaveGraph("state-test", g)

				err := repo.UpdateNodeState("state-test", "n1", graph.NodeStateRunning)
				Expect(err).NotTo(HaveOccurred())

				loaded, _ := repo.LoadGraph("state-test")
				n, _ := loaded.GetNode("n1")
				Expect(n.State).To(Equal(graph.NodeStateRunning))
			})

			It("should update through all state transitions", func() {
				g := graph.NewGraph("transition-test")
				node := &graph.Node{
					ID:    "n1",
					Type:  graph.NodeTypeStep,
					Name:  "Test",
					State: graph.NodeStateWaiting,
				}
				g.AddNode(node)
				repo.SaveGraph("transition-test", g)

				states := []graph.NodeState{
					graph.NodeStateRunning,
					graph.NodeStateSucceeded,
				}
				for _, state := range states {
					err := repo.UpdateNodeState("transition-test", "n1", state)
					Expect(err).NotTo(HaveOccurred())
				}

				loaded, _ := repo.LoadGraph("transition-test")
				n, _ := loaded.GetNode("n1")
				Expect(n.State).To(Equal(graph.NodeStateSucceeded))
			})
		})

		Context("when node does not exist", func() {
			It("should return error", func() {
				g := graph.NewGraph("missing-node")
				repo.SaveGraph("missing-node", g)

				err := repo.UpdateNodeState("missing-node", "non-existent", graph.NodeStateRunning)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when app does not exist", func() {
			It("should return error", func() {
				err := repo.UpdateNodeState("non-existent-app", "n1", graph.NodeStateRunning)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Graph Run Management", func() {
	var repo *storage.Repository
	var cleanup func()

	BeforeEach(func() {
		tmpFile, err := os.CreateTemp("", "bdd-runs-*.db")
		Expect(err).NotTo(HaveOccurred())

		db, err := storage.NewSQLiteConnection(tmpFile.Name())
		Expect(err).NotTo(HaveOccurred())

		err = storage.AutoMigrate(db)
		Expect(err).NotTo(HaveOccurred())

		repo = storage.NewRepository(db)
		cleanup = func() { os.Remove(tmpFile.Name()) }
	})

	AfterEach(func() {
		if cleanup != nil {
			cleanup()
		}
	})

	Describe("CreateGraphRun", func() {
		Context("when app exists", func() {
			It("should create run with pending status", func() {
				g := graph.NewGraph("run-test")
				repo.SaveGraph("run-test", g)

				run, err := repo.CreateGraphRun("run-test", 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(run).NotTo(BeNil())
				Expect(run.Status).To(Equal("pending"))
				Expect(run.Version).To(Equal(1))
			})

			It("should generate unique run ID", func() {
				g := graph.NewGraph("unique-id-test")
				repo.SaveGraph("unique-id-test", g)

				run1, _ := repo.CreateGraphRun("unique-id-test", 1)
				run2, _ := repo.CreateGraphRun("unique-id-test", 2)

				Expect(run1.ID).NotTo(Equal(run2.ID))
			})
		})

		Context("when app does not exist", func() {
			It("should return error", func() {
				_, err := repo.CreateGraphRun("non-existent-app", 1)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("UpdateGraphRun", func() {
		BeforeEach(func() {
			g := graph.NewGraph("update-run-test")
			repo.SaveGraph("update-run-test", g)
			_, _ = repo.CreateGraphRun("update-run-test", 1)
		})

		It("should update to running status", func() {
			run, _ := repo.CreateGraphRun("update-run-test", 2)
			err := repo.UpdateGraphRun(run.ID, "running", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update to completed status", func() {
			run, _ := repo.CreateGraphRun("update-run-test", 3)
			err := repo.UpdateGraphRun(run.ID, "completed", nil)
			Expect(err).NotTo(HaveOccurred())

			runs, _ := repo.GetGraphRuns("update-run-test")
			for _, r := range runs {
				if r.ID == run.ID {
					Expect(r.Status).To(Equal("completed"))
				}
			}
		})

		It("should update to failed with error message", func() {
			run, _ := repo.CreateGraphRun("update-run-test", 4)
			errMsg := "Something went wrong"
			err := repo.UpdateGraphRun(run.ID, "failed", &errMsg)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetGraphRuns", func() {
		It("should return all runs for an app", func() {
			g := graph.NewGraph("multi-run-test")
			repo.SaveGraph("multi-run-test", g)

			repo.CreateGraphRun("multi-run-test", 1)
			repo.CreateGraphRun("multi-run-test", 2)
			repo.CreateGraphRun("multi-run-test", 3)

			runs, err := repo.GetGraphRuns("multi-run-test")
			Expect(err).NotTo(HaveOccurred())
			Expect(runs).To(HaveLen(3))
		})

		It("should return empty list for app with no runs", func() {
			g := graph.NewGraph("no-runs")
			repo.SaveGraph("no-runs", g)

			runs, err := repo.GetGraphRuns("no-runs")
			Expect(err).NotTo(HaveOccurred())
			Expect(runs).To(BeEmpty())
		})

		It("should return error for non-existent app", func() {
			_, err := repo.GetGraphRuns("non-existent")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Concurrent Access", func() {
	var repo *storage.Repository
	var cleanup func()

	BeforeEach(func() {
		tmpFile, err := os.CreateTemp("", "bdd-concurrent-*.db")
		Expect(err).NotTo(HaveOccurred())

		db, err := storage.NewSQLiteConnection(tmpFile.Name())
		Expect(err).NotTo(HaveOccurred())

		err = storage.AutoMigrate(db)
		Expect(err).NotTo(HaveOccurred())

		repo = storage.NewRepository(db)
		cleanup = func() { os.Remove(tmpFile.Name()) }
	})

	AfterEach(func() {
		if cleanup != nil {
			cleanup()
		}
	})

	It("should handle concurrent reads safely", func() {
		g := graph.NewGraph("concurrent-read")
		for i := 0; i < 10; i++ {
			n := &graph.Node{
				ID:   fmt.Sprintf("step-%d", i),
				Type: graph.NodeTypeStep,
				Name: fmt.Sprintf("Step %d", i),
			}
			g.AddNode(n)
		}
		repo.SaveGraph("concurrent-read", g)

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := repo.LoadGraph("concurrent-read")
				if err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		Expect(errors).To(BeEmpty())
	})
})
