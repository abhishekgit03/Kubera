package collector

import (
	"time"
)

// NewMock returns a Collector backed by hardcoded data for local development.
func NewMock(cpuCost, memCost float64) *Collector {
	c := &Collector{
		cpuCostPerCoreHour: cpuCost,
		memCostPerGBHour:   memCost,
	}

	snap := &ClusterSnapshot{
		CollectedAt:      time.Now().UTC().Format(time.RFC3339),
		TotalNamespaces:  4,
		TotalPods:        32,
		MetricsAvailable: true,
		Workloads: []WorkloadCost{
			{
				Namespace: "production", WorkloadName: "api-server", WorkloadType: "Deployment",
				Replicas: 8, CPURequestMillis: 2000, MemRequestMB: 2048,
				CPULimitMillis: 4000, MemLimitMB: 4096,
				CPUActualMillis: 180, MemActualMB: 512,
				CPUWasteRatio: 0.09, MemWasteRatio: 0.25,
				EstimatedMonthlyCost: 560.64, EstimatedMonthlyWaste: 420.48,
			},
			{
				Namespace: "production", WorkloadName: "worker", WorkloadType: "Deployment",
				Replicas: 5, CPURequestMillis: 1000, MemRequestMB: 1024,
				CPULimitMillis: 2000, MemLimitMB: 2048,
				CPUActualMillis: 95, MemActualMB: 380,
				CPUWasteRatio: 0.095, MemWasteRatio: 0.37,
				EstimatedMonthlyCost: 175.2, EstimatedMonthlyWaste: 140.16,
			},
			{
				Namespace: "staging", WorkloadName: "api-server", WorkloadType: "Deployment",
				Replicas: 3, CPURequestMillis: 2000, MemRequestMB: 2048,
				CPULimitMillis: 4000, MemLimitMB: 4096,
				CPUActualMillis: 40, MemActualMB: 200,
				CPUWasteRatio: 0.02, MemWasteRatio: 0.10,
				EstimatedMonthlyCost: 210.24, EstimatedMonthlyWaste: 199.73,
			},
			{
				Namespace: "staging", WorkloadName: "postgres", WorkloadType: "StatefulSet",
				Replicas: 1, CPURequestMillis: 500, MemRequestMB: 512,
				CPULimitMillis: 1000, MemLimitMB: 1024,
				CPUActualMillis: 120, MemActualMB: 390,
				CPUWasteRatio: 0.24, MemWasteRatio: 0.76,
				EstimatedMonthlyCost: 17.52, EstimatedMonthlyWaste: 5.26,
			},
			{
				Namespace: "monitoring", WorkloadName: "prometheus", WorkloadType: "StatefulSet",
				Replicas: 1, CPURequestMillis: 1000, MemRequestMB: 4096,
				CPULimitMillis: 2000, MemLimitMB: 8192,
				CPUActualMillis: 420, MemActualMB: 3200,
				CPUWasteRatio: 0.42, MemWasteRatio: 0.78,
				EstimatedMonthlyCost: 64.68, EstimatedMonthlyWaste: 22.64,
			},
			{
				Namespace: "monitoring", WorkloadName: "grafana", WorkloadType: "Deployment",
				Replicas: 2, CPURequestMillis: 500, MemRequestMB: 512,
				CPULimitMillis: 1000, MemLimitMB: 1024,
				CPUActualMillis: 30, MemActualMB: 180,
				CPUWasteRatio: 0.06, MemWasteRatio: 0.35,
				EstimatedMonthlyCost: 35.04, EstimatedMonthlyWaste: 28.03,
			},
			{
				Namespace: "default", WorkloadName: "redis", WorkloadType: "Deployment",
				Replicas: 3, CPURequestMillis: 500, MemRequestMB: 1024,
				CPULimitMillis: 1000, MemLimitMB: 2048,
				CPUActualMillis: 85, MemActualMB: 900,
				CPUWasteRatio: 0.17, MemWasteRatio: 0.88,
				EstimatedMonthlyCost: 52.56, EstimatedMonthlyWaste: 5.26,
			},
		},
	}

	for _, w := range snap.Workloads {
		snap.TotalEstimatedCost += w.EstimatedMonthlyCost
		snap.TotalEstimatedWaste += w.EstimatedMonthlyWaste
	}
	snap.TotalWorkloads = len(snap.Workloads)

	c.latest = snap
	return c
}
