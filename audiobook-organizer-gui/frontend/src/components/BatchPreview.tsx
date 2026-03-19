import { useState, useEffect } from 'react'
import { main } from '../../wailsjs/go/models'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { GetBatchPreview } from '../../wailsjs/go/main/App'
import { ColoredPath } from './ColoredPath'
import { useSettings } from '../contexts/SettingsContext'

interface BatchPreviewProps {
  selectedIndices: Set<number>
  outputDir: string
}

export function BatchPreview({ selectedIndices, outputDir }: BatchPreviewProps) {
  const [expanded, setExpanded] = useState(false)
  const [previewItems, setPreviewItems] = useState<main.PreviewItem[]>([])
  const { settings } = useSettings()

  const selectedIndicesArray = Array.from(selectedIndices)

  // Re-fetch whenever selection, outputDir, or any relevant setting changes
  useEffect(() => {
    if (selectedIndicesArray.length === 0) {
      setPreviewItems([])
      return
    }
    GetBatchPreview(selectedIndicesArray, outputDir)
      .then(items => setPreviewItems(items || []))
      .catch(() => setPreviewItems([]))
  }, [selectedIndices, outputDir, settings.layout, settings.authorFormat, JSON.stringify(settings.fieldOptions)])

  if (previewItems.length === 0) {
    return null
  }

  const getPathParts = (item: main.PreviewItem) => {
    const effectiveOutputDir = item.output_dir || outputDir || '/output'
    const relPath = item.to.startsWith(effectiveOutputDir)
      ? item.to.slice(effectiveOutputDir.length)
      : item.to
    const relParts = relPath.split('/').filter(p => p.length > 0)
    return [effectiveOutputDir, ...relParts]
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
          <span className="text-xs text-muted-foreground">({previewItems.length} selected)</span>
        </div>
      </button>

      {expanded && (
        <div className="p-2 max-h-96 overflow-y-auto">
          <div className="grid grid-cols-2 gap-2 text-xs">
            {/* Header */}
            <div className="font-medium text-muted-foreground">Input</div>
            <div className="font-medium text-muted-foreground">Output</div>

            {/* File list */}
            {previewItems.map((item, idx) => {
              const parts = getPathParts(item)
              return (
                <div key={idx} className="contents">
                  <div className="p-1.5 bg-muted/20 rounded font-mono text-[10px] break-all border border-border">
                    {item.from}
                  </div>
                  <div className="p-1.5 font-mono text-[10px] break-all">
                    <ColoredPath
                      parts={parts}
                      author={item.author}
                      series={item.series || undefined}
                      title={item.title}
                      filename={item.filename}
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
