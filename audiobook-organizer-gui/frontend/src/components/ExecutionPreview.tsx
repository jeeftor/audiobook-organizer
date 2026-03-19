import { useState, useEffect } from 'react'
import { main } from '../../wailsjs/go/models'
import { ChevronDown, ChevronRight, Copy } from 'lucide-react'
import {
  UpdateFieldMappingField,
  UpdateLayout,
  GetBatchPreview,
} from '../../wailsjs/go/main/App'
import { ColoredPath } from './ColoredPath'
import { useSettings } from '../contexts/SettingsContext'

interface ExecutionPreviewProps {
  selectedIndices: Set<number>
  inputDir: string
  outputDir: string
  onExecute: (copyMode: boolean, operations: Array<{from: string, to: string}>) => void
  onCancel: () => void
  onFieldMappingChange?: () => void
}

export function ExecutionPreview({
  selectedIndices,
  inputDir,
  outputDir,
  onExecute,
  onCancel,
  onFieldMappingChange,
}: ExecutionPreviewProps) {
  const [copyMode, setCopyMode] = useState(false)
  const [showCommands, setShowCommands] = useState(false)
  const [previewItems, setPreviewItems] = useState<main.PreviewItem[]>([])
  const [previewLoading, setPreviewLoading] = useState(false)
  const [isExecuting, setIsExecuting] = useState(false)
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [viewMode, setViewMode] = useState<'books' | 'files'>('books')
  const { settings, refreshSettings } = useSettings()

  const selectedIndicesArray = Array.from(selectedIndices)

  // Reload batch preview whenever selection, outputDir, or any relevant setting changes
  useEffect(() => {
    if (selectedIndicesArray.length === 0) {
      setPreviewItems([])
      return
    }
    setPreviewLoading(true)
    GetBatchPreview(selectedIndicesArray, outputDir)
      .then(items => setPreviewItems(items || []))
      .catch(err => console.error('Failed to get batch preview:', err))
      .finally(() => setPreviewLoading(false))
  }, [selectedIndices, outputDir, settings.layout, settings.authorFormat, JSON.stringify(settings.fieldOptions)])

  const getFieldOption = (field: string) => {
    return settings.fieldOptions.find(opt => opt.field === field)
  }

  const handleFieldMappingChange = async (field: string, value: string) => {
    try {
      await UpdateFieldMappingField(field, value)
      refreshSettings()
      // Refresh preview with updated field mapping
      const items = await GetBatchPreview(selectedIndicesArray, outputDir)
      setPreviewItems(items || [])
      if (onFieldMappingChange) {
        onFieldMappingChange()
      }
    } catch (err) {
      console.error('Failed to update field mapping:', err)
    }
  }

  // Group by book: author + series (or title when no series).
  // This correctly collapses individual track files into their parent book.
  const bookGroups = (() => {
    const map = new Map<string, { key: string; bookLabel: string; author: string; series?: string; items: main.PreviewItem[] }>()
    for (const item of previewItems) {
      const bookName = item.series || item.title
      const key = `${item.author}|||${bookName}`
      if (!map.has(key)) {
        map.set(key, { key, bookLabel: bookName, author: item.author, series: item.series || undefined, items: [] })
      }
      map.get(key)!.items.push(item)
    }
    return Array.from(map.values())
  })()

  const hasMultipleFilesPerBook = bookGroups.some(g => g.items.length > 1)

  const toggleGroup = (key: string) => {
    setExpandedGroups(prev => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  // Build path parts from a PreviewItem for ColoredPath component
  const getPathParts = (item: main.PreviewItem) => {
    const effectiveOutputDir = item.output_dir || outputDir || '/output'
    const relPath = item.to.startsWith(effectiveOutputDir)
      ? item.to.slice(effectiveOutputDir.length)
      : item.to
    const relParts = relPath.split('/').filter(p => p.length > 0)
    return [effectiveOutputDir, ...relParts]
  }

  // Generate equivalent CLI command
  const generateCliCommand = () => {
    const parts = ['audiobook-organizer']
    if (inputDir) parts.push(`--dir="${inputDir}"`)
    if (outputDir) parts.push(`--out="${outputDir}"`)
    if (settings.layout) parts.push(`--layout=${settings.layout}`)

    const titleField = getFieldOption('title')?.current
    if (titleField && titleField !== 'title') parts.push(`--title-field=${titleField}`)

    const seriesField = getFieldOption('series')?.current
    if (seriesField && seriesField !== 'series') parts.push(`--series-field=${seriesField}`)

    const authorField = getFieldOption('authors')?.current
    if (authorField && authorField !== 'authors') parts.push(`--author-fields=${authorField}`)

    const trackField = getFieldOption('track')?.current
    if (trackField && trackField !== 'track') parts.push(`--track-field=${trackField}`)

    return parts.join(' \\\n  ')
  }

  // Generate bash commands
  const generateCommands = () => {
    const dirs = new Set(previewItems.map(item => {
      const parts = getPathParts(item)
      return parts.slice(0, -1).join('/')
    }))
    const mkdirCommands = Array.from(dirs).map(dir => `mkdir -p "${dir}"`)
    const fileCommands = previewItems.map(item =>
      copyMode
        ? `cp "${item.from}" "${item.to}"`
        : `mv "${item.from}" "${item.to}"`
    )
    return [...mkdirCommands, '', ...fileCommands].join('\n')
  }

  return (
    <div className="flex flex-col min-h-0 flex-1 overflow-hidden">
      {/* Header */}
      <div className="border-b border-border p-3 bg-card">
        <div className="flex items-center justify-between">
          <div className="text-lg font-semibold">
            Ready to Organize —{' '}
            {hasMultipleFilesPerBook
              ? <>{bookGroups.length} {bookGroups.length === 1 ? 'book' : 'books'} <span className="text-sm font-normal text-muted-foreground">({previewItems.length} files)</span></>
              : <>{previewItems.length} {previewItems.length === 1 ? 'file' : 'files'}</>
            }
          </div>
          {hasMultipleFilesPerBook && (
            <div className="flex rounded border border-border overflow-hidden text-xs">
              <button
                onClick={() => setViewMode('books')}
                className={`px-3 py-1 transition-colors ${viewMode === 'books' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted/50'}`}
              >Books</button>
              <button
                onClick={() => setViewMode('files')}
                className={`px-3 py-1 transition-colors ${viewMode === 'files' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted/50'}`}
              >Files</button>
            </div>
          )}
        </div>
        <div className="flex items-center gap-4 mt-2 text-sm">
          {/* Copy mode checkbox */}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="copyMode"
              checked={copyMode}
              onChange={(e) => setCopyMode(e.target.checked)}
              className="h-4 w-4 rounded border-border accent-primary cursor-pointer"
            />
            <label htmlFor="copyMode" className="text-sm cursor-pointer flex items-center gap-1.5">
              <Copy className="h-4 w-4" />
              Copy files (leave originals in place)
            </label>
          </div>

          <div className="text-muted-foreground">|</div>

          {/* Layout selector */}
          <div className="flex items-center gap-2 text-xs">
            <span className="text-muted-foreground">Layout:</span>
            <select
              value={settings.layout}
              onChange={async (e) => {
                await UpdateLayout(e.target.value)
                refreshSettings()
              }}
              className="text-xs p-1 rounded border border-border bg-background"
            >
              {settings.layoutOptions.map(opt => (
                <option key={opt.name} value={opt.name}>{opt.name}</option>
              ))}
            </select>
          </div>

          {settings.renameConfig.enabled && (
            <>
              <div className="text-muted-foreground">|</div>
              <div className="text-xs text-muted-foreground">
                Rename: <span className="text-foreground">{settings.renameConfig.template}</span>
              </div>
            </>
          )}
        </div>

        {/* Field Mapping bar */}
        <div className="mt-3 pt-2 border-t border-border">
          <div className="flex items-center gap-3 text-xs">
            <span className="font-medium text-muted-foreground">Field Mapping:</span>

            {/* Title */}
            <div className="flex items-center gap-1">
              <span className="text-green-600 font-medium">Title:</span>
              <select
                value={getFieldOption('title')?.current || 'title'}
                onChange={(e) => handleFieldMappingChange('title', e.target.value)}
                className="text-xs p-0.5 rounded border border-border bg-background"
              >
                {getFieldOption('title')?.options?.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            </div>

            {/* Series */}
            <div className="flex items-center gap-1">
              <span className="text-cyan-600 font-medium">Series:</span>
              <select
                value={getFieldOption('series')?.current || 'series'}
                onChange={(e) => handleFieldMappingChange('series', e.target.value)}
                className="text-xs p-0.5 rounded border border-border bg-background"
              >
                {getFieldOption('series')?.options?.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            </div>

            {/* Author */}
            <div className="flex items-center gap-1">
              <span className="text-orange-600 font-medium">Author:</span>
              <select
                value={getFieldOption('authors')?.current || 'artist'}
                onChange={(e) => handleFieldMappingChange('authors', e.target.value)}
                className="text-xs p-0.5 rounded border border-border bg-background"
              >
                {getFieldOption('authors')?.options?.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            </div>

            {/* Track */}
            <div className="flex items-center gap-1">
              <span className="text-blue-600 font-medium">Track:</span>
              <select
                value={getFieldOption('track')?.current || 'track'}
                onChange={(e) => handleFieldMappingChange('track', e.target.value)}
                className="text-xs p-0.5 rounded border border-border bg-background"
              >
                {getFieldOption('track')?.options?.map((opt) => (
                  <option key={opt} value={opt}>{opt}</option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </div>

      {/* Preview list */}
      <div className="flex-1 overflow-y-auto p-3">
        {viewMode === 'books' ? (
          /* Books view — one row per book, expandable to see individual files */
          <div className="space-y-1">
            {bookGroups.map(group => {
              const expanded = expandedGroups.has(group.key)
              const firstItem = group.items[0]
              const firstParts = getPathParts(firstItem)
              // Show path up to (but not including) the filename
              const folderParts = group.items.length > 1 ? firstParts.slice(0, -1) : firstParts
              return (
                <div key={group.key} className="border border-border rounded overflow-hidden">
                  {/* Book header row */}
                  <button
                    onClick={() => group.items.length > 1 && toggleGroup(group.key)}
                    className={`w-full grid grid-cols-2 gap-2 text-[10px] p-1.5 bg-muted/10 transition-colors text-left ${group.items.length > 1 ? 'hover:bg-muted/30 cursor-pointer' : 'cursor-default'}`}
                  >
                    <div className="font-mono text-muted-foreground flex items-center gap-1 min-w-0">
                      {group.items.length > 1
                        ? (expanded ? <ChevronDown className="h-3 w-3 shrink-0" /> : <ChevronRight className="h-3 w-3 shrink-0" />)
                        : <span className="w-3 shrink-0" />
                      }
                      <span className="truncate">{firstItem.from}</span>
                      {group.items.length > 1 && (
                        <span className="ml-1 text-[9px] bg-muted px-1 py-0.5 rounded whitespace-nowrap shrink-0">{group.items.length} files</span>
                      )}
                    </div>
                    <div className="font-mono break-all">
                      <ColoredPath
                        parts={folderParts}
                        author={group.author}
                        series={group.series}
                        title={group.bookLabel}
                        filename={group.items.length === 1 ? firstItem.filename : ''}
                      />
                    </div>
                  </button>

                  {/* File rows — shown when group expanded */}
                  {expanded && (
                    <div className="divide-y divide-border/50 border-t border-border">
                      {group.items.map((item, idx) => (
                        <div key={idx} className="grid grid-cols-2 gap-2 text-[10px] px-4 py-1 bg-background/50">
                          <div className="font-mono text-muted-foreground truncate">{item.filename}</div>
                          <div className="font-mono">
                            <ColoredPath
                              parts={getPathParts(item)}
                              author={item.author}
                              series={item.series || undefined}
                              title={item.title}
                              filename={item.filename}
                            />
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        ) : (
          /* Files view — flat list, one row per file */
          <div className="space-y-1">
            <div className="grid grid-cols-2 gap-2 text-xs mb-1 font-medium text-muted-foreground px-1">
              <div>From</div>
              <div>To</div>
            </div>
            {previewItems.map((item, idx) => (
              <div key={idx} className="grid grid-cols-2 gap-2 text-[10px]">
                <div className="p-1.5 bg-muted/20 rounded font-mono break-all border border-border">{item.from}</div>
                <div className="p-1.5 font-mono break-all">
                  <ColoredPath
                    parts={getPathParts(item)}
                    author={item.author}
                    series={item.series || undefined}
                    title={item.title}
                    filename={item.filename}
                  />
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Warnings */}
        {(settings.renameConfig.enabled || previewItems.length > 10) && (
          <div className="mt-3 p-2 bg-yellow-500/10 border border-yellow-500/20 rounded text-xs">
            <div className="font-medium text-yellow-700 dark:text-yellow-400 mb-1">Notice</div>
            <ul className="text-muted-foreground space-y-0.5 ml-4 list-disc">
              {settings.renameConfig.enabled && <li>Files will be renamed</li>}
              <li>{new Set(previewItems.map(item => getPathParts(item).slice(0, -1).join('/'))).size} directories will be created</li>
              {!copyMode && <li>Original files will be moved (not copied)</li>}
            </ul>
          </div>
        )}
      </div>

      {/* Bash commands panel */}
      <div className="border-t border-border">
        <button
          onClick={() => setShowCommands(!showCommands)}
          className="w-full p-2 flex items-center gap-2 hover:bg-muted/50 transition-colors text-xs"
        >
          {showCommands ? (
            <ChevronDown className="h-3 w-3 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-3 w-3 text-muted-foreground" />
          )}
          <span className="font-medium">Show Commands</span>
        </button>

        {showCommands && (
          <div className="p-3 bg-muted/20 border-t border-border max-h-64 overflow-y-auto space-y-3">
            {/* CLI command */}
            <div>
              <div className="text-[10px] font-medium text-muted-foreground mb-1">CLI Command (equivalent)</div>
              <pre className="text-[10px] font-mono bg-background p-2 rounded border border-border overflow-x-auto">
                {generateCliCommand()}
              </pre>
              <button
                onClick={() => navigator.clipboard.writeText(generateCliCommand())}
                className="mt-1 text-xs px-2 py-1 bg-primary text-primary-foreground rounded hover:bg-primary/90"
              >
                Copy CLI Command
              </button>
            </div>

            {/* Bash mv/cp commands */}
            <div>
              <div className="text-[10px] font-medium text-muted-foreground mb-1">Bash Commands (individual files)</div>
              <pre className="text-[10px] font-mono bg-background p-2 rounded border border-border overflow-x-auto">
                {generateCommands()}
              </pre>
              <button
                onClick={() => navigator.clipboard.writeText(generateCommands())}
                className="mt-1 text-xs px-2 py-1 bg-primary text-primary-foreground rounded hover:bg-primary/90"
              >
                Copy Bash Commands
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Action buttons */}
      <div className="border-t border-border p-3 bg-card flex justify-between">
        <button
          onClick={onCancel}
          className="px-4 py-2 rounded border border-border hover:bg-muted transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={() => {
            setIsExecuting(true)
            onExecute(copyMode, previewItems.map(item => ({ from: item.from, to: item.to })))
          }}
          disabled={previewLoading || isExecuting || previewItems.length === 0}
          className="px-6 py-2 rounded bg-green-600 text-white hover:bg-green-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {previewLoading ? 'Loading preview…' : isExecuting ? 'Executing…' : 'Execute Organization →'}
        </button>
      </div>
    </div>
  )
}
