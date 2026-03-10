package report

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/ciaranmcdonnell/go-api-server/benchmark/internal/k6"
)

var csvHeader = []string{
	"timestamp", "scenario", "profile", "label",
	"total_reqs", "rps",
	"duration_avg", "duration_p50", "duration_p95", "duration_p99",
	"errors",
}

func AppendCSV(path string, result *k6.Result, label string) error {
	exists := fileExists(path)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening CSV: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if !exists {
		if err := w.Write(csvHeader); err != nil {
			return fmt.Errorf("writing header: %w", err)
		}
	}

	row := []string{
		result.Timestamp,
		result.Scenario,
		result.Profile,
		label,
		fmt.Sprintf("%d", result.TotalReqs),
		fmt.Sprintf("%.2f", result.RPS),
		fmt.Sprintf("%.2f", result.DurationAvg),
		fmt.Sprintf("%.2f", result.DurationP50),
		fmt.Sprintf("%.2f", result.DurationP95),
		fmt.Sprintf("%.2f", result.DurationP99),
		fmt.Sprintf("%d", result.Errors),
	}

	if err := w.Write(row); err != nil {
		return fmt.Errorf("writing row: %w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
