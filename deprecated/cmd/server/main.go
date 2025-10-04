package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"idp-orchestrator/pkg/api"
	"idp-orchestrator/pkg/storage"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	port       int
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string
	dbSSLMode  string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "idp-orchestrator-server",
	Short: "IDP Orchestrator API Server",
	Long:  `HTTP server providing REST and GraphQL APIs for the IDP Orchestrator`,
	RunE:  runServer,
}

func init() {
	rootCmd.Flags().IntVar(&port, "port", 8080, "server port")
	rootCmd.Flags().StringVar(&dbHost, "db-host", "localhost", "database host")
	rootCmd.Flags().IntVar(&dbPort, "db-port", 5432, "database port")
	rootCmd.Flags().StringVar(&dbUser, "db-user", "postgres", "database user")
	rootCmd.Flags().StringVar(&dbPassword, "db-password", "", "database password")
	rootCmd.Flags().StringVar(&dbName, "db-name", "idp_orchestrator", "database name")
	rootCmd.Flags().StringVar(&dbSSLMode, "db-ssl-mode", "disable", "database SSL mode")

	viper.AutomaticEnv()
	viper.BindPFlags(rootCmd.Flags())
}

func runServer(cmd *cobra.Command, args []string) error {
	if dbPassword == "" {
		dbPassword = viper.GetString("db-password")
	}
	if dbPassword == "" {
		if envPassword := os.Getenv("POSTGRES_PASSWORD"); envPassword != "" {
			dbPassword = envPassword
		}
	}

	cfg := storage.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSSLMode,
	}

	db, err := storage.NewConnection(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	if err := storage.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	repository := storage.NewRepository(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	restHandler := api.NewRESTHandler(repository)
	defer restHandler.Close()

	restHandler.SetupRoutes(r)

	resolver := api.NewResolver(repository)
	srv := handler.NewDefaultServer(api.NewExecutableSchema(api.Config{Resolvers: resolver}))

	r.POST("/graphql", gin.WrapH(srv))
	r.GET("/graphql", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "IDP Orchestrator API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health":   "/health",
				"graphql":  "/graphql",
				"rest_api": "/api/v1",
			},
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on port %d", port)
		log.Printf("GraphQL playground available at http://localhost:%d/graphql", port)
		log.Printf("REST API available at http://localhost:%d/api/v1", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}