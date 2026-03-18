import { useState, useEffect } from 'react'
import { organizer, main } from '../../wailsjs/go/models'
import { GetFieldMappingOptions, UpdateFieldMappingField } from '../../wailsjs/go/main/App'
import { BatchPreview } from './BatchPreview'

interface MetadataEditorProps {
  book: organizer.Metadata | null
  bookIndex: number | null
  books: organizer.Metadata[]
  selectedIndices: Set<number>
  outputDir: string
  onFieldMappingChange?: () => void
}

export function MetadataEditor({ book, bookIndex, books, selectedIndices, outputDir, onFieldMappingChange }: MetadataEditorProps) {
  const [fieldOptions, setFieldOptions] = useState<main.FieldMappingOption[]>([])

  useEffect(() => {
    loadFieldOptions()
  }, [])

  // Reload field options when book changes (after field mapping update)
  useEffect(() => {
    if (book) {
      loadFieldOptions()
    }
  }, [book?.source_path])

  // Log book metadata for debugging
  useEffect(() => {
    if (book) {
      console.log('═══════════════════════════════════════════════════════')
      console.log('[MetadataEditor] SELECTED BOOK METADATA:')
      console.log('═══════════════════════════════════════════════════════')
      console.log('Source Path:', book.source_path)
      console.log('Source Type:', book.source_type)
      console.log('')
      console.log('MAPPED VALUES (after field mapping applied):')
      console.log('  Title:', book.title || '(empty)')
      console.log('  Album:', book.album || '(empty)')
      console.log('  Authors:', book.authors?.join(', ') || '(empty)')
      console.log('  Series:', book.series?.join(', ') || '(empty)')
      console.log('  Track Number:', book.track_number || '(empty)')
      console.log('  Track Title:', book.track_title || '(empty)')
      console.log('')
      console.log('RAW METADATA FIELDS (from file tags):')
      if (book.raw_data) {
        Object.keys(book.raw_data).sort().forEach(key => {
          const value = book.raw_data![key]
          console.log(`  ${key}:`, value || '(empty)')
        })
      } else {
        console.log('  (no raw_data available)')
      }
      console.log('═══════════════════════════════════════════════════════')
    }
  }, [book?.source_path])

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
      console.log(`[MetadataEditor] Updating field mapping: ${field} = ${value}`)
      await UpdateFieldMappingField(field, value)
      console.log(`[MetadataEditor] Field mapping updated, reloading options...`)

      // Reload field options to get updated current values
      await loadFieldOptions()

      console.log(`[MetadataEditor] Triggering rescan...`)
      if (onFieldMappingChange) {
        onFieldMappingChange()
        console.log(`[MetadataEditor] Rescan triggered`)
      } else {
        console.warn(`[MetadataEditor] No onFieldMappingChange callback provided!`)
      }
    } catch (err) {
      console.error('Failed to update field mapping:', err)
    }
  }

  const getFieldOption = (field: string) => {
    return fieldOptions.find(opt => opt.field === field)
  }

  if (!book) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        Select an audiobook to view and edit metadata
      </div>
    )
  }

  // Get current field mappings
  const titleMapping = fieldOptions.find(opt => opt.field === 'title')?.current || 'title'
  const seriesMapping = fieldOptions.find(opt => opt.field === 'series')?.current || 'series'
  const authorsMapping = fieldOptions.find(opt => opt.field === 'authors')?.current || 'artist'
  const trackMapping = fieldOptions.find(opt => opt.field === 'track')?.current || 'track'

  // Get all metadata fields with mapping annotations
  const allMetadataFields: { key: string; value: any; mappedTo: string | null }[] = []
  if (book.raw_data) {
    Object.keys(book.raw_data)
      .sort()
      .forEach(key => {
        const value = book.raw_data![key]
        let mappedTo: string | null = null

        if (key === titleMapping) mappedTo = 'Title'
        else if (key === seriesMapping) mappedTo = 'Series'
        else if (authorsMapping.split(',').includes(key)) mappedTo = 'Author'
        else if (key === trackMapping) mappedTo = 'Track'

        if (value !== null && value !== undefined && value !== '') {
          allMetadataFields.push({ key, value, mappedTo })
        }
      })
  }

  return (
    <div className="p-4 overflow-y-auto">
      <h2 className="text-base font-semibold mb-3">Metadata Editor</h2>
      <p className="text-xs text-muted-foreground mb-4">
        Source: {book.source_path}
      </p>


      {/* All Metadata Fields Display - Table Format */}
      <div>
        <h3 className="text-xs font-semibold mb-2 text-muted-foreground">All Metadata Fields</h3>
        <table className="w-full text-xs border-collapse">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left py-1 px-2 font-medium text-muted-foreground w-[30px]"></th>
              <th className="text-left py-1 px-2 font-medium text-muted-foreground w-[140px]">Field</th>
              <th className="text-left py-1 px-2 font-medium text-muted-foreground">Value</th>
              <th className="text-left py-1 px-2 font-medium text-muted-foreground w-[80px]">Used For</th>
            </tr>
          </thead>
          <tbody>
            {allMetadataFields.map(({ key, value, mappedTo }, index) => (
              <tr
                key={key}
                className={`border-b border-border/30 ${
                  mappedTo ? 'bg-primary/5' : (index % 2 === 0 ? 'bg-muted/10' : 'bg-transparent')
                } hover:bg-accent/50 transition-colors`}
              >
                <td className="py-1.5 px-2 text-center">
                  {mappedTo && (
                    <span className={`font-bold ${
                      mappedTo === 'Title' ? 'text-green-600' :
                      mappedTo === 'Series' ? 'text-cyan-600' :
                      mappedTo === 'Author' ? 'text-orange-600' :
                      mappedTo === 'Track' ? 'text-blue-600' :
                      'text-primary'
                    }`}>✓</span>
                  )}
                </td>
                <td className="py-1.5 px-2">
                  <span className={`font-medium ${
                    mappedTo ? (
                      mappedTo === 'Title' ? 'text-green-600' :
                      mappedTo === 'Series' ? 'text-cyan-600' :
                      mappedTo === 'Author' ? 'text-orange-600' :
                      mappedTo === 'Track' ? 'text-blue-600' :
                      'text-primary'
                    ) : 'text-foreground'
                  }`}>
                    {key}
                  </span>
                </td>
                <td className="py-1.5 px-2 font-mono break-all">
                  {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                </td>
                <td className="py-1.5 px-2 text-right">
                  {mappedTo && (
                    <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded inline-block ${
                      mappedTo === 'Title' ? 'text-green-600 bg-green-600/20' :
                      mappedTo === 'Series' ? 'text-cyan-600 bg-cyan-600/20' :
                      mappedTo === 'Author' ? 'text-orange-600 bg-orange-600/20' :
                      mappedTo === 'Track' ? 'text-blue-600 bg-blue-600/20' :
                      'text-primary bg-primary/20'
                    }`}>
                      ← {mappedTo}
                    </span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {allMetadataFields.length === 0 && (
        <div className="text-center text-muted-foreground text-xs py-8">
          No metadata available
        </div>
      )}

      {/* Batch Preview - collapsible bottom panel */}
      <BatchPreview
        books={books}
        selectedIndices={selectedIndices}
        outputDir={outputDir}
      />
    </div>
  )
}
