import { useState } from 'react'

function CardPreview({ card, onEdit, onDelete }) {
  const [isFlipped, setIsFlipped] = useState(false)

  const handleFlip = () => {
    setIsFlipped(!isFlipped)
  }

  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow">
      {/* Card Content */}
      <div
        onClick={handleFlip}
        className="p-6 min-h-[200px] cursor-pointer relative"
      >
        <div className="mb-4">
          <span className="text-xs font-medium text-gray-500 uppercase">
            {isFlipped ? 'Back' : 'Front'}
          </span>
        </div>

        <div className="text-gray-900">
          {isFlipped ? card.back : card.front}
        </div>

        {/* Flip indicator */}
        <div className="absolute bottom-4 right-4 text-xs text-gray-400">
          Click to flip
        </div>
      </div>

      {/* Tags */}
      {card.tags && card.tags.length > 0 && (
        <div className="px-6 py-3 bg-gray-50 border-t border-gray-100">
          <div className="flex flex-wrap gap-2">
            {card.tags.map((tag, index) => (
              <span
                key={index}
                className="px-2 py-1 text-xs font-medium bg-primary/10 text-primary rounded-md"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="px-6 py-3 bg-gray-50 border-t border-gray-100 flex gap-2">
        <button
          onClick={onEdit}
          className="flex-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
        >
          Edit
        </button>
        <button
          onClick={onDelete}
          className="px-3 py-2 text-sm font-medium text-red-600 bg-white border border-red-300 rounded-md hover:bg-red-50 transition-colors"
        >
          Delete
        </button>
      </div>
    </div>
  )
}

export default CardPreview
