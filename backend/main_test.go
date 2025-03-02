package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateSession(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["session_id"])
}

func TestJoinSession(t *testing.T) {
	router := setupRouter()

	// First, create a session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	sessionID := response["session_id"]

	// Now, join the session
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/session/"+sessionID+"/join?name=Player1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var joinResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &joinResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, joinResponse["player_id"])
	assert.Equal(t, sessionID, joinResponse["session_id"])
}

func TestGetSessions(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sessions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
}

func TestGetSession(t *testing.T) {
	router := setupRouter()

	// First, create a session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	sessionID := response["session_id"]

	// Now, get the session
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session/"+sessionID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var session GameSession
	err := json.Unmarshal(w.Body.Bytes(), &session)
	assert.NoError(t, err)
	assert.Equal(t, sessionID, session.ID)
}

func TestSetQuestion(t *testing.T) {
	router := setupRouter()

	// First, create a session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	sessionID := response["session_id"]

	// Now, set a question
	w = httptest.NewRecorder()
	question := map[string]string{"question": "What is your favorite color?"}
	jsonValue, _ := json.Marshal(question)
	req, _ = http.NewRequest("POST", "/session/"+sessionID+"/question", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var questionResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &questionResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Question set!", questionResponse["message"])
}

func TestSubmitAnswer(t *testing.T) {
	router := setupRouter()

	// First, create a session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	sessionID := response["session_id"]

	// Join the session
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/session/"+sessionID+"/join?name=Player1", nil)
	router.ServeHTTP(w, req)
	var joinResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &joinResponse)
	playerID := joinResponse["player_id"]

	// Now, submit an answer
	w = httptest.NewRecorder()
	answer := map[string]string{"player_id": playerID, "answer": "Blue"}
	jsonValue, _ := json.Marshal(answer)
	req, _ = http.NewRequest("POST", "/session/"+sessionID+"/answer", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var answerResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &answerResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Answer submitted!", answerResponse["message"])
}

func TestStartGame(t *testing.T) {
	router := setupRouter()

	// First, create a session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	sessionID := response["session_id"]

	// Join the session with 3 players
	for i := 1; i <= 3; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", fmt.Sprintf("/session/%s/join?name=Player%d", sessionID, i), nil)
		router.ServeHTTP(w, req)
	}

	// Now, start the game
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/session/"+sessionID+"/start", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var startResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &startResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Game started!", startResponse["message"])
}

func TestAskerSelectedOnGameStart(t *testing.T) {
	router := setupRouter()

	// First, create a session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/session", nil)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	sessionID := response["session_id"]

	// Join the session with 3 players
	for i := 1; i <= 3; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", fmt.Sprintf("/session/%s/join?name=Player%d", sessionID, i), nil)
		router.ServeHTTP(w, req)
	}

	// Start the game
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/session/"+sessionID+"/start", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Get the session to check if an Asker has been selected
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session/"+sessionID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var session GameSession
	err := json.Unmarshal(w.Body.Bytes(), &session)
	assert.NoError(t, err)
	assert.NotNil(t, session.Asker)
	assert.NotEmpty(t, session.Asker.ID)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/session", createSession)
	r.POST("/session/:sessionID/join", joinSession)
	r.GET("/sessions", getSessions)
	r.GET("/session/:sessionID", getSession)
	r.POST("/session/:sessionID/question", setQuestion)
	r.POST("/session/:sessionID/answer", submitAnswer)
	r.POST("/session/:sessionID/start", startGame)

	return r
}
