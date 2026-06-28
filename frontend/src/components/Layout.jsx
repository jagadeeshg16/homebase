import { NavLink, useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import './Layout.css'

const nav = [
  { to: '/',           label: 'Dashboard' },
  { to: '/subdomains', label: 'Subdomains' },
  { to: '/health',     label: 'Health' },
  { to: '/dns',        label: 'DNS' },
  { to: '/settings',   label: 'Settings' },
]

export default function Layout({ children }) {
  const navigate = useNavigate()

  async function logout() {
    await api.logout()
    navigate('/login')
  }

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-brand">jagadeeshg.in</div>
        <nav className="sidebar-nav">
          {nav.map(n => (
            <NavLink key={n.to} to={n.to} end={n.to === '/'} className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}>
              {n.label}
            </NavLink>
          ))}
        </nav>
        <button className="btn-ghost logout-btn" onClick={logout}>Logout</button>
      </aside>
      <main className="main">{children}</main>
    </div>
  )
}
