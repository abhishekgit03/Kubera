package prompt

import (
	"encoding/json"
	"fmt"

	"github.com/abhishekgit03/kubera/calculator"
	"github.com/abhishekgit03/kubera/collector"
)

type clusterSummary struct {
	TotalWorkloads   int     `json:"total_workloads"`
	TotalNamespaces  int     `json:"total_namespaces"`
	OverallWastePct  float64 `json:"overall_waste_pct"`
	MetricsAvailable bool    `json:"metrics_available"`
}

type offenderSummary struct {
	Namespace      string  `json:"namespace"`
	Workload       string  `json:"workload"`
	WorkloadType   string  `json:"type"`
	Replicas       int32   `json:"replicas"`
	CPURequestM    int64   `json:"cpu_request_millicores"`
	CPUActualM     int64   `json:"cpu_actual_millicores"`
	CPUUtilPct     float64 `json:"cpu_utilization_pct"`
	MemRequestMB   int64   `json:"mem_request_mb"`
	MemActualMB    int64   `json:"mem_actual_mb"`
	MemUtilPct     float64 `json:"mem_utilization_pct"`
	WasteSharePct  float64 `json:"waste_share_of_cluster_pct"` // this workload's waste as % of total cluster waste
}

type promptPayload struct {
	ClusterSummary clusterSummary    `json:"cluster_summary"`
	TopOffenders   []offenderSummary `json:"top_offenders"`
}

const systemPrompt = `You are Kubera, a Kubernetes resource efficiency analyst embedded inside a cluster.
You have been given structured utilization data from the Kubernetes API and Metrics Server.

Your job is to write a RESOURCE EFFICIENCY NARRATIVE in Markdown format.
Write it like a senior SRE briefing an engineering manager — specific, direct, no fluff.

## Rules

**Format:** Output valid Markdown. Use ## headings, **bold** for workload names, and bullet points for fix suggestions.

**Language:** Talk only in percentages and utilization ratios. Never mention dollar amounts or currency.
The cost fields in the data are for relative ranking only — they are not billing-accurate and must not be cited.
Instead of "this costs $X/month", say "this workload is consuming Y% of cluster CPU capacity while using only Z% of what it requests".

**Content:**
- State the cluster-wide waste picture first (one sentence with the overall_waste_pct)
- For each top offender: name it, state the exact utilization numbers, explain WHY it's wasteful
  (over-replicated, CPU request far exceeds actual usage, no limits set, staging env running at prod scale, etc.)
- For each offender, give the exact fix: specific millicores/Mi values to set, or replica count to reduce to
- End with a summary of how many workloads need attention and the aggregate waste percentage that could be reclaimed

**If metrics are not available** (metrics_available: false): note this clearly and base analysis on request sizes and replica counts only.

**Do not be vague.** Do not say "consider optimizing". Say "reduce cpu request from 2000m to 300m".

**Length:** Keep the narrative under 500 words.

After the narrative, output a fenced ` + "`" + `json` + "`" + ` block with this exact structure (no extra fields):
` + "```" + `json
{ "fixes": [ { "workload": "", "namespace": "", "current_cpu_request": "", "suggested_cpu_request": "", "current_replicas": 0, "suggested_replicas": 0, "monthly_saving_usd": 0 } ] }
` + "```"

// Build constructs the full prompt string to send to Gemini.
func Build(snap *collector.ClusterSnapshot) string {
	offenders := calculator.TopOffenders(snap, 5)

	var totalWaste, totalCost float64
	for _, w := range snap.Workloads {
		totalCost += w.EstimatedMonthlyCost
		totalWaste += w.EstimatedMonthlyWaste
	}
	overallWastePct := 0.0
	if totalCost > 0 {
		overallWastePct = (totalWaste / totalCost) * 100
	}

	summaries := make([]offenderSummary, len(offenders))
	for i, w := range offenders {
		cpuUtilPct := 0.0
		if w.CPURequestMillis > 0 {
			cpuUtilPct = float64(w.CPUActualMillis) / float64(w.CPURequestMillis) * 100
		}
		memUtilPct := 0.0
		if w.MemRequestMB > 0 {
			memUtilPct = float64(w.MemActualMB) / float64(w.MemRequestMB) * 100
		}
		wasteSharePct := 0.0
		if totalWaste > 0 {
			wasteSharePct = (w.EstimatedMonthlyWaste / totalWaste) * 100
		}
		summaries[i] = offenderSummary{
			Namespace:     w.Namespace,
			Workload:      w.WorkloadName,
			WorkloadType:  w.WorkloadType,
			Replicas:      w.Replicas,
			CPURequestM:   w.CPURequestMillis,
			CPUActualM:    w.CPUActualMillis,
			CPUUtilPct:    cpuUtilPct,
			MemRequestMB:  w.MemRequestMB,
			MemActualMB:   w.MemActualMB,
			MemUtilPct:    memUtilPct,
			WasteSharePct: wasteSharePct,
		}
	}

	payload := promptPayload{
		ClusterSummary: clusterSummary{
			TotalWorkloads:   snap.TotalWorkloads,
			TotalNamespaces:  snap.TotalNamespaces,
			OverallWastePct:  overallWastePct,
			MetricsAvailable: snap.MetricsAvailable,
		},
		TopOffenders: summaries,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return fmt.Sprintf("%s\n\nCluster data:\n%s", systemPrompt, string(data))
}
