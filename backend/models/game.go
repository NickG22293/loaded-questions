package models

type Phase string

const (
	PhaseAsking    Phase = "ASKING"
	PhaseAnswering Phase = "ANSWERING"
	PhaseAssigning Phase = "ASSIGNING"
	PhaseScoring   Phase = "SCORING"
)

type Answer struct {
	ID         string `json:"id"`
	PlayerID   string `json:"playerId"`
	Text       string `json:"text"`
	AssignedTo string `json:"assignedTo,omitempty"`
}

type Round struct {
	RoundNumber int       `json:"roundNumber"`
	AskerID     string    `json:"askerId"`
	Question    string    `json:"question,omitempty"`
	Answers     []*Answer `json:"answers,omitempty"`
	Phase       Phase     `json:"phase"`
}

type Game struct {
	ID           string    `json:"id"`
	LobbyID      string    `json:"lobbyId"`
	Players      []*Player `json:"players"`
	CurrentRound *Round    `json:"currentRound"`
	Rounds       []*Round  `json:"rounds"`
	TargetScore  int       `json:"targetScore"`
	WinnerID     string    `json:"winnerId,omitempty"`
}
