package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ciaranmcdonnell/go-api-server/benchmark/internal/k6"
	"github.com/ciaranmcdonnell/go-api-server/benchmark/internal/report"
)

const csvPath = "benchmark/results.csv"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "extract":
		runExtract()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runExtract() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: benchtools extract <pattern> [--label <label>]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  benchtools extract benchmark/results/*_opt1.json --label opt-v1")
		fmt.Fprintln(os.Stderr, "  benchtools extract benchmark/results/health_smoke_*.json --label baseline")
		fmt.Fprintln(os.Stderr, "  benchtools extract benchmark/results/specific_file.json --label test")
		os.Exit(1)
	}

	pattern := os.Args[2]
	label := parseLabel()

	// Expand glob pattern
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid pattern: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No files matched: %s\n", pattern)
		os.Exit(1)
	}

	sort.Strings(files)
	fmt.Fprintf(os.Stderr, "Matched %d file(s)\n\n", len(files))

	created := !fileExists(csvPath)
	errors := 0

	for _, file := range files {
		result, err := k6.ParseFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  SKIP %s: %v\n", file, err)
			errors++
			continue
		}

		if err := report.AppendCSV(csvPath, result, label); err != nil {
			fmt.Fprintf(os.Stderr, "  SKIP %s: %v\n", file, err)
			errors++
			continue
		}

		report.PrintSummary(result, label, csvPath, created)
		created = false
	}

	if errors > 0 {
		fmt.Fprintf(os.Stderr, "\n%d file(s) failed\n", errors)
		os.Exit(1)
	}
}

func parseLabel() string {
	for i, arg := range os.Args {
		if (arg == "--label" || arg == "-l") && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: benchtools <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  extract <pattern> [--label <label>]  Extract metrics to CSV")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Pattern supports globs:")
	fmt.Fprintln(os.Stderr, "  benchmark/results/*_opt1.json        All opt1 results")
	fmt.Fprintln(os.Stderr, "  benchmark/results/auth-flow_*.json   All auth-flow results")
	fmt.Fprintln(os.Stderr, "  benchmark/results/*.json             Everything")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
