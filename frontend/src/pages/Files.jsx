import { useState, useEffect } from 'react'
import { api } from '../lib/api'

function formatBytes(b) {
  if (b < 1024) return b + ' B'
  if (b < 1048576) return (b / 1024).toFixed(1) + ' KB'
  return (b / 1048576).toFixed(1) + ' MB'
}

function timeAgo(d) {
  const s = Math.floor((Date.now() - new Date(d).getTime()) / 1000)
  if (s < 60) return 'just now'
  if (s < 3600) return Math.floor(s / 60) + 'm ago'
  if (s < 86400) return Math.floor(s / 3600) + 'h ago'
  return Math.floor(s / 86400) + 'd ago'
}

const subjectColors = {
  DSA: '#25D366', OS: '#0ea5e9', CN: '#f59e0b', DBMS: '#8b5cf6',
  DAA: '#ef4444', COA: '#ec4899', General: '#6b7280', Uncategorized: '#6b7280',
}

export default function Files() {
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('all')
  const [selectedFile, setSelectedFile] = useState(null)
  const [versions, setVersions] = useState([])
  const [viewMode, setViewMode] = useState('list') // 'list' | 'grid'

  useEffect(() => {
    const fetchFiles = () => api.files('', 100).then(f => { setFiles(f || []); setLoading(false) })
    fetchFiles()
    const interval = setInterval(fetchFiles, 5000)
    return () => clearInterval(interval)
  }, [])

  const staged = files.filter(f => f.status === 'staged')
  const committed = files.filter(f => f.status === 'committed' || !f.status)
  const subjects = ['all', ...new Set(files.map(f => f.subject || 'General'))]
  const activeFiles = filter === 'all' ? files : files.filter(f => (f.subject || 'General') === filter)
  const filtered = activeFiles

  const showVersions = async (f) => {
    setSelectedFile(f)
    const v = await api.versions(f.id)
    setVersions(v || [])
  }

  if (loading) return (
    <div style={{ padding: 40, textAlign: 'center', color: '#888' }}>
      <div style={{ fontSize: 24, animation: 'spin 1s linear infinite', display: 'inline-block' }}>⏳</div>
    </div>
  )

  return (
    <div style={{ padding: '32px 40px', maxWidth: 1000 }} className="fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <h1 style={{ fontSize: 28, fontWeight: 800, letterSpacing: -1, marginBottom: 4 }}>Files 📁</h1>
          <p style={{ fontSize: 14, color: '#888' }}>{committed.length} committed · {staged.length} staged</p>
        </div>
        <div style={{ display: 'flex', gap: 4, background: '#f5f5f5', borderRadius: 8, padding: 3 }}>
          {['list', 'grid'].map(m => (
            <button key={m} onClick={() => setViewMode(m)} style={{
              padding: '6px 12px', border: 'none', borderRadius: 6, fontSize: 13, cursor: 'pointer',
              background: viewMode === m ? '#fff' : 'transparent', color: viewMode === m ? '#111' : '#888',
              fontWeight: viewMode === m ? 600 : 400, boxShadow: viewMode === m ? '0 1px 3px rgba(0,0,0,0.1)' : 'none',
            }}>
              {m === 'list' ? '☰' : '▦'}
            </button>
          ))}
        </div>
      </div>

      {/* Staging Area */}
      {staged.length > 0 && (
        <div style={{
          background: '#fffbeb', border: '1px solid #fde68a', borderRadius: 12, padding: 16, marginBottom: 20,
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
            <div style={{ fontSize: 14, fontWeight: 700, color: '#92400e' }}>
              📦 Staging Area — {staged.length} file(s) not yet committed
            </div>
            <span style={{ fontSize: 12, color: '#b45309', background: '#fef3c7', padding: '2px 10px', borderRadius: 10, fontWeight: 600 }}>
              Use /reyna commit in WhatsApp to push to Drive
            </span>
          </div>
          {staged.slice(0, 5).map((f, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '6px 0', fontSize: 13 }}>
              <span style={{ color: '#f59e0b' }}>●</span>
              <span style={{ fontWeight: 600, color: '#78350f' }}>{f.file_name}</span>
              <span style={{ color: '#92400e', fontSize: 12 }}>{f.subject || 'General'}</span>
              <span style={{ color: '#b45309', fontSize: 11 }}>by {f.shared_by_name || 'unknown'}</span>
            </div>
          ))}
          {staged.length > 5 && <div style={{ fontSize: 12, color: '#92400e', marginTop: 4 }}>+{staged.length - 5} more</div>}
        </div>
      )}

      {/* Subject filter tabs */}
      <div style={{ display: 'flex', gap: 8, marginBottom: 24, flexWrap: 'wrap' }}>
        {subjects.map(s => (
          <button key={s} onClick={() => setFilter(s)} style={{
            padding: '6px 14px', border: '1px solid', borderRadius: 20, fontSize: 13, cursor: 'pointer',
            fontWeight: 600, transition: 'all 0.2s',
            background: filter === s ? (subjectColors[s] || '#111') : '#fff',
            color: filter === s ? '#fff' : '#666',
            borderColor: filter === s ? (subjectColors[s] || '#111') : '#ddd',
          }}>
            {s === 'all' ? `All (${files.length})` : `${s} (${files.filter(f => (f.subject || 'General') === s).length})`}
          </button>
        ))}
      </div>

      {/* Files */}
      {filtered.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#bbb' }}>
          <div style={{ fontSize: 48, marginBottom: 12 }}>📭</div>
          <p style={{ fontSize: 16 }}>No files here yet.</p>
          <p style={{ fontSize: 13, marginTop: 4 }}>Use <code>/reyna add</code> in your WhatsApp group to store files.</p>
        </div>
      ) : viewMode === 'list' ? (
        <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, overflow: 'hidden' }}>
          {/* Table header */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 100px 100px 120px 60px', gap: 12, padding: '10px 20px', background: '#fafafa', borderBottom: '1px solid #eee', fontSize: 11, fontWeight: 700, color: '#999', textTransform: 'uppercase', letterSpacing: 1 }}>
            <span>File</span><span>Subject</span><span>Size</span><span>Shared by</span><span>Ver</span>
          </div>
          {filtered.map((f, i) => (
            <div key={i} onClick={() => showVersions(f)} style={{
              display: 'grid', gridTemplateColumns: '1fr 100px 100px 120px 60px', gap: 12,
              padding: '12px 20px', borderBottom: '1px solid #f5f5f5', cursor: 'pointer',
              transition: 'background 0.2s', alignItems: 'center',
            }}
              onMouseEnter={e => e.currentTarget.style.background = '#f9fafb'}
              onMouseLeave={e => e.currentTarget.style.background = '#fff'}
            >
              <div>
                <div style={{ fontSize: 14, fontWeight: 600, color: '#222' }}>📄 {f.file_name}</div>
                <div style={{ fontSize: 11, color: '#bbb' }}>{timeAgo(f.created_at)}</div>
              </div>
              <span style={{
                fontSize: 11, fontWeight: 700, padding: '2px 8px', borderRadius: 10,
                background: (subjectColors[f.subject] || '#6b7280') + '15',
                color: subjectColors[f.subject] || '#6b7280',
                display: 'inline-block', width: 'fit-content',
              }}>{f.subject || 'General'}</span>
              <span style={{ fontSize: 13, color: '#888' }}>{formatBytes(f.file_size)}</span>
              <span style={{ fontSize: 13, color: '#666' }}>{f.shared_by_name || '—'}</span>
              <span style={{ fontSize: 13, color: '#25D366', fontWeight: 700 }}>v{f.version}</span>
              <span style={{
                fontSize: 10, fontWeight: 700, padding: '1px 6px', borderRadius: 8,
                background: f.status === 'committed' ? '#dcfce7' : '#fef3c7',
                color: f.status === 'committed' ? '#166534' : '#92400e',
                marginLeft: 4,
              }}>{f.status === 'committed' ? '✓' : '○'}</span>
            </div>
          ))}
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 12 }}>
          {filtered.map((f, i) => (
            <div key={i} onClick={() => showVersions(f)} style={{
              background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 20, cursor: 'pointer',
              transition: 'all 0.2s', borderTop: `3px solid ${subjectColors[f.subject] || '#6b7280'}`,
            }}
              onMouseEnter={e => { e.currentTarget.style.transform = 'translateY(-2px)'; e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.08)' }}
              onMouseLeave={e => { e.currentTarget.style.transform = ''; e.currentTarget.style.boxShadow = '' }}
            >
              <div style={{ fontSize: 28, marginBottom: 8 }}>📄</div>
              <div style={{ fontSize: 13, fontWeight: 700, color: '#222', marginBottom: 4, wordBreak: 'break-word' }}>{f.file_name}</div>
              <div style={{ fontSize: 11, color: '#999' }}>{f.subject || 'General'} · v{f.version}</div>
              <div style={{ fontSize: 11, color: '#bbb', marginTop: 4 }}>{formatBytes(f.file_size)} · {f.shared_by_name || 'unknown'}</div>
            </div>
          ))}
        </div>
      )}

      {/* Version History Modal */}
      {selectedFile && (
        <div onClick={() => setSelectedFile(null)} style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000,
        }}>
          <div onClick={e => e.stopPropagation()} style={{
            background: '#fff', borderRadius: 16, padding: 32, maxWidth: 500, width: '90%',
            maxHeight: '80vh', overflow: 'auto',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
              <h3 style={{ fontSize: 18, fontWeight: 800 }}>📄 {selectedFile.file_name}</h3>
              <button onClick={() => setSelectedFile(null)} style={{
                background: '#f5f5f5', border: 'none', borderRadius: '50%', width: 32, height: 32,
                cursor: 'pointer', fontSize: 16,
              }}>✕</button>
            </div>
            <div style={{ fontSize: 13, color: '#888', marginBottom: 20 }}>
              Subject: <strong>{selectedFile.subject || 'General'}</strong> · Size: <strong>{formatBytes(selectedFile.file_size)}</strong> · Current: <strong>v{selectedFile.version}</strong>
            </div>

            <h4 style={{ fontSize: 14, fontWeight: 700, marginBottom: 12 }}>Version History</h4>
            {versions.length === 0 ? (
              <p style={{ fontSize: 13, color: '#bbb' }}>Single version</p>
            ) : (
              versions.map((v, i) => (
                <div key={i} style={{
                  display: 'flex', alignItems: 'center', gap: 12, padding: '10px 0',
                  borderBottom: '1px solid #f5f5f5',
                }}>
                  <div style={{
                    width: 10, height: 10, borderRadius: '50%',
                    background: i === 0 ? '#25D366' : '#ddd',
                  }} />
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 13, fontWeight: 600 }}>Version {v.version}</div>
                    <div style={{ fontSize: 11, color: '#999' }}>{formatBytes(v.file_size)} · {timeAgo(v.created_at)}</div>
                  </div>
                  {i === 0 && <span style={{ fontSize: 11, fontWeight: 700, color: '#25D366', background: '#f0fdf4', padding: '2px 8px', borderRadius: 4 }}>Latest</span>}
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
