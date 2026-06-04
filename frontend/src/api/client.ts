import type { ClusterSnapshot } from '../types'

export async function fetchSnapshot(): Promise<ClusterSnapshot> {
  const res = await fetch('/api/snapshot')
  if (!res.ok) throw new Error(`Failed to fetch snapshot: ${res.status}`)
  return res.json()
}

export function streamAnalysis(): EventSource {
  return new EventSource('/api/analyze')
}
