function ExportButton({ sessionId }) {
  const handleExport = async () => {
    try {
      const response = await fetch(`/api/export?format=anki&session_id=${sessionId}`)

      if (!response.ok) {
        const error = await response.json()
        throw new Error(error.error || 'Failed to export cards')
      }

      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `flashcards_${sessionId}.csv`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (error) {
      alert(`Error: ${error.message}`)
    }
  }

  return (
    <button
      onClick={handleExport}
      className="px-6 py-2 text-sm font-medium text-white bg-success rounded-md hover:bg-green-600 transition-colors"
    >
      Export to Anki
    </button>
  )
}

export default ExportButton
