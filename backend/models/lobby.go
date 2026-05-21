package models

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
