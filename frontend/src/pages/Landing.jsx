import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, isLoggedIn } from '../lib/api'

// Inline WhatsApp Demo component for the landing page
function LandingWhatsAppDemo() {
  const reynaResponses = {
    '/reyna add': [
      "Saved to your Drive ✅ Isse zyada organized toh tera room bhi nahi hoga 🗂️",
      "Done! File stored. Aur kitne bhejega? Free mein kaam karti hoon but unlimited nahi hoon 😤",
    ],
    '/reyna log': [
      "📋 47 files synced across 6 subjects. Tu toh version control pro ban gaya re 🚀",
    ],
    '/reyna find': [
      "🔍 3 results found. Raat ke 2 baje search? Exam kal hai na? Classic 🎯",
      "🔍 Found 2 files. Next time thoda specific search kar — main Google nahi hoon 😂",
    ],
    '/reyna status': [
      "📊 8 new files. 5 PYQs — everyone suddenly remembered exams exist 😏",
    ],
    '/reyna help': [
      "🤖 Reyna hoon — tera group ka sabse useful member.\n\n/reyna add . → file save\n/reyna find \"xyz\" → search\n/reyna log → history\n/reyna status → updates\n\nAb padh le bhai 📁",
    ],
    default: [
      "Bhai main bot hoon, tutor nahi. Commands de, gyaan mat. /reyna help 🙃",
      "Interesting. Filing under 'things Reyna doesn't care about.' /reyna help 🤷‍♀️",
    ],
  }
  const pick = a => a[Math.floor(Math.random() * a.length)]
  const getResp = msg => {
    const l = msg.toLowerCase().trim()
    if (l.startsWith('/reyna add')) return pick(reynaResponses['/reyna add'])
    if (l.startsWith('/reyna log')) return pick(reynaResponses['/reyna log'])
    if (l.startsWith('/reyna find')) return pick(reynaResponses['/reyna find'])
    if (l.startsWith('/reyna status')) return pick(reynaResponses['/reyna status'])
    if (l.startsWith('/reyna')) return pick(reynaResponses['/reyna help'])
    return pick(reynaResponses.default)
  }
  const timeNow = () => { const d=new Date(); let h=d.getHours(),m=d.getMinutes().toString().padStart(2,'0'),a=h>=12?'PM':'AM'; return `${h%12||12}:${m} ${a}` }

  const [msgs, setMsgs] = useState([
    { from:'Rahul', type:'other', text:'📄 CN_Unit4_Notes.pdf (2.3 MB)', time:'10:42 AM' },
    { from:'You', type:'sent', text:'/reyna add .', time:'10:42 AM' },
    { from:'🤖 Reyna', type:'reyna', text:'Saved CN_Unit4_Notes.pdf → Drive. That\'s 23 files. Tera Drive toh library ban raha hai 📚', time:'10:42 AM' },
  ])
  const [inp, setInp] = useState('')
  const [typing, setTyping] = useState(false)

  const send = () => {
    if (!inp.trim()) return
    const t = timeNow()
    setMsgs(p => [...p, { from:'You', type:'sent', text:inp.trim(), time:t }])
    const msg = inp.trim(); setInp(''); setTyping(true)
    setTimeout(() => {
      setTyping(false)
      setMsgs(p => [...p, { from:'🤖 Reyna', type:'reyna', text:getResp(msg), time:timeNow() }])
    }, 700 + Math.random()*800)
  }

  return (
    <div style={{ background:'#0b141a', borderRadius:16, overflow:'hidden', maxWidth:400, width:'100%', boxShadow:'0 25px 80px rgba(0,0,0,0.4)', border:'1px solid #1a2730' }}>
      <div style={{ background:'#075E54', padding:'12px 16px', display:'flex', alignItems:'center', gap:12 }}>
        <div style={{ width:38, height:38, borderRadius:'50%', background:'linear-gradient(135deg,#25D366,#128C7E)', display:'flex', alignItems:'center', justifyContent:'center', fontSize:16, fontWeight:800, color:'#fff' }}>R</div>
        <div><div style={{ fontSize:14, fontWeight:600, color:'#e9edef' }}>CSE 2026 — Section B</div><div style={{ fontSize:11, color:'#8696a0' }}>Reyna, Rahul, Priya, +47</div></div>
      </div>
      <div style={{ padding:'12px 14px', minHeight:300, maxHeight:340, overflowY:'auto', display:'flex', flexDirection:'column', gap:5 }}>
        {msgs.map((m,i) => (
          <div key={i} style={{ alignSelf:m.type==='sent'?'flex-end':'flex-start', background:m.type==='sent'?'#005c4b':'#1f2c34', borderLeft:m.type==='reyna'?'2px solid #25D366':'none', maxWidth:'85%', padding:'7px 11px', borderRadius:8 }}>
            <div style={{ fontSize:11, fontWeight:600, color:m.type==='sent'?'#a8d8a8':m.type==='reyna'?'#25D366':'#53bdeb', marginBottom:2 }}>{m.from}</div>
            <div style={{ fontSize:13, color:'#e9edef', whiteSpace:'pre-wrap', lineHeight:1.5 }}>{m.text}</div>
            <div style={{ fontSize:10, color:'rgba(255,255,255,0.35)', textAlign:'right', marginTop:3 }}>{m.time}</div>
          </div>
        ))}
        {typing && <div style={{ alignSelf:'flex-start', background:'#1f2c34', borderLeft:'2px solid #25D366', padding:'7px 11px', borderRadius:8 }}><div style={{ fontSize:11, fontWeight:600, color:'#25D366' }}>🤖 Reyna</div><div style={{ fontSize:12, color:'#8696a0', fontStyle:'italic' }}>typing...</div></div>}
      </div>
      <div style={{ padding:'8px 12px', background:'#075E54', display:'flex', gap:8, alignItems:'center' }}>
        <input value={inp} onChange={e=>setInp(e.target.value)} onKeyDown={e=>e.key==='Enter'&&send()} placeholder="Try /reyna help ..." style={{ flex:1, background:'#2a3942', border:'none', borderRadius:20, padding:'9px 16px', color:'#e9edef', fontSize:13, fontFamily:'JetBrains Mono, monospace', outline:'none' }}/>
        <button onClick={send} style={{ width:36, height:36, borderRadius:'50%', background:'#25D366', border:'none', cursor:'pointer', display:'flex', alignItems:'center', justifyContent:'center' }}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="white"><path d="M2 21l21-9L2 3v7l15 2-15 2z"/></svg>
        </button>
      </div>
    </div>
  )
}

export default function Landing() {
  const [contact, setContact] = useState('')
  const [joined, setJoined] = useState(false)
  const navigate = useNavigate()

  const joinWaitlist = async () => {
    if (!contact.trim()) return
    try { await api.joinWaitlist(contact.trim()) } catch {}
    setJoined(true)
  }

  const S = (props) => <section style={{ maxWidth:820, margin:'0 auto', padding:'80px 24px', ...props.style }}>{props.children}</section>

  return (
    <div style={{ fontFamily:"'Inter', sans-serif", color:'#222', background:'#fff' }}>
      {/* Nav */}
      <nav style={{ position:'sticky', top:0, zIndex:100, background:'rgba(255,255,255,0.92)', backdropFilter:'blur(12px)', borderBottom:'1px solid #eee', padding:'0 32px' }}>
        <div style={{ maxWidth:1100, margin:'0 auto', display:'flex', alignItems:'center', justifyContent:'space-between', height:56 }}>
          <a href="#" style={{ fontWeight:900, fontSize:20, letterSpacing:-0.5 }}>Reyna <span style={{ fontSize:11, color:'#999', fontWeight:400 }}>v0.1</span></a>
          <div style={{ display:'flex', gap:24, alignItems:'center' }}>
            {['How it works','Surfaces','Demo'].map(t => <a key={t} href={`#${t.toLowerCase().replace(/ /g,'-')}`} style={{ fontSize:13, color:'#666', fontWeight:500 }}>{t}</a>)}
            {isLoggedIn() ? (
              <button onClick={()=>navigate('/dashboard')} style={{ fontSize:13, fontWeight:700, color:'#fff', background:'#111', padding:'7px 18px', borderRadius:6, border:'none', cursor:'pointer' }}>Dashboard →</button>
            ) : (
              <button onClick={()=>navigate('/login')} style={{ fontSize:13, fontWeight:700, color:'#fff', background:'#111', padding:'7px 18px', borderRadius:6, border:'none', cursor:'pointer' }}>Sign in</button>
            )}
          </div>
        </div>
      </nav>

      {/* Hero */}
      <S style={{ padding:'80px 24px 40px', textAlign:'center', maxWidth:820 }}>
        <div style={{ display:'inline-block', fontSize:12, fontWeight:600, color:'#25D366', background:'#f0fdf4', border:'1px solid #bbf7d0', padding:'5px 14px', borderRadius:50, marginBottom:28 }}>Free & open source — always</div>
        <h1 style={{ fontSize:'clamp(36px,5.5vw,58px)', fontWeight:900, lineHeight:1.1, letterSpacing:-2, marginBottom:24 }}>
          Wrestling with <span style={{ textDecoration:'line-through', color:'#ccc' }}>projects</span> notes?<br/>
          <span style={{ color:'#25D366' }}>It doesn't have to be this messy.</span>
        </h1>
        <p style={{ fontSize:18, lineHeight:1.75, color:'#555', maxWidth:620, margin:'0 auto 20px' }}>
          There are lots of ways to share study material. WhatsApp is the most popular. And also the worst at keeping anything findable after 48 hours. <strong style={{ color:'#222' }}>You know?</strong>
        </p>
        <p style={{ fontSize:18, lineHeight:1.75, color:'#555', maxWidth:620, margin:'0 auto 24px' }}>
          Not Reyna. She's a WhatsApp bot that treats your group like a <strong style={{ color:'#222' }}>Git repository</strong>. Files → Drive. Versioned, searchable, never lost. With <span style={{ color:'#25D366', fontWeight:600 }}>desi attitude</span>.
        </p>
        <p style={{ fontSize:16, color:'#888', maxWidth:560, margin:'0 auto 36px', fontStyle:'italic' }}>
          "Git for your WhatsApp group. Your semester's notes, versioned, searchable, never lost."
        </p>
        <div style={{ display:'flex', gap:12, justifyContent:'center', flexWrap:'wrap' }}>
          <a href="#waitlist" style={{ fontSize:15, fontWeight:700, color:'#fff', background:'#111', padding:'14px 32px', borderRadius:8 }}>Join the waitlist →</a>
          <a href="#demo" style={{ fontSize:15, fontWeight:600, color:'#555', background:'#f5f5f5', padding:'14px 32px', borderRadius:8, border:'1px solid #ddd' }}>See it in action ↓</a>
        </div>
      </S>

      {/* Problem */}
      <S id="how-it-works">
        <h2 style={{ fontSize:14, fontWeight:700, color:'#25D366', letterSpacing:2, textTransform:'uppercase', marginBottom:16 }}>The Problem</h2>
        <h3 style={{ fontSize:'clamp(24px,3.5vw,36px)', fontWeight:800, lineHeight:1.2, marginBottom:24, letterSpacing:-1 }}>Your notes are drowning in "Good morning 🌅" messages.</h3>
        <p style={{ fontSize:17, lineHeight:1.8, color:'#555', marginBottom:16 }}>India's engineering students don't use Notion. Or Slack. They use <strong>WhatsApp groups</strong>. That's where notes go, assignments go, PYQs go. And that's also where <strong style={{ color:'#ff4444' }}>all of it dies</strong>.</p>
        <p style={{ fontSize:17, lineHeight:1.8, color:'#555' }}>Buried under memes, forwards, and 47 variations of "bhai notes bhej de." There's no search. No version history. <strong style={{ color:'#ff4444' }}>Message #4,847 is where your semester's knowledge goes to die.</strong></p>
      </S>

      {/* How it works */}
      <div style={{ background:'#fafafa', borderTop:'1px solid #eee', borderBottom:'1px solid #eee' }}>
        <S>
          <h2 style={{ fontSize:14, fontWeight:700, color:'#25D366', letterSpacing:2, textTransform:'uppercase', marginBottom:16 }}>Let's walk through it.</h2>
          <h3 style={{ fontSize:'clamp(24px,3.5vw,36px)', fontWeight:800, lineHeight:1.2, marginBottom:12, letterSpacing:-1 }}>10 seconds after you add Reyna, clarity sets in.</h3>
          <p style={{ fontSize:17, lineHeight:1.8, color:'#555', marginBottom:40 }}>Reyna joins your WhatsApp group as a bot. She sits silently. Someone wants to save a file? One command. No app download, no sign-up video nobody watches.</p>
          {[
            { cmd:'/reyna add .', desc:'Store the last shared file', detail:'Someone drops CN_Unit4.pdf. You reply /reyna add — boom, in your Drive, organized by subject.' },
            { cmd:'/reyna find "DSA notes"', desc:'Search everything stored', detail:'Instead of scrolling 3 months of chat, one command finds every DSA file ever saved.' },
            { cmd:'/reyna log', desc:'Full version history', detail:'Like git log but for study material. Who shared what, when, which version.' },
            { cmd:'/reyna status', desc:'What\'s new since last sync', detail:'Been bunking? See exactly what files you missed. No shame, just data.' },
          ].map((c,i) => (
            <div key={i} style={{ display:'flex', gap:24, marginBottom:24, padding:24, background:'#fff', borderRadius:12, border:'1px solid #eee', alignItems:'flex-start' }}>
              <div style={{ minWidth:48, height:48, borderRadius:10, background:'#111', display:'flex', alignItems:'center', justifyContent:'center', fontSize:18, fontWeight:800, color:'#25D366' }}>{i+1}</div>
              <div>
                <code style={{ fontFamily:'JetBrains Mono, monospace', fontSize:15, fontWeight:600, color:'#25D366', background:'#f0fdf4', padding:'3px 10px', borderRadius:4, display:'inline-block', marginBottom:6 }}>{c.cmd}</code>
                <div style={{ fontSize:14, fontWeight:700, color:'#222', marginBottom:4 }}>{c.desc}</div>
                <div style={{ fontSize:14, color:'#666', lineHeight:1.7 }}>{c.detail}</div>
              </div>
            </div>
          ))}
        </S>
      </div>

      {/* Surfaces */}
      <S id="surfaces" style={{ maxWidth:900 }}>
        <h2 style={{ fontSize:14, fontWeight:700, color:'#25D366', letterSpacing:2, textTransform:'uppercase', marginBottom:16 }}>Two Surfaces</h2>
        <h3 style={{ fontSize:'clamp(24px,3.5vw,36px)', fontWeight:800, lineHeight:1.2, marginBottom:12, letterSpacing:-1 }}>One bot. Two ways to use her.</h3>
        <p style={{ fontSize:17, lineHeight:1.8, color:'#555', marginBottom:40 }}>Reyna lives where you live — in the group chat. But when you need the full picture, there's Reyna Web.</p>
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:24 }}>
          {[
            { num:'01', icon:'💬', title:'WhatsApp Bot', sub:'The group member', color:'#25D366',
              desc:'Sits in your group. Intercepts commands, stores files to Drive, roasts anyone sending good morning images.',
              items:['Slash commands in chat','Auto-stores to Google Drive','Per-user scoping','Desi humour in every response','Works on any phone'] },
            { num:'02', icon:'🌐', title:'Reyna Web', sub:'The dashboard', color:'#0ea5e9',
              desc:'Full web dashboard to browse, search, and manage everything stored. GitHub for your group\'s notes.',
              items:['Browse with folder structure','Full-text search','Version history','Drive integration','Group analytics'] },
          ].map((s,i) => (
            <div key={i} style={{ background:'#fff', border:'2px solid #111', borderRadius:14, padding:32, position:'relative', overflow:'hidden' }}>
              <div style={{ position:'absolute', top:12, right:16, fontSize:64, fontWeight:900, color:'#f5f5f5', lineHeight:1 }}>{s.num}</div>
              <div style={{ fontSize:32, marginBottom:12 }}>{s.icon}</div>
              <h4 style={{ fontSize:22, fontWeight:800, marginBottom:4 }}>{s.title}</h4>
              <div style={{ fontSize:12, fontWeight:700, color:s.color, textTransform:'uppercase', letterSpacing:2, marginBottom:16 }}>{s.sub}</div>
              <p style={{ fontSize:14.5, color:'#555', lineHeight:1.8, marginBottom:20 }}>{s.desc}</p>
              {s.items.map((t,j) => <div key={j} style={{ fontSize:13.5, color:'#444', padding:'8px 0', borderBottom:'1px solid #f0f0f0', display:'flex', gap:8 }}><span style={{ color:s.color, fontWeight:700 }}>→</span> {t}</div>)}
            </div>
          ))}
        </div>
      </S>

      {/* Demo */}
      <div id="demo" style={{ background:'#111', borderTop:'1px solid #222' }}>
        <div style={{ maxWidth:900, margin:'0 auto', padding:'80px 24px', display:'flex', gap:48, alignItems:'center', flexWrap:'wrap', justifyContent:'center' }}>
          <div style={{ flex:'1 1 340px', maxWidth:440 }}>
            <h2 style={{ fontSize:14, fontWeight:700, color:'#25D366', letterSpacing:2, textTransform:'uppercase', marginBottom:16 }}>Try it yourself</h2>
            <h3 style={{ fontSize:32, fontWeight:800, lineHeight:1.2, marginBottom:16, color:'#fff', letterSpacing:-1 }}>Go ahead. Type a command.</h3>
            <p style={{ fontSize:16, lineHeight:1.8, color:'#aaa', marginBottom:24 }}>This is a working demo of Reyna's WhatsApp bot personality. Type commands and watch her respond — with results <em>and</em> roasts:</p>
            <div style={{ display:'flex', flexDirection:'column', gap:8 }}>
              {['/reyna add .', '/reyna find "DBMS"', '/reyna log', '/reyna status', '/reyna help'].map(c =>
                <code key={c} style={{ fontFamily:'JetBrains Mono, monospace', fontSize:13, color:'#25D366', background:'rgba(37,211,102,0.08)', padding:'6px 12px', borderRadius:6, border:'1px solid rgba(37,211,102,0.15)' }}>{c}</code>
              )}
            </div>
          </div>
          <LandingWhatsAppDemo />
        </div>
      </div>

      {/* Personality */}
      <S style={{ maxWidth:900 }}>
        <h2 style={{ fontSize:14, fontWeight:700, color:'#25D366', letterSpacing:2, textTransform:'uppercase', marginBottom:16 }}>Reyna's Personality</h2>
        <h3 style={{ fontSize:'clamp(24px,3.5vw,32px)', fontWeight:800, lineHeight:1.3, marginBottom:32, letterSpacing:-1 }}>
          She's not just a bot. She's your group's most useful <span style={{ color:'#25D366' }}>and most annoying</span> member.
        </h3>
        <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr 1fr', gap:16 }}>
          {[
            { when:'3 AM search', user:'/reyna find "DBMS"', reply:"Bhai 3 baje raat ko DBMS? Tera breakup hua hai kya? Anyway, 4 results 📂" },
            { when:'Duplicate file', user:'[sends CN_Unit2.pdf]', reply:"Ye file already 4 baar aa chuki hai. 5th time share karne se marks nahi badhenge ✅" },
            { when:'Asks for help', user:'/reyna help', reply:"Main karu toh kya karu — saves your files. Commands: add, find, log, status. Ab padh le 📁" },
            { when:'Exam panic', user:'/reyna status', reply:"12 new files. 8 PYQs. Everyone suddenly remembered exams exist 📚😂" },
            { when:'Meme uploaded', user:'[sends meme.jpg]', reply:"Nice meme. Not storing tho. I have standards. 🙃 /reyna add sirf study material." },
            { when:'Night before exam', user:'/reyna find "OS PYQ"', reply:"3 PYQs found. 11 PM, exam kal, aur tu abhi search. Classic 📄📄📄" },
          ].map((r,i) => (
            <div key={i} style={{ background:'#fafafa', border:'1px solid #eee', borderRadius:12, padding:20 }}>
              <div style={{ fontSize:11, fontWeight:700, color:'#999', textTransform:'uppercase', letterSpacing:1, marginBottom:10 }}>{r.when}</div>
              <div style={{ fontSize:13, color:'#888', marginBottom:10, paddingLeft:10, borderLeft:'2px solid #ddd' }}>{r.user}</div>
              <div style={{ fontSize:14, color:'#25D366', fontWeight:600, fontStyle:'italic', lineHeight:1.6 }}>"{r.reply}"</div>
            </div>
          ))}
        </div>
      </S>

      {/* The answer is YES */}
      <S style={{ textAlign:'center' }}>
        <h2 style={{ fontSize:'clamp(24px,3.5vw,36px)', fontWeight:800, marginBottom:32, letterSpacing:-1 }}>The answer is <span style={{ color:'#25D366' }}>YES!</span></h2>
        <div style={{ columns:2, columnGap:24, textAlign:'left' }}>
          {[
            "Can I save files without leaving WhatsApp?", "Can I search 3-month-old files in one command?",
            "Does it work on my ₹8000 phone?", "Can different members have separate Drive repos?",
            "Is there version history?", "Will Reyna roast me at 3 AM? (yes)",
            "Can I use it on the web too?", "Is my data private? (YOUR Drive)",
            "Can I see who shares most?", "Will she store memes? (no, she has standards)",
            "Is there any cost? (₹0, forever)", "Is it open source? (MIT License)",
          ].map((q,i) => (
            <div key={i} style={{ fontSize:15, color:'#444', padding:'10px 0', display:'flex', gap:10, breakInside:'avoid' }}>
              <span style={{ color:'#25D366', fontWeight:800 }}>✓</span><span>{q}</span>
            </div>
          ))}
        </div>
      </S>

      {/* CTA */}
      <section id="waitlist" style={{ background:'#111', padding:'80px 24px', textAlign:'center' }}>
        <h2 style={{ fontSize:'clamp(28px,4.5vw,44px)', fontWeight:900, lineHeight:1.2, marginBottom:16, color:'#fff', letterSpacing:-1 }}>
          You wouldn't be scrolling if your group was organized.
        </h2>
        <p style={{ fontSize:17, color:'#888', maxWidth:500, margin:'0 auto 32px', lineHeight:1.7 }}>
          Notes getting lost? <strong style={{ color:'#fff' }}>It's time for Reyna.</strong><br/>
          Same file shared 10 times? <strong style={{ color:'#fff' }}>It's time for Reyna.</strong><br/>
          Exam tomorrow, can't find PYQs? <strong style={{ color:'#25D366' }}>It's time for Reyna.</strong>
        </p>
        {!joined ? (
          <div style={{ maxWidth:440, margin:'0 auto', display:'flex', gap:8 }}>
            <input type="text" value={contact} onChange={e=>setContact(e.target.value)} onKeyDown={e=>e.key==='Enter'&&joinWaitlist()}
              placeholder="+91 your number or email"
              style={{ flex:1, background:'#1a1a1a', border:'1px solid #333', borderRadius:8, padding:'14px 18px', color:'#fff', fontSize:14, outline:'none' }} />
            <button onClick={joinWaitlist} style={{ fontSize:14, fontWeight:700, color:'#111', background:'#25D366', padding:'14px 24px', borderRadius:8, border:'none', cursor:'pointer', whiteSpace:'nowrap' }}>Join Beta</button>
          </div>
        ) : (
          <div style={{ fontSize:18, color:'#25D366', fontWeight:700 }}>✅ You're in. Reyna will ping you. (With attitude, obviously.)</div>
        )}
        <p style={{ fontSize:12, color:'#555', marginTop:16 }}>Free forever. Open source. Your data stays in YOUR Drive.</p>
      </section>

      {/* Footer */}
      <footer style={{ borderTop:'1px solid #eee', padding:'24px 32px', textAlign:'center' }}>
        <p style={{ fontSize:13, color:'#999', marginBottom:6 }}>
          <span style={{ fontWeight:800, color:'#111', letterSpacing:1 }}>REYNA</span>
          <span style={{ margin:'0 12px', color:'#ddd' }}>|</span>
          <a href="#" style={{ color:'#666' }}>GitHub</a> · <a href="#" style={{ color:'#666' }}>Twitter</a>
        </p>
        <p style={{ fontSize:12, color:'#bbb' }}>Built with ☕ and frustration by students who lost their notes one too many times.</p>
      </footer>
    </div>
  )
}
