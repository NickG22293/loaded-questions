import { useParams } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import { useSSE } from '@/api/useSSE'

export function Game() {
  const { id } = useParams<{ id: string }>()
  const { game, connected } = useSSE(id)

  if (!connected) {
    return (
      <Layout>
        <p className="text-muted-foreground">Connecting…</p>
      </Layout>
    )
  }

  if (!game) {
    return (
      <Layout>
        <p className="text-muted-foreground">Waiting for game to start…</p>
      </Layout>
    )
  }

  const phase = game.currentRound?.phase

  return (
    <Layout>
      <div className="max-w-2xl mx-auto">
        <p className="text-sm text-muted-foreground mb-4">
          Round {game.currentRound?.roundNumber} · Phase: {phase}
        </p>
        {/* Phase-specific views implemented in the next iteration */}
        <p className="text-center text-muted-foreground py-16">
          Game phase <span className="font-mono font-bold text-foreground">{phase}</span> — coming
          soon.
        </p>
      </div>
    </Layout>
  )
}
