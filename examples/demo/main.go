package main

import (
	"fmt"
	"log"
	"os"

	"github.com/philipsahli/innominatus-graph/pkg/storage"

	"github.com/philipsahli/innominatus-graph/pkg/export"

	"github.com/philipsahli/innominatus-graph/pkg/execution"

	"github.com/philipsahli/innominatus-graph/pkg/graph"

	"gorm.io/gorm"
)

// DemoObserver implements ExecutionObserver to log state changes
type DemoObserver struct{}

func (o *DemoObserver) OnNodeStateChange(node *graph.Node, oldState, newState graph.NodeState) {
	fmt.Printf("⚡ State Change: %s (%s) %s → %s\n", node.Name, node.Type, oldState, newState)
}

func main() {
	fmt.Println("🚀 Innominatus Graph SDK Demo")
	fmt.Println("========================================\n")

	// Step 1: Create a graph with workflow, steps, and resources
	fmt.Println("📊 Building graph with workflow → steps → resources...")
	g := graph.NewGraph("demo-app")

	// Create spec node
	specNode := &graph.Node{
		ID:          "app-spec",
		Type:        graph.NodeTypeSpec,
		Name:        "Application Spec",
		Description: "Score specification for demo app",
		Properties: map[string]interface{}{
			"version": "1.0",
		},
	}
	g.AddNode(specNode)

	// Create workflow node
	workflowNode := &graph.Node{
		ID:          "deploy-workflow",
		Type:        graph.NodeTypeWorkflow,
		Name:        "Deploy Application Workflow",
		Description: "Multi-step deployment workflow",
	}
	g.AddNode(workflowNode)

	// Create step nodes
	step1 := &graph.Node{
		ID:          "provision-infra-step",
		Type:        graph.NodeTypeStep,
		Name:        "Provision Infrastructure",
		Description: "Step to provision cloud infrastructure",
	}
	g.AddNode(step1)

	step2 := &graph.Node{
		ID:          "deploy-app-step",
		Type:        graph.NodeTypeStep,
		Name:        "Deploy Application",
		Description: "Step to deploy application containers",
	}
	g.AddNode(step2)

	step3 := &graph.Node{
		ID:          "configure-monitoring-step",
		Type:        graph.NodeTypeStep,
		Name:        "Configure Monitoring",
		Description: "Step to configure monitoring and alerts",
	}
	g.AddNode(step3)

	// Create resource nodes
	dbResource := &graph.Node{
		ID:          "postgres-db",
		Type:        graph.NodeTypeResource,
		Name:        "PostgreSQL Database",
		Description: "Managed PostgreSQL instance",
	}
	g.AddNode(dbResource)

	k8sResource := &graph.Node{
		ID:          "k8s-deployment",
		Type:        graph.NodeTypeResource,
		Name:        "Kubernetes Deployment",
		Description: "Application deployment in Kubernetes",
	}
	g.AddNode(k8sResource)

	prometheusResource := &graph.Node{
		ID:          "prometheus-scrape",
		Type:        graph.NodeTypeResource,
		Name:        "Prometheus Scrape Config",
		Description: "Monitoring configuration",
	}
	g.AddNode(prometheusResource)

	// Add edges
	// Spec → Workflow
	g.AddEdge(&graph.Edge{
		ID:          "spec-to-workflow",
		FromNodeID:  "app-spec",
		ToNodeID:    "deploy-workflow",
		Type:        graph.EdgeTypeDependsOn,
		Description: "Workflow depends on spec",
	})

	// Workflow contains steps
	g.AddEdge(&graph.Edge{
		ID:          "workflow-contains-step1",
		FromNodeID:  "deploy-workflow",
		ToNodeID:    "provision-infra-step",
		Type:        graph.EdgeTypeContains,
		Description: "Workflow contains infra provisioning step",
	})

	g.AddEdge(&graph.Edge{
		ID:          "workflow-contains-step2",
		FromNodeID:  "deploy-workflow",
		ToNodeID:    "deploy-app-step",
		Type:        graph.EdgeTypeContains,
		Description: "Workflow contains app deployment step",
	})

	g.AddEdge(&graph.Edge{
		ID:          "workflow-contains-step3",
		FromNodeID:  "deploy-workflow",
		ToNodeID:    "configure-monitoring-step",
		Type:        graph.EdgeTypeContains,
		Description: "Workflow contains monitoring config step",
	})

	// Steps configure resources
	g.AddEdge(&graph.Edge{
		ID:          "step1-configures-db",
		FromNodeID:  "provision-infra-step",
		ToNodeID:    "postgres-db",
		Type:        graph.EdgeTypeConfigures,
		Description: "Step configures database",
	})

	g.AddEdge(&graph.Edge{
		ID:          "step2-configures-k8s",
		FromNodeID:  "deploy-app-step",
		ToNodeID:    "k8s-deployment",
		Type:        graph.EdgeTypeConfigures,
		Description: "Step configures k8s deployment",
	})

	g.AddEdge(&graph.Edge{
		ID:          "step3-configures-prometheus",
		FromNodeID:  "configure-monitoring-step",
		ToNodeID:    "prometheus-scrape",
		Type:        graph.EdgeTypeConfigures,
		Description: "Step configures prometheus",
	})

	// Step dependencies
	g.AddEdge(&graph.Edge{
		ID:          "step2-depends-step1",
		FromNodeID:  "deploy-app-step",
		ToNodeID:    "provision-infra-step",
		Type:        graph.EdgeTypeDependsOn,
		Description: "App deployment depends on infrastructure",
	})

	g.AddEdge(&graph.Edge{
		ID:          "step3-depends-step2",
		FromNodeID:  "configure-monitoring-step",
		ToNodeID:    "deploy-app-step",
		Type:        graph.EdgeTypeDependsOn,
		Description: "Monitoring depends on app deployment",
	})

	fmt.Printf("✅ Graph created: %d nodes, %d edges\n\n", len(g.Nodes), len(g.Edges))

	// Step 2: Demonstrate helper methods
	fmt.Println("🔍 Querying graph nodes...")
	workflows := g.GetNodesByType(graph.NodeTypeWorkflow)
	steps := g.GetNodesByType(graph.NodeTypeStep)
	resources := g.GetNodesByType(graph.NodeTypeResource)

	fmt.Printf("  Workflows: %d\n", len(workflows))
	fmt.Printf("  Steps: %d\n", len(steps))
	fmt.Printf("  Resources: %d\n\n", len(resources))

	// Demonstrate parent-child relationships
	childSteps := g.GetChildSteps("deploy-workflow")
	fmt.Printf("  Workflow has %d child steps:\n", len(childSteps))
	for _, step := range childSteps {
		fmt.Printf("    - %s\n", step.Name)
	}
	fmt.Println()

	// Step 3: Export graph to DOT/SVG
	fmt.Println("📤 Exporting graph visualizations...")
	exporter := export.NewExporter()
	defer exporter.Close()

	// Export to DOT
	dotBytes, dotErr := exporter.ExportGraph(g, export.FormatDOT)
	if dotErr != nil {
		log.Fatalf("Failed to export DOT: %v", dotErr)
	}
	os.WriteFile("demo-graph.dot", dotBytes, 0644)
	fmt.Println("  ✅ Exported to demo-graph.dot")

	// Export to SVG
	svgBytes, svgErr := exporter.ExportGraph(g, export.FormatSVG)
	if svgErr != nil {
		log.Fatalf("Failed to export SVG: %v", svgErr)
	}
	os.WriteFile("demo-graph.svg", svgBytes, 0644)
	fmt.Println("  ✅ Exported to demo-graph.svg")

	fmt.Println()

	// Step 4: Demonstrate state management
	fmt.Println("🔄 Demonstrating state propagation...")

	// Simulate step failure
	fmt.Println("\n  Simulating step failure...")
	stateErr := g.UpdateNodeState("provision-infra-step", graph.NodeStateFailed)
	if stateErr != nil {
		log.Fatalf("Failed to update state: %v", stateErr)
	}

	// Check parent workflow state
	workflowNode, _ = g.GetNode("deploy-workflow")
	fmt.Printf("  Parent workflow state after step failure: %s\n", workflowNode.State)

	// Reset and simulate success
	fmt.Println("\n  Simulating successful execution...")
	for _, step := range childSteps {
		g.UpdateNodeState(step.ID, graph.NodeStateRunning)
		g.UpdateNodeState(step.ID, graph.NodeStateSucceeded)
	}

	// Query nodes by state
	succeededNodes := g.GetNodesByState(graph.NodeStateSucceeded)
	fmt.Printf("  Nodes in 'succeeded' state: %d\n", len(succeededNodes))

	// Step 5: Database persistence (SQLite or PostgreSQL)
	dbPassword := os.Getenv("DB_PASSWORD")
	useSQLite := os.Getenv("USE_SQLITE")

	var db *gorm.DB
	var dbErr error

	if useSQLite != "" || dbPassword == "" {
		fmt.Println("\n💾 Demonstrating SQLite persistence...")
		db, dbErr = storage.NewSQLiteConnection("demo-graph.db")
		if dbErr != nil {
			fmt.Printf("   ⚠️  Could not connect to SQLite: %v\n", dbErr)
		} else {
			// Auto-migrate schema
			storage.AutoMigrate(db)

			repo := storage.NewRepository(db)

			// Save graph
			saveErr := repo.SaveGraph("demo-app", g)
			if saveErr != nil {
				log.Fatalf("Failed to save graph: %v", saveErr)
			}
			fmt.Println("  ✅ Graph saved to SQLite (demo-graph.db)")

			// Load graph
			loadedGraph, loadErr := repo.LoadGraph("demo-app")
			if loadErr != nil {
				log.Fatalf("Failed to load graph: %v", loadErr)
			}
			fmt.Printf("  ✅ Graph loaded from SQLite (%d nodes, %d edges)\n",
				len(loadedGraph.Nodes), len(loadedGraph.Edges))

			// Step 6: Execution with observer (SQLite)
			fmt.Println("\n🎯 Demonstrating execution with observer...")
			runner := execution.NewMockWorkflowRunner()
			engine := execution.NewEngine(repo, runner)

			// Register observer
			observer := &DemoObserver{}
			engine.RegisterObserver(observer)

			fmt.Println("  Starting workflow execution with state change notifications...\n")
			// Note: This would execute the workflow and trigger observer callbacks
			// Skipped in demo to avoid complex setup
		}
	} else {
		// PostgreSQL mode
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}

		fmt.Println("\n💾 Demonstrating PostgreSQL persistence...")
		db, dbErr = storage.NewPostgresConnection(dbHost, "postgres", dbPassword, "idp_orchestrator", "disable", 5432)
		if dbErr != nil {
			fmt.Printf("   ⚠️  Could not connect to PostgreSQL: %v\n", dbErr)
		} else {
			repo := storage.NewRepository(db)

			// Save graph
			saveErr := repo.SaveGraph("demo-app", g)
			if saveErr != nil {
				log.Fatalf("Failed to save graph: %v", saveErr)
			}
			fmt.Println("  ✅ Graph saved to PostgreSQL database")

			// Load graph
			loadedGraph, loadErr := repo.LoadGraph("demo-app")
			if loadErr != nil {
				log.Fatalf("Failed to load graph: %v", loadErr)
			}
			fmt.Printf("  ✅ Graph loaded from PostgreSQL (%d nodes, %d edges)\n",
				len(loadedGraph.Nodes), len(loadedGraph.Edges))

			// Step 6: Execution with observer (PostgreSQL)
			fmt.Println("\n🎯 Demonstrating execution with observer...")
			runner := execution.NewMockWorkflowRunner()
			engine := execution.NewEngine(repo, runner)

			// Register observer
			observer := &DemoObserver{}
			engine.RegisterObserver(observer)

			fmt.Println("  Starting workflow execution with state change notifications...\n")
			// Note: This would execute the workflow and trigger observer callbacks
			// Skipped in demo to avoid complex setup
		}
	}

	fmt.Println("\n✨ Demo completed successfully!")
	fmt.Println("========================================")
	fmt.Println("\n📖 Next Steps:")
	fmt.Println("  1. Import this SDK in your orchestrator: go get github.com/innominatus/innominatus-graph")
	fmt.Println("  2. View demo-graph.svg for visual representation")
	fmt.Println("  3. Implement ExecutionObserver for real-time state tracking")
	fmt.Println("  4. Use SQLite for development or PostgreSQL for production")
	fmt.Println("\n💡 Database Options:")
	fmt.Println("  - SQLite (default): Just run 'go run main.go'")
	fmt.Println("  - PostgreSQL: DB_PASSWORD=yourpassword go run main.go")
}
