const BASE = '/api'

async function request(path, opts = {}) {
  const res = await fetch(BASE + path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  })
  if (res.status === 401) {
    window.location.href = '/login'
    return
  }
  return res
}

export const api = {
  login: (username, password) =>
    request('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),

  logout: () => request('/auth/logout', { method: 'POST' }),

  subdomains: () => request('/subdomains').then(r => r.json()),
  createSubdomain: (data) => request('/subdomains', { method: 'POST', body: JSON.stringify(data) }),
  deleteSubdomain: (id) => request(`/subdomains/${id}`, { method: 'DELETE' }),
  updatePrivacy: (id, data) => request(`/subdomains/${id}/privacy`, { method: 'PATCH', body: JSON.stringify(data) }),
  updateRateLimit: (id, data) => request(`/subdomains/${id}/ratelimit`, { method: 'PATCH', body: JSON.stringify(data) }),

  health: () => request('/health').then(r => r.json()),
  subdomainHealth: (id) => request(`/health/${id}`).then(r => r.json()),

  dnsUpdate: () => request('/dns/update', { method: 'POST' }),
}
