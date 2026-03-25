const API = '/api';

function getHeaders() {
  const h = { 'Content-Type': 'application/json' };
  const token = localStorage.getItem('reyna_token');
  if (token) h['Authorization'] = `Bearer ${token}`;
  return h;
}

async function req(path, opts = {}) {
  const res = await fetch(`${API}${path}`, { headers: getHeaders(), ...opts });
  if (res.status === 401) {
    localStorage.removeItem('reyna_token');
    localStorage.removeItem('reyna_user');
    window.location.href = '/login';
    return null;
  }
  return res.json();
}

export const api = {
  // Auth
  register: (phone, name) => req('/auth/register', { method: 'POST', body: JSON.stringify({ phone, name }) }),
  login: (phone) => req('/auth/login', { method: 'POST', body: JSON.stringify({ phone }) }),
  me: () => req('/me'),

  // Dashboard
  dashboard: () => req('/dashboard'),

  // Groups
  groups: () => req('/groups'),
  createGroup: (wa_id, name) => req('/groups', { method: 'POST', body: JSON.stringify({ wa_id, name }) }),

  // Files
  files: (groupId, limit) => req(`/files?group_id=${groupId || ''}&limit=${limit || 50}`),
  search: (q, groupId) => req(`/files/search?q=${encodeURIComponent(q)}&group_id=${groupId || ''}`),
  versions: (fileId) => req(`/files/versions?file_id=${fileId}`),
  upload: (data) => req('/files/upload', { method: 'POST', body: JSON.stringify(data) }),

  // Activity
  activity: (groupId) => req(`/activity?group_id=${groupId}`),

  // Bot command (simulates WhatsApp bot)
  botCommand: (groupWaId, command, userPhone) =>
    req('/bot/command', { method: 'POST', body: JSON.stringify({ group_wa_id: groupWaId, command, user_phone: userPhone }) }),

  // Waitlist
  joinWaitlist: (contact) => req('/waitlist', { method: 'POST', body: JSON.stringify({ contact }) }),
  waitlistCount: () => req('/waitlist'),

  // Health
  health: () => req('/health'),

  // Google Drive
  googleStatus: () => req('/auth/google/status'),
  googleConnect: () => req('/auth/google/connect'),
};

export function saveAuth(data) {
  if (data?.token) localStorage.setItem('reyna_token', data.token);
  if (data?.user) localStorage.setItem('reyna_user', JSON.stringify(data.user));
}

export function getUser() {
  try { return JSON.parse(localStorage.getItem('reyna_user')); } catch { return null; }
}

export function isLoggedIn() {
  return !!localStorage.getItem('reyna_token');
}

export function logout() {
  localStorage.removeItem('reyna_token');
  localStorage.removeItem('reyna_user');
}
