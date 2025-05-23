// +build ignore

// verify_test_setup.go - Verifies that the test infrastructure is properly set up
package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

func main() {
	fmt.Println("=== Verifying Test Setup for ExecuteTurn1Combined ===")
	fmt.Println()

	checks := []CheckResult{}

	// Check 1: Verify test directories exist
	testDirs := []string{
		"internal/handler",
		"internal/handler/mocks",
		"test_reports",
	}

	for _, dir := range testDirs {
		if _, err := os.Stat(dir); err == nil {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("Directory %s exists", dir),
				Passed:  true,
				Message: "✓",
			})
		} else {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("Directory %s exists", dir),
				Passed:  false,
				Message: "✗ Missing directory",
			})
		}
	}

	// Check 2: Verify test files exist
	testFiles := []string{
		"internal/handler/handler_happy_test.go",
		"internal/handler/mocks/mocks.go",
		"internal/handler/mocks/fixtures.go",
		"run_tests.go",
		"test.sh",
		"Makefile",
	}

	for _, file := range testFiles {
		if _, err := os.Stat(file); err == nil {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("File %s exists", filepath.Base(file)),
				Passed:  true,
				Message: "✓",
			})
		} else {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("File %s exists", filepath.Base(file)),
				Passed:  false,
				Message: "✗ Missing file",
			})
		}
	}

	// Check 3: Verify Go module
	if _, err := os.Stat("go.mod"); err == nil {
		checks = append(checks, CheckResult{
			Name:    "go.mod exists",
			Passed:  true,
			Message: "✓",
		})

		// Check for testify dependency
		content, _ := os.ReadFile("go.mod")
		if strings.Contains(string(content), "github.com/stretchr/testify") {
			checks = append(checks, CheckResult{
				Name:    "testify dependency present",
				Passed:  true,
				Message: "✓",
			})
		} else {
			checks = append(checks, CheckResult{
				Name:    "testify dependency present",
				Passed:  false,
				Message: "✗ Missing testify dependency",
			})
		}
	} else {
		checks = append(checks, CheckResult{
			Name:    "go.mod exists",
			Passed:  false,
			Message: "✗ Missing go.mod",
		})
	}

	// Check 4: Verify package imports
	pkg, err := build.ImportDir(".", 0)
	if err == nil {
		checks = append(checks, CheckResult{
			Name:    "Package builds",
			Passed:  true,
			Message: fmt.Sprintf("✓ Package: %s", pkg.ImportPath),
		})
	} else {
		checks = append(checks, CheckResult{
			Name:    "Package builds",
			Passed:  false,
			Message: fmt.Sprintf("✗ Build error: %v", err),
		})
	}

	// Check 5: Test for template files
	templateDirs := []string{
		"templates/turn1-layout-vs-checking",
		"templates/turn1-previous-vs-current",
	}

	for _, dir := range templateDirs {
		if _, err := os.Stat(dir); err == nil {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("Template dir %s exists", filepath.Base(dir)),
				Passed:  true,
				Message: "✓",
			})
		} else {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("Template dir %s exists", filepath.Base(dir)),
				Passed:  false,
				Message: "✗ Missing template directory",
			})
		}
	}

	// Print results
	fmt.Println("Check Results:")
	fmt.Println("==============")
	
	passedCount := 0
	failedCount := 0
	
	for _, check := range checks {
		status := "✅"
		if !check.Passed {
			status = "❌"
			failedCount++
		} else {
			passedCount++
		}
		fmt.Printf("%s %-40s %s\n", status, check.Name, check.Message)
	}

	fmt.Println()
	fmt.Printf("Summary: %d passed, %d failed\n", passedCount, failedCount)
	fmt.Println()

	// Provide setup instructions if needed
	if failedCount > 0 {
		fmt.Println("Setup Instructions:")
		fmt.Println("==================")
		fmt.Println("1. Create missing directories:")
		fmt.Println("   mkdir -p test_reports internal/handler/mocks")
		fmt.Println()
		fmt.Println("2. Install test dependencies:")
		fmt.Println("   go get github.com/stretchr/testify/assert")
		fmt.Println("   go get github.com/stretchr/testify/mock")
		fmt.Println("   go get github.com/stretchr/testify/require")
		fmt.Println()
		fmt.Println("3. Run tests:")
		fmt.Println("   make test                 # Using Makefile")
		fmt.Println("   ./test.sh -v -c          # Using shell script")
		fmt.Println("   go run run_tests.go -v   # Using Go test runner")
		fmt.Println()
	} else {
		fmt.Println("✅ All checks passed! Your test setup is ready.")
		fmt.Println()
		fmt.Println("You can now run tests using:")
		fmt.Println("  make test                 # Basic tests")
		fmt.Println("  make test-cover          # With coverage")
		fmt.Println("  make test-happy          # Happy path test only")
		fmt.Println("  ./test.sh -v -c          # Using shell script")
		fmt.Println("  go run run_tests.go -v   # Using Go test runner")
		fmt.Println()
	}

	if failedCount > 0 {
		os.Exit(1)
	}
}