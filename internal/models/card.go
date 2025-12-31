package models

import "time"

// Card represents a flashcard
type Card struct {
	ID                 string    `json:"id"`
	Front              string    `json:"front"`
	Back               string    `json:"back"`
	Tags               []string  `json:"tags"`
	SourceConversation string    `json:"source_conversation,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Session represents a card generation session
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Cards     []Card    `json:"cards,omitempty"`
}

// GenerateRequest represents the request to generate cards
type GenerateRequest struct {
	Conversation string          `json:"conversation"`
	Options      GenerateOptions `json:"options"`
}

// GenerateOptions contains options for card generation
type GenerateOptions struct {
	MaxCards   int      `json:"max_cards"`
	Difficulty string   `json:"difficulty"` // basic, intermediate, advanced, mixed
	Topics     []string `json:"topics"`
}

// GenerateResponse represents the response from card generation
type GenerateResponse struct {
	SessionID string `json:"session_id"`
	Cards     []Card `json:"cards"`
}

// SaveCardsRequest represents the request to save cards
type SaveCardsRequest struct {
	SessionID string `json:"session_id"`
	Cards     []Card `json:"cards"`
}

// UpdateCardRequest represents the request to update a card
type UpdateCardRequest struct {
	Front string   `json:"front"`
	Back  string   `json:"back"`
	Tags  []string `json:"tags"`
}
