import type { Player } from '@/types'
import { Badge } from '@/components/ui/badge'

interface PlayerListProps {
  players: Player[]
  currentPlayerId?: string
}

export function PlayerList({ players, currentPlayerId }: PlayerListProps) {
  return (
    <ul className="space-y-2">
      {players.map((player) => (
        <li
          key={player.id}
          className="flex items-center justify-between rounded-md border px-4 py-2"
        >
          <div className="flex items-center gap-2">
            <span className="font-medium">{player.name}</span>
            {player.isCreator && (
              <Badge variant="secondary" className="text-xs">
                Host
              </Badge>
            )}
            {player.id === currentPlayerId && (
              <Badge variant="outline" className="text-xs">
                You
              </Badge>
            )}
          </div>
          <span className="text-sm text-muted-foreground font-mono">{player.score} pts</span>
        </li>
      ))}
    </ul>
  )
}
