import { useState, useEffect } from 'react'
import { GetFieldMappingOptions, GetCurrentFieldMapping, UpdateFieldMappingField } from '../../wailsjs/go/main/App'
import { organizer, main } from '../../wailsjs/go/models'
import { ChevronDown, ChevronRight } from 'lucide-react'

interface FieldMappingEditorProps {
  onMappingChange?: () => void
}

export function FieldMappingEditor({ onMappingChange }: FieldMappingEditorProps) {
  const [expanded, setExpanded] = useState(false)
  const [currentMapping, setCurrentMapping] = useState<organizer.FieldMapping | null>(null)
  const [options, setOptions] = useState<main.FieldMappingOption[]>([])

  useEffect(() => {
    loadMapping()
    loadOptions()
  }, [])

  const loadMapping = async () => {
    try {
      const mapping = await GetCurrentFieldMapping()
      setCurrentMapping(mapping as organizer.FieldMapping)
    } catch (err) {
      console.error('Failed to load field mapping:', err)
    }
  }

  const loadOptions = async () => {
    try {
      const opts = await GetFieldMappingOptions()
      setOptions(opts as main.FieldMappingOption[])
    } catch (err) {
      console.error('Failed to load field options:', err)
    }
  }

  const handleFieldChange = async (field: string, value: string) => {
    try {
      await UpdateFieldMappingField(field, value)
      await loadMapping()
      if (onMappingChange) {
        onMappingChange()
      }
    } catch (err) {
      console.error('Failed to update field:', err)
    }
  }

  if (!currentMapping) {
    return null
  }

  return (
    <div className="border border-border rounded-lg bg-card">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full p-3 flex items-center justify-between hover:bg-accent transition-colors"
      >
        <div className="flex items-center gap-2">
          {expanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
          <span className="text-sm font-medium">Field Mapping Configuration</span>
        </div>
        <span className="text-xs text-muted-foreground">
          Configure metadata sources
        </span>
      </button>

      {expanded && (
        <div className="p-3 border-t border-border space-y-3">
          <div className="text-xs text-muted-foreground mb-2">
            Choose which metadata fields to use for each property
          </div>

          {/* Render each field mapping option */}
          {options.map((option) => (
            <div key={option.field}>
              <label className="text-xs font-medium mb-1 block">
                {option.label}
              </label>
              <select
                value={option.current}
                onChange={(e) => handleFieldChange(option.field, e.target.value)}
                className="w-full text-xs p-2 rounded border border-border bg-background"
              >
                {option.options?.map((opt) => (
                  <option key={opt} value={opt}>
                    {opt}
                  </option>
                ))}
              </select>
              {option.description && (
                <div className="text-[10px] text-muted-foreground mt-1">
                  {option.description}
                </div>
              )}
            </div>
          ))}

          <div className="pt-2 border-t border-border">
            <div className="text-[10px] text-muted-foreground">
              💡 Field mapping determines which ID3 tags are read from audio files.
              For example, "Album" tag can map to either Series or Title.
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
