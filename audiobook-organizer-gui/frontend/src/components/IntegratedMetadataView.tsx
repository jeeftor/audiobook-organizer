import { useState, useEffect } from 'react'
import { organizer, main } from '../../wailsjs/go/models'
import { GetFieldMappingOptions } from '../../wailsjs/go/main/App'

interface IntegratedMetadataViewProps {
  book: organizer.Metadata | null
}

export function IntegratedMetadataView({ book }: IntegratedMetadataViewProps) {
  const [fieldOptions, setFieldOptions] = useState<main.FieldMappingOption[]>([])

  useEffect(() => {
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

  if (!book) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground p-4 text-sm">
        Select an audiobook to view metadata
      </div>
    )
  }

  // Get field mapping annotations
  const getFieldAnnotation = (fieldName: string): string | null => {
    const titleMapping = fieldOptions.find(opt => opt.field === 'title')
    const seriesMapping = fieldOptions.find(opt => opt.field === 'series')
    const authorsMapping = fieldOptions.find(opt => opt.field === 'authors')

    if (titleMapping?.current === fieldName) return '← Title'
    if (seriesMapping?.current === fieldName) return '← Series'
    if (authorsMapping?.current === fieldName) return '← Author'

    return null
  }

  // Collect all metadata fields
  const allFields: { key: string; value: any; annotation: string | null }[] = []

  if (book.raw_data) {
    Object.keys(book.raw_data)
      .sort()
      .forEach(key => {
        const value = book.raw_data![key]
        if (value !== null && value !== undefined && value !== '') {
          allFields.push({
            key,
            value,
            annotation: getFieldAnnotation(key)
          })
        }
      })
  }

  return (
    <div className="p-4">
      <div className="mb-3">
        <h3 className="text-sm font-semibold">All Metadata Fields</h3>
        <p className="text-xs text-muted-foreground">
          Fields marked with arrows are used in organization
        </p>
      </div>

      <div className="space-y-1 text-xs">
        {allFields.map(({ key, value, annotation }) => (
          <div
            key={key}
            className={`flex justify-between gap-2 p-1.5 rounded ${
              annotation ? 'bg-primary/10 border border-primary/20' : 'hover:bg-muted/50'
            }`}
          >
            <div className="flex items-center gap-2 flex-shrink-0">
              <span className={`font-medium ${annotation ? 'text-primary' : 'text-muted-foreground'}`}>
                {key}:
              </span>
              {annotation && (
                <span className="text-[10px] font-semibold text-primary bg-primary/20 px-1.5 py-0.5 rounded">
                  {annotation}
                </span>
              )}
            </div>
            <span className="font-mono text-right break-all">
              {typeof value === 'object' ? JSON.stringify(value) : String(value)}
            </span>
          </div>
        ))}
      </div>

      {allFields.length === 0 && (
        <div className="text-center text-muted-foreground text-xs py-4">
          No metadata available
        </div>
      )}
    </div>
  )
}
