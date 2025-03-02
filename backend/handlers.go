package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
    Started  bool              `json:"started"`
    mu       sync.Mutex
}

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

func createSession(c *gin.Context) {
    sessionID := uuid.New().String()[:6]
    sessionsMu.Lock()
    sessions[sessionID] = &GameSession{
        ID:      sessionID,
        Players: make(map[PlayerID]Player),
        Answers: make(map[PlayerID]string),
        Guesses: make(map[PlayerID]string),
        Score:   make(map[PlayerID]int),
    }
    sessionsMu.Unlock()
    c.JSON(http.StatusOK, gin.H{"session_id": sessionID})
}

func startGame(c *gin.Context) {
    sessionID := c.Param("sessionID")

    sessionsMu.Lock()
    session, exists := sessions[sessionID]
    sessionsMu.Unlock()

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
        return
    }

    session.mu.Lock()
    defer session.mu.Unlock()
    if len(session.Players) < 3 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "At least 3 players required to start"})
        return
    }
    assignAsker(session)

    session.Started = true

    c.JSON(http.StatusOK, gin.H{"message": "Game started!"})
}

func assignAsker(session *GameSession) {
    rand.Seed(time.Now().UnixNano())
    playerIDs := make([]PlayerID, 0, len(session.Players))
    for playerID := range session.Players {
        playerIDs = append(playerIDs, playerID)
    }
    randomIndex := rand.Intn(len(session.Players))
    asker := session.Players[playerIDs[randomIndex]]
    session.Asker = &asker
    fmt.Printf("Asker: %s, Session ID: %s\n", asker.Name, session.ID)
}

func joinSession(c *gin.Context) {
    sessionID := c.Param("sessionID")
    playerName := c.Query("name")
    if playerName == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
        return
    }

    sessionsMu.Lock()
    session, exists := sessions[sessionID]
    sessionsMu.Unlock()
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
        return
    }

    playerID := PlayerID(uuid.New().String())
    player := Player{ID: playerID, Name: playerName}

    session.mu.Lock()
    session.Players[playerID] = player
    if session.Asker == nil {
        session.Asker = &player
    }
    session.mu.Unlock()

    c.JSON(http.StatusOK, gin.H{"player_id": playerID, "session_id": sessionID})
}

func getSessions(c *gin.Context) {
    jsonData, err := json.Marshal(sessions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, "Problem fetching sessions")
    }
    c.JSON(http.StatusFound, jsonData)
}

func getSession(c *gin.Context) {
    sessionID := c.Param("sessionID")

    sessionsMu.Lock()
    session, exists := sessions[sessionID]
    sessionsMu.Unlock()

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
        return
    }

    c.JSON(http.StatusOK, session)
}

func setQuestion(c *gin.Context) {
    sessionID := c.Param("sessionID")

    sessionsMu.Lock()
    session, exists := sessions[sessionID]
    sessionsMu.Unlock()

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
        return
    }

    var requestBody struct {
        Question string `json:"question"`
    }

    if err := c.ShouldBindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    session.mu.Lock()
    session.Question = requestBody.Question
    session.mu.Unlock()

    c.JSON(http.StatusOK, gin.H{"message": "Question set!"})
}

func submitAnswer(c *gin.Context) {
    sessionID := c.Param("sessionID")

    sessionsMu.Lock()
    session, exists := sessions[sessionID]
    sessionsMu.Unlock()

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
        return
    }

    var requestBody struct {
        PlayerID PlayerID `json:"player_id"`
        Answer   string `json:"answer"`
    }

    if err := c.ShouldBindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    session.mu.Lock()
    session.Answers[requestBody.PlayerID] = requestBody.Answer
    session.mu.Unlock()

    c.JSON(http.StatusOK, gin.H{"message": "Answer submitted!"})
}

func getAnswers(c *gin.Context) {
	sessionID := c.Param("sessionID")

    sessionsMu.Lock()
    session, exists := sessions[sessionID]
    sessionsMu.Unlock()

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
        return
    }

	c.JSON(http.StatusOK, session.Answers)
}