// +build ignore

// run_tests.go - Test runner entry point for ExecuteTurn1Combined Lambda
// Usage: go run run_tests.go [options]
//
// Options:
//   -v          Verbose output
//   -cover      Show code coverage
//   -bench      Run benchmarks
//   -short      Skip long-running tests
//   -specific   Run specific test pattern
//   -report     Generate HTML coverage report

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Parse command line flags
	var (
		verbose     = flag.Bool("v", false, "Verbose test output")
		coverage    = flag.Bool("cover", false, "Show code coverage")
		bench       = flag.Bool("bench", false, "Run benchmarks")
		short       = flag.Bool("short", false, "Skip long-running tests")
		specific    = flag.String("specific", "", "Run specific test pattern")
		htmlReport  = flag.Bool("report", false, "Generate HTML coverage report")
		integration = flag.Bool("integration", false, "Include integration tests")
		race        = flag.Bool("race", false, "Enable race detector")
		cpu         = flag.Int("cpu", 0, "Number of CPUs to use (0 = all)")
		timeout     = flag.Duration("timeout", 10*time.Minute, "Test timeout")
	)
	flag.Parse()

	// Build base command
	args := []string{"test"}

	// Add verbosity
	if *verbose {
		args = append(args, "-v")
	}

	// Add coverage
	if *coverage || *htmlReport {
		args = append(args, "-cover", "-coverprofile=coverage.out")
		args = append(args, "-covermode=atomic")
	}

	// Add benchmarks
	if *bench {
		args = append(args, "-bench=.")
		args = append(args, "-benchmem")
	}

	// Add short flag
	if *short {
		args = append(args, "-short")
	}

	// Add race detector
	if *race {
		args = append(args, "-race")
	}

	// Add CPU flag
	if *cpu > 0 {
		args = append(args, fmt.Sprintf("-cpu=%d", *cpu))
	}

	// Add timeout
	args = append(args, fmt.Sprintf("-timeout=%s", *timeout))

	// Add specific test pattern
	if *specific != "" {
		args = append(args, "-run", *specific)
	}

	// Determine which packages to test
	packages := []string{"./..."}
	if !*integration {
		// Exclude integration tests by default
		packages = []string{
			"./internal/handler",
			"./internal/config",
			"./internal/models",
			"./internal/utils",
			"./internal/validation",
		}
	}

	args = append(args, packages...)

	// Create test report directory
	reportDir := "test_reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		fmt.Printf("Failed to create report directory: %v\n", err)
	}

	// Log test configuration
	fmt.Println("=== ExecuteTurn1Combined Test Runner ===")
	fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Verbose: %v\n", *verbose)
	fmt.Printf("  Coverage: %v\n", *coverage || *htmlReport)
	fmt.Printf("  Benchmarks: %v\n", *bench)
	fmt.Printf("  Short: %v\n", *short)
	fmt.Printf("  Race Detector: %v\n", *race)
	fmt.Printf("  Integration: %v\n", *integration)
	if *specific != "" {
		fmt.Printf("  Pattern: %s\n", *specific)
	}
	fmt.Printf("  Timeout: %s\n", *timeout)
	fmt.Printf("  Packages: %s\n", strings.Join(packages, " "))
	fmt.Println("=====================================")
	fmt.Println()

	// Run tests
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	fmt.Println()
	fmt.Printf("=== Test Execution Complete ===\n")
	fmt.Printf("Duration: %s\n", duration)

	// Generate HTML coverage report if requested
	if *htmlReport && err == nil {
		fmt.Println("\nGenerating HTML coverage report...")
		htmlCmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o=test_reports/coverage.html")
		if err := htmlCmd.Run(); err != nil {
			fmt.Printf("Failed to generate HTML report: %v\n", err)
		} else {
			fmt.Println("HTML coverage report generated: test_reports/coverage.html")
		}
	}

	// Generate test summary
	if *coverage && err == nil {
		fmt.Println("\n=== Coverage Summary ===")
		summaryCmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
		summaryCmd.Stdout = os.Stdout
		summaryCmd.Run()
	}

	// Exit with appropriate code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
	
	fmt.Println("\n✅ All tests passed!")
	os.Exit(0)
}