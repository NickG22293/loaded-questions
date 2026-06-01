package daily

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("daily: not found")
	ErrAlreadyExists = errors.New("daily: already exists")
)

// DailyStore defines all persistence operations for the daily game.
// Business logic lives in the handlers; this layer is pure data access.
type DailyStore interface {
	// Users
	UpsertUser(ctx context.Context, u *User) error

	// Groups
	CreateGroup(ctx context.Context, g *Group) error
	GetGroup(ctx context.Context, groupID string) (*Group, error)
	GetGroupByInviteCode(ctx context.Context, code string) (*Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]Group, error)
	JoinGroup(ctx context.Context, groupID, userID string) error
	GetGroupMembers(ctx context.Context, groupID string) ([]GroupMember, error)

	// Questions + Answers
	GetTodayQuestion(ctx context.Context) (*DailyQuestion, error)
	GetUserAnswer(ctx context.Context, userID, questionID string) (*Answer, error)
	SubmitAnswer(ctx context.Context, a *Answer) error
	GetGroupAnswerProgress(ctx context.Context, groupID, questionID string) (AnswerProgress, error)
	GetGroupAnswers(ctx context.Context, groupID, questionID string) ([]Answer, error)

	// Guesses + Results
	HasUserSubmittedGuesses(ctx context.Context, guesserID, groupID, questionID string) (bool, error)
	SubmitGuesses(ctx context.Context, guesses []Guess) error
	GetGroupResults(ctx context.Context, guesserID, groupID, questionID string) ([]GuessResult, error)
	GetLeaderboard(ctx context.Context, groupID string) ([]LeaderboardEntry, error)

	// Chat
	GetChatMessages(ctx context.Context, groupID, questionID string) ([]ChatMessage, error)
	SendChatMessage(ctx context.Context, m *ChatMessage) error
}
