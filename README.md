# CardForge - AI-Powered Flashcard Generator

CardForge is a web application that takes conversation transcripts and uses Claude AI to automatically generate high-quality Anki flashcards. Built with Go (using Bedrock framework) and React, it provides a clean, fast, and focused workflow: paste → generate → export.

## Features

- **AI-Powered Generation**: Uses Claude Sonnet 4.5 to extract key concepts from conversations
- **Quality Flashcards**: Generates focused, spaced-repetition friendly cards
- **Easy Editing**: Edit cards inline before export
- **Anki Export**: Direct export to Anki CSV format
- **Session Management**: Track multiple generation sessions
- **Single Binary Deployment**: Everything bundled in one executable

## Technology Stack

### Backend
- **Framework**: [Bedrock v0.0.5](https://github.com/Jack4Code/bedrock)
- **Database**: SQLite
- **AI**: Anthropic Claude API (Sonnet 4.5) via direct HTTP
- **Language**: Go 1.25

### Frontend
- **Framework**: React with Vite
- **Styling**: Tailwind CSS
- **State**: React hooks only

## Prerequisites

- Go 1.25+
- Node.js 18+
- Anthropic API key ([get one here](https://console.anthropic.com/))

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/Jack4Code/cardforge.git
cd cardforge
```

### 2. Set up environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add your Anthropic API key
# ANTHROPIC_API_KEY=sk-ant-your-key-here
```

### 3. Install dependencies

```bash
# Backend dependencies
go mod download

# Frontend dependencies
cd web
npm install
cd ..
```

### 4. Run in development mode

**Terminal 1 - Backend:**
```bash
go run main.go
```

**Terminal 2 - Frontend:**
```bash
cd web
npm run dev
```

Open http://localhost:5173 in your browser.

### 5. Build for production

```bash
# Build frontend
cd web
npm run build
cd ..

# Build Go binary with embedded assets
go build -o cardforge main.go

# Run the binary
./cardforge
```

Access the application at http://localhost:8080

## Usage

### Generate Flashcards

1. **Paste Conversation**: Copy your conversation transcript into the text area
2. **Configure Options** (optional):
   - Max cards (default: 50)
   - Difficulty level (basic, intermediate, advanced, mixed)
3. **Generate**: Click "Generate Cards" or press Ctrl+Enter
4. **Review & Edit**: Preview generated cards, edit as needed
5. **Export**: Download as Anki CSV

### Keyboard Shortcuts

- `Ctrl+Enter` / `Cmd+Enter` - Generate cards from input

## API Endpoints

```
POST   /api/generate          - Generate cards from conversation
GET    /api/cards             - List all cards (pagination supported)
POST   /api/cards             - Save cards to database
PUT    /api/cards/:id         - Update a specific card
DELETE /api/cards/:id         - Delete a card
GET    /api/sessions          - List all sessions
GET    /api/sessions/:id      - Get session with cards
GET    /api/export            - Export cards (format=anki, session_id required)
GET    /health                - Health check endpoint
```

## Configuration

All configuration is done via environment variables:

```bash
# Required
ANTHROPIC_API_KEY=sk-ant-your-api-key-here

# Optional (with defaults)
DATABASE_PATH=./cardforge.db
PORT=8080
FRONTEND_DIR=./web/dist
ENV=development
```

## Project Structure

```
cardforge/
├── main.go                      # Application entry point
├── internal/
│   ├── handlers/                # HTTP request handlers
│   │   └── handlers.go
│   ├── models/                  # Data models
│   │   └── card.go
│   ├── storage/                 # Database layer
│   │   └── sqlite.go
│   └── claude/                  # Claude API client
│       └── client.go
├── web/
│   ├── src/
│   │   ├── components/          # React components
│   │   │   ├── ConversationInput.jsx
│   │   │   ├── CardPreview.jsx
│   │   │   ├── CardEditor.jsx
│   │   │   └── ExportButton.jsx
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── dist/                    # Built frontend (generated)
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
├── migrations/
│   └── 001_initial_schema.sql   # Database schema
├── build.sh                     # Production build script
├── go.mod
├── go.sum
├── README.md
└── .env.example
```

## Database Schema

CardForge uses SQLite with three main tables:

- **cards**: Stores flashcards (front, back, tags, metadata)
- **sessions**: Tracks generation sessions
- **session_cards**: Links cards to sessions

Migrations run automatically on startup.

## Development

### Running Tests

```bash
# Backend tests
go test ./...

# Frontend tests
cd web
npm test
```

### Building

```bash
# Build script
./build.sh
```

The build script:
1. Builds the React frontend
2. Compiles the Go binary with embedded assets
3. Creates a single executable: `./cardforge`

## Deployment

CardForge deploys as a single binary with minimal dependencies:

1. Build the application
2. Copy the binary and `.env` to your server
3. Run with `./cardforge`

### Systemd Service Example

```ini
[Unit]
Description=CardForge - AI Flashcard Generator
After=network.target

[Service]
Type=simple
User=cardforge
WorkingDirectory=/opt/cardforge
ExecStart=/opt/cardforge/cardforge
Restart=on-failure
Environment="ANTHROPIC_API_KEY=sk-ant-..."
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
```

## Example Card Quality

### Good Examples
- **Front**: "What is NAT?"
  **Back**: "Network Address Translation - allows multiple devices to share one public IP address by rewriting packet headers at the router level."

- **Front**: "Why do submarine cables dominate internet traffic over satellites?"
  **Back**: "Higher bandwidth (terabits/sec), lower latency (~80ms vs 600ms), better reliability, and lower cost per bit at scale."

### What to Avoid
- Yes/no questions
- Overly broad questions
- Multiple concepts in one card

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable is required"
- Make sure you've created a `.env` file with your API key
- Verify the API key is valid

### Database errors
- Check that the `migrations/` directory is accessible
- Ensure write permissions for the database file location

### Frontend build fails
- Run `npm install` in the `web/` directory
- Check Node.js version (18+ required)

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Bedrock](https://github.com/Jack4Code/bedrock)
- Powered by [Anthropic Claude](https://www.anthropic.com/)
- UI styled with [Tailwind CSS](https://tailwindcss.com/)

## Support

For issues, questions, or suggestions:
- Open an issue on GitHub
- Check existing issues for solutions

---

Built with ❤️ for better learning through spaced repetition
