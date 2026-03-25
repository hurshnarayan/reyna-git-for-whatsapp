import { useState, useEffect } from 'react'
import { api, getUser } from '../lib/api'
import { useNavigate } from 'react-router-dom'

function formatBytes(b) {
  if (b < 1024) return b + ' B'
  if (b < 1048576) return (b / 1024).toFixed(1) + ' KB'
  if (b < 1073741824) return (b / 1048576).toFixed(1) + ' MB'
  return (b / 1073741824).toFixed(1) + ' GB'
}

function timeAgo(d) {
  const s = Math.floor((Date.now() - new Date(d).getTime()) / 1000)
  if (s < 60) return 'just now'
  if (s < 3600) return Math.floor(s / 60) + 'm ago'
  if (s < 86400) return Math.floor(s / 3600) + 'h ago'
  return Math.floor(s / 86400) + 'd ago'
}

export default function Dashboard() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [driveStatus, setDriveStatus] = useState(null)
  const [connecting, setConnecting] = useState(false)
  const user = getUser()
  const navigate = useNavigate()

  useEffect(() => {
    const fetchData = () => {
      api.dashboard().then(d => { setData(d); setLoading(false) }).catch(() => setLoading(false))
    }
    fetchData()
    api.googleStatus().then(setDriveStatus).catch(() => {})

    // Poll every 5 seconds for live updates from WhatsApp bot
    const interval = setInterval(fetchData, 5000)

    // Listen for Google OAuth popup callback
    const handler = (e) => {
      if (e.data?.type === 'google_auth_success') {
        setDriveStatus(prev => ({ ...prev, connected: true, email: e.data.email }))
        setConnecting(false)
        api.googleStatus().then(setDriveStatus)
      }
      if (e.data?.type === 'google_auth_error') {
        setConnecting(false)
      }
    }
    window.addEventListener('message', handler)
    return () => { clearInterval(interval); window.removeEventListener('message', handler) }
  }, [])

  const connectDrive = async () => {
    setConnecting(true)
    const resp = await api.googleConnect()
    if (resp?.url) {
      window.open(resp.url, 'google_auth', 'width=500,height=600,left=200,top=100')
    } else {
      setConnecting(false)
    }
  }

  if (loading) return (
    <div style={{ padding: 40, textAlign: 'center', color: '#888' }}>
      <div style={{ fontSize: 24, animation: 'spin 1s linear infinite', display: 'inline-block' }}>⏳</div>
      <p style={{ marginTop: 12, fontSize: 14 }}>Loading your repo...</p>
    </div>
  )

  const stats = data?.stats || { total_files: 0, total_groups: 0, total_size: 0, recent_files: [], subject_breakdown: {}, top_contributors: [] }
  const storageUsed = data?.storage_used || 0
  const storageLimit = data?.storage_limit || 15 * 1024 * 1024 * 1024

  return (
    <div style={{ padding: '32px 40px', maxWidth: 1000 }} className="fade-in">
      <div style={{ marginBottom: 32 }}>
        <h1 style={{ fontSize: 28, fontWeight: 800, letterSpacing: -1, marginBottom: 4 }}>
          Welcome back{user?.name ? `, ${user.name}` : ''} 👋
        </h1>
        <p style={{ fontSize: 15, color: '#888' }}>Here's what Reyna has been up to in your repos.</p>
      </div>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16, marginBottom: 32 }}>
        {[
          { label: 'Total Files', value: stats.total_files, icon: '📄', color: '#25D366' },
          { label: 'Groups', value: stats.total_groups, icon: '💬', color: '#0ea5e9' },
          { label: 'Storage Used', value: formatBytes(storageUsed), icon: '💾', color: '#f59e0b' },
          { label: 'Drive Free', value: formatBytes(storageLimit - storageUsed), icon: '☁️', color: '#8b5cf6' },
        ].map((s, i) => (
          <div key={i} style={{
            background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 20,
            borderTop: `3px solid ${s.color}`,
          }}>
            <div style={{ fontSize: 22, marginBottom: 8 }}>{s.icon}</div>
            <div style={{ fontSize: 28, fontWeight: 800, color: '#111', letterSpacing: -1 }}>{s.value}</div>
            <div style={{ fontSize: 12, color: '#999', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 1, marginTop: 4 }}>{s.label}</div>
          </div>
        ))}
      </div>

      {/* Google Drive Connection */}
      <div style={{
        background: driveStatus?.connected ? '#f0fdf4' : '#fffbeb',
        border: `1px solid ${driveStatus?.connected ? '#bbf7d0' : '#fde68a'}`,
        borderRadius: 12, padding: 20, marginBottom: 24,
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{ fontSize: 28 }}>{driveStatus?.connected ? '☁️' : '🔗'}</div>
          <div>
            <div style={{ fontSize: 15, fontWeight: 700, color: '#111' }}>
              {driveStatus?.connected ? 'Google Drive Connected' : 'Connect Google Drive'}
            </div>
            <div style={{ fontSize: 13, color: '#888' }}>
              {driveStatus?.connected
                ? `Syncing to ${driveStatus.email} → Reyna/ folder`
                : driveStatus?.configured === false
                  ? 'Server needs GOOGLE_CLIENT_ID & SECRET env vars to enable Drive sync'
                  : 'Files are stored locally. Connect Drive to sync across devices.'}
            </div>
          </div>
        </div>
        {driveStatus?.connected ? (
          <span style={{ fontSize: 12, fontWeight: 700, color: '#25D366', background: '#dcfce7', padding: '4px 12px', borderRadius: 20 }}>✓ Connected</span>
        ) : driveStatus?.configured !== false ? (
          <button onClick={connectDrive} disabled={connecting} style={{
            background: '#111', color: '#fff', border: 'none', borderRadius: 8,
            padding: '10px 20px', fontSize: 13, fontWeight: 700, cursor: 'pointer',
            opacity: connecting ? 0.6 : 1,
          }}>
            {connecting ? 'Connecting...' : 'Connect Drive →'}
          </button>
        ) : null}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 24 }}>
        {/* Recent Files */}
        <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 24 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <h2 style={{ fontSize: 16, fontWeight: 700 }}>Recent Files</h2>
            <button onClick={() => navigate('/files')} style={{
              background: 'none', border: '1px solid #ddd', borderRadius: 6, padding: '4px 12px',
              fontSize: 12, color: '#666', cursor: 'pointer',
            }}>View all →</button>
          </div>
          {(stats.recent_files || []).length === 0 ? (
            <p style={{ fontSize: 14, color: '#bbb', textAlign: 'center', padding: 32 }}>No files yet. Use /reyna add in your WhatsApp group!</p>
          ) : (
            (stats.recent_files || []).slice(0, 8).map((f, i) => (
              <div key={i} style={{
                display: 'flex', alignItems: 'center', gap: 12, padding: '10px 0',
                borderBottom: i < 7 ? '1px solid #f5f5f5' : 'none',
              }}>
                <div style={{
                  width: 36, height: 36, borderRadius: 8, background: '#f0fdf4',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16, flexShrink: 0,
                }}>📄</div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 14, fontWeight: 600, color: '#222', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{f.file_name}</div>
                  <div style={{ fontSize: 12, color: '#999' }}>
                    {f.subject || 'General'} · v{f.version} · by {f.shared_by_name || 'unknown'}
                  </div>
                </div>
                <div style={{ fontSize: 11, color: '#bbb', flexShrink: 0 }}>{timeAgo(f.created_at)}</div>
              </div>
            ))
          )}
        </div>

        {/* Right Column */}
        <div>
          {/* Subject Breakdown */}
          <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 24, marginBottom: 16 }}>
            <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16 }}>By Subject</h2>
            {Object.entries(stats.subject_breakdown || {}).length === 0 ? (
              <p style={{ fontSize: 13, color: '#bbb' }}>No data yet</p>
            ) : (
              Object.entries(stats.subject_breakdown || {}).sort((a, b) => b[1] - a[1]).map(([sub, cnt], i) => (
                <div key={i} style={{ marginBottom: 10 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, marginBottom: 4 }}>
                    <span style={{ fontWeight: 600 }}>{sub}</span>
                    <span style={{ color: '#999' }}>{cnt}</span>
                  </div>
                  <div style={{ height: 6, background: '#f5f5f5', borderRadius: 3 }}>
                    <div style={{
                      height: '100%', borderRadius: 3, background: '#25D366',
                      width: `${(cnt / stats.total_files) * 100}%`, transition: 'width 0.5s',
                    }} />
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Top Contributors */}
          <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 24 }}>
            <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16 }}>Top Contributors 🏆</h2>
            {(stats.top_contributors || []).length === 0 ? (
              <p style={{ fontSize: 13, color: '#bbb' }}>No contributors yet</p>
            ) : (
              (stats.top_contributors || []).map((c, i) => (
                <div key={i} style={{
                  display: 'flex', alignItems: 'center', gap: 10, padding: '8px 0',
                  borderBottom: '1px solid #f5f5f5',
                }}>
                  <div style={{
                    width: 28, height: 28, borderRadius: '50%', background: i === 0 ? '#25D366' : i === 1 ? '#0ea5e9' : '#f5f5f5',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 12, fontWeight: 800, color: i < 2 ? '#fff' : '#999',
                  }}>{i + 1}</div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 13, fontWeight: 600 }}>{c.name || c.phone}</div>
                  </div>
                  <div style={{ fontSize: 13, fontWeight: 700, color: '#25D366' }}>{c.count} files</div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
