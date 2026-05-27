package sessions

import "time"

type Phase string

const (
	PhaseAsking    Phase = "ASKING"
	PhaseAnswering Phase = "ANSWERING"
	PhaseAssigning Phase = "ASSIGNING"
	PhaseScoring   Phase = "SCORING"
	PhaseGameOver  Phase = "GAME_OVER"
)

type Player struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	IsCreator bool   `json:"isCreator"`
}

type Lobby struct {
	ID              string    `json:"id"`
	Players         []*Player `json:"players"`
	CreatorID       string    `json:"creatorId"`
	TargetScore     int       `json:"targetScore"`
	AnswerTimerSecs int       `json:"answerTimerSecs"`
	GameStarted     bool      `json:"gameStarted"`
	GameID          string    `json:"gameId,omitempty"`
}

type Answer struct {
	ID         string `json:"id"`
	PlayerID   string `json:"playerId"`
	Text       string `json:"text"`
	AssignedTo string `json:"assignedTo,omitempty"`
}

type Round struct {
	RoundNumber    int        `json:"roundNumber"`
	AskerID        string     `json:"askerId"`
	Question       string     `json:"question,omitempty"`
	Answers        []*Answer  `json:"answers,omitempty"`
	Phase          Phase      `json:"phase"`
	PhaseStartedAt *time.Time `json:"phaseStartedAt,omitempty"`
}

type Game struct {
	ID              string    `json:"id"`
	LobbyID         string    `json:"lobbyId"`
	Players         []*Player `json:"players"`
	CurrentRound    *Round    `json:"currentRound"`
	Rounds          []*Round  `json:"rounds"`
	TargetScore     int       `json:"targetScore"`
	AnswerTimerSecs int       `json:"answerTimerSecs"`
	WinnerID        string    `json:"winnerId,omitempty"`
}
