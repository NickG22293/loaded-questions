package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type PlayerID string

type Player struct {
    ID   PlayerID `json:"id"`
    Name string `json:"name"`
}

type GameSession struct {
    ID       string            `json:"id"`
    Asker    *Player           `json:"asker"`
    Players  map[PlayerID]Player `json:"players"`
    Question string            `json:"question"`
    Answers  map[PlayerID]string `json:"answers"`
    Guesses  map[PlayerID]string `json:"guesses"`
    Score    map[PlayerID]int    `json:"score"`
	Status   Status            `json:"status"`
    mu       sync.Mutex
	playerConnections map[PlayerID]*websocket.Conn
}

type Status int

const (
	Lobby Status = iota
	Asking
	Answering
	Guessing
)

func (s Status) String() string {
	return [...]string{"Lobby", "Asking", "Answering", "Guessing"}[s]
}
