package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"innominatusrchestrator/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "innominatus-ctl",
	Short: "IDP Orchestrator CLI",
	Long:  `Command line tool for the IDP Orchestrator graph-based platform`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initdbCmd)

	rootCmd.PersistentFlags().StringVar(&config.ConfigFile, "config", "", "config file (default is $HOME/.innominatusrchestrator.yaml)")
	rootCmd.PersistentFlags().StringVar(&config.DatabaseHost, "db-host", "localhost", "database host")
	rootCmd.PersistentFlags().IntVar(&config.DatabasePort, "db-port", 5432, "database port")
	rootCmd.PersistentFlags().StringVar(&config.DatabaseUser, "db-user", "postgres", "database user")
	rootCmd.PersistentFlags().StringVar(&config.DatabasePassword, "db-password", "", "database password")
	rootCmd.PersistentFlags().StringVar(&config.DatabaseName, "db-name", "idp_orchestrator", "database name")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("innominatus-ctl version %s\n", version)
		fmt.Printf("Built on %s from commit %s\n", date, commit)
	},
}

var initdbCmd = &cobra.Command{
	Use:   "initdb",
	Short: "Initialize the database",
	Long:  `Initialize the PostgreSQL database by creating the database and running migrations`,
	RunE:  runInitDB,
}

func init() {
	initdbCmd.Flags().Bool("rm", false, "Remove existing database and all its objects before initialization")
}

func runInitDB(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing database...")

	// Get the --rm flag value
	rmFlag, _ := cmd.Flags().GetBool("rm")

	// Connect to postgres database to create the target database if it doesn't exist
	adminDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		config.DatabaseHost, config.DatabasePort, config.DatabaseUser, config.DatabasePassword)

	adminDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer adminDB.Close()

	// Check if database exists
	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", config.DatabaseName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Handle --rm flag: drop database if it exists
	if rmFlag && exists {
		fmt.Printf("Dropping existing database '%s'...\n", config.DatabaseName)

		// Terminate active connections to the database
		_, err = adminDB.Exec(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = $1
			  AND pid <> pg_backend_pid()
		`, config.DatabaseName)
		if err != nil {
			fmt.Printf("Warning: Could not terminate active connections: %v\n", err)
		}

		// Drop the database
		_, err = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", config.DatabaseName))
		if err != nil {
			return fmt.Errorf("failed to drop database: %w", err)
		}
		fmt.Printf("Database '%s' dropped successfully\n", config.DatabaseName)
		exists = false
	}

	// Create database if it doesn't exist
	if !exists {
		fmt.Printf("Creating database '%s'...\n", config.DatabaseName)
		_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DatabaseName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		fmt.Printf("Database '%s' created successfully\n", config.DatabaseName)
	} else {
		fmt.Printf("Database '%s' already exists\n", config.DatabaseName)
	}

	// Connect to the target database
	targetDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DatabaseHost, config.DatabasePort, config.DatabaseUser, config.DatabasePassword, config.DatabaseName)

	db, err := sql.Open("pgx", targetDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to target database: %w", err)
	}
	defer db.Close()

	// Read and execute migration file
	fmt.Println("Running database migrations...")
	migrationPath := filepath.Join("migrations", "001_create_tables.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Load sample data for helloworld app
	fmt.Println("Loading helloworld sample data...")
	sampleDataPath := filepath.Join("examples", "helloworld-data.sql")
	sampleDataSQL, err := os.ReadFile(sampleDataPath)
	if err != nil {
		fmt.Printf("Warning: Could not load helloworld sample data (%s): %v\n", sampleDataPath, err)
		fmt.Println("You can load it manually later with: psql -f examples/helloworld-data.sql")
	} else {
		_, err = db.Exec(string(sampleDataSQL))
		if err != nil {
			fmt.Printf("Warning: Failed to load helloworld sample data: %v\n", err)
			fmt.Println("You can load it manually later with: psql -f examples/helloworld-data.sql")
		} else {
			fmt.Println("HelloWorld sample data loaded successfully!")
		}
	}

	fmt.Println("Database initialization completed successfully!")
	fmt.Printf("Database: %s\n", config.DatabaseName)
	fmt.Printf("Host: %s:%d\n", config.DatabaseHost, config.DatabasePort)
	fmt.Printf("User: %s\n", config.DatabaseUser)

	return nil
}
