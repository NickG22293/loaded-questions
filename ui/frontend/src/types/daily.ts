export interface DailyQuestion {
  id: string
  question: string
  active_date: string
}

export interface Answer {
  id: string
  question_id: string
  user_id: string
  answer_text: string
  submitted_at: string
}

export interface Group {
  id: string
  name: string
  invite_code: string
  created_by: string
  created_at: string
}

export interface GroupMember {
  id: string
  display_name: string
  avatar_url: string
}

export interface AnswerProgress {
  answered: number
  total: number
}

export interface GroupDetail {
  group: Group
  members: GroupMember[]
  progress: AnswerProgress
  guessing_unlocked: boolean
  user_answered: boolean
  user_guessed: boolean
}

export interface GuessTarget {
  answer_id: string
  answer_text: string
}

export interface GuessResult {
  answer_id: string
  answer_text: string
  actual_user_id: string
  actual_name: string
  guessed_user_id: string
  guessed_name: string
  correct: boolean
}

export interface LeaderboardEntry {
  user_id: string
  display_name: string
  avatar_url: string
  score: number
}

export interface ChatMessage {
  id: string
  group_id: string
  question_id: string
  user_id: string
  display_name: string
  avatar_url: string
  message: string
  created_at: string
}
