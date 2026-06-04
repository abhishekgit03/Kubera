package collector

type WorkloadCost struct {
	Namespace    string `json:"namespace"`
	WorkloadName string `json:"workload_name"`
	WorkloadType string `json:"workload_type"` // Deployment, StatefulSet, DaemonSet

	Replicas int32 `json:"replicas"`

	CPURequestMillis int64 `json:"cpu_request_millicores"`
	MemRequestMB     int64 `json:"mem_request_mb"`

	CPULimitMillis int64 `json:"cpu_limit_millicores"`
	MemLimitMB     int64 `json:"mem_limit_mb"`

	CPUActualMillis int64 `json:"cpu_actual_millicores"`
	MemActualMB     int64 `json:"mem_actual_mb"`

	CPUWasteRatio float64 `json:"cpu_waste_ratio"`
	MemWasteRatio float64 `json:"mem_waste_ratio"`

	EstimatedMonthlyCost  float64 `json:"estimated_monthly_cost"`
	EstimatedMonthlyWaste float64 `json:"estimated_monthly_waste"`
}

type ClusterSnapshot struct {
	CollectedAt         string         `json:"collected_at"`
	TotalNamespaces     int            `json:"total_namespaces"`
	TotalWorkloads      int            `json:"total_workloads"`
	TotalPods           int            `json:"total_pods"`
	TotalEstimatedCost  float64        `json:"total_estimated_cost"`
	TotalEstimatedWaste float64        `json:"total_estimated_waste"`
	Workloads           []WorkloadCost `json:"workloads"`
	MetricsAvailable    bool           `json:"metrics_available"`
}
