import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { PreviewChanges as PreviewChangesAPI, ExecuteOrganize } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'

interface PreviewChangesProps {
  inputDir: string
  outputDir: string
  selectedIndices: number[]
  onBack: () => void
  onComplete: () => void
}

export function PreviewChanges({ inputDir, outputDir, selectedIndices, onBack, onComplete }: PreviewChangesProps) {
  const [preview, setPreview] = useState<main.PreviewItem[]>([])
  const [loading, setLoading] = useState(true)
  const [executing, setExecuting] = useState(false)
  const [error, setError] = useState('')
  const [useRename, setUseRename] = useState(false)
  const [showRenameDialog, setShowRenameDialog] = useState(false)
  const [renameTemplate, setRenameTemplate] = useState<string[]>(['', '', '', ''])
  const [renameSeparator, setRenameSeparator] = useState('-')

  useEffect(() => {
    const loadPreview = async () => {
      setLoading(true)
      setError('')
      console.log('PreviewChanges: Loading preview with:', { inputDir, outputDir, selectedIndices })
      try {
        const result = await PreviewChangesAPI(inputDir, outputDir, selectedIndices)
        console.log('PreviewChanges: Got result:', result)
        console.log('PreviewChanges: Result length:', result?.length || 0)
        setPreview(result || [])
      } catch (err) {
        console.error('PreviewChanges: Error:', err)
        setError(`Failed to generate preview: ${err}`)
      } finally {
        setLoading(false)
      }
    }
    loadPreview()
  }, [inputDir, outputDir, selectedIndices])

  const handleExecute = async () => {
    setExecuting(true)
    setError('')
    try {
      await ExecuteOrganize(false)
      onComplete()
    } catch (err) {
      setError(`Failed to organize files: ${err}`)
      setExecuting(false)
    }
  }

  // Helper to extract components from path
  const parseOutputPath = (path: string) => {
    const parts = path.split('/')
    if (parts.length < 3) return null

    // Assuming format: /output/Author/Series/Title/filename.ext
    // or: /output/Author/Title/filename.ext
    const filename = parts[parts.length - 1]
    const title = parts[parts.length - 2]
    const author = parts[parts.length - 3] === outputDir.split('/').pop() ? 'Unknown' : parts[parts.length - 3]
    const series = parts.length > 4 ? parts[parts.length - 3] : null

    return { author, series, title, filename }
  }

  // Format output filename based on rename template
  const formatOutputFilename = (item: main.PreviewItem) => {
    if (!useRename || renameTemplate.filter(f => f).length === 0) {
      // Use existing filename
      return item.to.split('/').pop() || 'unknown'
    }

    // Apply rename template (simplified for now)
    const parts = parseOutputPath(item.to)
    if (!parts) return item.to.split('/').pop() || 'unknown'

    const values: Record<string, string> = {
      author: parts.author,
      series: parts.series || 'Unknown Series',
      title: parts.title,
      track: '01' // TODO: get from metadata
    }

    const ext = parts.filename.split('.').pop()
    const renamed = renameTemplate
      .filter(f => f)
      .map(f => values[f] || f)
      .join(` ${renameSeparator} `)

    return renamed ? `${renamed}.${ext}` : parts.filename
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Card className="w-full max-w-4xl">
          <CardContent className="p-8 text-center">
            <div className="animate-pulse space-y-2">
              <div className="text-base font-medium">Generating preview...</div>
              <div className="text-xs text-muted-foreground">Calculating target paths</div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  const hasConflicts = preview.some(item => item.is_conflict)

  return (
    <div className="min-h-screen p-4 bg-background">
      <div className="max-w-7xl mx-auto space-y-4">
        {/* Header with controls */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-xl font-bold">Preview Changes ({preview.length} operations)</h1>
            {hasConflicts && (
              <span className="text-xs px-2 py-1 bg-yellow-500/20 text-yellow-600 rounded">
                ⚠️ {preview.filter(p => p.is_conflict).length} Conflicts
              </span>
            )}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={onBack} disabled={executing}>
              Back
            </Button>
            <Button
              onClick={handleExecute}
              disabled={executing || hasConflicts}
              size="sm"
            >
              {executing ? 'Organizing...' : 'Execute'}
            </Button>
          </div>
        </div>

        {error && (
          <div className="p-2 text-xs text-destructive bg-destructive/10 rounded">
            {error}
          </div>
        )}

        {/* Rename controls */}
        <Card>
          <CardHeader className="py-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">File Naming</CardTitle>
              <div className="flex gap-2">
                <Button
                  variant={!useRename ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setUseRename(false)}
                  className="h-7 text-xs"
                >
                  Keep Original Names
                </Button>
                <Button
                  variant={useRename ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setUseRename(true)}
                  className="h-7 text-xs"
                >
                  Rename Files
                </Button>
                {useRename && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowRenameDialog(true)}
                    className="h-7 text-xs"
                  >
                    Configure Template
                  </Button>
                )}
              </div>
            </div>
          </CardHeader>
          {useRename && renameTemplate.filter(f => f).length > 0 && (
            <CardContent className="py-2">
              <div className="text-xs font-mono text-muted-foreground">
                Template: {renameTemplate.filter(f => f).map(f => `{${f}}`).join(` ${renameSeparator} `)} + extension
              </div>
            </CardContent>
          )}
        </Card>

        {/* Preview list */}
        <Card>
          <CardContent className="py-2">
            {preview.length === 0 ? (
              <div className="p-4 text-center text-muted-foreground text-xs">
                No operations to preview
              </div>
            ) : (
              <div className="space-y-0.5 max-h-[70vh] overflow-y-auto font-mono text-[10px]">
                {preview.map((item, idx) => {
                  const sourcePath = item.from
                  const targetPath = item.to

                  // Parse output path to get components for coloring
                  const outputParts = targetPath.split('/')
                  const outputDirPart = outputDir.split('/').pop() || ''
                  const outputDirIndex = outputParts.indexOf(outputDirPart)

                  // Build path with colors
                  let pathPrefix = ''
                  let author = ''
                  let series = ''
                  let title = ''
                  let filename = ''

                  if (outputDirIndex >= 0) {
                    // Everything before the output dir (like /Users/jstein/output)
                    pathPrefix = outputParts.slice(0, outputDirIndex + 1).join('/')

                    if (outputParts.length > outputDirIndex + 1) {
                      author = outputParts[outputDirIndex + 1] || ''

                      if (outputParts.length > outputDirIndex + 4) {
                        // Layout: author-series-title -> output/Author/Series/Title/filename
                        series = outputParts[outputDirIndex + 2] || ''
                        title = outputParts[outputDirIndex + 3] || ''
                        filename = outputParts[outputDirIndex + 4] || ''
                      } else if (outputParts.length === outputDirIndex + 4) {
                        // Could be author-series (output/Author/Series/filename) OR author-title (output/Author/Title/filename)
                        // Check if the middle part looks like a series (same as album field would be)
                        const middlePart = outputParts[outputDirIndex + 2] || ''
                        // For now, treat this as author-series layout (series, no title)
                        series = middlePart
                        filename = outputParts[outputDirIndex + 3] || ''
                      } else if (outputParts.length === outputDirIndex + 3) {
                        // Layout: author-only -> output/Author/filename
                        filename = outputParts[outputDirIndex + 2] || ''
                      }
                    }
                  }

                  const finalFilename = useRename ? formatOutputFilename(item) : filename

                  return (
                    <div
                      key={idx}
                      className={`px-2 py-1 rounded ${
                        item.is_conflict ? 'bg-yellow-500/10' : ''
                      }`}
                    >
                      {/* Source - full path */}
                      <div className="text-muted-foreground mb-0.5">
                        {sourcePath}
                      </div>

                      {/* Destination - full path with colored components, NO spaces around slashes */}
                      <div className="ml-2">
                        <span className="text-muted-foreground">{pathPrefix}/</span>
                        <span className="text-orange-600">{author}</span>
                        {series && (
                          <>
                            <span className="text-muted-foreground">/</span>
                            <span className="text-cyan-600">{series}</span>
                          </>
                        )}
                        {title && (
                          <>
                            <span className="text-muted-foreground">/</span>
                            <span className="text-green-600">{title}</span>
                          </>
                        )}
                        <span className="text-muted-foreground">/</span>
                        <span className={useRename ? "text-yellow-600 font-bold" : "text-muted-foreground"}>
                          {finalFilename}
                        </span>
                      </div>

                      {item.is_conflict && (
                        <div className="ml-2 text-[9px] text-yellow-600 mt-0.5">
                          ⚠️ Conflict: Multiple files → same location
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Rename Template Dialog */}
      {showRenameDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setShowRenameDialog(false)}>
          <Card className="w-full max-w-2xl" onClick={(e) => e.stopPropagation()}>
            <CardHeader className="py-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">Configure Filename Template</CardTitle>
                <Button variant="outline" size="sm" onClick={() => setShowRenameDialog(false)}>
                  Close
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="text-sm font-medium mb-2 text-cyan-600">Available Fields:</div>
                <div className="space-y-1">
                  {['author', 'series', 'track', 'title'].map((field) => (
                    <div key={field} className="flex items-center gap-2">
                      <span className="text-yellow-600 font-mono text-sm w-20">{field}</span>
                      <div className="flex gap-1">
                        {[0, 1, 2, 3].map((slot) => (
                          <Button
                            key={slot}
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              const newTemplate = [...renameTemplate]
                              newTemplate[slot] = field
                              setRenameTemplate(newTemplate)
                            }}
                            className="h-6 w-6 p-0 text-xs"
                          >
                            {slot + 1}
                          </Button>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <div className="text-sm font-medium mb-2">Template Slots:</div>
                <div className="space-y-1">
                  {renameTemplate.map((field, idx) => (
                    <div key={idx} className="flex items-center gap-2 font-mono text-sm">
                      <span className="w-8">{idx + 1}.</span>
                      <span className={field ? 'text-yellow-600 font-bold' : 'text-muted-foreground'}>
                        {field ? `{${field}}` : '<empty>'}
                      </span>
                      {field && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            const newTemplate = [...renameTemplate]
                            newTemplate[idx] = ''
                            setRenameTemplate(newTemplate)
                          }}
                          className="h-6 text-xs"
                        >
                          Clear
                        </Button>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div>
                <div className="text-sm font-medium mb-2">Separator:</div>
                <div className="flex gap-2">
                  {['-', '/', ' ', '.', '_'].map((sep) => (
                    <Button
                      key={sep}
                      variant={renameSeparator === sep ? 'default' : 'outline'}
                      size="sm"
                      onClick={() => setRenameSeparator(sep)}
                      className="font-mono"
                    >
                      {sep === ' ' ? '<space>' : sep}
                    </Button>
                  ))}
                </div>
              </div>

              <div>
                <div className="text-sm font-medium mb-2">Preview:</div>
                <div className="p-3 bg-muted rounded font-mono text-sm">
                  {renameTemplate.filter(f => f).length > 0 ? (
                    renameTemplate
                      .filter(f => f)
                      .map(f => `{${f}}`)
                      .join(` ${renameSeparator} `) + '.{ext}'
                  ) : (
                    '<empty template>'
                  )}
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <Button variant="outline" onClick={() => setShowRenameDialog(false)}>
                  Done
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
