import { Outlet, NavLink, useNavigate } from 'react-router-dom'
import { isLoggedIn, logout, getUser } from '../lib/api'
import { useEffect } from 'react'

const nav = [
  { to: '/dashboard', label: 'Dashboard', icon: '📊' },
  { to: '/files', label: 'Files', icon: '📁' },
  { to: '/search', label: 'Search', icon: '🔍' },
  { to: '/bot', label: 'Bot Demo', icon: '🤖' },
]

export default function Layout() {
  const navigate = useNavigate()
  const user = getUser()

  useEffect(() => {
    if (!isLoggedIn()) navigate('/login')
  }, [])

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      {/* Sidebar */}
      <aside style={{
        width: 240, background: '#111', color: '#fff', padding: '24px 0',
        display: 'flex', flexDirection: 'column', position: 'fixed', top: 0, bottom: 0,
      }}>
        <div style={{ padding: '0 20px 24px', borderBottom: '1px solid #222' }}>
          <a href="/" style={{ fontWeight: 900, fontSize: 20, color: '#fff', letterSpacing: -0.5 }}>
            Reyna <span style={{ fontSize: 11, color: '#666', fontWeight: 400 }}>v0.1</span>
          </a>
        </div>

        <nav style={{ flex: 1, padding: '16px 0' }}>
          {nav.map(n => (
            <NavLink key={n.to} to={n.to} style={({ isActive }) => ({
              display: 'flex', alignItems: 'center', gap: 10, padding: '10px 20px',
              fontSize: 14, fontWeight: 500, color: isActive ? '#25D366' : '#888',
              background: isActive ? 'rgba(37,211,102,0.08)' : 'transparent',
              borderLeft: isActive ? '3px solid #25D366' : '3px solid transparent',
              transition: 'all 0.2s',
            })}>
              <span style={{ fontSize: 16 }}>{n.icon}</span>
              {n.label}
            </NavLink>
          ))}
        </nav>

        <div style={{ padding: '16px 20px', borderTop: '1px solid #222' }}>
          {user && (
            <div style={{ fontSize: 13, color: '#666', marginBottom: 8 }}>
              {user.name || user.phone}
            </div>
          )}
          <button onClick={() => { logout(); navigate('/login') }} style={{
            background: 'none', border: '1px solid #333', color: '#888', padding: '6px 14px',
            borderRadius: 6, fontSize: 12, cursor: 'pointer', width: '100%',
          }}>
            Logout
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main style={{ flex: 1, marginLeft: 240, background: '#fafafa', minHeight: '100vh' }}>
        <Outlet />
      </main>
    </div>
  )
}
