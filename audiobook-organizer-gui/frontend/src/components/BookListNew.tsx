import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import {
  ScanDirectory,
  GetAvailableScanModes,
  UpdateScanMode,
  GetCurrentScanMode,
  GetSampleMetadataPreviews,
  GetAvailableLayouts,
  GetCurrentLayout,
  UpdateLayout,
  GetFieldMappingOptions,
  UpdateFieldMappingField
} from '../../wailsjs/go/main/App'
import { organizer, main } from '../../wailsjs/go/models'

interface BookListProps {
  inputDir: string
  outputDir: string
  onNext: (selectedIndices: number[]) => void
  onBack: () => void
}

export function BookList({ inputDir, outputDir, onNext, onBack }: BookListProps) {
  const [books, setBooks] = useState<organizer.Metadata[]>([])
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [scanModes, setScanModes] = useState<main.ScanMode[]>([])
  const [currentMode, setCurrentMode] = useState('')
  const [metadataPreviews, setMetadataPreviews] = useState<any[]>([])
  const [currentPreviewIndex, setCurrentPreviewIndex] = useState(0)
  const [layouts, setLayouts] = useState<any[]>([])
  const [currentLayout, setCurrentLayout] = useState('author-series-title')
  const [hasMetadataJsonFiles, setHasMetadataJsonFiles] = useState(true)
  const [showFieldMappingDialog, setShowFieldMappingDialog] = useState(false)
  const [fieldMappingOptions, setFieldMappingOptions] = useState<any[]>([])
  const [metadataPreviewOffset, setMetadataPreviewOffset] = useState(0)

  useEffect(() => {
    GetAvailableScanModes().then(modes => setScanModes(modes))
    GetCurrentScanMode().then(mode => setCurrentMode(mode))
    GetAvailableLayouts().then(layouts => setLayouts(layouts))
    GetCurrentLayout().then(layout => setCurrentLayout(layout))
  }, [])

  // Keyboard navigation - cycle by 1
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft') {
        setMetadataPreviewOffset(prev => Math.max(0, prev - 1))
      } else if (e.key === 'ArrowRight') {
        setMetadataPreviewOffset(prev => Math.min(books.length - 3, prev + 1))
      }
    }
    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [books.length])

  const scanBooks = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await ScanDirectory(inputDir)

      console.log('BookList: scanBooks got', result?.length || 0, 'books, currentMode =', currentMode)

      // Auto-fallback: if metadata.json mode (or initial mode) finds 0 books, switch to embedded mode
      if ((!result || result.length === 0) && (currentMode === 'metadata.json' || currentMode === '')) {
        console.log('No audiobooks found in metadata.json mode, switching to embedded (directory)')
        setHasMetadataJsonFiles(false)
        await UpdateScanMode('embedded (directory)')
        setCurrentMode('embedded (directory)')
        await new Promise(resolve => setTimeout(resolve, 100))
        const retryResult = await ScanDirectory(inputDir)
        console.log('BookList: retry scan got', retryResult?.length || 0, 'books')
        setBooks(retryResult || [])
        const allIndices = new Set(retryResult?.map((_, idx) => idx) || [])
        setSelected(allIndices)
      } else {
        setBooks(result || [])
        const allIndices = new Set(result?.map((_, idx) => idx) || [])
        setSelected(allIndices)
        // Track if metadata.json mode has files
        if (currentMode === 'metadata.json' && result && result.length > 0) {
          setHasMetadataJsonFiles(true)
        }
      }

      // Load metadata previews (up to 3)
      try {
        const previews = await GetSampleMetadataPreviews(inputDir)
        if (previews) {
          setMetadataPreviews(previews)
        }
      } catch (previewErr) {
        console.error('Failed to load metadata previews:', previewErr)
      }

      // Load field mapping options after scan
      try {
        const options = await GetFieldMappingOptions()
        setFieldMappingOptions(options)
      } catch (optErr) {
        console.error('Failed to load field mapping options:', optErr)
      }
    } catch (err) {
      setError(`Failed to scan directory: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    scanBooks()
  }, [inputDir])

  const handleScanModeChange = async (modeName: string) => {
    try {
      await UpdateScanMode(modeName)
      setCurrentMode(modeName)
      // Wait a bit for config to settle
      await new Promise(resolve => setTimeout(resolve, 100))
      await scanBooks()
    } catch (err) {
      setError(`Failed to update scan mode: ${err}`)
    }
  }

  const handleLayoutChange = async (layout: string) => {
    try {
      await UpdateLayout(layout)
      setCurrentLayout(layout)
    } catch (err) {
      setError(`Failed to update layout: ${err}`)
    }
  }

  const toggleSelection = (index: number) => {
    const newSelected = new Set(selected)
    if (newSelected.has(index)) {
      newSelected.delete(index)
    } else {
      newSelected.add(index)
    }
    setSelected(newSelected)
  }

  const toggleAll = () => {
    if (selected.size === books.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(books.map((_, idx) => idx)))
    }
  }

  const handleNext = () => {
    const selectedIndices = Array.from(selected).sort((a, b) => a - b)
    onNext(selectedIndices)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Card className="w-full max-w-4xl">
          <CardContent className="p-12 text-center">
            <div className="animate-pulse space-y-4">
              <div className="text-lg font-medium">Scanning for audiobooks...</div>
              <div className="text-sm text-muted-foreground">This may take a moment</div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-4 bg-background">
      <div className="max-w-7xl mx-auto space-y-4">
        {/* Header with Mode Selector */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-xl font-bold">Metadata Mode: {currentMode}</h1>
            <div className="flex gap-2">
              {scanModes.map((mode) => {
                const isDisabled = loading || (mode.name === 'metadata.json' && !hasMetadataJsonFiles)
                return (
                  <Button
                    key={mode.name}
                    variant={currentMode === mode.name ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => handleScanModeChange(mode.name)}
                    disabled={isDisabled}
                    className={isDisabled && mode.name === 'metadata.json' ? 'opacity-50 cursor-not-allowed' : ''}
                    title={mode.name === 'metadata.json' && !hasMetadataJsonFiles ? 'No metadata.json files found' : ''}
                  >
                    {mode.name}
                  </Button>
                )
              })}
            </div>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowFieldMappingDialog(true)}>
              ⚙️ Field Mapping
            </Button>
            <Button variant="outline" onClick={onBack}>Back</Button>
            <Button onClick={handleNext} disabled={selected.size === 0}>
              Next ({selected.size})
            </Button>
          </div>
        </div>

        {error && !books.length && (
          <div className="p-2 text-xs text-destructive bg-destructive/10 rounded">
            {error}
          </div>
        )}

        {/* Panel 1: Metadata Preview (Top) - 3 Columns */}
        <Card>
          <CardHeader className="py-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium">Metadata Preview ({metadataPreviews.length} books)</CardTitle>
              <div className="flex gap-1 items-center">
                {/* Multi-step navigation */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setMetadataPreviewOffset(Math.max(0, metadataPreviewOffset - 3))}
                  disabled={metadataPreviewOffset === 0}
                  className="h-7 w-8 p-0 text-xs"
                >
                  ≪
                </Button>
                {/* Single-step navigation (also by 3) */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setMetadataPreviewOffset(Math.max(0, metadataPreviewOffset - 3))}
                  disabled={metadataPreviewOffset === 0}
                  className="h-7 w-7 p-0"
                >
                  ‹
                </Button>
                <span className="text-xs text-muted-foreground px-2">
                  {metadataPreviewOffset + 1}-{Math.min(metadataPreviewOffset + 3, books.length)} of {books.length}
                </span>
                {/* Single-step navigation (also by 3) */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setMetadataPreviewOffset(Math.min(books.length - 3, metadataPreviewOffset + 3))}
                  disabled={metadataPreviewOffset + 3 >= books.length}
                  className="h-7 w-7 p-0"
                >
                  ›
                </Button>
                {/* Multi-step navigation */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setMetadataPreviewOffset(Math.min(books.length - 3, metadataPreviewOffset + 3))}
                  disabled={metadataPreviewOffset + 3 >= books.length}
                  className="h-7 w-8 p-0 text-xs"
                >
                  ≫
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent className="py-2">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {metadataPreviews.slice(metadataPreviewOffset, metadataPreviewOffset + 3).map((preview: any, idx: number) => (
                <div key={idx} className="border rounded p-2">
                  <div className="text-[10px] text-muted-foreground mb-2">
                    {preview.filename} • {preview.source_type}
                  </div>
                  <div className="space-y-0 max-h-48 overflow-y-auto font-mono text-[10px]">
                    {preview.raw_fields?.map((field: any, fieldIdx: number) => (
                      <div key={fieldIdx} className="flex items-center gap-1 py-0.5">
                        <span className="text-muted-foreground w-24 shrink-0">{field.key}:</span>
                        <span className="flex-1 truncate text-[9px]">{field.value}</span>
                        {field.indicator && (
                          <span className={`text-[8px] font-bold px-1 py-0.5 rounded shrink-0 ${
                            field.indicator === 'TITLE' ? 'bg-green-500/20 text-green-600' :
                            field.indicator === 'AUTHOR' ? 'bg-orange-500/20 text-orange-600' :
                            field.indicator === 'SERIES' ? 'bg-cyan-500/20 text-cyan-600' :
                            field.indicator === 'TRACK' ? 'bg-blue-500/20 text-blue-600' :
                            'bg-gray-500/20 text-gray-600'
                          }`}>
                            ← {field.indicator}
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
            {metadataPreviews.length === 0 && (
              <div className="text-xs text-muted-foreground py-4 text-center">
                No metadata preview available
              </div>
            )}
          </CardContent>
        </Card>

        {/* Panel 2: Input Files (Middle) */}
        <Card>
          <CardHeader className="py-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">Input Files ({books.length} files)</CardTitle>
              <Button variant="outline" size="sm" onClick={toggleAll} className="h-7 text-xs">
                {selected.size === books.length ? 'Deselect All' : 'Select All'}
              </Button>
            </div>
          </CardHeader>
          <CardContent className="py-2">
            <div className="space-y-0.5 max-h-64 overflow-y-auto">
              {books.map((book, idx) => (
                <label
                  key={idx}
                  className="flex items-center gap-2 px-2 py-0.5 hover:bg-accent rounded cursor-pointer text-[10px] font-mono"
                >
                  <input
                    type="checkbox"
                    checked={selected.has(idx)}
                    onChange={() => toggleSelection(idx)}
                    className="rounded shrink-0"
                  />
                  <div className="flex-1 min-w-0 truncate">
                    {book.source_path || 'Unknown path'}
                  </div>
                </label>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Panel 3: Output Files (Bottom) */}
        <Card>
          <CardHeader className="py-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">Output Files ({books.length} files)</CardTitle>
              <select
                value={currentLayout}
                onChange={(e) => handleLayoutChange(e.target.value)}
                className="text-xs border rounded px-2 py-1"
              >
                {layouts.map((layout) => (
                  <option key={layout.name} value={layout.name}>
                    {layout.name}
                  </option>
                ))}
              </select>
            </div>
          </CardHeader>
          <CardContent className="py-2">
            <div className="space-y-0.5 font-mono text-[10px] max-h-64 overflow-y-auto">
              {books.map((book, idx) => {
                // Get field mapping configuration
                const titleField = fieldMappingOptions.find(o => o.field === 'title')?.current || 'title'
                const seriesField = fieldMappingOptions.find(o => o.field === 'series')?.current || 'album'
                const authorFieldsStr = fieldMappingOptions.find(o => o.field === 'authors')?.current || 'artist'
                const authorFields = authorFieldsStr.split(',').map((f: string) => f.trim())

                // Get values from book properties based on configured fields
                const titleValue = book[titleField as keyof typeof book]
                const seriesValue = book[seriesField as keyof typeof book]

                // Try each author field in priority order
                let authorValue = null
                for (const field of authorFields) {
                  const val = book[field as keyof typeof book]
                  if (val) {
                    authorValue = val
                    break
                  }
                }

                // Fallback to authors array if available
                if (!authorValue && book.authors && book.authors.length > 0) {
                  authorValue = book.authors[0]
                }

                const author = authorValue ? String(authorValue) : 'Unknown Author'
                const series = seriesValue ? String(seriesValue) : 'Unknown Series'
                const title = titleValue ? String(titleValue) : 'Unknown Title'
                const filename = book.source_path?.split('/').pop() || 'file'

                return (
                  <div key={idx} className="flex items-center gap-1 px-2 py-0.5">
                    {currentLayout === 'author-series-title' && (
                      <>
                        <span className="text-orange-600">{author}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-cyan-600">{series}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-green-600">{title}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-muted-foreground">{filename}</span>
                      </>
                    )}
                    {currentLayout === 'author-series-title-number' && (
                      <>
                        <span className="text-orange-600">{author}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-cyan-600">{series}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-cyan-600">#1 - </span>
                        <span className="text-green-600">{title}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-muted-foreground">{filename}</span>
                      </>
                    )}
                    {currentLayout === 'author-series' && (
                      <>
                        <span className="text-orange-600">{author}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-cyan-600">{series}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-muted-foreground">{filename}</span>
                      </>
                    )}
                    {currentLayout === 'author-title' && (
                      <>
                        <span className="text-orange-600">{author}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-green-600">{title}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-muted-foreground">{filename}</span>
                      </>
                    )}
                    {currentLayout === 'author-only' && (
                      <>
                        <span className="text-orange-600">{author}</span>
                        <span className="text-muted-foreground">/</span>
                        <span className="text-muted-foreground">{filename}</span>
                      </>
                    )}
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Field Mapping Dialog - 3-Column Vertical TUI-style */}
      {showFieldMappingDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setShowFieldMappingDialog(false)}>
          <Card className="w-full max-w-[95vw] max-h-[90vh] overflow-hidden" onClick={(e) => e.stopPropagation()}>
            <CardHeader className="py-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Field Mapping Configuration</CardTitle>
                <Button variant="outline" size="sm" onClick={() => setShowFieldMappingDialog(false)}>
                  Close
                </Button>
              </div>
            </CardHeader>
            <CardContent className="p-3 overflow-y-auto max-h-[80vh]">
              <div className="grid grid-cols-3 gap-3">
                {metadataPreviews.slice(metadataPreviewOffset, metadataPreviewOffset + 3).map((preview: any, colIdx: number) => (
                  <div key={colIdx} className="space-y-3">
                    {/* Metadata Preview */}
                    <div className="border rounded p-2">
                      <div className="text-xs font-medium mb-1">Metadata Preview (#{colIdx + 1})</div>
                      <div className="text-[9px] text-muted-foreground mb-2">
                        {preview.filename}
                      </div>
                      <div className="space-y-0 font-mono text-[9px] max-h-48 overflow-y-auto">
                        {preview.raw_fields?.map((field: any, idx: number) => (
                          <div key={idx} className="flex items-center gap-1 py-0.5">
                            <span className="text-muted-foreground w-20 shrink-0 text-[8px]">{field.key}:</span>
                            <span className="flex-1 truncate text-[8px]">{field.value}</span>
                            {field.indicator && (
                              <span className={`text-[7px] font-bold px-1 rounded shrink-0 ${
                                field.indicator === 'TITLE' ? 'bg-green-500/20 text-green-600' :
                                field.indicator === 'AUTHOR' ? 'bg-orange-500/20 text-orange-600' :
                                field.indicator === 'SERIES' ? 'bg-cyan-500/20 text-cyan-600' :
                                field.indicator === 'TRACK' ? 'bg-blue-500/20 text-blue-600' :
                                'bg-gray-500/20 text-gray-600'
                              }`}>
                                ← {field.indicator}
                              </span>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Selection Options */}
                    <div className="border rounded p-2 space-y-2">
                      <div className="text-xs font-medium mb-1">Select Fields</div>
                      {fieldMappingOptions.map((option: any) => (
                        <div key={option.field} className="space-y-1">
                          <label className="text-[10px] font-medium text-cyan-600">
                            {option.field === 'title' ? 'Title' :
                             option.field === 'series' ? 'Series' :
                             option.field === 'authors' ? 'Author' :
                             option.field === 'track' ? 'Track' : 'Disc'}:
                          </label>
                          <div className="space-y-0.5">
                            {option.options.slice(0, 5).map((opt: string) => {
                              const fieldValue = preview.raw_fields?.find((f: any) => f.key === opt)?.value || ''
                              const isSelected = option.current === opt || (option.field === 'authors' && option.current.includes(opt))

                              // Skip fields that don't exist in this book's metadata
                              if (!fieldValue) return null

                              return (
                                <button
                                  key={opt}
                                  onClick={async () => {
                                    await UpdateFieldMappingField(option.field, opt)
                                    const updated = await GetFieldMappingOptions()
                                    setFieldMappingOptions(updated)
                                    const previews = await GetSampleMetadataPreviews(inputDir)
                                    if (previews) setMetadataPreviews(previews)
                                  }}
                                  className={`w-full text-left px-2 py-1 rounded text-[9px] font-mono hover:bg-accent ${
                                    isSelected ? 'bg-primary/10 border border-primary' : 'border border-border'
                                  }`}
                                >
                                  <div className="flex items-center gap-1">
                                    {isSelected && <span className="text-primary text-[10px]">→</span>}
                                    <span className="font-bold text-yellow-600">{opt}:</span>
                                    <span className="truncate text-[8px]">{fieldValue}</span>
                                  </div>
                                </button>
                              )
                            })}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
