package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"loaded-questions/middleware"
	"loaded-questions/models"
)

type createLobbyRequest struct {
	PlayerName      string `json:"playerName"`
	TargetScore     int    `json:"targetScore"`
	AnswerTimerSecs int    `json:"answerTimerSecs"`
}

type createLobbyResponse struct {
	LobbyID  string `json:"lobbyId"`
	PlayerID string `json:"playerId"`
}

func (h *Handler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	var req createLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.PlayerName == "" {
		http.Error(w, "playerName required", http.StatusBadRequest)
		return
	}
	if req.TargetScore <= 0 {
		req.TargetScore = 10
	}
	if req.AnswerTimerSecs <= 0 {
		req.AnswerTimerSecs = 60
	}

	playerID := generateToken()[:8]
	lobbyID := generateLobbyID()
	token := generateToken()

	player := &models.Player{
		ID:        playerID,
		Name:      req.PlayerName,
		IsCreator: true,
	}
	lobby := &models.Lobby{
		ID:              lobbyID,
		Players:         []*models.Player{player},
		CreatorID:       playerID,
		TargetScore:     req.TargetScore,
		AnswerTimerSecs: req.AnswerTimerSecs,
	}

	if err := h.store.CreateLobby(lobby); err != nil {
		http.Error(w, "failed to create lobby", http.StatusInternalServerError)
		return
	}
	if err := h.store.SetPlayerToken(token, lobbyID, playerID); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)
	writeJSON(w, http.StatusCreated, createLobbyResponse{LobbyID: lobbyID, PlayerID: playerID})
}

type joinLobbyRequest struct {
	PlayerName string `json:"playerName"`
}

type joinLobbyResponse struct {
	PlayerID string `json:"playerId"`
}

func (h *Handler) JoinLobby(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")

	var req joinLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.PlayerName == "" {
		http.Error(w, "playerName required", http.StatusBadRequest)
		return
	}

	lobby, err := h.store.GetLobby(lobbyID)
	if err != nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}
	if lobby.GameStarted {
		http.Error(w, "game already started", http.StatusConflict)
		return
	}
	if len(lobby.Players) >= 10 {
		http.Error(w, "lobby is full", http.StatusConflict)
		return
	}
	for _, p := range lobby.Players {
		if p.Name == req.PlayerName {
			http.Error(w, "name already taken", http.StatusConflict)
			return
		}
	}

	playerID := generateToken()[:8]
	token := generateToken()
	player := &models.Player{
		ID:   playerID,
		Name: req.PlayerName,
	}
	lobby.Players = append(lobby.Players, player)

	if err := h.store.UpdateLobby(lobby); err != nil {
		http.Error(w, "failed to join lobby", http.StatusInternalServerError)
		return
	}
	if err := h.store.SetPlayerToken(token, lobbyID, playerID); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, token)
	h.store.BroadcastToLobby(lobbyID, sseEvent("player_joined", player))
	writeJSON(w, http.StatusOK, joinLobbyResponse{PlayerID: playerID})
}

func (h *Handler) GetLobby(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	lobby, err := h.store.GetLobby(lobbyID)
	if err != nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, lobby)
}

func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "id")
	playerID := middleware.GetPlayerID(r.Context())

	lobby, err := h.store.GetLobby(lobbyID)
	if err != nil {
		http.Error(w, "lobby not found", http.StatusNotFound)
		return
	}
	if lobby.CreatorID != playerID {
		http.Error(w, "only the creator can start the game", http.StatusForbidden)
		return
	}
	if lobby.GameStarted {
		http.Error(w, "game already started", http.StatusConflict)
		return
	}
	if len(lobby.Players) < 2 {
		http.Error(w, "need at least 2 players to start", http.StatusBadRequest)
		return
	}

	lobby.GameStarted = true
	if err := h.store.UpdateLobby(lobby); err != nil {
		http.Error(w, "failed to start game", http.StatusInternalServerError)
		return
	}

	h.store.BroadcastToLobby(lobbyID, sseEvent("game_started", lobby))
	w.WriteHeader(http.StatusNoContent)
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "player_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}
