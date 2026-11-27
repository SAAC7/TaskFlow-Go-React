import type { ApiErrorBody, Task, TaskInput, TaskPatch } from './types'

const BASE_URL = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080/tasks').replace(/\/$/, '')

/** Error thrown for every failed request, with the data sent back by the API. */
export class ApiError extends Error {
  status: number
  code: string
  details?: Record<string, string>

  /** Builds the error from the HTTP status, the API's error code, its message and the per-field details. */
  constructor(status: number, code: string, message: string, details?: Record<string, string>) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

/**
* Single place that talks to fetch: adds the base URL, sends JSON, parses the
* response and converts every failure into an ApiError.
*/
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  let res: Response
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      ...options,
      headers: options.body ? { 'Content-Type': 'application/json' } : undefined,
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError(0, 'network_error', 'No se pudo conectar con el servidor')
  }

  if (res.status === 204) return undefined as T // DELETE responde sin cuerpo

  const body = await res.json().catch(() => null)
  if (!res.ok) {
    const e = (body as ApiErrorBody | null)?.error
    throw new ApiError(res.status, e?.code ?? 'unknown', e?.message ?? 'Error inesperado', e?.details)
  }
  return body as T
}
/** One function per endpoint of the Go API. */
export const api = {

  /** GET /tasks: returns the tasks (newest first); the signal lets the caller cancel the request. */
  list: (signal?: AbortSignal) => request<Task[]>('/tasks?limit=100', { signal }),

  /** POST /tasks: creates a task and returns it with its generated id and dates. */
  create: (input: TaskInput) =>
    request<Task>('/tasks', { method: 'POST', body: JSON.stringify(input) }),
  /** PATCH /tasks/{id}: changes only the fields sent in the patch and returns the updated task. */
  patch: (id: string, patch: TaskPatch) =>
    request<Task>(`/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),
  /** DELETE /tasks/{id}: removes the task (the API answers 204 with no body). */
  remove: (id: string) => request<void>(`/tasks/${id}`, { method: 'DELETE' }),
}

// Converts any error into text to display to the user.
export function getErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.details) {
      return Object.entries(err.details).map(([campo, msg]) => `${campo}: ${msg}`).join(' · ')
    }
    return err.message
  }
  return 'Ocurrió un error inesperado'
}