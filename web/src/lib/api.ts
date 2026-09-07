// Petit client HTTP typé vers l'API Go. Les cookies de session sont envoyés
// automatiquement (même origine ; en dev, Vite proxifie /api vers :8080).

async function req(method: string, url: string, body?: unknown): Promise<any> {
  const opts: RequestInit = { method, headers: {} }
  if (body !== undefined) {
    opts.headers = { 'Content-Type': 'application/json' }
    opts.body = JSON.stringify(body)
  }
  const res = await fetch(url, opts)
  const text = await res.text()
  const data = text ? JSON.parse(text) : null
  if (!res.ok) {
    throw new Error((data && data.error) || `Erreur ${res.status}`)
  }
  return data
}

async function reqForm(url: string, formData: FormData): Promise<any> {
  const res = await fetch(url, { method: 'POST', body: formData })
  const text = await res.text()
  const data = text ? JSON.parse(text) : null
  if (!res.ok) {
    throw new Error((data && data.error) || `Erreur ${res.status}`)
  }
  return data
}

export const api = {
  get: (url: string) => req('GET', url),
  post: (url: string, body?: unknown) => req('POST', url, body),
  put: (url: string, body?: unknown) => req('PUT', url, body),
  patch: (url: string, body?: unknown) => req('PATCH', url, body),
  del: (url: string) => req('DELETE', url),
  postForm: (url: string, formData: FormData) => reqForm(url, formData),
}

export type User = {
  id: number
  email: string
  display_name: string
  is_admin: boolean
  created_at?: string
}
