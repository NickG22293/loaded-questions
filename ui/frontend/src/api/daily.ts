import { supabase } from '@/lib/supabase'
import type {
  Answer,
  ChatMessage,
  DailyQuestion,
  Group,
  GroupDetail,
  GroupMember,
  GuessResult,
  GuessTarget,
  LeaderboardEntry,
} from '@/types/daily'

async function authHeaders(): Promise<Record<string, string>> {
  const { data: { session } } = await supabase.auth.getSession()
  if (!session) throw new Error('Not authenticated')
  return {
    Authorization: `Bearer ${session.access_token}`,
    'Content-Type': 'application/json',
  }
}

async function get<T>(path: string): Promise<T> {
  const headers = await authHeaders()
  const res = await fetch(path, { headers })
  if (!res.ok) throw new Error(await res.text())
  return res.json() as Promise<T>
}

async function post<T>(path: string, body?: unknown): Promise<T> {
  const headers = await authHeaders()
  const res = await fetch(path, {
    method: 'POST',
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) throw new Error(await res.text())
  return res.json() as Promise<T>
}

export async function getToday(): Promise<{ question: DailyQuestion; answer: Answer | null }> {
  return get('/api/daily/today')
}

export async function submitAnswer(answerText: string): Promise<Answer> {
  return post('/api/daily/answers', { answer_text: answerText })
}

export async function listGroups(): Promise<Group[]> {
  return get('/api/daily/groups')
}

export async function createGroup(name: string): Promise<Group> {
  return post('/api/daily/groups', { name })
}

export async function joinGroup(inviteCode: string): Promise<Group> {
  return post('/api/daily/groups/join', { invite_code: inviteCode })
}

export async function getGroup(groupId: string): Promise<GroupDetail> {
  return get(`/api/daily/groups/${groupId}`)
}

export async function getGuessTargets(
  groupId: string
): Promise<{ answers: GuessTarget[]; members: GroupMember[] }> {
  return get(`/api/daily/groups/${groupId}/guesses`)
}

// Body is a flat array — the backend decodes []struct directly.
export async function submitGuesses(
  groupId: string,
  guesses: { answer_id: string; guessed_user_id: string }[]
): Promise<GuessResult[]> {
  return post(`/api/daily/groups/${groupId}/guesses`, guesses)
}

export async function getResults(groupId: string): Promise<GuessResult[]> {
  return get(`/api/daily/groups/${groupId}/results`)
}

export async function getLeaderboard(groupId: string): Promise<LeaderboardEntry[]> {
  return get(`/api/daily/groups/${groupId}/leaderboard`)
}

export async function getChatMessages(groupId: string): Promise<ChatMessage[]> {
  return get(`/api/daily/groups/${groupId}/chat`)
}

export async function sendChatMessage(groupId: string, message: string): Promise<ChatMessage> {
  return post(`/api/daily/groups/${groupId}/chat`, { message })
}

// EventSource can't set headers, so the token is passed as a query param.
// The backend auth middleware accepts ?token= as a fallback.
export function groupEventsUrl(groupId: string, accessToken: string): string {
  return `/api/daily/groups/${groupId}/events?token=${encodeURIComponent(accessToken)}`
}
