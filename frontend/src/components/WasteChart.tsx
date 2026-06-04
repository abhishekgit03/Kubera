import { Maximize2 } from 'lucide-react'
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import type { WorkloadCost } from '../types'

interface Props {
  workloads: WorkloadCost[]
  onExpand?: () => void
  expanded?: boolean
}

interface NsData { namespace: string; cost: number; waste: number }

function fmt(n: number) {
  return `$${n.toLocaleString('en-US', { maximumFractionDigits: 0 })}`
}

function truncate(s: string, max: number) {
  return s.length > max ? s.slice(0, max) + '…' : s
}

export function WasteChart({ workloads, onExpand, expanded }: Props) {
  const byNs: Record<string, NsData> = {}
  for (const w of workloads) {
    if (!byNs[w.namespace]) byNs[w.namespace] = { namespace: w.namespace, cost: 0, waste: 0 }
    byNs[w.namespace].cost += w.estimated_monthly_cost
    byNs[w.namespace].waste += w.estimated_monthly_waste
  }
  const data = Object.values(byNs).sort((a, b) => b.waste - a.waste)
  const chartHeight = expanded ? 480 : 280

  return (
    <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm h-full">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-semibold text-gray-700 uppercase tracking-widest">
          Cost by Namespace
        </h2>
        {onExpand && (
          <button onClick={onExpand} className="text-gray-400 hover:text-gray-600 transition-colors" title="Expand">
            <Maximize2 className="w-4 h-4" />
          </button>
        )}
      </div>
      <ResponsiveContainer width="100%" height={chartHeight}>
        <BarChart data={data} margin={{ top: 4, right: 8, bottom: expanded ? 24 : 4, left: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
          <XAxis
            dataKey="namespace"
            tick={{ fill: '#6B7280', fontSize: 11 }}
            axisLine={{ stroke: '#D1D5DB' }}
            tickLine={false}
            tickFormatter={(v) => expanded ? v : truncate(v, 8)}
          />
          <YAxis
            tickFormatter={fmt}
            tick={{ fill: '#6B7280', fontSize: 11 }}
            axisLine={false}
            tickLine={false}
            width={64}
          />
          <Tooltip
            contentStyle={{ background: '#fff', border: '1px solid #E5E7EB', borderRadius: 8, boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
            labelStyle={{ color: '#111827', fontWeight: 600 }}
            itemStyle={{ color: '#374151' }}
            formatter={(v: number) => fmt(v)}
          />
          {/* Cost bars — brand dark navy */}
          <Bar dataKey="cost" name="Monthly Cost" radius={[3, 3, 0, 0]}>
            {data.map((_, i) => <Cell key={i} fill="#1A3A5C" />)}
          </Bar>
          {/* Waste bars — brand teal accent */}
          <Bar dataKey="waste" name="Monthly Waste" radius={[3, 3, 0, 0]}>
            {data.map((_, i) => <Cell key={i} fill="#00B4D8" />)}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
      <div className="flex gap-4 mt-2 justify-center text-xs text-gray-500">
        <span className="flex items-center gap-1.5">
          <span className="w-3 h-3 rounded-sm inline-block" style={{ backgroundColor: '#1A3A5C' }} /> Monthly Cost
        </span>
        <span className="flex items-center gap-1.5">
          <span className="w-3 h-3 rounded-sm inline-block" style={{ backgroundColor: '#00B4D8' }} /> Monthly Waste
        </span>
      </div>
    </div>
  )
}
