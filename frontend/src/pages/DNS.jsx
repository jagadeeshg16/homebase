import { useEffect, useState } from 'react'
import './DNS.css'

export default function DNS() {
  const [currentIP, setCurrentIP] = useState('—')
  const [triggering, setTriggering] = useState(false)
  const [message, setMessage] = useState(null)
  const [events, setEvents] = useState([])

  useEffect(() => {
    fetch('https://api.ipify.org').then(r => r.text()).then(setCurrentIP).catch(() => setCurrentIP('unavailable'))
    loadEvents()
    const interval = setInterval(loadEvents, 15_000)
    return () => clearInterval(interval)
  }, [])

  async function loadEvents() {
    const res = await fetch('/api/dns/events', { credentials: 'include' })
    if (res.ok) setEvents(await res.json())
  }

  async function triggerUpdate() {
    setTriggering(true)
    setMessage(null)
    const res = await fetch('/api/dns/update', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ip: currentIP }),
    })
    setTriggering(false)
    setMessage({ ok: res.ok, text: res.ok ? 'DNS update triggered.' : 'Update failed.' })
    loadEvents()
  }

  async function retry(id) {
    await fetch(`/api/dns/retry/${id}`, { method: 'POST', credentials: 'include' })
    setTimeout(loadEvents, 1000)
  }

  return (
    <div>
      <h1 className="page-title">DNS</h1>

      <div className="card dns-card">
        <div className="dns-row">
          <span className="dns-label">Current public IP</span>
          <span className="dns-value mono">{currentIP}</span>
        </div>
        <div className="dns-row">
          <span className="dns-label">DDNS runs automatically every 5 min via cron</span>
          <button className="btn-primary" onClick={triggerUpdate} disabled={triggering || currentIP === '—'}>
            {triggering ? 'Updating…' : 'Update now'}
          </button>
        </div>
        {message && <p className={message.ok ? 'msg-ok' : 'msg-err'}>{message.text}</p>}
      </div>

      <h2 className="section-title" style={{ marginTop: '1.5rem' }}>DNS Events</h2>
      <div className="card">
        {events.length === 0 ? (
          <p className="muted">No DNS events yet.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Subdomain</th>
                <th>Operation</th>
                <th>Status</th>
                <th>Attempts</th>
                <th>Error</th>
                <th>Next retry</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {events.map(e => (
                <tr key={e.id}>
                  <td><strong>{e.subdomain}</strong></td>
                  <td><span className="badge badge-gray">{e.operation}</span></td>
                  <td><StatusBadge status={e.status} /></td>
                  <td>{e.attempts}</td>
                  <td className="error-cell">{e.error_msg || '—'}</td>
                  <td className="muted-sm">
                    {e.next_retry ? new Date(e.next_retry).toLocaleTimeString() : '—'}
                  </td>
                  <td>
                    {e.status === 'failed' && (
                      <button className="btn-ghost" onClick={() => retry(e.id)}>Retry</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

function StatusBadge({ status }) {
  const map = {
    success: 'badge-green',
    pending: 'badge-yellow',
    failed: 'badge-red',
  }
  return <span className={`badge ${map[status] || 'badge-gray'}`}>{status}</span>
}
