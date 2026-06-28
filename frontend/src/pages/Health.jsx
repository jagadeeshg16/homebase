import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import './Health.css'

export default function Health() {
  const [health, setHealth] = useState([])
  const [loading, setLoading] = useState(true)

  async function load() {
    const data = await api.health()
    setHealth(data || [])
    setLoading(false)
  }

  useEffect(() => {
    load()
    const interval = setInterval(load, 30_000)
    return () => clearInterval(interval)
  }, [])

  if (loading) return <p className="muted">Loading…</p>

  return (
    <div>
      <div className="health-header">
        <h1 className="page-title">Health</h1>
        <button className="btn-ghost" onClick={load}>Refresh</button>
      </div>

      {health.length === 0 ? (
        <p className="muted">No active subdomains to check.</p>
      ) : (
        <div className="health-list">
          {health.map(h => (
            <HealthRow key={h.subdomain_id} entry={h} />
          ))}
        </div>
      )}
    </div>
  )
}

function HealthRow({ entry }) {
  const ok = entry.status >= 200 && entry.status < 400
  const down = entry.status === 0 || entry.status >= 400

  return (
    <div className="card health-row">
      <div className="health-left">
        <span className={`dot ${ok ? 'dot-green' : down ? 'dot-red' : 'dot-gray'}`} />
        <strong>{entry.subdomain_name}</strong>
      </div>
      <div className="health-meta">
        <span className={`badge ${ok ? 'badge-green' : 'badge-red'}`}>
          {entry.status === 0 ? 'Down' : entry.status}
        </span>
        <span className="muted-sm">{entry.response_ms}ms</span>
        <span className="muted-sm">{new Date(entry.checked_at).toLocaleTimeString()}</span>
      </div>
    </div>
  )
}
