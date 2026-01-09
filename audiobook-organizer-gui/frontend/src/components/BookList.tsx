import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import {
  ScanDirectory,
  GetAvailableScanModes,
  UpdateScanMode,
  GetCurrentScanMode,
  GetFieldMappingPresets,
  GetFieldMappingOptions,
  UpdateFieldMapping,
  UpdateFieldMappingField,
  GetSampleMetadataPreviews
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
  const [showFieldMapping, setShowFieldMapping] = useState(false)
  const [fieldMappingPresets, setFieldMappingPresets] = useState<main.FieldMappingPreset[]>([])
  const [fieldMappingOptions, setFieldMappingOptions] = useState<main.FieldMappingOption[]>([])
  const [metadataPreview, setMetadataPreview] = useState<any>(null)

  useEffect(() => {
    // Load available scan modes
    GetAvailableScanModes().then(modes => setScanModes(modes))
    GetCurrentScanMode().then(mode => setCurrentMode(mode))
    // Load field mapping configuration
    GetFieldMappingPresets().then(presets => setFieldMappingPresets(presets))
    GetFieldMappingOptions().then(options => setFieldMappingOptions(options))
  }, [])

  const loadMetadataPreview = async () => {
    try {
      const previews = await GetSampleMetadataPreviews(inputDir)
      setMetadataPreview(previews?.[0] || null)
    } catch (err) {
      console.error('Failed to load metadata preview:', err)
    }
  }

  const scanBooks = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await ScanDirectory(inputDir)
      setBooks(result || [])
      const allIndices = new Set(result?.map((_, idx) => idx) || [])
      setSelected(allIndices)
    } catch (err) {
      setError(`Failed to scan directory: ${err}`)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    scanBooks()
  }, [inputDir])

  useEffect(() => {
    // Load metadata preview when field mapping is shown
    if (showFieldMapping && !metadataPreview) {
      loadMetadataPreview()
    }
  }, [showFieldMapping])

  const handleScanModeChange = async (modeName: string) => {
    try {
      await UpdateScanMode(modeName)
      setCurrentMode(modeName)
      // Re-scan with new mode
      await scanBooks()
    } catch (err) {
      setError(`Failed to update scan mode: ${err}`)
    }
  }

  const handlePresetSelect = async (preset: any) => {
    try {
      await UpdateFieldMapping(preset.mapping)
      // Reload field mapping options to show updated values
      const options = await GetFieldMappingOptions()
      setFieldMappingOptions(options)
      // Reload metadata preview with new mapping
      await loadMetadataPreview()
      // Re-scan with new field mapping
      await scanBooks()
    } catch (err) {
      setError(`Failed to apply preset: ${err}`)
    }
  }

  const handleFieldChange = async (field: string, value: string) => {
    try {
      await UpdateFieldMappingField(field, value)
      // Reload field mapping options to show updated values
      const options = await GetFieldMappingOptions()
      setFieldMappingOptions(options)
      // Reload metadata preview with new mapping
      await loadMetadataPreview()
      // Re-scan with new field mapping
      await scanBooks()
    } catch (err) {
      setError(`Failed to update field: ${err}`)
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
    <div className="min-h-screen p-4">
      <div className="max-w-7xl mx-auto space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Found {books.length} Audiobook{books.length !== 1 ? 's' : ''}</h1>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onBack}>Back</Button>
            <Button onClick={handleNext} disabled={selected.size === 0}>
              Next: Preview Changes ({selected.size})
            </Button>
          </div>
        </div>
          {/* Scanning Mode Selector */}
          <div className="p-4 bg-accent/50 rounded-lg space-y-3">
            <div className="text-sm font-medium">Scanning Mode</div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
              {scanModes.map((mode) => (
                <button
                  key={mode.name}
                  onClick={() => handleScanModeChange(mode.name)}
                  disabled={loading}
                  className={`p-3 text-left rounded-md border-2 transition-all ${
                    currentMode === mode.name
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:border-primary/50'
                  } ${loading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
                >
                  <div className="font-medium text-sm mb-1">
                    {mode.name === 'hierarchical' && '📁 Hierarchical'}
                    {mode.name === 'flat' && '📄 Flat'}
                    {mode.name === 'json' && '📋 Metadata.json'}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {mode.description}
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Field Mapping Configuration */}
          <div className="p-4 bg-accent/50 rounded-lg space-y-3">
            <button
              onClick={() => setShowFieldMapping(!showFieldMapping)}
              className="w-full flex items-center justify-between text-sm font-medium hover:text-primary transition-colors"
            >
              <span>⚙️ Field Mapping Configuration</span>
              <span className="text-xs">{showFieldMapping ? '▼' : '▶'}</span>
            </button>

            {showFieldMapping && (
              <div className="space-y-4 pt-2">
                {/* Metadata Preview - Show First */}
                {metadataPreview && (
                  <div className="p-3 bg-background border border-border rounded-lg">
                    <div className="text-xs font-medium mb-2">📋 Raw Metadata Fields</div>
                    <div className="text-[10px] text-muted-foreground mb-2">
                      File: {metadataPreview.filename} • Source: {metadataPreview.source_type}
                    </div>
                    <div className="space-y-1 max-h-48 overflow-y-auto">
                      {metadataPreview.raw_fields?.map((field: any, idx: number) => (
                        <div key={idx} className="flex items-start gap-2 text-xs font-mono">
                          <span className="text-muted-foreground min-w-[120px]">{field.key}:</span>
                          <span className="flex-1 truncate">{field.value}</span>
                          {field.indicator && (
                            <span className={`text-[10px] font-bold px-1.5 py-0.5 rounded ${
                              field.indicator === 'TITLE' ? 'bg-green-500/20 text-green-600' :
                              field.indicator === 'AUTHOR' ? 'bg-orange-500/20 text-orange-600' :
                              field.indicator === 'SERIES' ? 'bg-cyan-500/20 text-cyan-600' :
                              field.indicator === 'TRACK' ? 'bg-blue-500/20 text-blue-600' :
                              field.indicator === 'DISC' ? 'bg-purple-500/20 text-purple-600' :
                              'bg-gray-500/20 text-gray-600'
                            }`}>
                              ← {field.indicator}
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Field Mapping Configuration */}
                <div className="p-3 bg-background border border-border rounded-lg">
                  <div className="text-xs font-medium mb-3">🎯 Field Mapping</div>
                  <div className="space-y-4">
                    {fieldMappingOptions.map((option) => (
                      <div key={option.field}>
                        {option.field === 'authors' ? (
                          // Multi-select for Author fields
                          <div className="space-y-2">
                            <label className="text-xs font-medium">{option.label}</label>
                            <div className="text-[10px] text-muted-foreground mb-1">{option.description}</div>
                            <div className="grid grid-cols-2 gap-2 p-2 border border-border rounded">
                              {option.options.map((opt) => {
                                const currentFields = option.current.split(',').map(f => f.trim())
                                const isSelected = currentFields.includes(opt)
                                const priority = currentFields.indexOf(opt) + 1

                                return (
                                  <label key={opt} className="flex items-center gap-2 text-xs cursor-pointer hover:bg-accent/50 p-1 rounded">
                                    <input
                                      type="checkbox"
                                      checked={isSelected}
                                      onChange={(e) => {
                                        let newFields = [...currentFields.filter(f => f)]
                                        if (e.target.checked) {
                                          newFields.push(opt)
                                        } else {
                                          newFields = newFields.filter(f => f !== opt)
                                        }
                                        handleFieldChange('authors', newFields.join(','))
                                      }}
                                      disabled={loading}
                                      className="rounded"
                                    />
                                    <span className="flex-1">{opt}</span>
                                    {isSelected && (
                                      <span className="text-[10px] bg-primary/20 text-primary px-1 rounded">#{priority}</span>
                                    )}
                                  </label>
                                )
                              })}
                            </div>
                            <div className="text-[10px] text-muted-foreground">
                              Selected: {option.current || 'none'}
                            </div>
                          </div>
                        ) : (
                          // Simple dropdown for other fields
                          <div className="space-y-1">
                            <label className="text-xs font-medium">{option.label}</label>
                            <select
                              value={option.current}
                              onChange={(e) => handleFieldChange(option.field, e.target.value)}
                              disabled={loading}
                              className="w-full p-2 text-xs rounded border border-border bg-background disabled:opacity-50"
                            >
                              {option.options.map((opt) => (
                                <option key={opt} value={opt}>{opt}</option>
                              ))}
                            </select>
                            <div className="text-[10px] text-muted-foreground">{option.description}</div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                  <button
                    onClick={loadMetadataPreview}
                    className="mt-3 text-[10px] text-primary hover:underline"
                  >
                    🔄 Refresh Preview
                  </button>
                </div>
              </div>
            )}
          </div>

          {error && (
            <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
              {error}
            </div>
          )}

          {books.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">
              No audiobooks found in the selected directory
            </div>
          ) : (
            <>
              <div className="flex items-center justify-between py-2 border-b">
                <Button variant="outline" size="sm" onClick={toggleAll}>
                  {selected.size === books.length ? 'Deselect All' : 'Select All'}
                </Button>
                <span className="text-sm text-muted-foreground">
                  {selected.size} of {books.length} selected
                </span>
              </div>

              <div className="space-y-2 max-h-[60vh] overflow-y-auto">
                {books.map((book, idx) => (
                  <div
                    key={idx}
                    className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                      selected.has(idx)
                        ? 'bg-primary/5 border-primary'
                        : 'hover:bg-accent'
                    }`}
                    onClick={() => toggleSelection(idx)}
                  >
                    <div className="flex items-start gap-3">
                      <input
                        type="checkbox"
                        checked={selected.has(idx)}
                        onChange={() => toggleSelection(idx)}
                        className="mt-1"
                        onClick={(e) => e.stopPropagation()}
                      />
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">
                          {book.album || book.title || 'Unknown Title'}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {book.authors && book.authors.length > 0 ? (
                            <span>by {book.authors.join(', ')}</span>
                          ) : (
                            <span>Unknown Author</span>
                          )}
                          {book.series && book.series.length > 0 && (
                            <span className="ml-2">• {book.series.join(', ')}</span>
                          )}
                        </div>
                        {book.album && book.title && book.album !== book.title && (
                          <div className="text-xs text-muted-foreground mt-1">
                            TTrack: {book.title}
                          </div>
                        )}
                        <div className="text-xs text-muted-foreground mt-1 truncate">
                          {book.source_path}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}

      </div>
    </div>
  )
}
