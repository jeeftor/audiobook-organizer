import { useState, useEffect } from 'react'
import { organizer } from '../../wailsjs/go/models'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { GetCurrentLayout, GetCurrentAuthorFormat } from '../../wailsjs/go/main/App'
import { buildOutputParts } from '../utils/pathUtils'
import { ColoredPath } from './ColoredPath'

interface BatchPreviewProps {
  books: organizer.Metadata[]
  selectedIndices: Set<number>
  outputDir: string
}

export function BatchPreview({ books, selectedIndices, outputDir }: BatchPreviewProps) {
  const [expanded, setExpanded] = useState(false)
  const [layout, setLayout] = useState('author-series-title')
  const [authorFormat, setAuthorFormat] = useState('preserve')

  // Load layout and author format
  useEffect(() => {
    const loadSettings = () => {
      GetCurrentLayout().then(l => setLayout(l)).catch(err => {
        console.error('Failed to get layout:', err)
      })
      GetCurrentAuthorFormat().then(f => setAuthorFormat(f)).catch(err => {
        console.error('Failed to get author format:', err)
      })
    }
    loadSettings()
    const interval = setInterval(loadSettings, 500)
    return () => clearInterval(interval)
  }, [])

  // Filter to only selected books
  const selectedBooks = books.filter((_, idx) => selectedIndices.has(idx))

  if (selectedBooks.length === 0) {
    return null
  }

  return (
    <div className="border-t border-border mt-3">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full p-2 flex items-center justify-between hover:bg-muted/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-4 w-4 text-muted-foreground" />
          )}
          <span className="text-xs font-medium">Batch Preview</span>
          <span className="text-xs text-muted-foreground">({selectedBooks.length} selected)</span>
        </div>
      </button>

      {expanded && (
        <div className="p-2 max-h-96 overflow-y-auto">
          <div className="grid grid-cols-2 gap-2 text-xs">
            {/* Header */}
            <div className="font-medium text-muted-foreground">Input</div>
            <div className="font-medium text-muted-foreground">Output</div>

            {/* File list */}
            {selectedBooks.map((book, idx) => {
              const { parts, author, series, title, filename } = buildOutputParts(
                book,
                outputDir,
                layout,
                authorFormat
              )

              return (
                <div key={idx} className="contents">
                  {/* Input */}
                  <div className="p-1.5 bg-muted/20 rounded font-mono text-[10px] break-all border border-border">
                    {book.source_path}
                  </div>

                  {/* Output - color-coded */}
                  <div className="p-1.5 bg-green-500/10 rounded font-mono text-[10px] break-all border border-green-500/20">
                    <ColoredPath
                      parts={parts}
                      author={author}
                      series={series}
                      title={title}
                      filename={filename}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
