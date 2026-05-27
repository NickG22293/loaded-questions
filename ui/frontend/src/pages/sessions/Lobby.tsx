import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { PlayerList } from '@/components/sessions/PlayerList'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { getLobby, startGame } from '@/api/sessions'
import type { Lobby as LobbyType, Player } from '@/types/sessions'

export function Lobby() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [lobby, setLobby] = useState<LobbyType | null>(null)
  const [playerId, setPlayerId] = useState<string | null>(null)
  const [starting, setStarting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    getLobby(id).then(setLobby).catch(() => navigate('/sessions'))
    const stored = sessionStorage.getItem(`playerId:${id}`)
    if (stored) setPlayerId(stored)
  }, [id, navigate])

  useEffect(() => {
    if (!id) return
    const es = new EventSource(`/api/sessions/lobbies/${id}/events`, { withCredentials: true })

    es.addEventListener('player_joined', (e: MessageEvent) => {
      const player = JSON.parse(e.data) as Player
      setLobby((prev) =>
        prev ? { ...prev, players: [...prev.players, player] } : prev
      )
    })

    es.addEventListener('game_started', () => {
      navigate(`/sessions/game/${id}`)
    })

    return () => es.close()
  }, [id, navigate])

  useEffect(() => {
    if (lobby?.gameStarted) {
      navigate(`/sessions/game/${id}`)
    }
  }, [lobby, navigate, id])

  async function handleStart() {
    if (!id) return
    setStarting(true)
    setError('')
    try {
      await startGame(id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start game')
    } finally {
      setStarting(false)
    }
  }

  if (!lobby) {
    return (
      <Layout>
        <p className="text-muted-foreground">Loading lobby…</p>
      </Layout>
    )
  }

  const isCreator = playerId === lobby.creatorId

  return (
    <Layout>
      <div className="max-w-lg mx-auto space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Lobby</CardTitle>
            <CardDescription>
              Share code <span className="font-mono font-bold text-foreground">{id}</span> with
              friends to join.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-6 text-sm text-muted-foreground">
              <span>First to {lobby.targetScore} pts wins</span>
              <span>Answer timer: {lobby.answerTimerSecs}s</span>
            </div>
            <PlayerList players={lobby.players} currentPlayerId={playerId ?? undefined} />
          </CardContent>
        </Card>

        {isCreator && (
          <div className="space-y-2">
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button
              onClick={handleStart}
              disabled={starting || lobby.players.length < 2}
              className="w-full"
              size="lg"
            >
              {starting ? 'Starting…' : 'Start Game'}
            </Button>
            {lobby.players.length < 2 && (
              <p className="text-xs text-muted-foreground text-center">
                Need at least 2 players to start.
              </p>
            )}
          </div>
        )}

        {!isCreator && (
          <p className="text-center text-muted-foreground text-sm">
            Waiting for the host to start the game…
          </p>
        )}
      </div>
    </Layout>
  )
}
