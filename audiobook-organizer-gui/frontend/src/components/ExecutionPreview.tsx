import { useState, useEffect } from 'react'
import { organizer } from '../../wailsjs/go/models'
import { main } from '../../wailsjs/go/models'
import { ChevronDown, ChevronRight, Copy } from 'lucide-react'
import {
  GetCurrentLayout,
  GetCurrentAuthorFormat,
  GetRenameConfig,
  GetFieldMappingOptions,
  UpdateFieldMappingField,
} from '../../wailsjs/go/main/App'
import { buildOutputParts } from '../utils/pathUtils'
import { ColoredPath } from './ColoredPath'

interface ExecutionPreviewProps {
  books: organizer.Metadata[]
  selectedIndices: Set<number>
  outputDir: string
  onExecute: (copyMode: boolean, operations: Array<{from: string, to: string}>) => void
  onCancel: () => void
  onFieldMappingChange?: () => void
}

export function ExecutionPreview({
  books,
  selectedIndices,
  outputDir,
  onExecute,
  onCancel,
  onFieldMappingChange,
}: ExecutionPreviewProps) {
  const [copyMode, setCopyMode] = useState(false)
  const [showCommands, setShowCommands] = useState(false)
  const [layout, setLayout] = useState('author-series-title')
  const [authorFormat, setAuthorFormat] = useState('preserve')
  const [renameEnabled, setRenameEnabled] = useState(false)
  const [renameTemplate, setRenameTemplate] = useState('')
  const [fieldOptions, setFieldOptions] = useState<main.FieldMappingOption[]>([])

  // Load layout, authorFormat, renameConfig, and field mapping options from backend
  useEffect(() => {
    const loadSettings = () => {
      GetCurrentLayout()
        .then(l => setLayout(l))
        .catch(err => console.error('Failed to get layout:', err))

      GetCurrentAuthorFormat()
        .then(f => setAuthorFormat(f))
        .catch(err => console.error('Failed to get author format:', err))

      GetRenameConfig()
        .then(cfg => {
          setRenameEnabled(cfg.enabled)
          setRenameTemplate(cfg.template)
        })
        .catch(err => console.error('Failed to get rename config:', err))

      GetFieldMappingOptions()
        .then(opts => setFieldOptions(opts))
        .catch(err => console.error('Failed to get field mapping options:', err))
    }

    loadSettings()
    const interval = setInterval(loadSettings, 500)
    return () => clearInterval(interval)
  }, [])

  const selectedBooks = books.filter((_, idx) => selectedIndices.has(idx))

  const getFieldOption = (field: string) => {
    return fieldOptions.find(opt => opt.field === field)
  }

  const handleFieldMappingChange = async (field: string, value: string) => {
    try {
      await UpdateFieldMappingField(field, value)
      // Reload options immediately
      const opts = await GetFieldMappingOptions()
      setFieldOptions(opts)
      if (onFieldMappingChange) {
        onFieldMappingChange()
      }
    } catch (err) {
      console.error('Failed to update field mapping:', err)
    }
  }

  // Build operations using shared utility
  const operations = selectedBooks.map((book) => {
    const { parts, author, series, title, filename } = buildOutputParts(
      book,
      outputDir,
      layout,
      authorFormat
    )
    // Use renamed filename if enabled, otherwise original
    const outputFilename = renameEnabled ? `${title}.mp3` : filename
    // Replace the last part (filename) with outputFilename
    const finalParts = [...parts.slice(0, -1), outputFilename]
    return {
      from: book.source_path,
      to: finalParts.join('/'),
      directory: finalParts.slice(0, -1).join('/'),
      parts: finalParts,
      author,
      series,
      title,
      filename: outputFilename,
    }
  })

  // Generate bash commands
  const generateCommands = () => {
    const dirs = new Set(operations.map(op => op.directory))
    const mkdirCommands = Array.from(dirs).map(dir => `mkdir -p "${dir}"`)
    const fileCommands = operations.map(op =>
      copyMode
        ? `cp "${op.from}" "${op.to}"`
        : `mv "${op.from}" "${op.to}"`
    )
    return [...mkdirCommands, '', ...fileCommands].join('\n')
  }

  return (
    <div className="flex flex-col min-h-0 flex-1 overflow-hidden">
      {/* Header */}
      <div className="border-b border-border p-3 bg-card">
        <div className="text-lg font-semibold">Ready to Organize ({selectedBooks.length} files)</div>
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

          {/* Settings summary */}
          <div className="text-xs text-muted-foreground">
            Layout: <span className="text-foreground">{layout}</span>
            {renameEnabled && (
              <> • Rename: <span className="text-foreground">{renameTemplate}</span></>
            )}
          </div>
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
        <div className="grid grid-cols-2 gap-2 text-xs mb-2 font-medium text-muted-foreground">
          <div>From</div>
          <div>To</div>
        </div>

        <div className="space-y-1">
          {operations.map((op, idx) => (
            <div key={idx} className="grid grid-cols-2 gap-2 text-[10px]">
              <div className="p-1.5 bg-muted/20 rounded font-mono break-all border border-border">
                {op.from}
              </div>
              <div className="p-1.5 bg-green-500/10 rounded font-mono break-all border border-green-500/20">
                <ColoredPath
                  parts={op.parts}
                  author={op.author}
                  series={op.series}
                  title={op.title}
                  filename={op.filename}
                />
              </div>
            </div>
          ))}
        </div>

        {/* Warnings */}
        {(renameEnabled || operations.length > 10) && (
          <div className="mt-3 p-2 bg-yellow-500/10 border border-yellow-500/20 rounded text-xs">
            <div className="font-medium text-yellow-700 dark:text-yellow-400 mb-1">Notice</div>
            <ul className="text-muted-foreground space-y-0.5 ml-4 list-disc">
              {renameEnabled && <li>Files will be renamed</li>}
              <li>{new Set(operations.map(op => op.directory)).size} directories will be created</li>
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
          <span className="font-medium">Show Bash Commands</span>
        </button>

        {showCommands && (
          <div className="p-3 bg-muted/20 border-t border-border max-h-48 overflow-y-auto">
            <pre className="text-[10px] font-mono bg-background p-2 rounded border border-border overflow-x-auto">
              {generateCommands()}
            </pre>
            <button
              onClick={() => {
                const commands = generateCommands()
                // navigator.clipboard requires HTTPS; use execCommand fallback for Wails
                const ta = document.createElement('textarea')
                ta.value = commands
                ta.style.position = 'fixed'
                ta.style.opacity = '0'
                document.body.appendChild(ta)
                ta.focus()
                ta.select()
                document.execCommand('copy')
                document.body.removeChild(ta)
              }}
              className="mt-2 text-xs px-2 py-1 bg-primary text-primary-foreground rounded hover:bg-primary/90"
            >
              Copy to Clipboard
            </button>
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
          onClick={() => onExecute(copyMode, operations.map(op => ({ from: op.from, to: op.to })))}
          className="px-6 py-2 rounded bg-green-600 text-white hover:bg-green-700 transition-colors font-medium"
        >
          Execute Organization →
        </button>
      </div>
    </div>
  )
}
