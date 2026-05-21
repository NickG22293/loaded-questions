import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { createLobby, joinLobby } from '@/api/client'

export function Home() {
  const navigate = useNavigate()
  const [view, setView] = useState<'menu' | 'create' | 'join'>('menu')

  // Create form state
  const [createName, setCreateName] = useState('')
  const [targetScore, setTargetScore] = useState(10)
  const [timerSecs, setTimerSecs] = useState(60)
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  // Join form state
  const [joinName, setJoinName] = useState('')
  const [lobbyCode, setLobbyCode] = useState('')
  const [joining, setJoining] = useState(false)
  const [joinError, setJoinError] = useState('')

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    setCreateError('')
    try {
      const { lobbyId } = await createLobby(createName, targetScore, timerSecs)
      navigate(`/lobby/${lobbyId}`)
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create lobby')
    } finally {
      setCreating(false)
    }
  }

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault()
    setJoining(true)
    setJoinError('')
    try {
      await joinLobby(lobbyCode.toUpperCase(), joinName)
      navigate(`/lobby/${lobbyCode.toUpperCase()}`)
    } catch (err) {
      setJoinError(err instanceof Error ? err.message : 'Failed to join lobby')
    } finally {
      setJoining(false)
    }
  }

  if (view === 'create') {
    return (
      <Layout>
        <div className="max-w-md mx-auto">
          <Card>
            <CardHeader>
              <CardTitle>Create a Lobby</CardTitle>
              <CardDescription>Set up a new game for your group.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleCreate} className="space-y-4">
                <div className="space-y-1">
                  <label className="text-sm font-medium">Your name</label>
                  <Input
                    value={createName}
                    onChange={(e) => setCreateName(e.target.value)}
                    placeholder="Enter your name"
                    required
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium">Points to win</label>
                  <Input
                    type="number"
                    min={1}
                    max={50}
                    value={targetScore}
                    onChange={(e) => setTargetScore(Number(e.target.value))}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium">Answer timer (seconds)</label>
                  <Input
                    type="number"
                    min={15}
                    max={300}
                    value={timerSecs}
                    onChange={(e) => setTimerSecs(Number(e.target.value))}
                  />
                </div>
                {createError && <p className="text-sm text-destructive">{createError}</p>}
                <div className="flex gap-2">
                  <Button type="button" variant="outline" onClick={() => setView('menu')}>
                    Back
                  </Button>
                  <Button type="submit" disabled={creating} className="flex-1">
                    {creating ? 'Creating…' : 'Create Lobby'}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </Layout>
    )
  }

  if (view === 'join') {
    return (
      <Layout>
        <div className="max-w-md mx-auto">
          <Card>
            <CardHeader>
              <CardTitle>Join a Lobby</CardTitle>
              <CardDescription>Enter the lobby code shared by the host.</CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleJoin} className="space-y-4">
                <div className="space-y-1">
                  <label className="text-sm font-medium">Lobby code</label>
                  <Input
                    value={lobbyCode}
                    onChange={(e) => setLobbyCode(e.target.value.toUpperCase())}
                    placeholder="e.g. XYZ123"
                    maxLength={6}
                    required
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium">Your name</label>
                  <Input
                    value={joinName}
                    onChange={(e) => setJoinName(e.target.value)}
                    placeholder="Enter your name"
                    required
                  />
                </div>
                {joinError && <p className="text-sm text-destructive">{joinError}</p>}
                <div className="flex gap-2">
                  <Button type="button" variant="outline" onClick={() => setView('menu')}>
                    Back
                  </Button>
                  <Button type="submit" disabled={joining} className="flex-1">
                    {joining ? 'Joining…' : 'Join Lobby'}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-md mx-auto text-center space-y-8 py-16">
        <div>
          <h1 className="text-4xl font-bold mb-2">Loaded Questions</h1>
          <p className="text-muted-foreground">A party game about how well you know each other.</p>
        </div>
        <div className="flex flex-col gap-4">
          <Button size="lg" onClick={() => setView('create')}>
            Create Lobby
          </Button>
          <Button size="lg" variant="outline" onClick={() => setView('join')}>
            Join Lobby
          </Button>
        </div>
      </div>
    </Layout>
  )
}
