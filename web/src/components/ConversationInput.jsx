import { useState } from 'react'

function ConversationInput({ onGenerate }) {
  const [conversation, setConversation] = useState('')
  const [maxCards, setMaxCards] = useState(50)
  const [difficulty, setDifficulty] = useState('mixed')
  const [showOptions, setShowOptions] = useState(false)

  const handleSubmit = (e) => {
    e.preventDefault()

    if (!conversation.trim()) {
      alert('Please enter a conversation')
      return
    }

    onGenerate(conversation, {
      max_cards: maxCards,
      difficulty,
      topics: [],
    })
  }

  const handleClear = () => {
    setConversation('')
  }

  const handleKeyDown = (e) => {
    // Ctrl/Cmd + Enter to submit
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      handleSubmit(e)
    }
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">
          Paste Conversation
        </h2>

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <textarea
              value={conversation}
              onChange={(e) => setConversation(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Paste your conversation transcript here..."
              className="w-full h-96 px-4 py-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
              autoFocus
            />
            <div className="flex items-center justify-between mt-2 text-sm text-gray-600">
              <span>{conversation.length.toLocaleString()} characters</span>
              {conversation.length > 0 && (
                <button
                  type="button"
                  onClick={handleClear}
                  className="text-primary hover:text-blue-700"
                >
                  Clear
                </button>
              )}
            </div>
          </div>

          {/* Options */}
          <div className="mb-4">
            <button
              type="button"
              onClick={() => setShowOptions(!showOptions)}
              className="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-1"
            >
              <span>{showOptions ? '▼' : '▶'}</span>
              <span>Options</span>
            </button>

            {showOptions && (
              <div className="mt-3 p-4 bg-gray-50 rounded-md space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Max Cards
                  </label>
                  <input
                    type="number"
                    value={maxCards}
                    onChange={(e) => setMaxCards(parseInt(e.target.value))}
                    min="1"
                    max="100"
                    className="w-32 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Difficulty
                  </label>
                  <select
                    value={difficulty}
                    onChange={(e) => setDifficulty(e.target.value)}
                    className="w-48 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary focus:border-transparent"
                  >
                    <option value="basic">Basic</option>
                    <option value="intermediate">Intermediate</option>
                    <option value="advanced">Advanced</option>
                    <option value="mixed">Mixed</option>
                  </select>
                </div>
              </div>
            )}
          </div>

          <button
            type="submit"
            disabled={!conversation.trim()}
            className="w-full px-6 py-3 bg-primary text-white font-medium rounded-md hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
          >
            Generate Cards
          </button>

          <p className="mt-2 text-sm text-gray-500 text-center">
            Press Ctrl+Enter to generate
          </p>
        </form>
      </div>
    </div>
  )
}

export default ConversationInput
