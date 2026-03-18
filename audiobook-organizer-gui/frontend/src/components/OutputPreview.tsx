import { organizer } from '../../wailsjs/go/models'
import { FolderTree, File } from 'lucide-react'

interface OutputPreviewProps {
  book: organizer.Metadata | null
  outputDir: string
  onSelectInput: (index: number) => void
}

export function OutputPreview({ book, outputDir, onSelectInput }: OutputPreviewProps) {
  if (!book) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        Select a file to preview output structure
      </div>
    )
  }

  const author = book.authors?.[0] || 'Unknown Author'
  const series = book.series?.[0]
  const title = book.title || book.album || 'Unknown Title'
  const filename = book.source_path?.split('/').pop() || 'file.mp3'

  return (
    <div className="p-4">
      {/* Output Path Tree */}
      <div className="space-y-1 font-mono text-xs">
        <div className="flex items-center gap-2 text-muted-foreground">
          <FolderTree className="h-3 w-3" />
          <span>{outputDir || '/output'}/</span>
        </div>

        <div className="ml-4 flex items-center gap-2 text-orange-600">
          <FolderTree className="h-3 w-3" />
          <span>{author}/</span>
        </div>

        {series && (
          <div className="ml-8 flex items-center gap-2 text-cyan-600">
            <FolderTree className="h-3 w-3" />
            <span>{series}/</span>
          </div>
        )}

        <div className={`${series ? 'ml-12' : 'ml-8'} flex items-center gap-2 text-green-600`}>
          <FolderTree className="h-3 w-3" />
          <span>{title}/</span>
        </div>

        <div className={`${series ? 'ml-16' : 'ml-12'} flex items-center gap-2 text-blue-600`}>
          <File className="h-3 w-3" />
          <span>{filename}</span>
        </div>
      </div>

      {/* Full Path */}
      <div className="mt-4 p-3 bg-muted rounded-lg">
        <div className="text-[10px] text-muted-foreground mb-1">Full Output Path:</div>
        <div className="text-xs font-mono break-all">
          {outputDir || '/output'}/{author}/{series ? `${series}/` : ''}{title}/{filename}
        </div>
      </div>

      {/* Metadata Summary */}
      <div className="mt-4 space-y-2">
        <div className="text-xs font-semibold">Metadata</div>
        <div className="space-y-1 text-xs">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Author:</span>
            <span className="text-orange-600 font-medium">{author}</span>
          </div>
          {series && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Series:</span>
              <span className="text-cyan-600 font-medium">{series}</span>
            </div>
          )}
          <div className="flex justify-between">
            <span className="text-muted-foreground">Title:</span>
            <span className="text-green-600 font-medium">{title}</span>
          </div>
          {book.track_number && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Track:</span>
              <span className="font-medium">{book.track_number}</span>
            </div>
          )}
        </div>
      </div>

      {/* Source File */}
      <div className="mt-4 p-3 bg-muted/50 rounded-lg">
        <div className="text-[10px] text-muted-foreground mb-1">Source File:</div>
        <div className="text-xs font-mono break-all text-muted-foreground">
          {book.source_path}
        </div>
      </div>
    </div>
  )
}
