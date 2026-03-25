const {
  makeWASocket, useMultiFileAuthState, DisconnectReason,
  downloadMediaMessage, getContentType,
  makeCacheableSignalKeyStore, fetchLatestBaileysVersion,
} = require('baileys');
const pino = require('pino');
const qrcode = require('qrcode-terminal');
const fs = require('fs');
const path = require('path');

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';
const AUTH_DIR = process.env.AUTH_DIR || './auth_state';
const logger = pino({ level: 'silent' });

// ─── Backend API ───

// Send a command (no file data)
async function sendCommand(groupJid, command, userPhone, userName) {
  try {
    const res = await fetch(`${BACKEND_URL}/api/bot/command`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        group_wa_id: groupJid, command,
        user_phone: userPhone, user_name: userName || userPhone,
      }),
    });
    return await res.json();
  } catch (err) {
    console.error('Backend error:', err.message);
    return { reply: 'Backend se baat nahi ho rahi 😢' };
  }
}

// Upload a file with binary data to backend
async function uploadFile(groupJid, userPhone, userName, fileInfo, fileBuffer) {
  try {
    const b64 = fileBuffer.toString('base64');
    const res = await fetch(`${BACKEND_URL}/api/bot/upload`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        group_wa_id: groupJid,
        user_phone: userPhone,
        user_name: userName || userPhone,
        file_name: fileInfo.fileName,
        file_size: fileInfo.fileSize,
        mime_type: fileInfo.mimeType,
        subject: fileInfo.subject || '',
        file_data: b64,
      }),
    });
    return await res.json();
  } catch (err) {
    console.error('Upload error:', err.message);
    return { reply: 'File upload fail ho gaya 😢 Backend check karo.' };
  }
}

// ─── Pending Files ───
const pendingFiles = new Map();

function trackFile(groupJid, info) {
  pendingFiles.set(groupJid, { ...info, ts: Date.now() });
  setTimeout(() => {
    const c = pendingFiles.get(groupJid);
    if (c && c.ts === info.ts) pendingFiles.delete(groupJid);
  }, 10 * 60 * 1000);
}

function phoneFromJid(jid) {
  return '+' + jid.split('@')[0].split(':')[0];
}

function guessSubject(fileName) {
  const f = fileName.toLowerCase();
  const map = {
    'dsa': 'DSA', 'data structure': 'DSA', 'algorithm': 'DAA',
    'os ': 'OS', 'operating system': 'OS',
    'cn ': 'CN', 'computer network': 'CN', 'networking': 'CN',
    'dbms': 'DBMS', 'database': 'DBMS', 'sql': 'DBMS',
    'daa': 'DAA', 'coa': 'COA', 'computer organization': 'COA',
    'math': 'Maths', 'physics': 'Physics', 'chemistry': 'Chemistry',
    'module': 'Module', 'unit': 'Notes', 'lecture': 'Notes',
    'lab': 'Lab', 'assignment': 'Assignment', 'pyq': 'PYQ',
    'previous year': 'PYQ', 'question paper': 'PYQ',
  };
  for (const [key, val] of Object.entries(map)) {
    if (f.includes(key)) return val;
  }
  return '';
}

// ─── Message Handler ───
async function handleMessage(sock, msg) {
  const chat = msg.key.remoteJid;
  if (!chat?.endsWith('@g.us')) return;
  if (msg.key.fromMe) return;

  const sender = msg.key.participant || chat;
  const senderPhone = phoneFromJid(sender);
  const pushName = msg.pushName || senderPhone;
  const messageType = getContentType(msg.message);

  // Track documents
  if (messageType === 'documentMessage' || messageType === 'documentWithCaptionMessage') {
    const doc = msg.message.documentMessage ||
      msg.message?.documentWithCaptionMessage?.message?.documentMessage;
    if (doc) {
      try {
        const buffer = await downloadMediaMessage(msg, 'buffer', {}, {
          logger, reuploadRequest: sock.updateMediaMessage,
        });
        const fileName = doc.fileName || 'file';
        trackFile(chat, {
          fileName, buffer,
          mimeType: doc.mimetype || 'application/octet-stream',
          fileSize: Number(doc.fileLength || buffer.length),
          sender: senderPhone, senderName: pushName,
          subject: guessSubject(fileName),
        });
        console.log(`📎 Tracked: ${fileName} (${(buffer.length/1024).toFixed(0)}KB) from ${pushName}`);
      } catch (err) {
        console.error('Download failed:', err.message);
      }
    }
    return;
  }

  const text = msg.message?.conversation || msg.message?.extendedTextMessage?.text || '';
  if (!text.trim().toLowerCase().startsWith('/reyna')) return;

  console.log(`💬 ${pushName}: ${text}`);
  const parts = text.trim().split(/\s+/);
  const action = parts[1]?.toLowerCase();

  // ── /reyna add — stages file with actual binary data ──
  if (action === 'add') {
    const arg = parts.slice(2).join(' ').trim();
    const pending = pendingFiles.get(chat);

    if (arg === '.' || arg === '') {
      if (!pending) {
        await sock.sendMessage(chat, {
          text: '🤖 *Reyna:* Koi file toh bhejo pehle! Hawa mein se save nahi kar sakti. 💨\n\nFile bhejo → phir `/reyna add .`',
        });
        return;
      }

      // Upload file WITH binary data to backend
      console.log(`📤 Uploading ${pending.fileName} (${(pending.buffer.length/1024).toFixed(0)}KB) to backend...`);
      const resp = await uploadFile(chat, senderPhone, pushName, {
        fileName: pending.fileName,
        fileSize: pending.fileSize,
        mimeType: pending.mimeType,
        subject: pending.subject,
      }, pending.buffer);

      await sock.sendMessage(chat, { text: `🤖 *Reyna:* ${resp.reply}` });
      pendingFiles.delete(chat);
      return;
    }

    // /reyna add FileName.pdf
    if (pending && pending.fileName.toLowerCase().includes(arg.toLowerCase().split('.')[0])) {
      console.log(`📤 Uploading ${pending.fileName} to backend...`);
      const resp = await uploadFile(chat, senderPhone, pushName, {
        fileName: pending.fileName,
        fileSize: pending.fileSize,
        mimeType: pending.mimeType,
        subject: pending.subject || guessSubject(arg),
      }, pending.buffer);
      await sock.sendMessage(chat, { text: `🤖 *Reyna:* ${resp.reply}` });
      pendingFiles.delete(chat);
    } else {
      // No matching pending file — just record metadata
      const resp = await sendCommand(chat, text, senderPhone, pushName);
      await sock.sendMessage(chat, { text: `🤖 *Reyna:* ${resp.reply}` });
    }
    return;
  }

  // ── All other commands: commit, rm, staged, find, log, status, help ──
  const resp = await sendCommand(chat, text, senderPhone, pushName);
  await sock.sendMessage(chat, { text: `🤖 *Reyna:* ${resp.reply}` });
}

// ─── Start ───
async function startBot() {
  const { state, saveCreds } = await useMultiFileAuthState(AUTH_DIR);
  const { version } = await fetchLatestBaileysVersion();
  console.log(`   WA Web version: ${version.join('.')}`);

  const sock = makeWASocket({
    version,
    auth: { creds: state.creds, keys: makeCacheableSignalKeyStore(state.keys, logger) },
    logger, browser: ['Reyna', 'Chrome', '4.0.0'],
    syncFullHistory: false, markOnlineOnConnect: false,
  });

  sock.ev.on('connection.update', (update) => {
    const { connection, lastDisconnect, qr } = update;
    if (qr) {
      console.log('\n📱 Scan with WhatsApp → Linked Devices → Link a Device:\n');
      qrcode.generate(qr, { small: true });
    }
    if (connection === 'open') {
      console.log('\n╔══════════════════════════════════════════╗');
      console.log('║  🤖 Reyna Bot — CONNECTED                 ║');
      console.log('║  Files upload to backend with binary data  ║');
      console.log('╚══════════════════════════════════════════╝\n');
    }
    if (connection === 'close') {
      const code = lastDisconnect?.error?.output?.statusCode;
      console.log(`❌ Disconnected (${code})`);
      if ([401, 403, 405].includes(code)) {
        fs.rmSync(AUTH_DIR, { recursive: true, force: true });
        setTimeout(startBot, 2000);
      } else {
        setTimeout(startBot, 5000);
      }
    }
  });

  sock.ev.on('creds.update', saveCreds);
  sock.ev.on('messages.upsert', async ({ messages }) => {
    for (const msg of messages) {
      try { await handleMessage(sock, msg); }
      catch (err) { console.error('Error:', err.message); }
    }
  });
}

console.log('\n🤖 Reyna WhatsApp Bot starting...');
console.log(`   Backend: ${BACKEND_URL}`);
console.log(`   Auth: ${AUTH_DIR}`);

if (fs.existsSync(path.join(AUTH_DIR, 'creds.json'))) {
  try { JSON.parse(fs.readFileSync(path.join(AUTH_DIR, 'creds.json'), 'utf8')); }
  catch { console.log('⚠️ Cleaning stale auth...'); fs.rmSync(AUTH_DIR, { recursive: true, force: true }); }
}

startBot().catch(err => { console.error('Fatal:', err); process.exit(1); });
