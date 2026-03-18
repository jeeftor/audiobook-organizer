import { organizer } from '../../wailsjs/go/models'
import { FileAudio } from 'lucide-react'

interface InputFileListProps {
  books: organizer.Metadata[]
  selectedIndex: number | null
  onSelect: (index: number) => void
  loading: boolean
}

export function InputFileList({ books, selectedIndex, onSelect, loading }: InputFileListProps) {
  if (loading) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        <div className="animate-pulse">Scanning files...</div>
      </div>
    )
  }

  if (books.length === 0) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        No files found. Click "Open Folder" to scan a directory.
      </div>
    )
  }

  return (
    <div className="divide-y divide-border">
      {books.map((book, index) => {
        const isSelected = selectedIndex === index
        const filename = book.source_path?.split('/').pop() || 'Unknown file'
        const fileExt = filename.split('.').pop()?.toUpperCase() || 'FILE'

        return (
          <div
            key={index}
            onClick={() => onSelect(index)}
            className={`p-2 cursor-pointer transition-colors ${
              isSelected
                ? 'bg-primary/10 border-l-2 border-l-primary'
                : 'hover:bg-accent border-l-2 border-l-transparent'
            }`}
          >
            <div className="flex items-start gap-2">
              <FileAudio className="h-4 w-4 mt-0.5 text-muted-foreground flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <div className="text-xs font-mono truncate" title={filename}>
                  {filename}
                </div>
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-[10px] px-1 py-0.5 rounded bg-muted text-muted-foreground font-mono">
                    {fileExt}
                  </span>
                  {book.title && (
                    <span className="text-[10px] text-muted-foreground truncate">
                      {book.title}
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
