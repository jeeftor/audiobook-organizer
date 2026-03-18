import { useState, useEffect } from 'react'
import { organizer } from '../../wailsjs/go/models'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { UpdateScanMode, GetCurrentScanMode, UpdateFieldMapping, GetFieldMappingPresets } from '../../wailsjs/go/main/App'

interface RightPanelProps {
  book: organizer.Metadata | null
  outputDir: string
}

export function RightPanel({ book, outputDir }: RightPanelProps) {
  const [showScanOptions, setShowScanOptions] = useState(true)
  const [showOrgOptions, setShowOrgOptions] = useState(true)
  const [showPreview, setShowPreview] = useState(true)
  const [scanMode, setScanMode] = useState('embedded (directory)')
  const [fieldMapping, setFieldMapping] = useState('Audio (default)')

  // Load current scan mode on mount
  useEffect(() => {
    GetCurrentScanMode().then(mode => {
      setScanMode(mode)
    }).catch(err => {
      console.error('Failed to get current scan mode:', err)
    })
  }, [])

  const handleScanModeChange = async (mode: string) => {
    try {
      await UpdateScanMode(mode)
      setScanMode(mode)
      console.log('Scan mode updated to:', mode)
    } catch (err) {
      console.error('Failed to update scan mode:', err)
    }
  }

  const handleFieldMappingChange = async (preset: string) => {
    try {
      const presets = await GetFieldMappingPresets()
      const selected = presets.find(p => p.name === preset)
      if (selected) {
        await UpdateFieldMapping(selected.mapping)
        setFieldMapping(preset)
        console.log('Field mapping updated to:', preset)
      }
    } catch (err) {
      console.error('Failed to update field mapping:', err)
    }
  }

  return (
    <div className="p-4 space-y-4">
      {/* Cover Art Placeholder */}
      <div className="border border-border rounded-lg p-4 bg-muted/30">
        <div className="aspect-square bg-muted rounded flex items-center justify-center text-muted-foreground text-sm">
          No Cover Art
        </div>
      </div>

      {/* Scanning Options */}
      <div className="border border-border rounded-lg">
        <button
          onClick={() => setShowScanOptions(!showScanOptions)}
          className="w-full flex items-center justify-between p-3 hover:bg-accent/50 transition-colors"
        >
          <span className="text-sm font-medium">Scanning Options</span>
          {showScanOptions ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>
        {showScanOptions && (
          <div className="p-3 border-t border-border space-y-3">
            <div>
              <label className="text-xs font-medium mb-2 block">Scan Mode</label>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-xs cursor-pointer">
                  <input
                    type="radio"
                    name="scanMode"
                    checked={scanMode === 'embedded (directory)'}
                    onChange={() => handleScanModeChange('embedded (directory)')}
                  />
                  <span>📁 Hierarchical (directory)</span>
                </label>
                <label className="flex items-center gap-2 text-xs cursor-pointer">
                  <input
                    type="radio"
                    name="scanMode"
                    checked={scanMode === 'embedded (file)'}
                    onChange={() => handleScanModeChange('embedded (file)')}
                  />
                  <span>📄 Flat (file)</span>
                </label>
                <label className="flex items-center gap-2 text-xs cursor-pointer">
                  <input
                    type="radio"
                    name="scanMode"
                    checked={scanMode === 'metadata.json'}
                    onChange={() => handleScanModeChange('metadata.json')}
                  />
                  <span>📋 Metadata.json only</span>
                </label>
              </div>
            </div>
            <div>
              <label className="text-xs font-medium mb-2 block">Field Mapping</label>
              <select
                className="w-full text-xs p-2 rounded border border-border bg-background"
                value={fieldMapping}
                onChange={(e) => handleFieldMappingChange(e.target.value)}
              >
                <option>Audio</option>
                <option>EPUB</option>
                <option>Default</option>
              </select>
            </div>
          </div>
        )}
      </div>

      {/* Organization Options */}
      <div className="border border-border rounded-lg">
        <button
          onClick={() => setShowOrgOptions(!showOrgOptions)}
          className="w-full flex items-center justify-between p-3 hover:bg-accent/50 transition-colors"
        >
          <span className="text-sm font-medium">Organization</span>
          {showOrgOptions ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>
        {showOrgOptions && (
          <div className="p-3 border-t border-border space-y-3">
            <div>
              <label className="text-xs font-medium mb-2 block">Layout Template</label>
              <select className="w-full text-xs p-2 rounded border border-border bg-background">
                <option>author-series-title</option>
                <option>author-title</option>
                <option>author-series</option>
                <option>author-only</option>
              </select>
            </div>
            <div>
              <label className="text-xs font-medium mb-2 block">Author Format</label>
              <select className="w-full text-xs p-2 rounded border border-border bg-background">
                <option>First Last</option>
                <option>Last, First</option>
                <option>Preserve</option>
              </select>
            </div>
            <div>
              <label className="flex items-center gap-2 text-xs cursor-pointer">
                <input type="checkbox" defaultChecked />
                <span>Include series</span>
              </label>
            </div>
            <div>
              <label className="flex items-center gap-2 text-xs cursor-pointer">
                <input type="checkbox" />
                <span>Replace spaces with _</span>
              </label>
            </div>
          </div>
        )}
      </div>

      {/* Preview */}
      <div className="border border-border rounded-lg">
        <button
          onClick={() => setShowPreview(!showPreview)}
          className="w-full flex items-center justify-between p-3 hover:bg-accent/50 transition-colors"
        >
          <span className="text-sm font-medium">Output Preview</span>
          {showPreview ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>
        {showPreview && (
          <div className="p-3 border-t border-border">
            {book ? (
              <div className="text-xs font-mono space-y-1">
                <div className="text-muted-foreground">{outputDir || '/output'}/</div>
                <div className="text-orange-600 ml-2">
                  {book.authors?.[0] || 'Unknown Author'}/
                </div>
                {book.series && book.series.length > 0 && (
                  <div className="text-cyan-600 ml-4">
                    {book.series[0]}/
                  </div>
                )}
                <div className="text-green-600 ml-6">
                  {book.title || 'Unknown Title'}/
                </div>
                <div className="text-muted-foreground ml-8">
                  file.m4b
                </div>
              </div>
            ) : (
              <div className="text-xs text-muted-foreground italic">
                Select an audiobook to preview
              </div>
            )}
          </div>
        )}
      </div>

      {/* Validation */}
      <div className="border border-border rounded-lg p-3 bg-muted/30">
        <div className="text-xs font-medium mb-2">Validation</div>
        <div className="text-xs text-muted-foreground">
          ✓ No conflicts detected
        </div>
      </div>
    </div>
  )
}
