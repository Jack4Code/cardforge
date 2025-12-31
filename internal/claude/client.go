package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropic-ai/anthropic-sdk-go"
	"github.com/anthropic-ai/anthropic-sdk-go/option"

	"github.com/Jack4Code/cardforge/internal/models"
)

// Client wraps the Anthropic Claude API client
type Client struct {
	client *anthropic.Client
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{client: client}
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

	// Call Claude API
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude_4_5_Sonnet_20250929),
		MaxTokens: anthropic.Int(16000),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("claude api error: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		if block.Type == anthropic.ContentBlockTypeText {
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
