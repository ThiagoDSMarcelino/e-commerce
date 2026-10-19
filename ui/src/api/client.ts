// Única base URL conhecida pelo frontend: o API gateway (main).
const API_URL = (import.meta.env.VITE_API_URL ?? '/api').replace(/\/$/, '')

export function apiUrl(path: string) {
  return `${API_URL}${path}`
}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

type ApiFetchInit = Omit<RequestInit, 'body'> & { json?: unknown }

export async function apiFetch<T>(path: string, init: ApiFetchInit = {}): Promise<T> {
  const { json, headers: extraHeaders, ...rest } = init

  const headers = new Headers(extraHeaders)
  if (!headers.has('Accept')) headers.set('Accept', 'application/json')
  if (json !== undefined) headers.set('Content-Type', 'application/json')

  let res: Response
  try {
    res = await fetch(apiUrl(path), {
      ...rest,
      headers,
      body: json !== undefined ? JSON.stringify(json) : undefined,
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError(0, 'Não foi possível conectar à API')
  }

  if (!res.ok) {
    throw new ApiError(res.status, await readErrorMessage(res))
  }

  const text = await res.text()
  return (text ? JSON.parse(text) : undefined) as T
}

async function readErrorMessage(res: Response) {
  try {
    const body: unknown = await res.json()
    if (body && typeof body === 'object' && 'error' in body && typeof body.error === 'string') {
      return body.error
    }
  } catch {
    // corpo vazio ou não-JSON
  }
  return `Erro ${res.status}${res.statusText ? ` - ${res.statusText}` : ''}`
}
