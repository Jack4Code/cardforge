package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Jack4Code/cardforge/internal/models"
)

// SQLite represents the SQLite storage implementation
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite storage instance
func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &SQLite{db: db}

	// Run migrations
	if err := s.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *SQLite) Close() error {
	return s.db.Close()
}

// runMigrations runs database migrations
func (s *SQLite) runMigrations() error {
	migrationFile := "migrations/001_initial_schema.sql"

	data, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err := s.db.Exec(string(data)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// CreateSession creates a new session
func (s *SQLite) CreateSession(sessionID string) error {
	query := `INSERT INTO sessions (id, created_at) VALUES (?, ?)`
	_, err := s.db.Exec(query, sessionID, time.Now())
	return err
}

// SaveCard saves a card to the database
func (s *SQLite) SaveCard(card *models.Card) error {
	tagsJSON, err := json.Marshal(card.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO cards (id, front, back, tags, source_conversation, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.Exec(query, card.ID, card.Front, card.Back, string(tagsJSON),
		card.SourceConversation, card.CreatedAt, card.UpdatedAt)
	return err
}

// LinkCardToSession links a card to a session
func (s *SQLite) LinkCardToSession(sessionID, cardID string) error {
	query := `INSERT INTO session_cards (session_id, card_id) VALUES (?, ?)`
	_, err := s.db.Exec(query, sessionID, cardID)
	return err
}

// GetCard retrieves a card by ID
func (s *SQLite) GetCard(id string) (*models.Card, error) {
	query := `SELECT id, front, back, tags, source_conversation, created_at, updated_at
			  FROM cards WHERE id = ?`

	var card models.Card
	var tagsJSON string

	err := s.db.QueryRow(query, id).Scan(
		&card.ID, &card.Front, &card.Back, &tagsJSON,
		&card.SourceConversation, &card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &card.Tags); err != nil {
		card.Tags = []string{}
	}

	return &card, nil
}

// ListCards retrieves all cards with pagination
func (s *SQLite) ListCards(limit, offset int) ([]models.Card, error) {
	query := `SELECT id, front, back, tags, source_conversation, created_at, updated_at
			  FROM cards ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var tagsJSON string

		err := rows.Scan(
			&card.ID, &card.Front, &card.Back, &tagsJSON,
			&card.SourceConversation, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &card.Tags); err != nil {
			card.Tags = []string{}
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// UpdateCard updates an existing card
func (s *SQLite) UpdateCard(id string, front, back string, tags []string) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `UPDATE cards SET front = ?, back = ?, tags = ?, updated_at = ? WHERE id = ?`
	_, err = s.db.Exec(query, front, back, string(tagsJSON), time.Now(), id)
	return err
}

// DeleteCard deletes a card
func (s *SQLite) DeleteCard(id string) error {
	// Delete from session_cards first (foreign key constraint)
	if _, err := s.db.Exec(`DELETE FROM session_cards WHERE card_id = ?`, id); err != nil {
		return err
	}

	// Delete the card
	_, err := s.db.Exec(`DELETE FROM cards WHERE id = ?`, id)
	return err
}

// ListSessions retrieves all sessions
func (s *SQLite) ListSessions(limit, offset int) ([]models.Session, error) {
	query := `SELECT id, created_at FROM sessions ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		if err := rows.Scan(&session.ID, &session.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// GetSession retrieves a session with its cards
func (s *SQLite) GetSession(id string) (*models.Session, error) {
	// Get session
	var session models.Session
	query := `SELECT id, created_at FROM sessions WHERE id = ?`
	if err := s.db.QueryRow(query, id).Scan(&session.ID, &session.CreatedAt); err != nil {
		return nil, err
	}

	// Get cards for this session
	query = `
		SELECT c.id, c.front, c.back, c.tags, c.source_conversation, c.created_at, c.updated_at
		FROM cards c
		JOIN session_cards sc ON c.id = sc.card_id
		WHERE sc.session_id = ?
		ORDER BY c.created_at ASC
	`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var tagsJSON string

		err := rows.Scan(
			&card.ID, &card.Front, &card.Back, &tagsJSON,
			&card.SourceConversation, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &card.Tags); err != nil {
			card.Tags = []string{}
		}

		cards = append(cards, card)
	}

	session.Cards = cards
	return &session, rows.Err()
}

// GetCardsBySession retrieves all cards for a session
func (s *SQLite) GetCardsBySession(sessionID string) ([]models.Card, error) {
	query := `
		SELECT c.id, c.front, c.back, c.tags, c.source_conversation, c.created_at, c.updated_at
		FROM cards c
		JOIN session_cards sc ON c.id = sc.card_id
		WHERE sc.session_id = ?
		ORDER BY c.created_at ASC
	`

	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var tagsJSON string

		err := rows.Scan(
			&card.ID, &card.Front, &card.Back, &tagsJSON,
			&card.SourceConversation, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &card.Tags); err != nil {
			card.Tags = []string{}
		}

		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// ExportCardsToAnkiCSV exports cards to Anki CSV format
func (s *SQLite) ExportCardsToAnkiCSV(sessionID string) (string, error) {
	cards, err := s.GetCardsBySession(sessionID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Front,Back,Tags\n")

	for _, card := range cards {
		// Escape quotes and wrap in quotes
		front := strings.ReplaceAll(card.Front, `"`, `""`)
		back := strings.ReplaceAll(card.Back, `"`, `""`)
		tags := strings.Join(card.Tags, " ")

		sb.WriteString(fmt.Sprintf(`"%s","%s","%s"`+"\n", front, back, tags))
	}

	return sb.String(), nil
}
