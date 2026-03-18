import { useState, useEffect } from 'react'
import { organizer } from '../../wailsjs/go/models'
import { GetCurrentLayout, GetCurrentAuthorFormat } from '../../wailsjs/go/main/App'
import { ArrowRight } from 'lucide-react'
import { RenameTemplateBuilder } from './RenameTemplateBuilder'
import { GetRenameConfig } from '../../wailsjs/go/main/App'

interface OutputPreviewSimpleProps {
  book: organizer.Metadata | null
  outputDir: string
}

export function OutputPreviewSimple({ book, outputDir }: OutputPreviewSimpleProps) {
  const [layout, setLayout] = useState('author-series-title')
  const [authorFormat, setAuthorFormat] = useState('preserve')
  const [renameEnabled, setRenameEnabled] = useState(false)

  // Load layout and author format on mount and when book changes
  useEffect(() => {
    const loadSettings = () => {
      GetCurrentLayout().then(l => setLayout(l)).catch(err => {
        console.error('Failed to get layout:', err)
      })
      GetCurrentAuthorFormat().then(f => setAuthorFormat(f)).catch(err => {
        console.error('Failed to get author format:', err)
      })
      GetRenameConfig().then(cfg => {
        setRenameEnabled(cfg.enabled || false)
      }).catch(err => {
        console.error('Failed to get rename config:', err)
      })
    }

    loadSettings()

    // Poll for changes every 500ms to catch layout/format updates
    const interval = setInterval(loadSettings, 500)
    return () => clearInterval(interval)
  }, [book])

  if (!book) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        Select a file to preview output path
      </div>
    )
  }

  const formatAuthorName = (name: string): string => {
    if (authorFormat === 'preserve') {
      return name
    }

    const parts = name.split(' ')
    if (parts.length < 2) return name

    if (authorFormat === 'first-last') {
      // Already in First Last format, return as-is
      return name
    } else if (authorFormat === 'last-first') {
      // Convert to Last, First
      const lastName = parts[parts.length - 1]
      const firstName = parts.slice(0, -1).join(' ')
      return `${lastName}, ${firstName}`
    }

    return name
  }

  const rawAuthor = book.authors?.[0] || 'Unknown Author'
  const author = formatAuthorName(rawAuthor)
  const series = book.series?.[0]
  const title = book.title || book.album || 'Unknown Title'
  const filename = book.source_path?.split('/').pop() || 'file.mp3'

  // Build output path based on layout
  const pathParts = [outputDir || '/output']

  switch (layout) {
    case 'author-series-title':
      pathParts.push(author)
      if (series) pathParts.push(series)
      pathParts.push(title)
      break
    case 'author-title':
      pathParts.push(author, title)
      break
    case 'series-title':
      if (series) pathParts.push(series)
      pathParts.push(title)
      break
    case 'author-only':
      pathParts.push(author)
      break
  }

  pathParts.push(filename)
  const outputPath = pathParts.join('/')

  return (
    <div className="p-4 space-y-3">
      {/* Before → After */}
      <div className="space-y-2">
        <div>
          <div className="text-xs font-medium text-muted-foreground mb-1">INPUT</div>
          <div className="p-2 bg-muted/50 rounded text-xs font-mono break-all">
            {book.source_path}
          </div>
        </div>

        <div className="flex justify-center">
          <ArrowRight className="h-4 w-4 text-primary" />
        </div>

        <div>
          <div className="text-xs font-medium text-muted-foreground mb-1">OUTPUT</div>
          <div className="p-2 bg-green-500/10 border border-green-500/20 rounded text-xs font-mono break-all">
            {/* Color-coded path segments matching Path Structure */}
            {pathParts.map((part, idx) => {
              let color = 'text-foreground'
              if (idx === 0) {
                // Output dir - default color
                color = 'text-muted-foreground'
              } else if (part === author) {
                color = 'text-orange-600'
              } else if (part === series) {
                color = 'text-cyan-600'
              } else if (part === title) {
                color = 'text-green-600'
              } else if (part === filename) {
                color = 'text-blue-600'
              }
              return (
                <span key={idx}>
                  <span className={color}>{part}</span>
                  {idx < pathParts.length - 1 && <span className="text-muted-foreground">/</span>}
                </span>
              )
            })}
          </div>
        </div>
      </div>

      {/* Path Structure */}
      <div className="border-t border-border pt-2 mt-2">
        <div className="text-xs font-medium mb-2">Path Structure</div>
        <div className="space-y-1.5 text-xs">
          <div className="flex items-start gap-2">
            <span className="text-muted-foreground w-16 flex-shrink-0">Author:</span>
            <span className="font-medium text-orange-600">{author}</span>
          </div>
          {series && (
            <div className="flex items-start gap-2">
              <span className="text-muted-foreground w-16 flex-shrink-0">Series:</span>
              <span className="font-medium text-cyan-600">{series}</span>
            </div>
          )}
          <div className="flex items-start gap-2">
            <span className="text-muted-foreground w-16 flex-shrink-0">Title:</span>
            <span className="font-medium text-green-600">{title}</span>
          </div>
          {!renameEnabled && (
            <div className="flex items-start gap-2">
              <span className="text-muted-foreground w-16 flex-shrink-0">Filename:</span>
              <span className="font-medium text-blue-600">{filename}</span>
            </div>
          )}
        </div>
      </div>

      {/* Rename Template Builder */}
      <RenameTemplateBuilder book={book} />

    </div>
  )
}
