export type Phase = 'ASKING' | 'ANSWERING' | 'ASSIGNING' | 'SCORING' | 'GAME_OVER'

export interface Player {
  id: string
  name: string
  score: number
  isCreator: boolean
}

export interface Lobby {
  id: string
  players: Player[]
  creatorId: string
  targetScore: number
  answerTimerSecs: number
  gameStarted: boolean
  gameId?: string
}

export interface Answer {
  id: string
  playerId: string
  text: string
  assignedTo?: string
}

export interface Round {
  roundNumber: number
  askerId: string
  question?: string
  answers?: Answer[]
  phase: Phase
  phaseStartedAt?: string
}

export interface Game {
  id: string
  lobbyId: string
  players: Player[]
  currentRound: Round | null
  rounds: Round[]
  targetScore: number
  answerTimerSecs: number
  winnerId?: string
}

export interface AnswerCountEvent {
  submitted: number
  total: number
  submittedPlayerIds: string[]
}

export type SSEEventName =
  | 'connected'
  | 'player_joined'
  | 'game_started'
  | 'phase_changed'
  | 'answer_count'
  | 'assignment_updated'
  | 'round_result'
  | 'game_over'

export interface SSEEvent<T = unknown> {
  event: SSEEventName
  data: T
}
