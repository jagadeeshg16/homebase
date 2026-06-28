import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import './Login.css'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  async function submit(e) {
    e.preventDefault()
    setLoading(true)
    setError('')
    const res = await api.login(username, password)
    setLoading(false)
    if (res?.ok) {
      navigate('/')
    } else {
      setError('Invalid credentials')
    }
  }

  return (
    <div className="login-wrap">
      <form className="login-box" onSubmit={submit}>
        <h1 className="login-title">Admin</h1>
        <p className="login-sub">jagadeeshg.in</p>
        {error && <div className="login-error">{error}</div>}
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={e => setUsername(e.target.value)}
          autoFocus
          required
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={e => setPassword(e.target.value)}
          required
        />
        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
    </div>
  )
}
