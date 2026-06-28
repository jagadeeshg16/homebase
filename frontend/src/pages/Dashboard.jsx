import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import './Dashboard.css'

export default function Dashboard() {
  const [subdomains, setSubdomains] = useState([])
  const [health, setHealth] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([api.subdomains(), api.health()]).then(([s, h]) => {
      setSubdomains(s || [])
      setHealth(h || [])
      setLoading(false)
    })
  }, [])

  const active   = subdomains.filter(s => s.is_active).length
  const inactive = subdomains.filter(s => !s.is_active).length
  const healthy  = health.filter(h => h.status >= 200 && h.status < 400).length
  const down     = health.filter(h => h.status === 0 || h.status >= 400).length

  if (loading) return <p className="muted">Loading…</p>

  return (
    <div>
      <h1 className="page-title">Dashboard</h1>
      <div className="stat-grid">
        <StatCard label="Active subdomains" value={active} color="green" />
        <StatCard label="Inactive" value={inactive} color="gray" />
        <StatCard label="Healthy" value={healthy} color="green" />
        <StatCard label="Down" value={down} color={down > 0 ? 'red' : 'gray'} />
      </div>

      <h2 className="section-title">Subdomains</h2>
      <div className="card">
        {subdomains.length === 0 ? (
          <p className="muted">No subdomains yet — drop a folder in sites/ to get started.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Domain</th>
                <th>Status</th>
                <th>Visibility</th>
              </tr>
            </thead>
            <tbody>
              {subdomains.map(s => (
                <tr key={s.id}>
                  <td>{s.name}</td>
                  <td><span className="domain">{s.full_domain}</span></td>
                  <td>
                    <span className={`badge ${s.is_active ? 'badge-green' : 'badge-gray'}`}>
                      {s.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td>
                    <span className={`badge ${s.is_public ? 'badge-green' : 'badge-yellow'}`}>
                      {s.is_public ? 'Public' : 'Private'}
                    </span>
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

function StatCard({ label, value, color }) {
  const colors = { green: 'var(--success)', red: 'var(--danger)', gray: 'var(--muted)' }
  return (
    <div className="stat-card card">
      <div className="stat-value" style={{ color: colors[color] }}>{value}</div>
      <div className="stat-label">{label}</div>
    </div>
  )
}
