import type { Game, Lobby } from '@/types/sessions'

const BASE = '/api/sessions'

function getPlayerToken(lobbyId: string): string {
  return sessionStorage.getItem(`token:${lobbyId}`) ?? ''
}

async function request<T>(path: string, init?: RequestInit, lobbyId?: string): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (lobbyId) {
    const token = getPlayerToken(lobbyId)
    if (token) headers['X-Player-Token'] = token
  }
  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    headers,
    ...init,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

// ── Lobby ──────────────────────────────────────────────────────────────────

export function createLobby(playerName: string, targetScore: number, answerTimerSecs: number) {
  return request<{ lobbyId: string; playerId: string; token: string }>('/lobbies', {
    method: 'POST',
    body: JSON.stringify({ playerName, targetScore, answerTimerSecs }),
  })
}

export function joinLobby(lobbyId: string, playerName: string) {
  return request<{ playerId: string; token: string }>(`/lobbies/${lobbyId}/join`, {
    method: 'POST',
    body: JSON.stringify({ playerName }),
  })
}

export function getLobby(lobbyId: string) {
  return request<Lobby>(`/lobbies/${lobbyId}`)
}

export function startGame(lobbyId: string) {
  return request<void>(`/lobbies/${lobbyId}/start`, { method: 'POST' }, lobbyId)
}

// ── Game ───────────────────────────────────────────────────────────────────

export function getGame(lobbyId: string) {
  return request<Game>(`/lobbies/${lobbyId}/game`)
}

export function submitQuestion(lobbyId: string, question: string) {
  return request<void>(`/lobbies/${lobbyId}/question`, {
    method: 'POST',
    body: JSON.stringify({ question }),
  }, lobbyId)
}

export function submitAnswer(lobbyId: string, answer: string) {
  return request<void>(`/lobbies/${lobbyId}/answer`, {
    method: 'POST',
    body: JSON.stringify({ answer }),
  }, lobbyId)
}

export function assignAnswer(lobbyId: string, answerId: string, playerId: string) {
  return request<void>(`/lobbies/${lobbyId}/assign`, {
    method: 'POST',
    body: JSON.stringify({ answerId, playerId }),
  }, lobbyId)
}

export function lockAssignments(lobbyId: string) {
  return request<void>(`/lobbies/${lobbyId}/lock`, { method: 'POST' }, lobbyId)
}

export function nextRound(lobbyId: string) {
  return request<void>(`/lobbies/${lobbyId}/next`, { method: 'POST' }, lobbyId)
}

export type { Game, Lobby }
