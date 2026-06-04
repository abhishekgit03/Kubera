import { ChevronDown, ChevronUp, Maximize2 } from 'lucide-react'
import { useState } from 'react'
import type { WorkloadCost } from '../types'

interface Props {
  workloads: WorkloadCost[]
  onExpand?: () => void
  expanded?: boolean
}

type SortKey = keyof WorkloadCost

function wasteBadge(ratio: number) {
  if (ratio < 0.2) return 'bg-red-50 text-red-600 border border-red-200'
  if (ratio < 0.5) return 'bg-amber-50 text-amber-700 border border-amber-200'
  return 'bg-emerald-50 text-emerald-700 border border-emerald-200'
}

function fmt(n: number) {
  return `$${n.toLocaleString('en-US', { maximumFractionDigits: 0 })}`
}

export function WorkloadTable({ workloads, onExpand, expanded }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('estimated_monthly_waste')
  const [asc, setAsc] = useState(false)

  function toggleSort(key: SortKey) {
    if (sortKey === key) setAsc(!asc)
    else { setSortKey(key); setAsc(false) }
  }

  const sorted = [...workloads].sort((a, b) => {
    const av = a[sortKey], bv = b[sortKey]
    if (typeof av === 'number' && typeof bv === 'number') return asc ? av - bv : bv - av
    return asc ? String(av).localeCompare(String(bv)) : String(bv).localeCompare(String(av))
  })

  function Th({ label, col, className = '' }: { label: string; col: SortKey; className?: string }) {
    const active = sortKey === col
    return (
      <th
        className={`px-3 py-2.5 text-left text-xs font-semibold text-gray-500 uppercase tracking-widest cursor-pointer hover:text-gray-800 whitespace-nowrap select-none ${className}`}
        onClick={() => toggleSort(col)}
      >
        <span className="flex items-center gap-1">
          {label}
          {active
            ? (asc ? <ChevronUp className="w-3 h-3" style={{ color: '#00B4D8' }} /> : <ChevronDown className="w-3 h-3" style={{ color: '#00B4D8' }} />)
            : null}
        </span>
      </th>
    )
  }

  const tableContent = (
    <table className="w-full text-sm">
      <thead className="sticky top-0 bg-white/95 backdrop-blur z-10">
        <tr className="border-b border-gray-200">
          <Th label="Namespace" col="namespace" />
          <Th label="Workload" col="workload_name" />
          <Th label="Type" col="workload_type" className="hidden sm:table-cell" />
          <Th label="Rep." col="replicas" className="hidden md:table-cell" />
          <Th label="CPU Req" col="cpu_request_millicores" className="hidden lg:table-cell" />
          <Th label="CPU Act" col="cpu_actual_millicores" className="hidden lg:table-cell" />
          <Th label="Mem Req" col="mem_request_mb" className="hidden xl:table-cell" />
          <Th label="Mem Act" col="mem_actual_mb" className="hidden xl:table-cell" />
          <Th label="CPU Util" col="cpu_waste_ratio" />
          <Th label="Cost/mo" col="estimated_monthly_cost" className="hidden md:table-cell" />
          <Th label="Waste/mo" col="estimated_monthly_waste" />
        </tr>
      </thead>
      <tbody>
        {sorted.map((w, i) => (
          <tr
            key={`${w.namespace}/${w.workload_name}`}
            className={`border-t border-gray-100 hover:bg-gray-50 transition-colors ${i % 2 === 0 ? 'bg-white' : 'bg-gray-50/50'}`}
          >
            <td className="px-3 py-2.5">
              <span className="inline-block bg-gray-100 text-gray-600 text-xs px-1.5 py-0.5 rounded max-w-[90px] truncate font-mono" title={w.namespace}>
                {w.namespace}
              </span>
            </td>
            <td className="px-3 py-2.5 font-semibold text-gray-900">
              <span className="block truncate max-w-[120px]" title={w.workload_name}>{w.workload_name}</span>
            </td>
            <td className="px-3 py-2.5 hidden sm:table-cell text-xs text-gray-500">{w.workload_type}</td>
            <td className="px-3 py-2.5 text-center text-gray-700 hidden md:table-cell">{w.replicas}</td>
            <td className="px-3 py-2.5 text-gray-500 text-xs hidden lg:table-cell font-mono">{w.cpu_request_millicores}m</td>
            <td className="px-3 py-2.5 text-gray-500 text-xs hidden lg:table-cell font-mono">{w.cpu_actual_millicores}m</td>
            <td className="px-3 py-2.5 text-gray-500 text-xs hidden xl:table-cell font-mono">{w.mem_request_mb} Mi</td>
            <td className="px-3 py-2.5 text-gray-500 text-xs hidden xl:table-cell font-mono">{w.mem_actual_mb} Mi</td>
            <td className="px-3 py-2.5">
              <span className={`inline-block text-xs px-2 py-0.5 rounded-full font-medium ${wasteBadge(w.cpu_waste_ratio)}`}>
                {(w.cpu_waste_ratio * 100).toFixed(0)}%
              </span>
            </td>
            <td className="px-3 py-2.5 text-gray-600 text-xs hidden md:table-cell">{fmt(w.estimated_monthly_cost)}</td>
            <td className="px-3 py-2.5 text-xs font-semibold" style={{ color: w.cpu_waste_ratio < 0.3 ? '#DC2626' : '#374151' }}>
              {fmt(w.estimated_monthly_waste)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )

  if (expanded) return tableContent

  return (
    <div className="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
      <div className="px-5 py-4 border-b border-gray-200 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-700 uppercase tracking-widest">Workloads</h2>
        {onExpand && (
          <button onClick={onExpand} className="text-gray-400 hover:text-gray-600 transition-colors" title="Expand">
            <Maximize2 className="w-4 h-4" />
          </button>
        )}
      </div>
      <div className="overflow-auto max-h-72">
        {tableContent}
      </div>
    </div>
  )
}
