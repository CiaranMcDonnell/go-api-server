package report

import (
	"fmt"

	"github.com/ciaranmcdonnell/go-api-server/benchmark/internal/k6"
)

func PrintSummary(result *k6.Result, label, csvPath string, created bool) {
	fmt.Printf("%-16s %s\n", "scenario:", result.Scenario)
	fmt.Printf("%-16s %s\n", "profile:", result.Profile)
	if label != "" {
		fmt.Printf("%-16s %s\n", "label:", label)
	}
	fmt.Printf("%-16s %d\n", "total reqs:", result.TotalReqs)
	fmt.Printf("%-16s %.2f\n", "rps:", result.RPS)
	fmt.Printf("%-16s %.2f ms\n", "avg:", result.DurationAvg)
	fmt.Printf("%-16s %.2f ms\n", "p50:", result.DurationP50)
	fmt.Printf("%-16s %.2f ms\n", "p95:", result.DurationP95)
	fmt.Printf("%-16s %.2f ms\n", "p99:", result.DurationP99)
	fmt.Printf("%-16s %d\n", "errors:", result.Errors)

	if created {
		fmt.Printf("\nCreated %s\n", csvPath)
	} else {
		fmt.Printf("\nAppended to %s\n", csvPath)
	}
}
