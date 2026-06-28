import { useEffect, useState } from 'react'
import './DNS.css'

export default function DNS() {
  const [currentIP, setCurrentIP] = useState('—')
  const [triggering, setTriggering] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    fetch('https://api.ipify.org')
      .then(r => r.text())
      .then(setCurrentIP)
      .catch(() => setCurrentIP('Unable to fetch'))
  }, [])

  async function triggerUpdate() {
    setTriggering(true)
    setMessage('')
    const res = await fetch('/api/dns/update', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ip: currentIP }),
    })
    setTriggering(false)
    setMessage(res.ok ? 'DNS updated successfully.' : 'Update failed — check logs.')
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
        </div>
        <div className="dns-row">
          <span className="dns-label">Manual trigger</span>
          <button className="btn-primary" onClick={triggerUpdate} disabled={triggering || currentIP === '—'}>
            {triggering ? 'Updating…' : 'Update DNS now'}
          </button>
        </div>
        {message && (
          <p className={message.includes('success') ? 'msg-ok' : 'msg-err'}>{message}</p>
        )}
      </div>

      <div className="card" style={{ marginTop: '1rem' }}>
        <p className="hint">
          The DDNS script runs every 5 minutes via cron. It detects IP changes and calls the backend,
          which updates all active subdomain A records via your configured DNS provider.
        </p>
      </div>
    </div>
  )
}
