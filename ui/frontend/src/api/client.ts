import type { Game, Lobby } from '@/types'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
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
  return request<{ lobbyId: string; playerId: string }>('/lobbies', {
    method: 'POST',
    body: JSON.stringify({ playerName, targetScore, answerTimerSecs }),
  })
}

export function joinLobby(lobbyId: string, playerName: string) {
  return request<{ playerId: string }>(`/lobbies/${lobbyId}/join`, {
    method: 'POST',
    body: JSON.stringify({ playerName }),
  })
}

export function getLobby(lobbyId: string) {
  return request<Lobby>(`/lobbies/${lobbyId}`)
}

export function startGame(lobbyId: string) {
  return request<void>(`/lobbies/${lobbyId}/start`, { method: 'POST' })
}

// ── Game ───────────────────────────────────────────────────────────────────

export function submitQuestion(gameId: string, question: string) {
  return request<void>(`/games/${gameId}/question`, {
    method: 'POST',
    body: JSON.stringify({ question }),
  })
}

export function submitAnswer(gameId: string, answer: string) {
  return request<void>(`/games/${gameId}/answer`, {
    method: 'POST',
    body: JSON.stringify({ answer }),
  })
}

export function assignAnswer(gameId: string, answerId: string, playerId: string) {
  return request<void>(`/games/${gameId}/assign`, {
    method: 'POST',
    body: JSON.stringify({ answerId, playerId }),
  })
}

export function lockAssignments(gameId: string) {
  return request<void>(`/games/${gameId}/lock`, { method: 'POST' })
}

export function nextRound(gameId: string) {
  return request<void>(`/games/${gameId}/next`, { method: 'POST' })
}

export type { Game, Lobby }
