import { useState, useEffect } from 'react'
import { organizer } from '../../wailsjs/go/models'
import { GetRenameConfig, UpdateRenameConfig, PreviewRename, GetRenamePresets, GetAvailableTemplateFields } from '../../wailsjs/go/main/App'

interface RenameTemplateBuilderProps {
  book: organizer.Metadata | null
}

interface RenameConfig {
  enabled: boolean
  template: string
  preset: string
  separator: string
  author_format: string
  replace_spaces: boolean
  space_char: string
}

export function RenameTemplateBuilder({ book }: RenameTemplateBuilderProps) {
  const [config, setConfig] = useState<RenameConfig>({
    enabled: false,
    template: '{track} - {title}',
    preset: 'Track - Title',
    separator: '-',
    author_format: 'first-last',
    replace_spaces: false,
    space_char: '.',
  })
  const [preview, setPreview] = useState<string>('')
  const [presets, setPresets] = useState<Array<{name: string, template: string}>>([])
  const [fields, setFields] = useState<Array<{name: string, description: string, example: string}>>([])

  useEffect(() => {
    loadConfig()
    loadPresets()
    loadFields()
  }, [])

  useEffect(() => {
    if (book && config.enabled && config.template) {
      updatePreview()
    } else {
      setPreview('')
    }
  }, [book, config.enabled, config.template, config.separator, config.author_format, config.replace_spaces, config.space_char])

  const loadConfig = async () => {
    try {
      const cfg = await GetRenameConfig()
      if (cfg) {
        setConfig(cfg as RenameConfig)
      }
    } catch (err) {
      console.error('Failed to load rename config:', err)
    }
  }

  const loadPresets = async () => {
    try {
      const p = await GetRenamePresets()
      setPresets(p as Array<{name: string, template: string}>)
    } catch (err) {
      console.error('Failed to load presets:', err)
    }
  }

  const loadFields = async () => {
    try {
      const f = await GetAvailableTemplateFields()
      setFields(f as Array<{name: string, description: string, example: string}>)
    } catch (err) {
      console.error('Failed to load fields:', err)
    }
  }

  const updatePreview = async () => {
    if (!book) return
    try {
      const result = await PreviewRename(book)
      setPreview(result || '')
    } catch (err) {
      setPreview('Error: ' + String(err))
    }
  }

  const handleConfigChange = async (updates: Partial<RenameConfig>) => {
    const newConfig = { ...config, ...updates }
    setConfig(newConfig)
    try {
      await UpdateRenameConfig(newConfig)
    } catch (err) {
      console.error('Failed to update rename config:', err)
    }
  }

  const handlePresetChange = (presetName: string) => {
    const preset = presets.find(p => p.name === presetName)
    if (preset) {
      handleConfigChange({
        preset: presetName,
        template: preset.template || config.template,
      })
    }
  }

  const toggleField = (fieldName: string) => {
    const fieldPattern = `{${fieldName}}`

    // Check if field already exists in template
    if (config.template.includes(fieldPattern)) {
      // Remove the field and clean up separators
      let newTemplate = config.template

      // Remove the field
      newTemplate = newTemplate.replace(fieldPattern, '')

      // Clean up multiple separators or leading/trailing separators
      const sepPattern = new RegExp(`\\s*${config.separator.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*`, 'g')
      newTemplate = newTemplate.replace(sepPattern, ` ${config.separator} `)

      // Clean up leading/trailing separators
      newTemplate = newTemplate.replace(new RegExp(`^\\s*${config.separator.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*`), '')
      newTemplate = newTemplate.replace(new RegExp(`\\s*${config.separator.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*$`), '')

      // Clean up double separators
      const doubleSep = ` ${config.separator} ${config.separator} `
      while (newTemplate.includes(doubleSep)) {
        newTemplate = newTemplate.replace(doubleSep, ` ${config.separator} `)
      }

      handleConfigChange({ template: newTemplate.trim(), preset: 'Custom' })
    } else {
      // Add the field
      const newTemplate = config.template + (config.template ? ` ${config.separator} ` : '') + fieldPattern
      handleConfigChange({ template: newTemplate, preset: 'Custom' })
    }
  }

  // Render color-coded preview by matching template fields to their values
  const renderColorCodedPreview = (preview: string, template: string, book: any) => {
    // Extract field names from template
    const fieldRegex = /\{([^}]+)\}/g
    const fields: string[] = []
    let match
    while ((match = fieldRegex.exec(template)) !== null) {
      fields.push(match[1])
    }

    // Build a map of field values
    const fieldValues: Record<string, string> = {
      track: book.track_number ? String(book.track_number).padStart(2, '0') : '',
      title: book.title || '',
      author: book.authors?.[0] || '',
      authors: book.authors?.join(', ') || '',
      series: book.series?.[0] || '',
      album: book.album || '',
    }

    // Color map for fields
    const colorMap: Record<string, string> = {
      track: 'bg-blue-600/20 text-blue-600',
      title: 'bg-green-600/20 text-green-600',
      author: 'bg-orange-600/20 text-orange-600',
      authors: 'bg-orange-600/20 text-orange-600',
      series: 'bg-cyan-600/20 text-cyan-600',
      album: 'bg-cyan-600/20 text-cyan-600',
    }

    // Try to match parts of the preview to field values
    let result: JSX.Element[] = []
    let remaining = preview
    let key = 0

    for (const field of fields) {
      const value = fieldValues[field]
      if (!value) continue

      const idx = remaining.indexOf(value)
      if (idx !== -1) {
        // Add text before the field value
        if (idx > 0) {
          result.push(
            <span key={key++} className="text-muted-foreground">
              {remaining.substring(0, idx)}
            </span>
          )
        }
        // Add the colored field value
        result.push(
          <span key={key++} className={`${colorMap[field] || ''} px-1 rounded`}>
            {value}
          </span>
        )
        remaining = remaining.substring(idx + value.length)
      }
    }

    // Add any remaining text
    if (remaining) {
      result.push(
        <span key={key++} className="text-muted-foreground">
          {remaining}
        </span>
      )
    }

    return result.length > 0 ? <>{result}</> : preview
  }

  if (!book) {
    return null
  }

  return (
    <div className="border-t border-border pt-3 mt-3">
      <div className="flex items-center gap-2 mb-3">
        <span className="text-xs font-medium">📝 File Rename Template</span>
      </div>

      {/* Enable checkbox */}
      <label className="flex items-center gap-2 text-xs mb-3 cursor-pointer">
        <input
          type="checkbox"
          checked={config.enabled}
          onChange={(e) => handleConfigChange({ enabled: e.target.checked })}
          className="rounded"
        />
        <span>Enable file renaming</span>
      </label>

      {config.enabled && (
        <div className="space-y-3">
          {/* Preset selector */}
          <div>
            <label className="text-xs font-medium text-muted-foreground mb-1 block">Preset</label>
            <select
              value={config.preset}
              onChange={(e) => handlePresetChange(e.target.value)}
              className="w-full text-xs p-1.5 rounded border border-border bg-background"
            >
              {presets.map((preset) => (
                <option key={preset.name} value={preset.name}>
                  {preset.name}
                </option>
              ))}
            </select>
          </div>

          {/* Template input */}
          <div>
            <label className="text-xs font-medium text-muted-foreground mb-1 block">Template</label>
            <input
              type="text"
              value={config.template}
              onChange={(e) => handleConfigChange({ template: e.target.value, preset: 'Custom' })}
              className="w-full text-xs p-1.5 rounded border border-border bg-background font-mono"
              placeholder="{track} - {title}"
            />
          </div>

          {/* Available fields */}
          <div>
            <label className="text-xs font-medium text-muted-foreground mb-1 block">
              Available Fields (click to insert)
            </label>
            <div className="flex flex-wrap gap-1">
              {fields.map((field) => {
                // Determine color based on field name
                let colorClass = 'bg-primary/10 hover:bg-primary/20 text-primary border-primary/20'
                if (field.name === 'track') {
                  colorClass = 'bg-blue-600/20 hover:bg-blue-600/30 text-blue-600 border-blue-600/30'
                } else if (field.name === 'title') {
                  colorClass = 'bg-green-600/20 hover:bg-green-600/30 text-green-600 border-green-600/30'
                } else if (field.name === 'author' || field.name === 'authors') {
                  colorClass = 'bg-orange-600/20 hover:bg-orange-600/30 text-orange-600 border-orange-600/30'
                } else if (field.name === 'series' || field.name === 'series_full' || field.name === 'series_number' || field.name === 'album') {
                  colorClass = 'bg-cyan-600/20 hover:bg-cyan-600/30 text-cyan-600 border-cyan-600/30'
                }

                return (
                  <button
                    key={field.name}
                    onClick={() => toggleField(field.name)}
                    className={`text-[10px] px-2 py-1 rounded border ${colorClass} ${config.template.includes(`{${field.name}}`) ? 'ring-2 ring-offset-1 ring-current' : ''}`}
                    title={`${field.description}\nExample: ${field.example}${field.name === 'author' ? '\n\nNote: Maps from \'artist\' metadata field' : ''}\n\nClick to ${config.template.includes(`{${field.name}}`) ? 'remove from' : 'add to'} template`}
                  >
                    {field.name}
                  </button>
                )
              })}
            </div>
          </div>

          {/* Options row */}
          <div className="grid grid-cols-2 gap-2">
            {/* Separator */}
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">Separator</label>
              <select
                value={config.separator}
                onChange={(e) => handleConfigChange({ separator: e.target.value })}
                className="w-full text-xs p-1.5 rounded border border-border bg-background"
              >
                <option value="-">- (dash)</option>
                <option value="_">_ (underscore)</option>
                <option value=".">. (dot)</option>
                <option value=" ">(space)</option>
              </select>
            </div>

            {/* Author format */}
            <div>
              <label className="text-xs font-medium text-muted-foreground mb-1 block">Author Format</label>
              <select
                value={config.author_format}
                onChange={(e) => handleConfigChange({ author_format: e.target.value })}
                className="w-full text-xs p-1.5 rounded border border-border bg-background"
              >
                <option value="first-last">First Last</option>
                <option value="last-first">Last, First</option>
                <option value="preserve">Preserve</option>
              </select>
            </div>
          </div>

          {/* Replace spaces */}
          <div className="flex items-center gap-2">
            <label className="flex items-center gap-2 text-xs cursor-pointer">
              <input
                type="checkbox"
                checked={config.replace_spaces}
                onChange={(e) => handleConfigChange({ replace_spaces: e.target.checked })}
                className="rounded"
              />
              <span>Replace spaces with:</span>
            </label>
            <select
              value={config.space_char}
              onChange={(e) => handleConfigChange({ space_char: e.target.value })}
              disabled={!config.replace_spaces}
              className="text-xs p-1 rounded border border-border bg-background disabled:opacity-50"
            >
              <option value=".">. (dot)</option>
              <option value="_">_ (underscore)</option>
              <option value="-">- (dash)</option>
            </select>
          </div>

          {/* Preview - show renamed filename with color-coded components */}
          {preview && (
            <div className="border-t border-border pt-2 mt-2">
              <div className="text-xs font-medium text-muted-foreground mb-1">Renamed Filename</div>
              <div className="p-2 bg-muted/20 border border-border rounded text-xs font-mono break-all">
                {renderColorCodedPreview(preview, config.template, book)}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
