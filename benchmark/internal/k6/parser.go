package k6

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/benchmark/internal/stats"
)

type point struct {
	Type   string `json:"type"`
	Metric string `json:"metric"`
	Data   struct {
		Value float64 `json:"value"`
	} `json:"data"`
}

type Result struct {
	Timestamp   string
	Scenario    string
	Profile     string
	TotalReqs   int
	RPS         float64
	DurationAvg float64
	DurationP50 float64
	DurationP95 float64
	DurationP99 float64
	Errors      int
}

func ParseFile(path string) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	metrics := map[string][]float64{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var p point
		if err := json.Unmarshal(scanner.Bytes(), &p); err != nil {
			continue
		}
		if p.Type == "Point" && p.Metric != "" {
			metrics[p.Metric] = append(metrics[p.Metric], p.Data.Value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	scenario, profile, ts := parseFilename(path)
	durations := metrics["http_req_duration"]

	return &Result{
		Timestamp:   ts,
		Scenario:    scenario,
		Profile:     profile,
		TotalReqs:   len(metrics["http_reqs"]),
		RPS:         calculateRPS(metrics),
		DurationAvg: stats.Avg(durations),
		DurationP50: stats.Percentile(durations, 50),
		DurationP95: stats.Percentile(durations, 95),
		DurationP99: stats.Percentile(durations, 99),
		Errors:      len(metrics["errors"]),
	}, nil
}

func calculateRPS(metrics map[string][]float64) float64 {
	iters, ok := metrics["iteration_duration"]
	if !ok || len(iters) == 0 {
		return 0
	}

	var totalDuration float64
	for _, d := range iters {
		totalDuration += d
	}

	avgIterMs := totalDuration / float64(len(iters))
	if avgIterMs <= 0 {
		return 0
	}

	totalReqs := len(metrics["http_reqs"])
	return float64(totalReqs) / (float64(len(iters)) * avgIterMs / 1000)
}

var knownProfiles = map[string]bool{
	"smoke": true, "load": true, "stress": true, "spike": true, "breakpoint": true,
}

func parseFilename(path string) (scenario, profile, timestamp string) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	parts := strings.Split(base, "_")

	// Find the profile keyword — everything before it is scenario, after is timestamp/label
	for i, part := range parts {
		if knownProfiles[part] {
			scenario = strings.Join(parts[:i], "-")
			profile = part
			rest := parts[i+1:]
			if len(rest) >= 2 {
				timestamp = rest[0] + "_" + rest[1]
			} else if len(rest) == 1 {
				timestamp = rest[0]
			} else {
				timestamp = time.Now().Format("20060102_150405")
			}
			return
		}
	}

	return base, "unknown", time.Now().Format("20060102_150405")
}
