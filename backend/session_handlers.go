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
	"github.com/gorilla/websocket"
)

var sessions = make(map[string]*GameSession)
var sessionsMu sync.Mutex

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

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

func handlePlayerConnection(session *GameSession, playerID PlayerID, conn *websocket.Conn) {
    defer func() {
        session.mu.Lock()
        delete(session.Players, playerID)
        delete(session.playerConnections, playerID)
        session.mu.Unlock()
        conn.Close()
        notifyPlayersUpdate(session)
    }()

    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            break
        }
    }
}

func notifyPlayersUpdate(session *GameSession) {
    session.mu.Lock()
    defer session.mu.Unlock()

    players := make(map[string]Player)
    for id, player := range session.Players {
        players[string(id)] = player
    }

    message := map[string]interface{}{
        "type":    "PLAYER_UPDATE",
        "players": players,
    }

    for _, conn := range session.playerConnections {
        err := conn.WriteJSON(message)
        if err != nil {
            conn.Close()
            // delete(session.playerConnections, conn)
        }
    }
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
