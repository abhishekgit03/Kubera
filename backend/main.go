package main

import (
	"embed"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/abhishekgit03/kubera/api"
	"github.com/abhishekgit03/kubera/collector"
	"github.com/abhishekgit03/kubera/gemini"
)

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	cpuCost := parseFloat(os.Getenv("CPU_COST_PER_CORE_HOUR"), 0.048)
	memCost := parseFloat(os.Getenv("MEM_COST_PER_GB_HOUR"), 0.006)
	refreshSecs := parseInt(os.Getenv("REFRESH_INTERVAL_SECONDS"), 300)

	var col *collector.Collector
	var err error

	if os.Getenv("USE_MOCK") == "true" {
		col = collector.NewMock(cpuCost, memCost)
		log.Println("running with mock data (USE_MOCK=true)")
	} else {
		col, err = collector.New(cpuCost, memCost, time.Duration(refreshSecs)*time.Second)
		if err != nil {
			log.Fatalf("failed to initialize collector: %v", err)
		}
	}

	geminiClient, _ := gemini.New()

	srv := api.NewServer(col, geminiClient, frontendFS)
	log.Println("kubera listening on :8080")
	if err := srv.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
