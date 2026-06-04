import { Check, Copy, Wrench } from 'lucide-react'
import { useState } from 'react'
import type { Fix } from '../types'

interface Props { fixes: Fix[] }

function patchCommand(fix: Fix): string {
  const patch = {
    spec: {
      replicas: fix.suggested_replicas,
      template: {
        spec: {
          containers: [{
            name: fix.workload,
            resources: { requests: { cpu: fix.suggested_cpu_request } },
          }],
        },
      },
    },
  }
  return `kubectl patch deployment ${fix.workload} -n ${fix.namespace} -p '${JSON.stringify(patch)}'`
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  function handleCopy() {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }
  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-700 transition-colors"
    >
      {copied
        ? <><Check className="w-3.5 h-3.5 text-emerald-500" /> Copied</>
        : <><Copy className="w-3.5 h-3.5" /> Copy</>}
    </button>
  )
}

export function FixSuggestions({ fixes }: Props) {
  return (
    <div className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
      <div className="flex items-center gap-2 mb-4">
        <Wrench className="w-4 h-4 text-emerald-500" />
        <h2 className="text-sm font-semibold text-gray-700 uppercase tracking-widest">Suggested Fixes</h2>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {fixes.map((fix, i) => (
          <div key={i} className="bg-gray-50 border border-gray-200 rounded-lg p-4 space-y-3">
            <div className="flex items-start justify-between">
              <div>
                <p className="font-semibold text-gray-900">{fix.workload}</p>
                <p className="text-xs text-gray-500 font-mono">{fix.namespace}</p>
              </div>
              <span className="text-emerald-600 font-bold text-sm">
                -${fix.monthly_saving_usd.toLocaleString('en-US', { maximumFractionDigits: 0 })}/mo
              </span>
            </div>

            <div className="space-y-1 text-xs">
              <div className="flex justify-between">
                <span className="text-gray-500">CPU request</span>
                <span>
                  <span className="text-gray-400 line-through mr-1">{fix.current_cpu_request}</span>
                  <span className="font-medium" style={{ color: '#00B4D8' }}>→ {fix.suggested_cpu_request}</span>
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Replicas</span>
                <span>
                  <span className="text-gray-400 line-through mr-1">{fix.current_replicas}</span>
                  <span className="font-medium" style={{ color: '#00B4D8' }}>→ {fix.suggested_replicas}</span>
                </span>
              </div>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-2.5">
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs text-gray-400 font-mono">kubectl patch</span>
                <CopyButton text={patchCommand(fix)} />
              </div>
              <code className="text-xs text-gray-600 break-all leading-relaxed font-mono block">
                {patchCommand(fix)}
              </code>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
