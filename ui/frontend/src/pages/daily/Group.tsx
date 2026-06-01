import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useAuth } from '@/contexts/AuthContext'
import {
  getChatMessages,
  getGroup,
  getGuessTargets,
  getLeaderboard,
  getResults,
  groupEventsUrl,
  sendChatMessage,
  submitGuesses,
} from '@/api/daily'
import type {
  ChatMessage,
  GroupDetail,
  GroupMember,
  GuessResult,
  GuessTarget,
  LeaderboardEntry,
} from '@/types/daily'

export function DailyGroup() {
  const { groupId } = useParams<{ groupId: string }>()
  const navigate = useNavigate()
  const { session } = useAuth()

  const [detail, setDetail] = useState<GroupDetail | null>(null)
  const [loading, setLoading] = useState(true)

  const [guessTargets, setGuessTargets] = useState<GuessTarget[]>([])
  const [members, setMembers] = useState<GroupMember[]>([])
  const [assignments, setAssignments] = useState<Record<string, string>>({})
  const [selectedAnswerId, setSelectedAnswerId] = useState<string | null>(null)
  const [submittingGuesses, setSubmittingGuesses] = useState(false)
  const [guessError, setGuessError] = useState('')

  const [results, setResults] = useState<GuessResult[] | null>(null)
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([])
  const [chatInput, setChatInput] = useState('')
  const [sendingMessage, setSendingMessage] = useState(false)
  const chatEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!groupId) return
    getGroup(groupId)
      .then(async (d) => {
        setDetail(d)
        setMembers(d.members)
        if (d.guessing_unlocked && !d.user_guessed) {
          const { answers } = await getGuessTargets(groupId)
          setGuessTargets(answers)
        }
        if (d.user_guessed) {
          const [r, lb, msgs] = await Promise.all([
            getResults(groupId),
            getLeaderboard(groupId),
            getChatMessages(groupId),
          ])
          setResults(r)
          setLeaderboard(lb)
          setChatMessages(msgs)
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [groupId])

  useEffect(() => {
    const token = session?.access_token
    if (!groupId || !token) return

    const es = new EventSource(groupEventsUrl(groupId, token))

    es.addEventListener('answer_progress', (e: MessageEvent) => {
      const progress = JSON.parse(e.data) as GroupDetail['progress']
      setDetail((prev) => (prev ? { ...prev, progress } : prev))
    })

    es.addEventListener('guessing_unlocked', () => {
      setDetail((prev) => (prev ? { ...prev, guessing_unlocked: true } : prev))
      if (!groupId) return
      getGuessTargets(groupId)
        .then(({ answers }) => setGuessTargets(answers))
        .catch(() => {})
    })

    es.addEventListener('chat_message', (e: MessageEvent) => {
      const msg = JSON.parse(e.data) as ChatMessage
      setChatMessages((prev) => [...prev, msg])
    })

    return () => es.close()
  }, [groupId, session])

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatMessages])

  async function handleAssign(memberId: string) {
    if (!selectedAnswerId) return
    setAssignments((prev) => ({ ...prev, [selectedAnswerId]: memberId }))
    setSelectedAnswerId(null)
  }

  async function handleLockIn() {
    if (!groupId) return
    const guesses = guessTargets.map((t) => ({
      answer_id: t.answer_id,
      guessed_user_id: assignments[t.answer_id] ?? '',
    }))
    if (guesses.some((g) => !g.guessed_user_id)) return
    setSubmittingGuesses(true)
    setGuessError('')
    try {
      const r = await submitGuesses(groupId, guesses)
      setResults(r)
      setDetail((prev) => (prev ? { ...prev, user_guessed: true } : prev))
      const [lb, msgs] = await Promise.all([getLeaderboard(groupId), getChatMessages(groupId)])
      setLeaderboard(lb)
      setChatMessages(msgs)
    } catch (err) {
      setGuessError(err instanceof Error ? err.message : 'Failed to submit guesses')
      setSubmittingGuesses(false)
    }
  }

  async function handleSendMessage(e: React.FormEvent) {
    e.preventDefault()
    if (!groupId || !chatInput.trim()) return
    setSendingMessage(true)
    try {
      await sendChatMessage(groupId, chatInput.trim())
      setChatInput('')
    } catch {}
    finally {
      setSendingMessage(false)
    }
  }

  if (loading) {
    return (
      <Layout>
        <div className="max-w-2xl mx-auto flex justify-center py-20">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        </div>
      </Layout>
    )
  }

  if (!detail) {
    return (
      <Layout>
        <div className="max-w-2xl mx-auto space-y-4">
          <p className="text-muted-foreground">Group not found.</p>
          <Button variant="outline" onClick={() => navigate('/daily')}>
            Back
          </Button>
        </div>
      </Layout>
    )
  }

  const { group, progress, guessing_unlocked } = detail
  const allAssigned =
    guessTargets.length > 0 && guessTargets.every((t) => assignments[t.answer_id])

  return (
    <Layout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('/daily')}
            className="text-muted-foreground hover:text-foreground"
          >
            ←
          </button>
          <div>
            <h2 className="text-2xl font-bold">{group.name}</h2>
            <p className="text-xs text-muted-foreground font-mono">{group.invite_code}</p>
          </div>
        </div>

        {/* ── Pre-guessing: progress ── */}
        {!guessing_unlocked && results === null && (
          <Card>
            <CardContent className="pt-6 space-y-4">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Answers submitted</p>
                <span className="text-lg font-bold tabular-nums">
                  {progress.answered} / {progress.total}
                </span>
              </div>
              <div className="h-2 rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-500"
                  style={{
                    width:
                      progress.total > 0
                        ? `${(progress.answered / progress.total) * 100}%`
                        : '0%',
                  }}
                />
              </div>
              <p className="text-sm text-muted-foreground text-center">
                Guessing unlocks when everyone answers or at 6pm.
              </p>
            </CardContent>
          </Card>
        )}

        {/* ── Guessing phase ── */}
        {guessing_unlocked && results === null && (
          <>
            <div>
              <h3 className="text-lg font-semibold">Assign the answers</h3>
              <p className="text-sm text-muted-foreground mt-1">
                Select an answer, then tap a group member to assign it.
              </p>
            </div>

            <div className="space-y-3">
              {guessTargets.map((t) => {
                const isSelected = selectedAnswerId === t.answer_id
                const assignedMember = assignments[t.answer_id]
                  ? members.find((m) => m.id === assignments[t.answer_id])
                  : null
                return (
                  <button
                    key={t.answer_id}
                    onClick={() => setSelectedAnswerId(isSelected ? null : t.answer_id)}
                    className={`w-full text-left rounded-lg border p-4 transition-colors ${
                      isSelected
                        ? 'border-primary bg-primary/10 ring-2 ring-primary'
                        : 'border-border hover:border-primary/50 bg-card'
                    }`}
                  >
                    <p className="text-sm">{t.answer_text}</p>
                    {assignedMember && (
                      <p className="mt-1.5 text-xs text-muted-foreground">
                        Assigned to:{' '}
                        <span className="font-semibold text-foreground">
                          {assignedMember.display_name}
                        </span>
                      </p>
                    )}
                  </button>
                )
              })}
            </div>

            {selectedAnswerId && (
              <Card>
                <CardContent className="pt-4 pb-4">
                  <p className="text-sm font-medium mb-3">Who wrote this?</p>
                  <div className="grid grid-cols-2 gap-2">
                    {members.map((m) => (
                      <Button
                        key={m.id}
                        variant="outline"
                        onClick={() => handleAssign(m.id)}
                        className="justify-start"
                      >
                        {m.display_name}
                      </Button>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}

            {guessError && <p className="text-sm text-destructive">{guessError}</p>}

            <Button
              onClick={handleLockIn}
              disabled={!allAssigned || submittingGuesses}
              className="w-full"
              size="lg"
            >
              {submittingGuesses
                ? 'Submitting…'
                : allAssigned
                  ? 'Lock In'
                  : 'Assign all answers to lock in'}
            </Button>
          </>
        )}

        {/* ── Results ── */}
        {results !== null && (
          <>
            <div>
              <h3 className="text-lg font-semibold">Results</h3>
              <p className="text-sm text-muted-foreground mt-1">
                {results.filter((r) => r.correct).length} / {results.length} correct
              </p>
            </div>

            <div className="space-y-3">
              {results.map((r) => (
                <Card
                  key={r.answer_id}
                  className={`border-2 ${r.correct ? 'border-green-500' : 'border-red-300'}`}
                >
                  <CardContent className="pt-4 pb-4 space-y-2">
                    <p className="italic text-sm">"{r.answer_text}"</p>
                    <div className="text-xs space-y-0.5 text-muted-foreground">
                      <p>
                        You guessed →{' '}
                        <span className="text-foreground font-medium">{r.guessed_name}</span>
                      </p>
                      <p>
                        Actually written by →{' '}
                        <span className="text-foreground font-medium">{r.actual_name}</span>
                      </p>
                    </div>
                    <p
                      className={`text-sm font-semibold ${r.correct ? 'text-green-600' : 'text-red-500'}`}
                    >
                      {r.correct ? '✓ Correct!' : '✗ Wrong'}
                    </p>
                  </CardContent>
                </Card>
              ))}
            </div>

            {leaderboard.length > 0 && (
              <Card>
                <CardContent className="pt-6 space-y-3">
                  <h4 className="font-semibold">Leaderboard</h4>
                  <ul className="divide-y">
                    {leaderboard.map((entry, i) => (
                      <li
                        key={entry.user_id}
                        className="flex items-center justify-between py-2 text-sm"
                      >
                        <span className="flex items-center gap-2">
                          <span className="text-muted-foreground w-5 tabular-nums">{i + 1}.</span>
                          {entry.display_name}
                        </span>
                        <span className="tabular-nums font-medium">{entry.score}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            )}

            <div className="space-y-3">
              <h4 className="font-semibold">Group Chat</h4>
              <Card>
                <CardContent className="pt-4 pb-4">
                  <div className="space-y-3 max-h-64 overflow-y-auto">
                    {chatMessages.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-4">
                        No messages yet. Say something!
                      </p>
                    ) : (
                      chatMessages.map((msg) => (
                        <div key={msg.id} className="space-y-0.5">
                          <p className="text-xs font-medium text-muted-foreground">
                            {msg.display_name}
                          </p>
                          <p className="text-sm">{msg.message}</p>
                        </div>
                      ))
                    )}
                    <div ref={chatEndRef} />
                  </div>
                </CardContent>
              </Card>
              <form onSubmit={handleSendMessage} className="flex gap-2">
                <Input
                  value={chatInput}
                  onChange={(e) => setChatInput(e.target.value)}
                  placeholder="Message your group…"
                  className="flex-1"
                />
                <Button type="submit" disabled={sendingMessage || !chatInput.trim()}>
                  Send
                </Button>
              </form>
            </div>
          </>
        )}
      </div>
    </Layout>
  )
}
