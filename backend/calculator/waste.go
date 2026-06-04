package calculator

import (
	"sort"

	"github.com/abhishekgit03/kubera/collector"
)

const hoursPerMonth = 730.0

// CPUWasteRatio returns actual/request. Returns 1.0 if request is 0 (no waste can be computed).
func CPUWasteRatio(actual, request int64) float64 {
	if request == 0 {
		return 1.0
	}
	r := float64(actual) / float64(request)
	if r > 1.0 {
		return 1.0
	}
	return r
}

// MemWasteRatio returns actual/request for memory.
func MemWasteRatio(actual, request int64) float64 {
	return CPUWasteRatio(actual, request)
}

// MonthlyCost computes USD/month for a workload based on resource requests × replicas.
func MonthlyCost(cpuRequestMillis, memRequestMB int64, replicas int32, cpuCostPerCoreHour, memCostPerGBHour float64) float64 {
	cpuCores := float64(cpuRequestMillis) / 1000.0
	memGB := float64(memRequestMB) / 1024.0
	r := float64(replicas)
	return (cpuCores*cpuCostPerCoreHour + memGB*memCostPerGBHour) * r * hoursPerMonth
}

// MonthlyWaste computes the idle portion cost: (request - actual) × replicas.
func MonthlyWaste(w *collector.WorkloadCost, cpuCostPerCoreHour, memCostPerGBHour float64) float64 {
	idleCPUMillis := w.CPURequestMillis - w.CPUActualMillis
	if idleCPUMillis < 0 {
		idleCPUMillis = 0
	}
	idleMemMB := w.MemRequestMB - w.MemActualMB
	if idleMemMB < 0 {
		idleMemMB = 0
	}

	idleCPUCores := float64(idleCPUMillis) / 1000.0
	idleMemGB := float64(idleMemMB) / 1024.0
	r := float64(w.Replicas)

	return (idleCPUCores*cpuCostPerCoreHour + idleMemGB*memCostPerGBHour) * r * hoursPerMonth
}

// EnrichSnapshot mutates a snapshot by computing all cost/waste fields.
func EnrichSnapshot(snap *collector.ClusterSnapshot, cpuCostPerCoreHour, memCostPerGBHour float64) {
	snap.TotalEstimatedCost = 0
	snap.TotalEstimatedWaste = 0

	for i := range snap.Workloads {
		w := &snap.Workloads[i]
		w.CPUWasteRatio = CPUWasteRatio(w.CPUActualMillis, w.CPURequestMillis)
		w.MemWasteRatio = MemWasteRatio(w.MemActualMB, w.MemRequestMB)
		w.EstimatedMonthlyCost = MonthlyCost(w.CPURequestMillis, w.MemRequestMB, w.Replicas, cpuCostPerCoreHour, memCostPerGBHour)
		w.EstimatedMonthlyWaste = MonthlyWaste(w, cpuCostPerCoreHour, memCostPerGBHour)

		snap.TotalEstimatedCost += w.EstimatedMonthlyCost
		snap.TotalEstimatedWaste += w.EstimatedMonthlyWaste
	}
}

// TopOffenders returns workloads sorted by EstimatedMonthlyWaste descending, capped at n.
func TopOffenders(snap *collector.ClusterSnapshot, n int) []collector.WorkloadCost {
	sorted := make([]collector.WorkloadCost, len(snap.Workloads))
	copy(sorted, snap.Workloads)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EstimatedMonthlyWaste > sorted[j].EstimatedMonthlyWaste
	})

	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}
