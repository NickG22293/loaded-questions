import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { PlayerList } from '@/components/sessions/PlayerList'
import { Countdown } from '@/components/sessions/Countdown'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import {
  getGame,
  submitQuestion,
  submitAnswer,
  assignAnswer,
  lockAssignments,
  nextRound,
} from '@/api/sessions'
import type { Game as GameType, Round, AnswerCountEvent } from '@/types/sessions'

export function Game() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [game, setGame] = useState<GameType | null>(null)

  // Phase 1
  const [question, setQuestion] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  // Phase 2
  const [answer, setAnswer] = useState('')
  const [answerSubmitted, setAnswerSubmitted] = useState(false)
  const [answerSubmitting, setAnswerSubmitting] = useState(false)
  const [answerError, setAnswerError] = useState('')
  const [answerCount, setAnswerCount] = useState<AnswerCountEvent | null>(null)
  const hasSubmittedRef = useRef(false)

  // Phase 3
  const [selectedAnswerId, setSelectedAnswerId] = useState<string | null>(null)
  const [assignError, setAssignError] = useState('')
  const [locking, setLocking] = useState(false)

  const currentPlayerId = id ? (sessionStorage.getItem(`playerId:${id}`) ?? '') : ''

  useEffect(() => {
    if (!id) return
    getGame(id).then(setGame).catch(console.error)
  }, [id])

  useEffect(() => {
    if (!id) return
    const es = new EventSource(`/api/sessions/lobbies/${id}/events`, { withCredentials: true })

    es.addEventListener('phase_changed', (e: MessageEvent) => {
      const updatedGame = JSON.parse(e.data) as GameType
      setGame(updatedGame)
      hasSubmittedRef.current = false
      setAnswerSubmitted(false)
      setAnswer('')
      setAnswerCount(null)
      setSelectedAnswerId(null)
      setAssignError('')
    })

    es.addEventListener('answer_count', (e: MessageEvent) => {
      const data = JSON.parse(e.data) as AnswerCountEvent
      setAnswerCount(data)
    })

    es.addEventListener('assignment_updated', (e: MessageEvent) => {
      const round = JSON.parse(e.data) as Round
      setGame((prev) => (prev ? { ...prev, currentRound: round } : prev))
    })

    return () => es.close()
  }, [id])

  async function handleSubmitQuestion(e: React.FormEvent) {
    e.preventDefault()
    if (!id || !question.trim()) return
    setSubmitting(true)
    setError('')
    try {
      await submitQuestion(id, question.trim())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit question')
      setSubmitting(false)
    }
  }

  async function handleSubmitAnswer(e: React.FormEvent) {
    e.preventDefault()
    if (!id || !answer.trim()) return
    setAnswerSubmitting(true)
    setAnswerError('')
    try {
      await submitAnswer(id, answer.trim())
      hasSubmittedRef.current = true
      setAnswerSubmitted(true)
    } catch (err) {
      setAnswerError(err instanceof Error ? err.message : 'Failed to submit answer')
    } finally {
      setAnswerSubmitting(false)
    }
  }

  async function handleAssign(targetPlayerId: string) {
    if (!id || !selectedAnswerId) return
    setAssignError('')
    try {
      await assignAnswer(id, selectedAnswerId, targetPlayerId)
      setSelectedAnswerId(null)
    } catch (err) {
      setAssignError(err instanceof Error ? err.message : 'Failed to assign answer')
    }
  }

  async function handleLockAssignments() {
    if (!id) return
    setLocking(true)
    try {
      await lockAssignments(id)
    } catch (err) {
      setAssignError(err instanceof Error ? err.message : 'Failed to lock assignments')
      setLocking(false)
    }
  }

  async function handleNextRound() {
    if (!id) return
    try {
      await nextRound(id)
    } catch {
      // first-click-wins; ignore 409 conflicts from other players clicking first
    }
  }

  if (!game) {
    return (
      <Layout>
        <p className="text-muted-foreground">Loading game…</p>
      </Layout>
    )
  }

  const { currentRound, players } = game
  const phase = currentRound?.phase
  const isAsker = currentRound?.askerId === currentPlayerId
  const askerName = players.find((p) => p.id === currentRound?.askerId)?.name ?? 'Someone'

  // ── GAME_OVER ─────────────────────────────────────────────────────────────

  if (phase === 'GAME_OVER') {
    const winner = players.find((p) => p.id === game.winnerId)
    return (
      <Layout>
        <div className="max-w-2xl mx-auto space-y-8 text-center">
          <div className="space-y-2">
            <p className="text-6xl">🎉</p>
            <h2 className="text-3xl font-bold">{winner?.name ?? 'Someone'} wins!</h2>
            <p className="text-muted-foreground">Final scores</p>
          </div>

          <Card>
            <CardContent className="pt-6">
              <ul className="divide-y">
                {[...players]
                  .sort((a, b) => b.score - a.score)
                  .map((p) => (
                    <li
                      key={p.id}
                      className={`flex items-center justify-between py-3 ${
                        p.id === game.winnerId ? 'font-bold text-primary' : ''
                      }`}
                    >
                      <span>
                        {p.name}
                        {p.id === currentPlayerId ? ' (you)' : ''}
                        {p.id === game.winnerId ? ' 🏆' : ''}
                      </span>
                      <span className="text-lg tabular-nums">{p.score}</span>
                    </li>
                  ))}
              </ul>
            </CardContent>
          </Card>

          <Button variant="outline" onClick={() => navigate('/sessions')}>
            Back to Home
          </Button>
        </div>
      </Layout>
    )
  }

  // ── Phase 1: ASKING ───────────────────────────────────────────────────────

  if (phase === 'ASKING') {
    if (isAsker) {
      return (
        <Layout>
          <div className="max-w-2xl mx-auto space-y-6">
            <div>
              <p className="text-sm text-muted-foreground mb-1">
                Round {currentRound?.roundNumber}
              </p>
              <h2 className="text-2xl font-bold">You're the Asker!</h2>
              <p className="text-muted-foreground mt-1">
                Think of an introspective question for the group.
              </p>
            </div>

            <Card>
              <CardContent className="pt-6">
                <form onSubmit={handleSubmitQuestion} className="space-y-4">
                  <div className="space-y-1.5">
                    <label className="text-sm font-medium">Your question</label>
                    <Textarea
                      value={question}
                      onChange={(e) => setQuestion(e.target.value)}
                      placeholder="e.g. If you could have one superpower, what would it be?"
                      rows={3}
                      autoFocus
                    />
                  </div>
                  {error && <p className="text-sm text-destructive">{error}</p>}
                  <Button
                    type="submit"
                    disabled={submitting || !question.trim()}
                    className="w-full"
                    size="lg"
                  >
                    {submitting ? 'Submitting…' : 'Ask the Group'}
                  </Button>
                </form>
              </CardContent>
            </Card>

            <PlayerList players={players} currentPlayerId={currentPlayerId} />
          </div>
        </Layout>
      )
    }

    return (
      <Layout>
        <div className="max-w-2xl mx-auto space-y-6">
          <div>
            <p className="text-sm text-muted-foreground mb-1">Round {currentRound?.roundNumber}</p>
            <h2 className="text-2xl font-bold">Get ready to answer…</h2>
          </div>

          <Card>
            <CardContent className="flex flex-col items-center py-14 gap-4">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
              <p className="text-muted-foreground text-center">
                Waiting for{' '}
                <span className="font-semibold text-foreground">{askerName}</span> to ask a
                question…
              </p>
            </CardContent>
          </Card>

          <PlayerList players={players} currentPlayerId={currentPlayerId} />
        </div>
      </Layout>
    )
  }

  // ── Phase 2: ANSWERING ────────────────────────────────────────────────────

  if (phase === 'ANSWERING') {
    const roundQuestion = currentRound?.question ?? ''
    const timerStartedAt = currentRound?.phaseStartedAt
    const submitted = answerCount?.submitted ?? 0
    const total = answerCount?.total ?? game.players.length - 1
    const submittedIds = answerCount?.submittedPlayerIds ?? []

    if (isAsker) {
      return (
        <Layout>
          <div className="max-w-2xl mx-auto space-y-6">
            <div>
              <p className="text-sm text-muted-foreground mb-1">Round {currentRound?.roundNumber}</p>
              <h2 className="text-2xl font-bold">Sit tight — you asked the question.</h2>
              <p className="text-muted-foreground mt-1">
                Wait for everyone to submit their answers.
              </p>
            </div>

            <Card>
              <CardContent className="pt-6 space-y-4">
                <blockquote className="border-l-4 pl-4 italic text-lg">{roundQuestion}</blockquote>

                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <span>
                    <span className="font-semibold text-foreground">{submitted}</span> / {total}{' '}
                    answered
                  </span>
                  {timerStartedAt && (
                    <Countdown totalSecs={game.answerTimerSecs} startedAt={timerStartedAt} />
                  )}
                </div>

                <ul className="divide-y">
                  {players.map((p) => {
                    if (p.id === currentRound?.askerId) return null
                    const hasAnswered = submittedIds.includes(p.id)
                    return (
                      <li key={p.id} className="flex items-center justify-between py-2 text-sm">
                        <span>{p.name}</span>
                        {hasAnswered ? (
                          <span className="text-green-600 font-medium">✓</span>
                        ) : (
                          <span className="text-muted-foreground">…</span>
                        )}
                      </li>
                    )
                  })}
                </ul>
              </CardContent>
            </Card>
          </div>
        </Layout>
      )
    }

    return (
      <Layout>
        <div className="max-w-2xl mx-auto space-y-6">
          <div>
            <p className="text-sm text-muted-foreground mb-1">Round {currentRound?.roundNumber}</p>
            <h2 className="text-2xl font-bold">Answer the question</h2>
          </div>

          <Card>
            <CardContent className="pt-6 space-y-4">
              <blockquote className="border-l-4 pl-4 italic text-lg">{roundQuestion}</blockquote>

              {answerSubmitted ? (
                <div className="space-y-3">
                  <p className="text-sm text-green-600 font-medium">Your answer is in!</p>
                  <div className="rounded-md bg-muted px-4 py-3 text-sm">{answer}</div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setAnswerSubmitted(false)}
                  >
                    Edit answer
                  </Button>
                </div>
              ) : (
                <form onSubmit={handleSubmitAnswer} className="space-y-4">
                  <Textarea
                    value={answer}
                    onChange={(e) => setAnswer(e.target.value)}
                    placeholder="Be honest — your answer is anonymous until the end."
                    rows={4}
                    autoFocus
                  />
                  {answerError && <p className="text-sm text-destructive">{answerError}</p>}
                  <Button
                    type="submit"
                    disabled={answerSubmitting || !answer.trim()}
                    className="w-full"
                    size="lg"
                  >
                    {answerSubmitting
                      ? 'Submitting…'
                      : hasSubmittedRef.current
                      ? 'Update Answer'
                      : 'Submit Answer'}
                  </Button>
                </form>
              )}

              <div className="flex items-center justify-between text-sm text-muted-foreground border-t pt-3">
                <span>
                  <span className="font-semibold text-foreground">{submitted}</span> / {total}{' '}
                  answered
                </span>
                {timerStartedAt && (
                  <Countdown totalSecs={game.answerTimerSecs} startedAt={timerStartedAt} />
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      </Layout>
    )
  }

  // ── Phase 3: ASSIGNING ────────────────────────────────────────────────────

  if (phase === 'ASSIGNING') {
    const answers = currentRound?.answers ?? []
    const allAssigned = answers.length > 0 && answers.every((a) => a.assignedTo)
    const answerers = players.filter((p) => p.id !== currentRound?.askerId)

    if (isAsker) {
      return (
        <Layout>
          <div className="max-w-2xl mx-auto space-y-6">
            <div>
              <p className="text-sm text-muted-foreground mb-1">Round {currentRound?.roundNumber}</p>
              <h2 className="text-2xl font-bold">Assign the answers</h2>
              <p className="text-muted-foreground mt-1">
                Select an answer, then click a player name to assign it.
              </p>
            </div>

            <blockquote className="border-l-4 pl-4 italic text-lg">
              {currentRound?.question}
            </blockquote>

            <div className="space-y-3">
              {answers.map((a) => {
                const isSelected = selectedAnswerId === a.id
                const assignedPlayer = a.assignedTo
                  ? players.find((p) => p.id === a.assignedTo)
                  : null
                return (
                  <button
                    key={a.id}
                    onClick={() => setSelectedAnswerId(isSelected ? null : a.id)}
                    className={`w-full text-left rounded-lg border p-4 transition-colors ${
                      isSelected
                        ? 'border-primary bg-primary/10 ring-2 ring-primary'
                        : 'border-border hover:border-primary/50 bg-card'
                    }`}
                  >
                    <p className="text-sm">{a.text}</p>
                    {assignedPlayer && (
                      <p className="mt-1.5 text-xs text-muted-foreground">
                        Assigned to:{' '}
                        <span className="font-semibold text-foreground">{assignedPlayer.name}</span>
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
                    {answerers.map((p) => (
                      <Button
                        key={p.id}
                        variant="outline"
                        onClick={() => handleAssign(p.id)}
                        className="justify-start"
                      >
                        {p.name}
                      </Button>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}

            {assignError && <p className="text-sm text-destructive">{assignError}</p>}

            <Button
              onClick={handleLockAssignments}
              disabled={!allAssigned || locking}
              className="w-full"
              size="lg"
            >
              {locking ? 'Locking in…' : `Lock In${!allAssigned ? ' (assign all answers first)' : ''}`}
            </Button>
          </div>
        </Layout>
      )
    }

    const assignedAnswers = answers.filter((a) => a.assignedTo)
    return (
      <Layout>
        <div className="max-w-2xl mx-auto space-y-6">
          <div>
            <p className="text-sm text-muted-foreground mb-1">Round {currentRound?.roundNumber}</p>
            <h2 className="text-2xl font-bold">The Asker is reviewing your answers…</h2>
          </div>

          <blockquote className="border-l-4 pl-4 italic text-lg">
            {currentRound?.question}
          </blockquote>

          <Card>
            <CardContent className="pt-6">
              {assignedAnswers.length === 0 ? (
                <div className="flex flex-col items-center py-8 gap-3">
                  <div className="h-6 w-6 animate-spin rounded-full border-4 border-primary border-t-transparent" />
                  <p className="text-sm text-muted-foreground">Waiting for first assignment…</p>
                </div>
              ) : (
                <ul className="divide-y">
                  {assignedAnswers.map((a) => {
                    const assignedPlayer = players.find((p) => p.id === a.assignedTo)
                    const isMe = a.assignedTo === currentPlayerId
                    return (
                      <li
                        key={a.id}
                        className={`py-3 space-y-0.5 ${isMe ? 'text-primary' : ''}`}
                      >
                        <p className={`text-sm italic ${isMe ? 'font-medium' : ''}`}>
                          "{a.text}"
                        </p>
                        <p className="text-xs text-muted-foreground">
                          →{' '}
                          <span
                            className={
                              isMe ? 'text-primary font-semibold' : 'text-foreground font-medium'
                            }
                          >
                            {assignedPlayer?.name ?? '?'}
                          </span>
                          {isMe && ' (you)'}
                        </p>
                      </li>
                    )
                  })}
                </ul>
              )}
            </CardContent>
          </Card>

          <p className="text-sm text-muted-foreground text-center">
            {assignedAnswers.length} / {answers.length} assigned
          </p>
        </div>
      </Layout>
    )
  }

  // ── Phase 4: SCORING ──────────────────────────────────────────────────────

  if (phase === 'SCORING') {
    const answers = currentRound?.answers ?? []
    const roundScore = answers.filter((a) => a.assignedTo === a.playerId).length
    const askerPlayer = players.find((p) => p.id === currentRound?.askerId)

    return (
      <Layout>
        <div className="max-w-2xl mx-auto space-y-6">
          <div>
            <p className="text-sm text-muted-foreground mb-1">
              Round {currentRound?.roundNumber} results
            </p>
            <h2 className="text-2xl font-bold">Round over!</h2>
          </div>

          <blockquote className="border-l-4 pl-4 italic text-lg">
            {currentRound?.question}
          </blockquote>

          <div className="space-y-3">
            {answers.map((a) => {
              const author = players.find((p) => p.id === a.playerId)
              const assignedTo = players.find((p) => p.id === a.assignedTo)
              const correct = a.assignedTo === a.playerId
              return (
                <Card
                  key={a.id}
                  className={`border-2 ${correct ? 'border-green-500' : 'border-red-300'}`}
                >
                  <CardContent className="pt-4 pb-4 space-y-2">
                    <p className="italic text-sm">"{a.text}"</p>
                    <div className="text-xs space-y-0.5 text-muted-foreground">
                      <p>
                        Asker guessed →{' '}
                        <span className="text-foreground font-medium">
                          {assignedTo?.name ?? '?'}
                        </span>
                      </p>
                      <p>
                        Actually written by →{' '}
                        <span className="text-foreground font-medium">{author?.name ?? '?'}</span>
                      </p>
                    </div>
                    <p
                      className={`text-sm font-semibold ${
                        correct ? 'text-green-600' : 'text-red-500'
                      }`}
                    >
                      {correct ? '✓ Correct!' : '✗ Wrong'}
                    </p>
                  </CardContent>
                </Card>
              )
            })}
          </div>

          <Card>
            <CardContent className="pt-6 space-y-3">
              <p className="font-semibold">
                {askerPlayer?.name ?? 'The Asker'} scored{' '}
                <span className="text-primary">
                  +{roundScore} point{roundScore !== 1 ? 's' : ''}
                </span>{' '}
                this round!
              </p>

              <div className="border-t pt-3">
                <h4 className="text-sm font-medium text-muted-foreground mb-2">Scoreboard</h4>
                <ul className="divide-y">
                  {[...players]
                    .sort((a, b) => b.score - a.score)
                    .map((p) => (
                      <li
                        key={p.id}
                        className={`flex items-center justify-between py-2 text-sm ${
                          p.id === currentPlayerId ? 'font-bold' : ''
                        }`}
                      >
                        <span>
                          {p.name}
                          {p.id === currentPlayerId ? ' (you)' : ''}
                        </span>
                        <span className="tabular-nums">{p.score}</span>
                      </li>
                    ))}
                </ul>
              </div>
            </CardContent>
          </Card>

          <Button onClick={handleNextRound} className="w-full" size="lg">
            Next Round
          </Button>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-2xl mx-auto">
        <p className="text-muted-foreground">Unknown phase: {phase}</p>
      </div>
    </Layout>
  )
}
