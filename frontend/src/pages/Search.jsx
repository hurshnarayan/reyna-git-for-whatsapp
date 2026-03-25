import { useState } from 'react'
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

export default function Search() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState(null)
  const [loading, setLoading] = useState(false)

  const doSearch = async (e) => {
    e?.preventDefault()
    if (!query.trim()) return
    setLoading(true)
    const files = await api.search(query.trim())
    setResults(files || [])
    setLoading(false)
  }

  return (
    <div style={{ padding: '32px 40px', maxWidth: 800 }} className="fade-in">
      <h1 style={{ fontSize: 28, fontWeight: 800, letterSpacing: -1, marginBottom: 4 }}>Search 🔍</h1>
      <p style={{ fontSize: 14, color: '#888', marginBottom: 24 }}>
        Like <code style={{ fontSize: 13, background: '#f0fdf4', padding: '2px 8px', borderRadius: 4, color: '#25D366' }}>/reyna find</code> but with a nicer UI.
      </p>

      <form onSubmit={doSearch} style={{ display: 'flex', gap: 8, marginBottom: 32 }}>
        <input
          value={query} onChange={e => setQuery(e.target.value)}
          placeholder='Search files... e.g. "DSA notes", "PYQ 2024", "OS lab manual"'
          style={{
            flex: 1, padding: '14px 18px', border: '2px solid #eee', borderRadius: 10,
            fontSize: 15, outline: 'none', fontFamily: 'Inter, sans-serif',
            transition: 'border-color 0.2s',
          }}
          onFocus={e => e.target.style.borderColor = '#25D366'}
          onBlur={e => e.target.style.borderColor = '#eee'}
        />
        <button type="submit" disabled={loading} style={{
          padding: '14px 24px', background: '#111', color: '#fff', border: 'none',
          borderRadius: 10, fontSize: 15, fontWeight: 700, cursor: 'pointer',
          opacity: loading ? 0.6 : 1,
        }}>
          {loading ? '...' : 'Search'}
        </button>
      </form>

      {results === null ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#ccc' }}>
          <div style={{ fontSize: 64, marginBottom: 16 }}>🔍</div>
          <p style={{ fontSize: 16 }}>Search for files across all your groups</p>
          <p style={{ fontSize: 13, marginTop: 8 }}>Pro tip: search by subject, file name, or tags</p>
        </div>
      ) : results.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: '#ccc' }}>
          <div style={{ fontSize: 48, marginBottom: 12 }}>🤷‍♀️</div>
          <p style={{ fontSize: 16 }}>Nothing found for "{query}"</p>
          <p style={{ fontSize: 13, color: '#999', marginTop: 8, fontStyle: 'italic' }}>
            Reyna says: "Zero results. Group mein notes daalte toh milte na? Memes toh bohot hain. 😂"
          </p>
        </div>
      ) : (
        <div>
          <p style={{ fontSize: 13, color: '#888', marginBottom: 16 }}>
            {results.length} result{results.length !== 1 ? 's' : ''} for "{query}"
          </p>
          <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, overflow: 'hidden' }}>
            {results.map((f, i) => (
              <div key={i} style={{
                display: 'flex', alignItems: 'center', gap: 14, padding: '14px 20px',
                borderBottom: i < results.length - 1 ? '1px solid #f5f5f5' : 'none',
              }}>
                <div style={{
                  width: 40, height: 40, borderRadius: 10, background: '#f0fdf4',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, flexShrink: 0,
                }}>📄</div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 15, fontWeight: 600, color: '#222' }}>{f.file_name}</div>
                  <div style={{ fontSize: 12, color: '#999', display: 'flex', gap: 12, marginTop: 2 }}>
                    <span>{f.subject || 'General'}</span>
                    <span>v{f.version}</span>
                    <span>{formatBytes(f.file_size)}</span>
                    <span>by {f.shared_by_name || 'unknown'}</span>
                  </div>
                </div>
                <div style={{ fontSize: 12, color: '#bbb', flexShrink: 0 }}>{timeAgo(f.created_at)}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
