import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, saveAuth, isLoggedIn } from '../lib/api'
import { useEffect } from 'react'

export default function Login() {
  const [phone, setPhone] = useState('')
  const [name, setName] = useState('')
  const [isRegister, setIsRegister] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => { if (isLoggedIn()) navigate('/dashboard') }, [])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = isRegister
        ? await api.register(phone, name)
        : await api.login(phone)
      if (data?.error) {
        setError(data.error)
        if (data.error.includes('not found')) setIsRegister(true)
      } else {
        saveAuth(data)
        navigate('/dashboard')
      }
    } catch (err) {
      setError('Connection failed. Is the backend running?')
    }
    setLoading(false)
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#fafafa' }}>
      <div style={{ maxWidth: 400, width: '100%', padding: 32 }}>
        <div style={{ textAlign: 'center', marginBottom: 40 }}>
          <h1 style={{ fontSize: 28, fontWeight: 900, letterSpacing: -1, marginBottom: 8 }}>
            Reyna <span style={{ fontSize: 13, color: '#999', fontWeight: 400 }}>v0.1</span>
          </h1>
          <p style={{ fontSize: 15, color: '#666' }}>Git for your WhatsApp group.</p>
        </div>

        <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 32 }}>
          <h2 style={{ fontSize: 18, fontWeight: 700, marginBottom: 4 }}>
            {isRegister ? 'Create account' : 'Sign in'}
          </h2>
          <p style={{ fontSize: 13, color: '#888', marginBottom: 24 }}>
            {isRegister ? 'Register with your WhatsApp number' : 'Use your WhatsApp number to log in'}
          </p>

          <form onSubmit={handleSubmit}>
            <label style={{ fontSize: 13, fontWeight: 600, color: '#444', display: 'block', marginBottom: 6 }}>
              WhatsApp Number
            </label>
            <input
              type="text" value={phone} onChange={e => setPhone(e.target.value)}
              placeholder="+91 9876543210"
              style={{
                width: '100%', padding: '12px 14px', border: '1px solid #ddd', borderRadius: 8,
                fontSize: 15, marginBottom: 16, outline: 'none', fontFamily: 'JetBrains Mono, monospace',
              }}
            />

            {isRegister && (
              <>
                <label style={{ fontSize: 13, fontWeight: 600, color: '#444', display: 'block', marginBottom: 6 }}>
                  Name
                </label>
                <input
                  type="text" value={name} onChange={e => setName(e.target.value)}
                  placeholder="Your name"
                  style={{
                    width: '100%', padding: '12px 14px', border: '1px solid #ddd', borderRadius: 8,
                    fontSize: 15, marginBottom: 16, outline: 'none',
                  }}
                />
              </>
            )}

            {error && (
              <div style={{ fontSize: 13, color: '#ff4444', marginBottom: 12, padding: '8px 12px', background: '#fff5f5', borderRadius: 6 }}>
                {error}
              </div>
            )}

            <button type="submit" disabled={loading || !phone} style={{
              width: '100%', padding: '12px', background: '#111', color: '#fff', border: 'none',
              borderRadius: 8, fontSize: 15, fontWeight: 700, cursor: 'pointer', opacity: loading ? 0.6 : 1,
            }}>
              {loading ? 'Loading...' : isRegister ? 'Register' : 'Sign in'}
            </button>
          </form>

          <div style={{ textAlign: 'center', marginTop: 16 }}>
            <button onClick={() => { setIsRegister(!isRegister); setError('') }} style={{
              background: 'none', border: 'none', color: '#25D366', fontSize: 13, cursor: 'pointer', fontWeight: 600,
            }}>
              {isRegister ? '← Back to login' : "Don't have an account? Register"}
            </button>
          </div>
        </div>

        <p style={{ textAlign: 'center', fontSize: 12, color: '#bbb', marginTop: 24 }}>
          Demo: use +919876543210 to login with seeded data
        </p>
      </div>
    </div>
  )
}
