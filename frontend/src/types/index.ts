export interface WorkloadCost {
  namespace: string
  workload_name: string
  workload_type: string
  replicas: number
  cpu_request_millicores: number
  cpu_limit_millicores: number
  cpu_actual_millicores: number
  cpu_waste_ratio: number
  mem_request_mb: number
  mem_limit_mb: number
  mem_actual_mb: number
  mem_waste_ratio: number
  estimated_monthly_cost: number
  estimated_monthly_waste: number
}

export interface ClusterSnapshot {
  collected_at: string
  total_namespaces: number
  total_workloads: number
  total_pods: number
  total_estimated_cost: number
  total_estimated_waste: number
  workloads: WorkloadCost[]
  metrics_available: boolean
}

export interface Fix {
  workload: string
  namespace: string
  current_cpu_request: string
  suggested_cpu_request: string
  current_replicas: number
  suggested_replicas: number
  monthly_saving_usd: number
}
