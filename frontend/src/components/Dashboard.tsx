import { AlertTriangle, BarChart3, DollarSign, Sparkles, TrendingDown } from 'lucide-react'
import { useState } from 'react'
import type { ClusterSnapshot, Fix } from '../types'
import { FixSuggestions } from './FixSuggestions'
import { KuberaIcon } from './KuberaIcon'
import { Modal } from './Modal'
import { NarrativePanel } from './NarrativePanel'
import { WasteChart } from './WasteChart'
import { WorkloadTable } from './WorkloadTable'

interface Props {
  snapshot: ClusterSnapshot | null
  narrative: string
  fixes: Fix[]
  analyzing: boolean
  lastAnalyzed: Date | null
  error: string | null
  onAnalyze: () => void
}

function MetricCard({ label, value, sub, icon }: {
  label: string; value: string; sub?: string; icon: React.ReactNode
}) {
  return (
    <div className="bg-white border border-gray-200 rounded-xl p-5 flex items-start gap-4 shadow-sm">
      <div className="p-2 rounded-lg bg-gray-50 border border-gray-100">{icon}</div>
      <div>
        <p className="text-xs text-gray-500 uppercase tracking-widest font-medium">{label}</p>
        <p className="text-2xl font-bold mt-1 text-gray-900">{value}</p>
        {sub && <p className="text-xs text-gray-400 mt-1">{sub}</p>}
      </div>
    </div>
  )
}

function fmt(n: number) {
  return n.toLocaleString('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 })
}

function dataAge(collectedAt: string): string {
  const ago = Math.floor((Date.now() - new Date(collectedAt).getTime()) / 1000)
  if (ago < 60) return `${ago}s ago`
  if (ago < 3600) return `${Math.floor(ago / 60)}m ago`
  return `${Math.floor(ago / 3600)}h ago`
}

export function Dashboard({ snapshot, narrative, fixes, analyzing, lastAnalyzed, error, onAnalyze }: Props) {
  const [chartExpanded, setChartExpanded] = useState(false)
  const [tableExpanded, setTableExpanded] = useState(false)

  const stale = snapshot && (Date.now() - new Date(snapshot.collected_at).getTime()) > 10 * 60 * 1000
  const wastePct = snapshot
    ? ((snapshot.total_estimated_waste / (snapshot.total_estimated_cost || 1)) * 100).toFixed(0)
    : null

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header — full brand navy, one place for the palette */}
      <header className="sticky top-0 z-10 shadow-md" style={{ backgroundColor: '#0F2744' }}>
        <div className="max-w-7xl mx-auto px-6 py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <KuberaIcon size={30} />
            <div>
              <span className="text-lg font-bold tracking-tight text-white">Kubera</span>
              <p className="text-xs leading-none mt-0.5" style={{ color: '#7BAFD4' }}>
                Know exactly what your cluster costs. And why.
              </p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            {snapshot && (
              <span className={`text-xs ${stale ? 'text-amber-300' : 'text-white/50'}`}>
                {stale && <AlertTriangle className="inline w-3 h-3 mr-1" />}
                Data: {dataAge(snapshot.collected_at)}
              </span>
            )}
            <button
              onClick={onAnalyze}
              disabled={analyzing || !snapshot}
              className="flex items-center gap-2 text-white text-sm font-semibold px-4 py-2 rounded-lg transition-all disabled:opacity-40 disabled:cursor-not-allowed"
              style={{ backgroundColor: analyzing || !snapshot ? undefined : '#00B4D8' }}
              onMouseEnter={e => { if (!analyzing && snapshot) (e.currentTarget as HTMLButtonElement).style.backgroundColor = '#0097b8' }}
              onMouseLeave={e => { if (!analyzing && snapshot) (e.currentTarget as HTMLButtonElement).style.backgroundColor = '#00B4D8' }}
            >
              <Sparkles className="w-4 h-4" />
              {analyzing ? 'Analyzing…' : 'Analyze'}
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8 space-y-6">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        {/* Metric cards */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <MetricCard
            label="Est. Monthly Cost"
            value={snapshot ? fmt(snapshot.total_estimated_cost) : '—'}
            sub={snapshot ? `${snapshot.total_workloads} workloads · ${snapshot.total_pods} pods` : undefined}
            icon={<DollarSign className="w-5 h-5 text-gray-400" />}
          />
          <MetricCard
            label="Est. Monthly Waste"
            value={snapshot ? fmt(snapshot.total_estimated_waste) : '—'}
            sub={snapshot && wastePct ? `${wastePct}% of budget is idle` : undefined}
            icon={<TrendingDown className="w-5 h-5 text-gray-400" />}
          />
          <MetricCard
            label="Savings Opportunity"
            value={snapshot ? fmt(snapshot.total_estimated_waste) : '—'}
            sub="if top offenders are right-sized"
            icon={<BarChart3 className="w-5 h-5 text-emerald-500" />}
          />
        </div>

        {/* Chart + Table */}
        {snapshot && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <WasteChart workloads={snapshot.workloads} onExpand={() => setChartExpanded(true)} />
            <WorkloadTable workloads={snapshot.workloads} onExpand={() => setTableExpanded(true)} />
          </div>
        )}

        {/* Narrative */}
        {(narrative || analyzing) && (
          <NarrativePanel narrative={narrative} analyzing={analyzing} lastAnalyzed={lastAnalyzed} />
        )}

        {/* Fix suggestions */}
        {fixes.length > 0 && <FixSuggestions fixes={fixes} />}
      </main>

      {chartExpanded && snapshot && (
        <Modal title="Cost by Namespace" onClose={() => setChartExpanded(false)}>
          <WasteChart workloads={snapshot.workloads} expanded />
        </Modal>
      )}
      {tableExpanded && snapshot && (
        <Modal title="All Workloads" onClose={() => setTableExpanded(false)}>
          <WorkloadTable workloads={snapshot.workloads} expanded />
        </Modal>
      )}
    </div>
  )
}
