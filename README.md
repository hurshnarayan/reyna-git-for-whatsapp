# Reyna - Git for your WhatsApp Group

> Your semester's notes, versioned, searchable, never lost.

India has 40,000+ engineering colleges. Every single one runs on WhatsApp groups. Notes, assignments, PYQs - shared, buried, and lost forever under memes and good morning messages.

**Reyna is a WhatsApp bot that treats your group chat like a Git repository.** She stages files, commits them to Google Drive, and responds with enough desi sass to keep the group entertained.

```
/reyna add .          → stage the last shared file
/reyna staged         → see what's waiting
/reyna commit         → push staged files to Google Drive
/reyna find "DSA"     → search across everything
/reyna log            → full history
/reyna rm notes.pdf   → unstage a file
```

---

## How It Works

```
WhatsApp Group                    Reyna Backend                  Google Drive
┌──────────────┐                 ┌──────────────┐              ┌──────────────┐
│ Someone sends│                 │              │              │              │
│ notes.pdf    │                 │   Go API     │              │  Reyna/      │
│              │  /reyna add .   │   + SQLite   │  /reyna      │  ├── DSA/    │
│ ────────────►│────────────────►│   + Sass     │──commit────► │  │  notes.pdf│
│              │  (file bytes    │              │  (uploads)   │  ├── OS/     │
│ Bot responds │   via base64)   │  Stores to   │              │  └── CN/    │
│ with sass 🤖 │◄────────────────│  local +DB   │              │              │
└──────────────┘                 └──────┬───────┘              └──────────────┘
                                        │
                                        │ polls every 5s
                                        ▼
                                 ┌──────────────┐
                                 │  React Web   │
                                 │  Dashboard   │
                                 │  :5173       │
                                 └──────────────┘
```

### Two Surfaces

| Surface | What it does |
|---|---|
| **WhatsApp Bot** | Lives in your group, intercepts `/reyna` commands, downloads files, uploads to Drive |
| **Web Dashboard** | Browse files, search, see staging area, version history, connect Google Drive |

---

## Commands

| Command | Description |
|---|---|
| `/reyna add .` | Stage the last shared file |
| `/reyna add File.pdf` | Stage a specific file |
| `/reyna staged` | View staged (uncommitted) files |
| `/reyna commit` | Commit all staged → Google Drive |
| `/reyna commit File` | Commit a specific file |
| `/reyna rm File` | Remove a staged file |
| `/reyna rm .` | Remove all staged files |
| `/reyna find "query"` | Search stored files |
| `/reyna log` | Show file history |
| `/reyna status` | What's new in the last 24 hours |
| `/reyna help` | Show commands (with attitude) |

---

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- Google Cloud OAuth credentials (free — [setup guide](#google-drive-setup))

### 1. Clone

```bash
git clone https://github.com/yourusername/reyna-git-for-whatsapp.git
cd reyna-git-for-whatsapp
```

### 2. Configure

```bash
cp .env.example .env
# Edit .env with your Google OAuth credentials
```

### 3. Install

```bash
cd frontend && npm install && cd ..
cd whatsapp-bot && npm install && cd ..
```

### 4. Run (three terminals)

```bash
# Terminal 1 — Backend
cd backend
source <(grep -v '^#' ../.env | sed 's/^/export /')
go run ./cmd/server/

# Terminal 2 — Frontend
cd frontend
npm run dev

# Terminal 3 — WhatsApp Bot
cd whatsapp-bot
node bot.js
# Scan QR code with WhatsApp → Linked Devices → Link a Device
```

### 5. Use

1. Open `http://localhost:5173` → Register with your WhatsApp number
2. Connect Google Drive from the dashboard
3. In any WhatsApp group the bot is in: share a file → `/reyna add .` → `/reyna commit`
4. File appears in your Google Drive under `Reyna/` folder

---

## Google Drive Setup

All free. No billing needed.

1. Go to [console.cloud.google.com](https://console.cloud.google.com)
2. Create a project called `Reyna`
3. Enable **Google Drive API** (APIs & Services → Library)
4. Go to **OAuth consent screen** → External → fill app name + emails → save through all tabs
5. Add your email as a **Test User**
6. Go to **Credentials** → Create OAuth Client ID → Web Application
7. Add redirect URI: `http://localhost:8080/api/auth/google/callback`
8. Copy Client ID and Client Secret into `.env`

---

## Tech Stack

| Component | Tech | Cost |
|---|---|---|
| Backend | Go + net/http + SQLite (WAL) | Free |
| Frontend | React 19 + Vite + React Router | Free |
| WhatsApp | Baileys v7 (WhatsApp Web protocol) | Free |
| Storage | Google Drive API (15 GB/user) | Free |
| Auth | JWT (phone-based) + Google OAuth 2.0 | Free |

**Zero paid APIs. Zero vendor lock-in. Your data stays in YOUR Drive.**

---

## Project Structure

```
reyna-git-for-whatsapp/
├── backend/
│   ├── cmd/server/main.go           # Entry point
│   └── internal/
│       ├── api/handlers.go          # REST API (20+ endpoints)
│       ├── auth/jwt.go              # JWT + token validation
│       ├── config/config.go         # Env config
│       ├── db/store.go              # SQLite (files, users, groups, versions)
│       ├── gdrive/service.go        # Google Drive OAuth + REST API + local fallback
│       ├── models/models.go         # All data types
│       └── reyna/personality.go     # The sass engine 🤖
├── frontend/
│   ├── src/
│   │   ├── main.jsx                 # React entry + router
│   │   ├── lib/api.js               # API client
│   │   ├── components/Layout.jsx    # Dashboard sidebar
│   │   └── pages/
│   │       ├── Landing.jsx          # Pitch page (Basecamp-style)
│   │       ├── Login.jsx            # Phone auth
│   │       ├── Dashboard.jsx        # Stats + Drive connect + live polling
│   │       ├── Files.jsx            # File browser + staging area + versions
│   │       ├── Search.jsx           # Full-text search
│   │       └── BotDemo.jsx          # Interactive bot simulator
│   ├── index.html
│   └── vite.config.js
├── whatsapp-bot/
│   ├── bot.js                       # Baileys v7 bot + file download + base64 upload
│   └── package.json
├── .env.example
├── .gitignore
├── Makefile
└── README.md
```

---

## API Endpoints

### Public

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/health` | Health check |
| `POST` | `/api/auth/register` | Register (phone + name) |
| `POST` | `/api/auth/login` | Login (phone) |
| `POST` | `/api/bot/command` | WhatsApp bot commands |
| `POST` | `/api/bot/upload` | WhatsApp bot file upload (base64) |
| `POST` | `/api/waitlist` | Join beta waitlist |

### Protected (Bearer token)

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/me` | Current user |
| `GET` | `/api/dashboard` | Stats, recent files, contributors |
| `GET` | `/api/files` | List files (with group filter) |
| `GET` | `/api/files/search?q=` | Full-text search |
| `GET` | `/api/files/versions?file_id=` | Version history |
| `GET` | `/api/auth/google/status` | Drive connection status |
| `GET` | `/api/auth/google/connect` | Get Google OAuth URL |

---

## Reyna's Personality

She's not just a bot. She's the group's most useful *and most annoying* member.

```
User: /reyna find "DBMS notes"  (at 3 AM)
Reyna: Bhai 3 baje raat ko DBMS? Tera breakup hua hai kya? Anyway, 4 results 📂

User: /reyna status  (during exams)
Reyna: 12 new files. 8 PYQs. Everyone suddenly remembered exams exist 📚😂

User: /reyna help
Reyna: Main karu toh kya karu — basically save karti hoon.
       Commands: add, staged, commit, rm, find, log, status.
       Ab padh le bhai. 📖

User: /reyna commit
Reyna: Boom! 3 files committed. Ab ye permanent hain.
       Exam ke time thank karega mujhe 😎
       ☁️ 3 file(s) pushed to Google Drive!
```

---

## License

MIT — Free forever. Your data stays in YOUR Drive.

---

*Built with ☕ and frustration by students who lost their notes one too many times.*
