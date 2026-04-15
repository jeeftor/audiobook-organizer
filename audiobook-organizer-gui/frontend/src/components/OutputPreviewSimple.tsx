import { useState, useEffect } from 'react'
import { organizer, main } from '../../wailsjs/go/models'
import { GetLivePreviewPath } from '../../wailsjs/go/main/App'
import { ArrowRight } from 'lucide-react'
import { RenameTemplateBuilder } from './RenameTemplateBuilder'
import { ColoredPath } from './ColoredPath'
import { CoverArt } from './CoverArt'
import { ValidationPanel } from './ValidationPanel'
import { useSettings } from '../contexts/SettingsContext'

interface OutputPreviewSimpleProps {
  book: organizer.Metadata | null
  bookIdx: number | null
  outputDir: string
  showCoverArt?: boolean
  scanVersion?: number
}

export function OutputPreviewSimple({ book, bookIdx, outputDir, showCoverArt = true, scanVersion = 0 }: OutputPreviewSimpleProps) {
  const [preview, setPreview] = useState<main.PreviewItem | null>(null)
  const { settings } = useSettings()

  const renameEnabled = settings.renameConfig.enabled

  // Re-fetch preview whenever bookIdx, outputDir, or any relevant setting changes
  useEffect(() => {
    if (bookIdx === null || bookIdx < 0) {
      setPreview(null)
      return
    }
    GetLivePreviewPath(bookIdx, outputDir).then(item => {
      setPreview(item)
    }).catch(err => {
      console.error('Failed to get live preview path:', err)
      setPreview(null)
    })
  }, [bookIdx, outputDir, settings.layout, settings.authorFormat, JSON.stringify(settings.fieldOptions)])

  if (!book || !preview) {
    return (
      <div className="space-y-0">
        {showCoverArt && (
          <div className="p-4 pb-2">
            <CoverArt bookIdx={bookIdx} />
          </div>
        )}
        <div className="p-4 text-center text-sm text-muted-foreground">
          Select a file to preview output path
        </div>
        <div className="border-t border-border">
          <ValidationPanel scanVersion={scanVersion} />
        </div>
      </div>
    )
  }

  // Build path parts for color-coded display from backend-provided fields
  const effectiveOutputDir = preview.output_dir || outputDir || '/output'
  const relPath = preview.to.startsWith(effectiveOutputDir)
    ? preview.to.slice(effectiveOutputDir.length)
    : preview.to
  const relParts = relPath.split('/').filter(p => p.length > 0)
  const pathParts = [effectiveOutputDir, ...relParts]

  const author = preview.author || 'Unknown Author'
  const series = preview.series || undefined
  const title = preview.title || 'Unknown Title'
  const filename = preview.filename || book.source_path?.split('/').pop() || 'file.mp3'

  return (
    <div className="p-4 space-y-3">
      {/* Cover Art */}
      {showCoverArt && (
        <div className="pb-1">
          <CoverArt bookIdx={bookIdx} />
        </div>
      )}

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
          <div className="p-2 text-xs font-mono break-all">
            <ColoredPath
              parts={pathParts}
              author={author}
              series={series}
              title={title}
              filename={filename}
            />
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

      {/* Validation Panel */}
      <div className="border-t border-border -mx-4 mt-2">
        <ValidationPanel scanVersion={scanVersion} />
      </div>

    </div>
  )
}
