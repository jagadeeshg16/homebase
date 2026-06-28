import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import './Subdomains.css'

export default function Subdomains() {
  const [subs, setSubs] = useState([])
  const [loading, setLoading] = useState(true)
  const [newName, setNewName] = useState('')
  const [newType, setNewType] = useState('static')
  const [newProxyURL, setNewProxyURL] = useState('')
  const [creating, setCreating] = useState(false)
  const [privacyModal, setPrivacyModal] = useState(null)

  async function load() {
    const data = await api.subdomains()
    setSubs(data || [])
    setLoading(false)
  }

  useEffect(() => { load() }, [])

  async function create(e) {
    e.preventDefault()
    if (!newName.trim()) return
    setCreating(true)
    await api.createSubdomain({
      name: newName.trim(),
      type: newType,
      proxy_url: newType === 'proxy' ? newProxyURL.trim() : '',
      is_public: false,
      rate_limit: 100,
    })
    setNewName('')
    setNewProxyURL('')
    setNewType('static')
    setCreating(false)
    load()
  }

  async function remove(id) {
    const res = await api.deleteSubdomain(id)
    if (res && (res.ok || res.status === 204)) {
      load()
    }
  }

  async function toggleActive(s) {
    await api.updatePrivacy(s.id, { is_public: s.is_public, is_active: !s.is_active })
    load()
  }

  async function updateRL(id, value) {
    await api.updateRateLimit(id, { rate_limit: parseInt(value) || 0 })
  }

  if (loading) return <p className="muted">Loading…</p>

  return (
    <div>
      <div className="subdomains-header">
        <h1 className="page-title">Subdomains</h1>
      </div>

      <form className="create-form card" onSubmit={create}>
        <div className="create-row">
          <input
            value={newName}
            onChange={e => setNewName(e.target.value)}
            placeholder="subdomain-name"
            pattern="[a-z0-9\-]+"
            title="Lowercase letters, numbers, hyphens only"
            required
          />
          <select value={newType} onChange={e => setNewType(e.target.value)}>
            <option value="static">Static files</option>
            <option value="proxy">Proxy (app)</option>
          </select>
          {newType === 'proxy' && (
            <input
              value={newProxyURL}
              onChange={e => setNewProxyURL(e.target.value)}
              placeholder="http://localhost:9001"
              required
            />
          )}
          <button type="submit" className="btn-primary" disabled={creating}>Add</button>
        </div>
      </form>

      <div className="card" style={{ marginTop: '1rem' }}>
        {subs.length === 0 ? (
          <p className="muted">No subdomains yet. Drop a folder in sites/ or add one above.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Domain</th>
                <th>Active</th>
                <th>Visibility</th>
                <th>Rate limit (rpm)</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {subs.map(s => (
                <tr key={s.id}>
                  <td><strong>{s.name}</strong></td>
                  <td>
                    <span className={`badge ${s.type === 'proxy' ? 'badge-yellow' : 'badge-gray'}`}>
                      {s.type === 'proxy' ? `proxy → ${s.proxy_url}` : 'static'}
                    </span>
                  </td>
                  <td><span className="mono">{s.full_domain}</span></td>
                  <td>
                    <Toggle checked={s.is_active} onChange={() => toggleActive(s)} />
                  </td>
                  <td>
                    <button
                      className={`badge-btn ${s.is_public ? 'badge-green' : 'badge-yellow'}`}
                      onClick={() => setPrivacyModal(s)}
                    >
                      {s.is_public ? 'Public' : 'Private'}
                    </button>
                  </td>
                  <td>
                    <input
                      type="number"
                      defaultValue={s.rate_limit}
                      min="0"
                      style={{ width: 80 }}
                      onBlur={e => updateRL(s.id, e.target.value)}
                    />
                  </td>
                  <td>
                    <button className="btn-danger" onClick={() => remove(s.id)}>Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {privacyModal && (
        <PrivacyModal
          sub={privacyModal}
          onClose={() => setPrivacyModal(null)}
          onSave={async (isPublic, password) => {
            await api.updatePrivacy(privacyModal.id, { is_public: isPublic, password })
            setPrivacyModal(null)
            load()
          }}
        />
      )}
    </div>
  )
}

function Toggle({ checked, onChange }) {
  return (
    <label className="toggle">
      <input type="checkbox" checked={checked} onChange={onChange} />
      <span className="toggle-track" />
    </label>
  )
}

function PrivacyModal({ sub, onClose, onSave }) {
  const [isPublic, setIsPublic] = useState(sub.is_public)
  const [password, setPassword] = useState('')

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2>Privacy — {sub.name}</h2>
        <div className="modal-options">
          <label className={`option ${isPublic ? 'selected' : ''}`} onClick={() => setIsPublic(true)}>
            <strong>Public</strong>
            <span>Anyone can access</span>
          </label>
          <label className={`option ${!isPublic ? 'selected' : ''}`} onClick={() => setIsPublic(false)}>
            <strong>Private</strong>
            <span>Password protected</span>
          </label>
        </div>
        {!isPublic && (
          <input
            type="password"
            placeholder="New password (leave blank to keep current)"
            value={password}
            onChange={e => setPassword(e.target.value)}
            style={{ width: '100%', marginTop: 12 }}
          />
        )}
        <div className="modal-actions">
          <button className="btn-ghost" onClick={onClose}>Cancel</button>
          <button className="btn-primary" onClick={() => onSave(isPublic, password)}>Save</button>
        </div>
      </div>
    </div>
  )
}
