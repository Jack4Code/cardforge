package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Jack4Code/cardforge/internal/models"
)

const (
	apiURL      = "https://api.anthropic.com/v1/messages"
	apiVersion  = "2023-06-01"
	modelSonnet = "claude-sonnet-4-5-20250929"
)

// Client wraps the Anthropic Claude API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// API request/response structures
type messageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type messageResponse struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
	Model   string         `json:"model"`
	Error   *apiError      `json:"error,omitempty"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// GeneratedCard represents a card from Claude's response
type GeneratedCard struct {
	Front string   `json:"front"`
	Back  string   `json:"back"`
	Tags  []string `json:"tags"`
}

// GenerateCards generates flashcards from a conversation using Claude
func (c *Client) GenerateCards(ctx context.Context, conversation string, options models.GenerateOptions) ([]GeneratedCard, error) {
	if conversation == "" {
		return nil, fmt.Errorf("conversation cannot be empty")
	}

	// Set defaults
	if options.MaxCards == 0 {
		options.MaxCards = 50
	}
	if options.Difficulty == "" {
		options.Difficulty = "mixed"
	}

	// Build the prompt
	prompt := c.buildPrompt(conversation, options)

	// Create API request
	reqBody := messageRequest{
		Model:     modelSonnet,
		MaxTokens: 16000,
		Messages: []message{
			{
				Role: "user",
				Content: []contentBlock{
					{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errResp messageResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
			return nil, fmt.Errorf("claude api error: %s - %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("claude api error: status %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp messageResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("claude api error: %s - %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	// Extract text from response
	var responseText string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("no text content in Claude response")
	}

	// Parse the JSON response
	cards, err := c.parseCardsFromResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cards: %w", err)
	}

	// Limit to max cards
	if len(cards) > options.MaxCards {
		cards = cards[:options.MaxCards]
	}

	return cards, nil
}

// buildPrompt constructs the prompt for Claude
func (c *Client) buildPrompt(conversation string, options models.GenerateOptions) string {
	var sb strings.Builder

	sb.WriteString("You are a flashcard generation expert. Given a conversation transcript, extract key facts, concepts, and knowledge into high-quality flashcards suitable for spaced repetition learning.\n\n")

	sb.WriteString("Guidelines:\n")
	sb.WriteString("- Create clear, concise questions that test understanding\n")
	sb.WriteString("- Answers should be complete but not overly verbose\n")
	sb.WriteString("- Include context where needed\n")
	sb.WriteString("- Tag appropriately by topic (use lowercase, hyphenated tags)\n")

	// Add difficulty-specific instructions
	switch options.Difficulty {
	case "basic":
		sb.WriteString("- Focus on fundamental concepts and definitions\n")
		sb.WriteString("- Keep questions simple and straightforward\n")
	case "intermediate":
		sb.WriteString("- Focus on application and understanding of concepts\n")
		sb.WriteString("- Include some complexity in questions\n")
	case "advanced":
		sb.WriteString("- Focus on complex concepts and relationships\n")
		sb.WriteString("- Include challenging questions that require deep understanding\n")
	default: // mixed
		sb.WriteString("- Vary difficulty levels from basic to advanced\n")
	}

	sb.WriteString("- Avoid yes/no questions\n")
	sb.WriteString("- Prefer \"What/How/Why\" questions\n")
	sb.WriteString("- One concept per card\n")
	sb.WriteString("- Extract the most important and memorable information\n")
	sb.WriteString("- Focus on concepts that would benefit from spaced repetition\n\n")

	// Add topic filtering if specified
	if len(options.Topics) > 0 {
		sb.WriteString(fmt.Sprintf("Focus on these topics: %s\n\n", strings.Join(options.Topics, ", ")))
	}

	sb.WriteString("Conversation:\n")
	sb.WriteString(conversation)
	sb.WriteString("\n\n")

	sb.WriteString("Generate flashcards in JSON format:\n")
	sb.WriteString("[\n")
	sb.WriteString("  {\n")
	sb.WriteString("    \"front\": \"question\",\n")
	sb.WriteString("    \"back\": \"answer\",\n")
	sb.WriteString("    \"tags\": [\"tag1\", \"tag2\"]\n")
	sb.WriteString("  }\n")
	sb.WriteString("]\n\n")

	sb.WriteString("Return ONLY the JSON array, no additional text or markdown formatting.")

	return sb.String()
}

// parseCardsFromResponse parses the JSON response from Claude
func (c *Client) parseCardsFromResponse(response string) ([]GeneratedCard, error) {
	// Clean up the response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Parse JSON
	var cards []GeneratedCard
	if err := json.Unmarshal([]byte(response), &cards); err != nil {
		return nil, fmt.Errorf("json parse error: %w (response: %s)", err, response)
	}

	return cards, nil
}
