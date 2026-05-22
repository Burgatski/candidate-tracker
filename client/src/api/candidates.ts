const API_BASE = 'http://localhost:8080'

export interface Candidate {
  id: number
  first_name: string
  last_name: string
  email: string
  phone: string
  picture: string
  skills: string[]
}

export interface Pagination {
  per_page: number
  page: number
  total_pages: number
}

export interface CandidateListData {
  total: number
  candidates: Candidate[]
  pagination: Pagination
}

export class ApiError extends Error {
  constructor(public errors: string[]) {
    super(errors.join(', '))
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const json = await res.json()
  if (json.status >= 400) throw new ApiError(json.errors ?? ['Unknown error'])
  return json.data as T
}

export const api = {
  list: (page: number) =>
    request<CandidateListData>(`/candidates?page=${page}`),

  get: (id: number) =>
    request<Candidate>(`/candidates/${id}`),

  create: (data: Omit<Candidate, 'id'>) =>
    request<Candidate>('/candidates', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Omit<Candidate, 'id'>) =>
    request<Candidate>(`/candidates/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    request<void>(`/candidates/${id}`, { method: 'DELETE' }),
}
