package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Hook functions for Claude Code automation
// These hooks run automatically on file changes to maintain code quality

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hooks.go <hook-name> [args...]")
		os.Exit(1)
	}

	hookName := os.Args[1]
	args := os.Args[2:]

	switch hookName {
	case "on-file-save":
		onFileSave(args)
	case "on-file-edit":
		onFileEdit(args)
	case "pre-commit":
		preCommit(args)
	case "on-test":
		onTest(args)
	default:
		fmt.Printf("Unknown hook: %s\n", hookName)
		os.Exit(1)
	}
}

// onFileSave runs when a file is saved
func onFileSave(args []string) {
	if len(args) == 0 {
		return
	}

	filePath := args[0]
	if !strings.HasSuffix(filePath, ".go") {
		return // Only process Go files
	}

	fmt.Printf("🔧 Running hooks for: %s\n", filePath)

	// 1. Format with gofmt
	if err := runCommand("gofmt", "-w", filePath); err != nil {
		fmt.Printf("⚠️  gofmt failed: %v\n", err)
	} else {
		fmt.Println("✅ Formatted with gofmt")
	}

	// 2. Run go vet on the file
	packagePath := filepath.Dir(filePath)
	if err := runCommand("go", "vet", packagePath); err != nil {
		fmt.Printf("⚠️  go vet found issues: %v\n", err)
	} else {
		fmt.Println("✅ go vet passed")
	}
}

// onFileEdit runs when a file is edited (can run tests)
func onFileEdit(args []string) {
	if len(args) == 0 {
		return
	}

	filePath := args[0]
	if !strings.HasSuffix(filePath, ".go") {
		return
	}

	// Skip test files and deprecated code
	if strings.HasSuffix(filePath, "_test.go") || strings.Contains(filePath, "deprecated/") {
		return
	}

	fmt.Printf("🧪 Running tests for: %s\n", filePath)

	// Find package directory
	packagePath := filepath.Dir(filePath)

	// Run tests for the package
	if err := runCommand("go", "test", packagePath, "-v"); err != nil {
		fmt.Printf("⚠️  Tests failed: %v\n", err)
	} else {
		fmt.Println("✅ Tests passed")
	}
}

// preCommit runs before git commit
func preCommit(args []string) {
	fmt.Println("🔍 Running pre-commit checks...")

	checks := []struct {
		name    string
		command string
		args    []string
	}{
		{"Format check", "gofmt", []string{"-l", "."}},
		{"Go vet", "go", []string{"vet", "./..."}},
		{"Tests", "go", []string{"test", "./...", "-short"}},
		{"Mod tidy check", "go", []string{"mod", "tidy"}},
	}

	failed := false
	for _, check := range checks {
		fmt.Printf("\n▶️  %s...\n", check.name)
		if err := runCommand(check.command, check.args...); err != nil {
			fmt.Printf("❌ %s failed: %v\n", check.name, err)
			failed = true
		} else {
			fmt.Printf("✅ %s passed\n", check.name)
		}
	}

	// Check test coverage (must be >80%)
	fmt.Printf("\n▶️  Coverage check (target: >80%%)...\n")
	if err := checkCoverage(); err != nil {
		fmt.Printf("❌ Coverage check failed: %v\n", err)
		failed = true
	} else {
		fmt.Println("✅ Coverage check passed")
	}

	if failed {
		fmt.Println("\n❌ Pre-commit checks failed. Fix issues before committing.")
		os.Exit(1)
	}

	fmt.Println("\n✅ All pre-commit checks passed!")
}

// onTest runs after tests complete
func onTest(args []string) {
	fmt.Println("📊 Running post-test checks...")

	// Check coverage
	if err := checkCoverage(); err != nil {
		fmt.Printf("⚠️  Coverage warning: %v\n", err)
	}
}

// checkCoverage verifies test coverage is above threshold
func checkCoverage() error {
	const coverageThreshold = 80.0

	// Run coverage
	cmd := exec.Command("go", "test", "-cover", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run coverage: %w", err)
	}

	// Parse coverage from output
	// Output format: "coverage: XX.X% of statements"
	lines := strings.Split(string(output), "\n")
	var totalCoverage float64
	var packageCount int

	for _, line := range lines {
		if strings.Contains(line, "coverage:") && strings.Contains(line, "%") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "coverage:" && i+1 < len(parts) {
					coverageStr := strings.TrimSuffix(parts[i+1], "%")
					var coverage float64
					if _, err := fmt.Sscanf(coverageStr, "%f", &coverage); err == nil {
						totalCoverage += coverage
						packageCount++
					}
				}
			}
		}
	}

	if packageCount == 0 {
		return fmt.Errorf("no coverage data found")
	}

	avgCoverage := totalCoverage / float64(packageCount)
	fmt.Printf("   Average coverage: %.1f%% (threshold: %.1f%%)\n", avgCoverage, coverageThreshold)

	if avgCoverage < coverageThreshold {
		return fmt.Errorf("coverage %.1f%% is below threshold %.1f%%", avgCoverage, coverageThreshold)
	}

	return nil
}

// runCommand executes a command and prints output
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
