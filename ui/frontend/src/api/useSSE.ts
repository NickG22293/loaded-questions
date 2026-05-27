import { useEffect, useReducer } from 'react'
import type { Game, Lobby, Player, Round, SSEEventName } from '@/types/sessions'

interface GameState {
  lobby: Lobby | null
  game: Game | null
  connected: boolean
}

type Action =
  | { type: 'connected' }
  | { type: 'player_joined'; player: Player }
  | { type: 'game_started'; lobby: Lobby }
  | { type: 'phase_changed'; round: Round }
  | { type: 'answer_count'; count: number; total: number }
  | { type: 'round_result'; game: Game }
  | { type: 'game_over'; game: Game }
  | { type: 'set_lobby'; lobby: Lobby }

function reducer(state: GameState, action: Action): GameState {
  switch (action.type) {
    case 'connected':
      return { ...state, connected: true }
    case 'set_lobby':
      return { ...state, lobby: action.lobby }
    case 'player_joined':
      if (!state.lobby) return state
      return {
        ...state,
        lobby: { ...state.lobby, players: [...state.lobby.players, action.player] },
      }
    case 'game_started':
      return { ...state, lobby: action.lobby }
    case 'phase_changed':
      if (!state.game) return state
      return { ...state, game: { ...state.game, currentRound: action.round } }
    case 'round_result':
    case 'game_over':
      return { ...state, game: action.game }
    default:
      return state
  }
}

export function useSSE(lobbyId: string | undefined) {
  const [state, dispatch] = useReducer(reducer, {
    lobby: null,
    game: null,
    connected: false,
  })

  useEffect(() => {
    if (!lobbyId) return

    const es = new EventSource(`/api/sessions/lobbies/${lobbyId}/events`, { withCredentials: true })

    const handle = (eventName: SSEEventName) => (e: MessageEvent) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const data = JSON.parse(e.data) as any
      switch (eventName) {
        case 'connected':
          dispatch({ type: 'connected' })
          break
        case 'player_joined':
          dispatch({ type: 'player_joined', player: data })
          break
        case 'game_started':
          dispatch({ type: 'game_started', lobby: data })
          break
        case 'phase_changed':
          dispatch({ type: 'phase_changed', round: data })
          break
        case 'round_result':
          dispatch({ type: 'round_result', game: data })
          break
        case 'game_over':
          dispatch({ type: 'game_over', game: data })
          break
      }
    }

    const events: SSEEventName[] = [
      'connected',
      'player_joined',
      'game_started',
      'phase_changed',
      'round_result',
      'game_over',
    ]
    events.forEach((name) => es.addEventListener(name, handle(name) as EventListener))

    return () => es.close()
  }, [lobbyId])

  return state
}
