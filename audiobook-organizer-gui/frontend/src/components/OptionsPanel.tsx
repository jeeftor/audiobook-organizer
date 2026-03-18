import { useState, useEffect } from 'react'
import { UpdateScanMode, GetCurrentScanMode, UpdateFieldMapping, GetFieldMappingPresets, UpdateLayout, GetCurrentLayout, UpdateAuthorFormat, GetCurrentAuthorFormat, GetFieldMappingOptions, UpdateFieldMappingField } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'

interface OptionsPanelProps {
  outputDir: string
  selectedBook?: any
  onScanModeChange?: () => void
  onFieldMappingChange?: () => void
  onLayoutChange?: () => void
}

export function OptionsPanel({ outputDir, selectedBook, onScanModeChange, onFieldMappingChange, onLayoutChange }: OptionsPanelProps) {
  const [scanMode, setScanMode] = useState('embedded (directory)')
  const [fieldMapping, setFieldMapping] = useState('Audio')
  const [layout, setLayout] = useState('author-series-title')
  const [authorFormat, setAuthorFormat] = useState('First Last')
  const [fieldOptions, setFieldOptions] = useState<main.FieldMappingOption[]>([])

  useEffect(() => {
    GetCurrentScanMode().then(mode => {
      setScanMode(mode)
    }).catch(err => {
      console.error('Failed to get current scan mode:', err)
    })

    GetCurrentLayout().then(layout => {
      setLayout(layout)
    }).catch(err => {
      console.error('Failed to get current layout:', err)
    })

    GetCurrentAuthorFormat().then(format => {
      setAuthorFormat(format)
    }).catch(err => {
      console.error('Failed to get current author format:', err)
    })

    loadFieldOptions()
  }, [])

  const loadFieldOptions = async () => {
    try {
      const options = await GetFieldMappingOptions()
      setFieldOptions(options)
    } catch (err) {
      console.error('Failed to load field options:', err)
    }
  }

  const handleFieldMappingChange = async (field: string, value: string) => {
    try {
      await UpdateFieldMappingField(field, value)
      await loadFieldOptions()
      if (onFieldMappingChange) {
        onFieldMappingChange()
      }
    } catch (err) {
      console.error('Failed to update field mapping:', err)
    }
  }

  const getFieldOption = (field: string) => {
    return fieldOptions.find(opt => opt.field === field)
  }

  const handleScanModeChange = async (mode: string) => {
    try {
      await UpdateScanMode(mode)
      setScanMode(mode)
      console.log('Scan mode updated to:', mode)

      // Trigger re-scan with new mode
      if (onScanModeChange) {
        onScanModeChange()
      }
    } catch (err) {
      console.error('Failed to update scan mode:', err)
    }
  }

  const handleFieldMappingPresetChange = async (preset: string) => {
    try {
      const presets = await GetFieldMappingPresets()
      const selected = presets.find(p => p.name === preset)
      if (selected) {
        await UpdateFieldMapping(selected.mapping)
        setFieldMapping(preset)
        console.log('Field mapping updated to:', preset)

        // Trigger re-scan with new mapping
        if (onFieldMappingChange) {
          onFieldMappingChange()
        }
      }
    } catch (err) {
      console.error('Failed to update field mapping:', err)
    }
  }

  const handleLayoutChange = async (newLayout: string) => {
    try {
      await UpdateLayout(newLayout)
      setLayout(newLayout)
      console.log('Layout updated to:', newLayout)

      // Trigger preview update
      if (onLayoutChange) {
        onLayoutChange()
      }
    } catch (err) {
      console.error('Failed to update layout:', err)
    }
  }

  const handleAuthorFormatChange = async (format: string) => {
    try {
      await UpdateAuthorFormat(format)
      setAuthorFormat(format)
      console.log('Author format updated to:', format)

      // Trigger preview update
      if (onLayoutChange) {
        onLayoutChange()
      }
    } catch (err) {
      console.error('Failed to update author format:', err)
    }
  }

  return (
    <div className="border-b border-border bg-muted/30">
      <div className="p-3 flex items-center gap-6 max-w-6xl mx-auto">
        {/* Scan Mode */}
        <div className="flex items-center gap-3">
          <span className="text-xs font-medium">Scan Mode:</span>
          <div className="flex gap-2">
            <label className="flex items-center gap-1.5 text-xs cursor-pointer">
              <input
                type="radio"
                name="scanMode"
                checked={scanMode === 'embedded (directory)'}
                onChange={() => handleScanModeChange('embedded (directory)')}
              />
              <span>📁 Hierarchical</span>
            </label>
            <label className="flex items-center gap-1.5 text-xs cursor-pointer">
              <input
                type="radio"
                name="scanMode"
                checked={scanMode === 'embedded (file)'}
                onChange={() => handleScanModeChange('embedded (file)')}
              />
              <span>📄 Flat</span>
            </label>
            <label className="flex items-center gap-1.5 text-xs cursor-pointer">
              <input
                type="radio"
                name="scanMode"
                checked={scanMode === 'metadata.json'}
                onChange={() => handleScanModeChange('metadata.json')}
              />
              <span>📋 JSON</span>
            </label>
          </div>
        </div>

        {/* Field Mapping Preset */}
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium">Preset:</span>
          <select
            className="text-xs p-1 rounded border border-border bg-background"
            value={fieldMapping}
            onChange={(e) => handleFieldMappingPresetChange(e.target.value)}
          >
            <option>Audio</option>
            <option>EPUB</option>
            <option>Default</option>
          </select>
        </div>

        {/* Layout */}
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium">Layout:</span>
          <select
            className="text-xs p-1 rounded border border-border bg-background"
            value={layout}
            onChange={(e) => handleLayoutChange(e.target.value)}
          >
            <option value="author-series-title">author-series-title</option>
            <option value="author-title">author-title</option>
            <option value="series-title">series-title</option>
            <option value="author-only">author-only</option>
          </select>
        </div>

        {/* Author Format */}
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium">Author:</span>
          <select
            className="text-xs p-1 rounded border border-border bg-background"
            value={authorFormat}
            onChange={(e) => handleAuthorFormatChange(e.target.value)}
          >
            <option value="preserve">Preserve</option>
            <option value="first-last">First Last</option>
            <option value="last-first">Last, First</option>
          </select>
        </div>
      </div>

      {/* Field Mapping - Compact horizontal row */}
      <div className="px-3 pb-2 max-w-6xl mx-auto">
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
              {getFieldOption('title')?.options?.map((opt) => {
                const preview = String(selectedBook?.raw_data?.[opt] || '')
                return (
                  <option key={opt} value={opt}>
                    {opt}{preview ? `: ${preview.substring(0, 30)}${preview.length > 30 ? '...' : ''}` : ''}
                  </option>
                )
              })}
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
              {getFieldOption('series')?.options?.map((opt) => {
                const preview = String(selectedBook?.raw_data?.[opt] || '')
                return (
                  <option key={opt} value={opt}>
                    {opt}{preview ? `: ${preview.substring(0, 30)}${preview.length > 30 ? '...' : ''}` : ''}
                  </option>
                )
              })}
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
              {getFieldOption('authors')?.options?.map((opt) => {
                const preview = String(selectedBook?.raw_data?.[opt] || '')
                return (
                  <option key={opt} value={opt}>
                    {opt}{preview ? `: ${preview.substring(0, 30)}${preview.length > 30 ? '...' : ''}` : ''}
                  </option>
                )
              })}
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
              {getFieldOption('track')?.options?.map((opt) => {
                const preview = selectedBook?.raw_data?.[opt] || ''
                return (
                  <option key={opt} value={opt}>
                    {opt}{preview ? `: ${preview}` : ''}
                  </option>
                )
              })}
            </select>
          </div>
        </div>
      </div>
    </div>
  )
}
