import { useState, useRef, useEffect } from 'react'
import { api, getUser } from '../lib/api'

function timeNow() {
  const d = new Date()
  let h = d.getHours(), m = d.getMinutes().toString().padStart(2, '0'), a = h >= 12 ? 'PM' : 'AM'
  return `${h % 12 || 12}:${m} ${a}`
}

export default function BotDemo() {
  const user = getUser()
  const [messages, setMessages] = useState([
    { from: 'Rahul', type: 'other', text: '📄 CN_Unit4_Notes.pdf (2.3 MB)', time: '10:42 AM' },
    { from: 'You', type: 'sent', text: '/reyna add CN_Unit4_Notes.pdf', time: '10:42 AM' },
    { from: '🤖 Reyna', type: 'reyna', text: 'Saved CN_Unit4_Notes.pdf → your Drive repo. That\'s 12 files total. Tera Drive toh library ban raha hai 📚', time: '10:42 AM' },
  ])
  const [input, setInput] = useState('')
  const [typing, setTyping] = useState(false)
  const chatRef = useRef(null)

  useEffect(() => {
    if (chatRef.current) chatRef.current.scrollTop = chatRef.current.scrollHeight
  }, [messages, typing])

  const send = async () => {
    if (!input.trim()) return
    const text = input.trim()
    setInput('')
    const t = timeNow()

    setMessages(p => [...p, { from: 'You', type: 'sent', text, time: t }])
    setTyping(true)

    try {
      const resp = await api.botCommand(
        '120363xxxxx@g.us',
        text,
        user?.phone || '+919876543210'
      )
      setTimeout(() => {
        setTyping(false)
        setMessages(p => [...p, {
          from: '🤖 Reyna', type: 'reyna',
          text: resp?.reply || 'Bhai server se baat nahi ho rahi. Try again? 🤷‍♀️',
          time: timeNow(),
        }])
      }, 600 + Math.random() * 800)
    } catch {
      setTyping(false)
      setMessages(p => [...p, {
        from: '🤖 Reyna', type: 'reyna',
        text: 'Server down hai bhai. Backend chalu kar pehle. 🔌',
        time: timeNow(),
      }])
    }
  }

  const quickCmds = ['/reyna help', '/reyna add .', '/reyna find "DSA"', '/reyna log', '/reyna status']

  return (
    <div style={{ padding: '32px 40px', maxWidth: 900 }} className="fade-in">
      <h1 style={{ fontSize: 28, fontWeight: 800, letterSpacing: -1, marginBottom: 4 }}>Bot Demo 🤖</h1>
      <p style={{ fontSize: 14, color: '#888', marginBottom: 32 }}>
        This hits the <strong>real Go backend</strong> — same API the WhatsApp bot will use. Try Reyna's commands:
      </p>

      <div style={{ display: 'flex', gap: 24, alignItems: 'flex-start' }}>
        {/* Chat */}
        <div style={{
          background: '#0b141a', borderRadius: 16, overflow: 'hidden', width: 420, flexShrink: 0,
          boxShadow: '0 20px 60px rgba(0,0,0,0.3)', border: '1px solid #1a2730',
        }}>
          {/* Header */}
          <div style={{ background: '#075E54', padding: '12px 16px', display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{
              width: 38, height: 38, borderRadius: '50%', background: 'linear-gradient(135deg, #25D366, #128C7E)',
              display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16, fontWeight: 800, color: '#fff',
            }}>R</div>
            <div>
              <div style={{ fontSize: 14, fontWeight: 600, color: '#e9edef' }}>CSE 2026 — Section B</div>
              <div style={{ fontSize: 11, color: '#8696a0' }}>Reyna, Rahul, Priya, +47</div>
            </div>
          </div>

          {/* Messages */}
          <div ref={chatRef} style={{ padding: '12px 14px', minHeight: 400, maxHeight: 500, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 5 }}>
            {messages.map((m, i) => (
              <div key={i} style={{
                alignSelf: m.type === 'sent' ? 'flex-end' : 'flex-start',
                background: m.type === 'sent' ? '#005c4b' : '#1f2c34',
                borderLeft: m.type === 'reyna' ? '2px solid #25D366' : 'none',
                maxWidth: '85%', padding: '7px 11px', borderRadius: 8,
                borderBottomRightRadius: m.type === 'sent' ? 2 : 8,
                borderBottomLeftRadius: m.type !== 'sent' ? 2 : 8,
                animation: 'fadeIn 0.3s ease',
              }}>
                <div style={{
                  fontSize: 11, fontWeight: 600, marginBottom: 2,
                  color: m.type === 'sent' ? '#a8d8a8' : m.type === 'reyna' ? '#25D366' : '#53bdeb',
                }}>{m.from}</div>
                <div style={{ fontSize: 13, color: '#e9edef', whiteSpace: 'pre-wrap', lineHeight: 1.5 }}>{m.text}</div>
                <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.35)', textAlign: 'right', marginTop: 3 }}>{m.time}</div>
              </div>
            ))}
            {typing && (
              <div style={{
                alignSelf: 'flex-start', background: '#1f2c34', borderLeft: '2px solid #25D366',
                padding: '7px 11px', borderRadius: 8, borderBottomLeftRadius: 2,
              }}>
                <div style={{ fontSize: 11, fontWeight: 600, color: '#25D366' }}>🤖 Reyna</div>
                <div style={{ fontSize: 12, color: '#8696a0', fontStyle: 'italic' }}>typing...</div>
              </div>
            )}
          </div>

          {/* Quick commands */}
          <div style={{ padding: '6px 12px', display: 'flex', gap: 6, overflowX: 'auto', borderTop: '1px solid #1a2730' }}>
            {quickCmds.map(c => (
              <button key={c} onClick={() => { setInput(c) }} style={{
                background: '#1a2730', border: '1px solid #2a3942', borderRadius: 14, padding: '4px 10px',
                color: '#25D366', fontSize: 11, cursor: 'pointer', whiteSpace: 'nowrap',
                fontFamily: 'JetBrains Mono, monospace',
              }}>{c}</button>
            ))}
          </div>

          {/* Input */}
          <div style={{ padding: '8px 12px', background: '#075E54', display: 'flex', gap: 8, alignItems: 'center' }}>
            <input
              value={input} onChange={e => setInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && send()}
              placeholder="Type a command..."
              style={{
                flex: 1, background: '#2a3942', border: 'none', borderRadius: 20, padding: '9px 16px',
                color: '#e9edef', fontSize: 13, fontFamily: 'JetBrains Mono, monospace', outline: 'none',
              }}
            />
            <button onClick={send} style={{
              width: 36, height: 36, borderRadius: '50%', background: '#25D366', border: 'none',
              cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="white"><path d="M2 21l21-9L2 3v7l15 2-15 2z" /></svg>
            </button>
          </div>
        </div>

        {/* Command Reference */}
        <div style={{ flex: 1 }}>
          <div style={{ background: '#fff', border: '1px solid #eee', borderRadius: 12, padding: 24 }}>
            <h3 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16 }}>Command Reference</h3>
            {[
              { cmd: '/reyna add .', desc: 'Stage the last shared file' },
              { cmd: '/reyna add File.pdf', desc: 'Stage a specific file by name' },
              { cmd: '/reyna staged', desc: 'View staged (uncommitted) files' },
              { cmd: '/reyna commit', desc: 'Commit all staged → Drive' },
              { cmd: '/reyna commit File', desc: 'Commit a specific file' },
              { cmd: '/reyna rm File', desc: 'Remove a staged file' },
              { cmd: '/reyna rm .', desc: 'Remove all staged files' },
              { cmd: '/reyna find "query"', desc: 'Search stored files' },
              { cmd: '/reyna log', desc: 'Show file history' },
              { cmd: '/reyna status', desc: 'Show new files since last check' },
              { cmd: '/reyna help', desc: 'Show commands (with attitude)' },
            ].map((c, i) => (
              <div key={i} onClick={() => setInput(c.cmd)} style={{
                padding: '10px 12px', borderBottom: '1px solid #f5f5f5', cursor: 'pointer',
                transition: 'background 0.2s', borderRadius: 6,
              }}
                onMouseEnter={e => e.currentTarget.style.background = '#f9fafb'}
                onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
              >
                <code style={{
                  fontSize: 13, fontWeight: 600, color: '#25D366', background: '#f0fdf4',
                  padding: '2px 8px', borderRadius: 4,
                }}>{c.cmd}</code>
                <div style={{ fontSize: 12, color: '#888', marginTop: 4 }}>{c.desc}</div>
              </div>
            ))}
          </div>

          <div style={{
            background: '#fffbeb', border: '1px solid #fde68a', borderRadius: 12, padding: 16, marginTop: 16,
          }}>
            <div style={{ fontSize: 13, fontWeight: 700, color: '#92400e', marginBottom: 6 }}>💡 How it connects</div>
            <p style={{ fontSize: 12, color: '#78716c', lineHeight: 1.7 }}>
              This demo hits <code style={{ fontSize: 11 }}>POST /api/bot/command</code> — the same endpoint the WhatsApp bot (Baileys) will call.
              Every command here creates real database entries. Check the Files and Dashboard pages to see the data Reyna processes.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
