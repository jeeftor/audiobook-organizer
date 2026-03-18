import { organizer } from '../../wailsjs/go/models'

interface FileListProps {
  books: organizer.Metadata[]
  selectedIndex: number | null
  onSelect: (index: number) => void
  loading: boolean
}

export function FileList({ books, selectedIndex, onSelect, loading }: FileListProps) {
  if (loading) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        <div className="animate-pulse">Scanning...</div>
      </div>
    )
  }

  if (books.length === 0) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        No audiobooks found. Click "Add Folder" to get started.
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      {/* Header */}
      <div className="sticky top-0 bg-card border-b border-border p-2">
        <div className="text-sm font-medium">
          Unclustered Files ({books.length})
        </div>
      </div>

      {/* File List */}
      <div className="divide-y divide-border">
        {books.map((book, index) => {
          const isSelected = selectedIndex === index
          const title = book.album || book.title || 'Unknown Title'
          const author = book.authors && book.authors.length > 0
            ? book.authors.join(', ')
            : 'Unknown Author'

          return (
            <div
              key={index}
              onClick={() => onSelect(index)}
              className={`p-3 cursor-pointer transition-colors ${
                isSelected
                  ? 'bg-primary/10 border-l-2 border-l-primary'
                  : `${index % 2 === 0 ? 'bg-muted/20' : 'bg-background'} hover:bg-accent border-l-2 border-l-transparent`
              }`}
            >
              <div className="text-sm font-medium truncate" title={title}>
                {title}
              </div>
              <div className="text-xs text-muted-foreground truncate mt-1" title={author}>
                {author}
              </div>
              {book.series && book.series.length > 0 && (
                <div className="text-xs text-cyan-600 truncate mt-0.5">
                  {book.series.join(', ')}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
