package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/Jack4Code/cardforge/internal/claude"
	"github.com/Jack4Code/cardforge/internal/models"
	"github.com/Jack4Code/cardforge/internal/storage"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	storage *storage.SQLite
	claude  *claude.Client
}

// NewHandler creates a new handler instance
func NewHandler(storage *storage.SQLite, claude *claude.Client) *Handler {
	return &Handler{
		storage: storage,
		claude:  claude,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// sendError sends an error response
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// GenerateCards handles POST /api/generate
func (h *Handler) GenerateCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Conversation == "" {
		sendError(w, "Conversation cannot be empty", http.StatusBadRequest)
		return
	}

	// Generate cards using Claude
	generatedCards, err := h.claude.GenerateCards(r.Context(), req.Conversation, req.Options)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate cards: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a session
	sessionID := uuid.New().String()
	if err := h.storage.CreateSession(sessionID); err != nil {
		sendError(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to models.Card and save
	now := time.Now()
	cards := make([]models.Card, len(generatedCards))
	for i, gc := range generatedCards {
		card := models.Card{
			ID:                 uuid.New().String(),
			Front:              gc.Front,
			Back:               gc.Back,
			Tags:               gc.Tags,
			SourceConversation: req.Conversation,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		// Save card
		if err := h.storage.SaveCard(&card); err != nil {
			sendError(w, fmt.Sprintf("Failed to save card: %v", err), http.StatusInternalServerError)
			return
		}

		// Link to session
		if err := h.storage.LinkCardToSession(sessionID, card.ID); err != nil {
			sendError(w, fmt.Sprintf("Failed to link card to session: %v", err), http.StatusInternalServerError)
			return
		}

		cards[i] = card
	}

	response := models.GenerateResponse{
		SessionID: sessionID,
		Cards:     cards,
	}

	sendJSON(w, response, http.StatusOK)
}

// ListCards handles GET /api/cards
func (h *Handler) ListCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	cards, err := h.storage.ListCards(limit, offset)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to list cards: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSON(w, cards, http.StatusOK)
}

// SaveCards handles POST /api/cards
func (h *Handler) SaveCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SaveCardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, card := range req.Cards {
		if err := h.storage.SaveCard(&card); err != nil {
			sendError(w, fmt.Sprintf("Failed to save card: %v", err), http.StatusInternalServerError)
			return
		}

		if req.SessionID != "" {
			if err := h.storage.LinkCardToSession(req.SessionID, card.ID); err != nil {
				sendError(w, fmt.Sprintf("Failed to link card to session: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	sendJSON(w, map[string]string{"status": "success"}, http.StatusOK)
}

// UpdateCard handles PUT /api/cards/:id
func (h *Handler) UpdateCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := r.URL.Path[len("/api/cards/"):]
	if id == "" {
		sendError(w, "Card ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateCard(id, req.Front, req.Back, req.Tags); err != nil {
		sendError(w, fmt.Sprintf("Failed to update card: %v", err), http.StatusInternalServerError)
		return
	}

	card, err := h.storage.GetCard(id)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get updated card: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSON(w, card, http.StatusOK)
}

// DeleteCard handles DELETE /api/cards/:id
func (h *Handler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := r.URL.Path[len("/api/cards/"):]
	if id == "" {
		sendError(w, "Card ID is required", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteCard(id); err != nil {
		sendError(w, fmt.Sprintf("Failed to delete card: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "success"}, http.StatusOK)
}

// ListSessions handles GET /api/sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	sessions, err := h.storage.ListSessions(limit, offset)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to list sessions: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSON(w, sessions, http.StatusOK)
}

// GetSession handles GET /api/sessions/:id
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := r.URL.Path[len("/api/sessions/"):]
	if id == "" {
		sendError(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	session, err := h.storage.GetSession(id)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get session: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSON(w, session, http.StatusOK)
}

// ExportCards handles GET /api/export
func (h *Handler) ExportCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	sessionID := r.URL.Query().Get("session_id")

	if format == "" {
		format = "anki"
	}

	if sessionID == "" {
		sendError(w, "session_id is required", http.StatusBadRequest)
		return
	}

	switch format {
	case "anki":
		csv, err := h.storage.ExportCardsToAnkiCSV(sessionID)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to export cards: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=flashcards_%s.csv", sessionID))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(csv))

	default:
		sendError(w, fmt.Sprintf("Unsupported format: %s", format), http.StatusBadRequest)
	}
}
