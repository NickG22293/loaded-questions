import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useAuth } from '@/contexts/AuthContext'
import { createGroup, getToday, joinGroup, listGroups, submitAnswer } from '@/api/daily'
import type { Answer, DailyQuestion, Group } from '@/types/daily'

export function DailyHome() {
  const { user } = useAuth()
  const navigate = useNavigate()

  const [question, setQuestion] = useState<DailyQuestion | null>(null)
  const [myAnswer, setMyAnswer] = useState<Answer | null>(null)
  const [answerText, setAnswerText] = useState('')
  const [submittingAnswer, setSubmittingAnswer] = useState(false)
  const [answerError, setAnswerError] = useState('')
  const [loadingToday, setLoadingToday] = useState(true)

  const [groups, setGroups] = useState<Group[]>([])
  const [loadingGroups, setLoadingGroups] = useState(true)

  const [newGroupName, setNewGroupName] = useState('')
  const [creatingGroup, setCreatingGroup] = useState(false)
  const [groupError, setGroupError] = useState('')

  const [joinCode, setJoinCode] = useState('')
  const [joiningGroup, setJoiningGroup] = useState(false)
  const [joinError, setJoinError] = useState('')

  useEffect(() => {
    getToday()
      .then(({ question, answer }) => {
        setQuestion(question)
        setMyAnswer(answer)
        if (answer) setAnswerText(answer.answer_text)
      })
      .catch(() => {})
      .finally(() => setLoadingToday(false))
  }, [])

  useEffect(() => {
    listGroups()
      .then(setGroups)
      .catch(() => {})
      .finally(() => setLoadingGroups(false))
  }, [])

  async function handleSubmitAnswer(e: React.FormEvent) {
    e.preventDefault()
    if (!answerText.trim()) return
    setSubmittingAnswer(true)
    setAnswerError('')
    try {
      const a = await submitAnswer(answerText.trim())
      setMyAnswer(a)
    } catch (err) {
      setAnswerError(err instanceof Error ? err.message : 'Failed to submit')
    } finally {
      setSubmittingAnswer(false)
    }
  }

  async function handleCreateGroup(e: React.FormEvent) {
    e.preventDefault()
    if (!newGroupName.trim()) return
    setCreatingGroup(true)
    setGroupError('')
    try {
      const g = await createGroup(newGroupName.trim())
      setGroups((prev) => [...prev, g])
      setNewGroupName('')
    } catch (err) {
      setGroupError(err instanceof Error ? err.message : 'Failed to create group')
    } finally {
      setCreatingGroup(false)
    }
  }

  async function handleJoinGroup(e: React.FormEvent) {
    e.preventDefault()
    if (!joinCode.trim()) return
    setJoiningGroup(true)
    setJoinError('')
    try {
      const g = await joinGroup(joinCode.trim().toUpperCase())
      setGroups((prev) => (prev.find((x) => x.id === g.id) ? prev : [...prev, g]))
      setJoinCode('')
    } catch (err) {
      setJoinError(err instanceof Error ? err.message : 'Group not found')
    } finally {
      setJoiningGroup(false)
    }
  }

  const firstName = (user?.user_metadata as { full_name?: string } | undefined)
    ?.full_name?.split(' ')[0]

  return (
    <Layout>
      <div className="max-w-2xl mx-auto space-y-8">
        <div>
          <h2 className="text-2xl font-bold">Hey, {firstName ?? user?.email ?? 'there'}!</h2>
          <p className="text-muted-foreground mt-1">Here's today's question.</p>
        </div>

        {loadingToday ? (
          <Card>
            <CardContent className="flex justify-center py-10">
              <div className="h-6 w-6 animate-spin rounded-full border-4 border-primary border-t-transparent" />
            </CardContent>
          </Card>
        ) : !question ? (
          <Card>
            <CardContent className="py-10 text-center text-muted-foreground">
              No question scheduled for today.
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="pt-6 space-y-4">
              <blockquote className="border-l-4 pl-4 italic text-lg">{question.question}</blockquote>

              {myAnswer ? (
                <div className="space-y-2">
                  <p className="text-sm font-medium text-green-600">Your answer is in!</p>
                  <div className="rounded-md bg-muted px-4 py-3 text-sm">{myAnswer.answer_text}</div>
                </div>
              ) : (
                <form onSubmit={handleSubmitAnswer} className="space-y-3">
                  <Textarea
                    value={answerText}
                    onChange={(e) => setAnswerText(e.target.value)}
                    placeholder="Your answer (anonymous to your groups until guessing opens)"
                    rows={3}
                    autoFocus
                  />
                  {answerError && <p className="text-sm text-destructive">{answerError}</p>}
                  <Button
                    type="submit"
                    disabled={submittingAnswer || !answerText.trim()}
                    className="w-full"
                  >
                    {submittingAnswer ? 'Submitting…' : 'Submit Answer'}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>
        )}

        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Your Groups</h3>

          {loadingGroups ? (
            <p className="text-sm text-muted-foreground">Loading groups…</p>
          ) : groups.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No groups yet — create one or join with a code below.
            </p>
          ) : (
            <div className="space-y-2">
              {groups.map((g) => (
                <button
                  key={g.id}
                  onClick={() => navigate(`/daily/groups/${g.id}`)}
                  className="w-full text-left rounded-lg border border-border bg-card p-4 hover:border-primary/50 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{g.name}</span>
                    <span className="text-xs text-muted-foreground font-mono">{g.invite_code}</span>
                  </div>
                </button>
              ))}
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 pt-2">
            <form onSubmit={handleCreateGroup} className="space-y-2">
              <p className="text-sm font-medium">Create a group</p>
              <Input
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                placeholder="Group name"
              />
              {groupError && <p className="text-xs text-destructive">{groupError}</p>}
              <Button
                type="submit"
                variant="outline"
                className="w-full"
                disabled={creatingGroup || !newGroupName.trim()}
              >
                {creatingGroup ? 'Creating…' : 'Create'}
              </Button>
            </form>

            <form onSubmit={handleJoinGroup} className="space-y-2">
              <p className="text-sm font-medium">Join a group</p>
              <Input
                value={joinCode}
                onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
                placeholder="Invite code"
                className="font-mono"
              />
              {joinError && <p className="text-xs text-destructive">{joinError}</p>}
              <Button
                type="submit"
                variant="outline"
                className="w-full"
                disabled={joiningGroup || !joinCode.trim()}
              >
                {joiningGroup ? 'Joining…' : 'Join'}
              </Button>
            </form>
          </div>
        </div>
      </div>
    </Layout>
  )
}
