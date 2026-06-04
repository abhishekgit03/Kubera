import { useEffect, useState } from 'react'
import { fetchSnapshot, streamAnalysis } from './api/client'
import { Dashboard } from './components/Dashboard'
import type { ClusterSnapshot, Fix } from './types'

export default function App() {
  const [snapshot, setSnapshot] = useState<ClusterSnapshot | null>(null)
  const [narrative, setNarrative] = useState<string>('')
  const [fixes, setFixes] = useState<Fix[]>([])
  const [analyzing, setAnalyzing] = useState(false)
  const [lastAnalyzed, setLastAnalyzed] = useState<Date | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const refresh = () =>
      fetchSnapshot()
        .then(setSnapshot)
        .catch((e) => setError(String(e)))

    refresh()
    const id = setInterval(refresh, 60_000)
    return () => clearInterval(id)
  }, [])

  function handleAnalyze() {
    if (analyzing) return
    setAnalyzing(true)
    setNarrative('')
    setFixes([])

    const es = streamAnalysis()
    let buffer = ''

    es.addEventListener('token', (e: MessageEvent) => {
      buffer += e.data
      // Try to extract the JSON fix block in real-time
      const jsonMatch = buffer.match(/```json\n([\s\S]*?)\n```/)
      if (jsonMatch) {
        try {
          const parsed = JSON.parse(jsonMatch[1])
          setFixes(parsed.fixes ?? [])
        } catch {
          // incomplete JSON yet, keep buffering
        }
        setNarrative(buffer.replace(/```json[\s\S]*?```/, '').trim())
      } else {
        setNarrative(buffer)
      }
    })

    es.addEventListener('done', () => {
      es.close()
      setAnalyzing(false)
      setLastAnalyzed(new Date())
    })

    es.addEventListener('error', (e: MessageEvent) => {
      es.close()
      setAnalyzing(false)
      if (e.data) setError(`Analysis error: ${e.data}`)
    })

    es.onerror = () => {
      es.close()
      setAnalyzing(false)
    }
  }

  return (
    <Dashboard
      snapshot={snapshot}
      narrative={narrative}
      fixes={fixes}
      analyzing={analyzing}
      lastAnalyzed={lastAnalyzed}
      error={error}
      onAnalyze={handleAnalyze}
    />
  )
}
