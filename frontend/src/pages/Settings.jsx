import { useState } from 'react'
import './Settings.css'

export default function Settings() {
  const [current, setCurrent] = useState('')
  const [next, setNext]       = useState('')
  const [confirm, setConfirm] = useState('')
  const [msg, setMsg]         = useState(null)
  const [saving, setSaving]   = useState(false)

  async function changePassword(e) {
    e.preventDefault()
    if (next !== confirm) { setMsg({ ok: false, text: 'Passwords do not match.' }); return }
    if (next.length < 8)  { setMsg({ ok: false, text: 'Password must be at least 8 characters.' }); return }
    setSaving(true)
    const res = await fetch('/api/settings/password', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ current_password: current, new_password: next }),
    })
    setSaving(false)
    if (res.ok) {
      setMsg({ ok: true, text: 'Password updated.' })
      setCurrent(''); setNext(''); setConfirm('')
    } else {
      setMsg({ ok: false, text: 'Failed — check current password.' })
    }
  }

  return (
    <div>
      <h1 className="page-title">Settings</h1>

      <div className="card settings-section">
        <h2 className="settings-title">Change Password</h2>
        <form onSubmit={changePassword} className="settings-form">
          <label>Current password
            <input type="password" value={current} onChange={e => setCurrent(e.target.value)} required />
          </label>
          <label>New password
            <input type="password" value={next} onChange={e => setNext(e.target.value)} required />
          </label>
          <label>Confirm new password
            <input type="password" value={confirm} onChange={e => setConfirm(e.target.value)} required />
          </label>
          {msg && <p className={msg.ok ? 'msg-ok' : 'msg-err'}>{msg.text}</p>}
          <button type="submit" className="btn-primary" disabled={saving}>
            {saving ? 'Saving…' : 'Update password'}
          </button>
        </form>
      </div>
    </div>
  )
}
