export type HealthResponse = {
  status: string
}

const token = new URLSearchParams(window.location.search).get('token') ?? ''

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: token ? { 'X-Audiobook-Organizer-Token': token } : undefined,
  })
  return decode<T>(response)
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'X-Audiobook-Organizer-Token': token } : {}),
    },
    body: JSON.stringify(body),
  })
  return decode<T>(response)
}

async function decode<T>(response: Response): Promise<T> {
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    const message = typeof payload.error === 'string' ? payload.error : response.statusText
    throw new Error(message)
  }
  return payload as T
}
