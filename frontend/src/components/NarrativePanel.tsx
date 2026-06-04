import { Clock } from 'lucide-react'
import ReactMarkdown from 'react-markdown'

interface Props {
  narrative: string
  analyzing: boolean
  lastAnalyzed: Date | null
}

function timeAgo(d: Date): string {
  const mins = Math.floor((Date.now() - d.getTime()) / 60000)
  if (mins < 1) return 'just now'
  if (mins === 1) return '1 minute ago'
  return `${mins} minutes ago`
}

export function NarrativePanel({ narrative, analyzing, lastAnalyzed }: Props) {
  const markdownText = narrative.replace(/```json[\s\S]*?```/g, '').trim()

  return (
    <div className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
      <div className="flex items-center justify-between mb-5">
        <h2 className="text-sm font-semibold text-gray-700 uppercase tracking-widest">
          AI Cost Narrative
        </h2>
        {lastAnalyzed && !analyzing && (
          <span className="flex items-center gap-1 text-xs text-gray-400">
            <Clock className="w-3 h-3" />
            Last analyzed: {timeAgo(lastAnalyzed)}
          </span>
        )}
        {analyzing && (
          <span className="text-xs font-medium animate-pulse" style={{ color: '#00B4D8' }}>
            Generating…
          </span>
        )}
      </div>

      <div className="prose prose-sm max-w-none text-gray-800
        prose-headings:text-gray-900 prose-headings:font-semibold
        prose-h2:text-base prose-h2:mt-5 prose-h2:mb-2 prose-h2:border-b prose-h2:border-gray-100 prose-h2:pb-1
        prose-h3:text-sm prose-h3:mt-4 prose-h3:mb-1
        prose-p:text-gray-700 prose-p:leading-relaxed prose-p:my-2
        prose-strong:text-gray-900 prose-strong:font-semibold
        prose-ul:my-2 prose-li:my-0.5 prose-li:text-gray-700
        prose-code:bg-gray-100 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-xs prose-code:font-mono prose-code:text-gray-800
        prose-pre:bg-gray-50 prose-pre:border prose-pre:border-gray-200 prose-pre:rounded-lg">
        <ReactMarkdown>{markdownText}</ReactMarkdown>
        {analyzing && (
          <span
            className="inline-block w-2 h-4 ml-0.5 animate-pulse align-text-bottom rounded-sm"
            style={{ backgroundColor: '#00B4D8' }}
          />
        )}
      </div>
    </div>
  )
}
