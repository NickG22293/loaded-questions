package daily

import "time"

// --- Domain types (written to/from DB) ---

type User struct {
	ID          string
	Email       string
	DisplayName string
	AvatarURL   string
	IsAdmin     bool
}

type Group struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type DailyQuestion struct {
	ID         string    `json:"id"`
	Question   string    `json:"question"`
	ActiveDate time.Time `json:"active_date"`
}

type Answer struct {
	ID          string    `json:"id"`
	QuestionID  string    `json:"question_id"`
	UserID      string    `json:"user_id"`
	AnswerText  string    `json:"answer_text"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type Guess struct {
	QuestionID    string
	GroupID       string
	GuesserID     string
	AnswerID      string
	GuessedUserID string
}

type ChatMessage struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	QuestionID  string    `json:"question_id"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// --- View types (API responses, never written directly) ---

type GroupMember struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type AnswerProgress struct {
	Answered int `json:"answered"`
	Total    int `json:"total"`
}

// GuessTarget is an anonymous answer shown during the guessing phase.
type GuessTarget struct {
	AnswerID   string `json:"answer_id"`
	AnswerText string `json:"answer_text"`
}

type GuessResult struct {
	AnswerID      string `json:"answer_id"`
	AnswerText    string `json:"answer_text"`
	ActualUserID  string `json:"actual_user_id"`
	ActualName    string `json:"actual_name"`
	GuessedUserID string `json:"guessed_user_id"`
	GuessedName   string `json:"guessed_name"`
	Correct       bool   `json:"correct"`
}

type LeaderboardEntry struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Score       int    `json:"score"`
}
