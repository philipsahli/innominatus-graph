package main

import (
	"fmt"
	"io"
	"os"

	"innominatusrchestrator/internal/config"

	"github.com/philipsahli/innominatus-graph/pkg/storage"

	"github.com/philipsahli/innominatus-graph/pkg/export"

	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Graph operations",
	Long:  `Commands for working with IDP Orchestrator graphs`,
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export graph to various formats",
	Long:  `Export a graph to DOT, SVG, or PNG format`,
	RunE:  runExport,
}

var (
	appName    string
	format     string
	outputFile string
	nodeIDs    []string
)

func init() {
	graphCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVar(&appName, "app", "", "application name (required)")
	exportCmd.Flags().StringVar(&format, "format", "dot", "output format: dot, svg, png")
	exportCmd.Flags().StringVar(&outputFile, "output", "", "output file path (default: stdout for DOT)")
	exportCmd.Flags().StringSliceVar(&nodeIDs, "nodes", nil, "specific node IDs to include in export")

	exportCmd.MarkFlagRequired("app")
}

func runExport(cmd *cobra.Command, args []string) error {
	cfg := storage.Config{
		Host:     config.DatabaseHost,
		Port:     config.DatabasePort,
		User:     config.DatabaseUser,
		Password: config.DatabasePassword,
		DBName:   config.DatabaseName,
		SSLMode:  "disable",
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

	repository := storage.NewRepository(db)
	exporter := export.NewExporter()
	defer exporter.Close()

	graph, err := repository.LoadGraph(appName)
	if err != nil {
		return fmt.Errorf("failed to load graph for app %s: %w", appName, err)
	}

	exportGraph := graph
	if len(nodeIDs) > 0 {
		exportGraph, err = exporter.CreateSubgraph(graph, nodeIDs)
		if err != nil {
			return fmt.Errorf("failed to create subgraph: %w", err)
		}
	}

	var exportFormat export.Format
	switch format {
	case "dot":
		exportFormat = export.FormatDOT
	case "svg":
		exportFormat = export.FormatSVG
	case "png":
		exportFormat = export.FormatPNG
	default:
		return fmt.Errorf("unsupported format: %s (supported: dot, svg, png)", format)
	}

	data, err := exporter.ExportGraph(exportGraph, exportFormat)
	if err != nil {
		return fmt.Errorf("failed to export graph: %w", err)
	}

	var writer io.Writer
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
		fmt.Printf("Graph exported to %s\n", outputFile)
	} else {
		writer = os.Stdout
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
