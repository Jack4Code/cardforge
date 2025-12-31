import { useState } from 'react'
import ConversationInput from './components/ConversationInput'
import CardPreview from './components/CardPreview'
import CardEditor from './components/CardEditor'
import ExportButton from './components/ExportButton'

function App() {
  const [view, setView] = useState('input') // 'input', 'loading', 'review'
  const [cards, setCards] = useState([])
  const [sessionId, setSessionId] = useState(null)
  const [editingCard, setEditingCard] = useState(null)
  const [conversation, setConversation] = useState('')

  const handleGenerate = async (conversationText, options) => {
    setView('loading')
    setConversation(conversationText)

    try {
      const response = await fetch('/api/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          conversation: conversationText,
          options,
        }),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to generate cards')
      }

      const data = await response.json()
      setCards(data.cards)
      setSessionId(data.session_id)
      setView('review')
    } catch (error) {
      alert(`Error: ${error.message}`)
      setView('input')
    }
  }

  const handleUpdateCard = async (cardId, updates) => {
    try {
      const response = await fetch(`/api/cards/${cardId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to update card')
      }

      const updatedCard = await response.json()
      setCards(cards.map(c => c.id === cardId ? updatedCard : c))
      setEditingCard(null)
    } catch (error) {
      alert(`Error: ${error.message}`)
    }
  }

  const handleDeleteCard = async (cardId) => {
    if (!confirm('Are you sure you want to delete this card?')) {
      return
    }

    try {
      const response = await fetch(`/api/cards/${cardId}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to delete card')
      }

      setCards(cards.filter(c => c.id !== cardId))
    } catch (error) {
      alert(`Error: ${error.message}`)
    }
  }

  const handleGenerateMore = () => {
    setView('input')
  }

  const handleBack = () => {
    setView('input')
    setCards([])
    setSessionId(null)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-primary">CardForge</h1>
            {view === 'review' && (
              <button
                onClick={handleBack}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-900"
              >
                ← Back to Input
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {view === 'input' && (
          <ConversationInput onGenerate={handleGenerate} />
        )}

        {view === 'loading' && (
          <div className="flex flex-col items-center justify-center min-h-[400px]">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
            <p className="mt-4 text-gray-600">Generating flashcards...</p>
          </div>
        )}

        {view === 'review' && (
          <div>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">
                Generated Flashcards ({cards.length})
              </h2>
              <div className="flex gap-3">
                <button
                  onClick={handleGenerateMore}
                  className="px-4 py-2 text-sm font-medium text-primary border border-primary rounded-md hover:bg-primary hover:text-white transition-colors"
                >
                  Generate More
                </button>
                <ExportButton sessionId={sessionId} />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {cards.map(card => (
                <CardPreview
                  key={card.id}
                  card={card}
                  onEdit={() => setEditingCard(card)}
                  onDelete={() => handleDeleteCard(card.id)}
                />
              ))}
            </div>
          </div>
        )}
      </main>

      {/* Card Editor Modal */}
      {editingCard && (
        <CardEditor
          card={editingCard}
          onSave={(updates) => handleUpdateCard(editingCard.id, updates)}
          onCancel={() => setEditingCard(null)}
        />
      )}
    </div>
  )
}

export default App
