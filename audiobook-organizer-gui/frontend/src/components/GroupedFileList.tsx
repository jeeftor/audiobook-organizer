import { useState, useEffect } from 'react'
import { organizer } from '../../wailsjs/go/models'
import { GetScanStatistics } from '../../wailsjs/go/main/App'
import { ChevronRight, ChevronDown, FolderTree, FileAudio, AlertCircle } from 'lucide-react'

interface GroupedFileListProps {
  books: organizer.Metadata[]
  selectedIndex: number | null
  selectedIndices: Set<number>
  onSelect: (index: number) => void
  onToggleSelection: (index: number) => void
  onSelectAll: () => void
  onSelectNone: () => void
  onSelectBook: (indices: number[]) => void
  loading: boolean
}

interface AlbumGroup {
  name: string
  author: string
  series: string
  file_count: number
  file_indices: number[]
  files: organizer.Metadata[]
}

export function GroupedFileList({
  books,
  selectedIndex,
  selectedIndices,
  onSelect,
  onToggleSelection,
  onSelectAll,
  onSelectNone,
  onSelectBook,
  loading
}: GroupedFileListProps) {
  const [groups, setGroups] = useState<AlbumGroup[]>([])
  const [ungrouped, setUngrouped] = useState<organizer.Metadata[]>([])
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [viewMode, setViewMode] = useState<'grouped' | 'flat'>('grouped')

  useEffect(() => {
    if (books.length > 0) {
      loadGroups()
    }
  }, [books])

  const loadGroups = async () => {
    try {
      const stats = await GetScanStatistics()
      setGroups(stats.album_groups || [])
      setUngrouped(stats.ungrouped_files || [])

      // Don't auto-expand any groups - let user expand what they want
    } catch (err) {
      console.error('Failed to load groups:', err)
    }
  }

  const toggleGroup = (author: string, name: string) => {
    const key = `${author}|${name}`
    const newExpanded = new Set(expandedGroups)
    if (newExpanded.has(key)) {
      newExpanded.delete(key)
    } else {
      newExpanded.add(key)
    }
    setExpandedGroups(newExpanded)
  }

  if (loading) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        <div className="animate-pulse">Scanning files...</div>
      </div>
    )
  }

  if (books.length === 0) {
    return (
      <div className="p-4 text-center text-sm text-muted-foreground">
        No files found. Click "Open Folder" to scan a directory.
      </div>
    )
  }

  // View mode toggle
  const renderViewToggle = () => (
    <div className="flex gap-1">
      <button
        onClick={() => setViewMode('grouped')}
        className={`p-1.5 rounded ${
          viewMode === 'grouped'
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-muted-foreground hover:bg-muted/80'
        }`}
        title="Grouped by album"
      >
        <FolderTree className="h-4 w-4" />
      </button>
      <button
        onClick={() => setViewMode('flat')}
        className={`p-1.5 rounded ${
          viewMode === 'flat'
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-muted-foreground hover:bg-muted/80'
        }`}
        title="Flat list"
      >
        <FileAudio className="h-4 w-4" />
      </button>
    </div>
  )

  // Selection toolbar
  const renderSelectionToolbar = () => (
    <div className="flex gap-1 text-[10px] border-t border-border p-1.5 bg-muted/50">
      <button
        onClick={onSelectAll}
        className="px-2 py-0.5 rounded bg-background hover:bg-accent transition-colors"
      >
        Select All
      </button>
      <button
        onClick={onSelectNone}
        className="px-2 py-0.5 rounded bg-background hover:bg-accent transition-colors"
      >
        Clear
      </button>
      <div className="flex-1" />
      <span className="text-muted-foreground px-2 py-0.5">
        {selectedIndices.size} selected
      </span>
    </div>
  )

  if (viewMode === 'flat') {
    // Flat file list view
    return (
      <div>
        <div className="sticky top-0 bg-card border-b border-border p-2 z-10 flex justify-between items-center">
          <div>
            <div className="text-sm font-semibold">All Files ({books.length})</div>
            <div className="text-xs text-muted-foreground">Flat list view</div>
          </div>
          {renderViewToggle()}
        </div>
        {renderSelectionToolbar()}
        <div className="divide-y divide-border">
          {books.map((book, index) => {
            const isSelected = selectedIndex === index
            const isChecked = selectedIndices.has(index)
            const filename = book.source_path?.split('/').pop() || 'Unknown file'

            return (
              <div
                key={index}
                className={`p-2 transition-colors ${
                  isChecked
                    ? 'bg-green-500/10 border-l-2 border-l-green-500'
                    : isSelected
                    ? 'bg-primary/10 border-l-2 border-l-primary'
                    : `${index % 2 === 0 ? 'bg-muted/10' : 'bg-transparent'} hover:bg-accent/50 border-l-2 border-l-transparent`
                }`}
              >
                <div className="flex items-start gap-2">
                  <input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => onToggleSelection(index)}
                    onClick={(e) => e.stopPropagation()}
                    className="mt-1 cursor-pointer accent-green-600"
                  />
                  <FileAudio
                    className={`h-4 w-4 mt-0.5 flex-shrink-0 cursor-pointer ${
                      isChecked ? 'text-green-600' : 'text-muted-foreground'
                    }`}
                    onClick={() => onSelect(index)}
                  />
                  <div className="flex-1 min-w-0 cursor-pointer" onClick={() => onSelect(index)}>
                    <div className={`text-xs font-mono truncate ${isChecked ? 'font-semibold' : ''}`} title={filename}>
                      {filename}
                    </div>
                    {book.title && (
                      <div className="text-[10px] text-muted-foreground truncate mt-0.5">
                        {book.title}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    )
  }

  // Grouped view
  return (
    <div>
      <div className="sticky top-0 bg-card border-b border-border p-2 z-10 flex justify-between items-center">
        <div>
          <div className="text-sm font-semibold">Audiobooks ({groups.length})</div>
          <div className="text-xs text-muted-foreground">Grouped by album</div>
        </div>
        {renderViewToggle()}
      </div>
      {renderSelectionToolbar()}
      <div className="divide-y divide-border">
        {groups.map((group, groupIdx) => {
          const groupKey = `${group.author}|${group.name}`
          const isExpanded = expandedGroups.has(groupKey)
          const hasSelectedFile = group.file_indices.some(idx => idx === selectedIndex)
          const allFilesSelected = group.file_indices.every(idx => selectedIndices.has(idx))
          const someFilesSelected = group.file_indices.some(idx => selectedIndices.has(idx))

          return (
            <div key={groupIdx} className={`${hasSelectedFile ? 'bg-primary/5' : ''} ${someFilesSelected ? 'border-l-2 border-l-green-500' : ''}`}>
              {/* Group Header */}
              <div className={`p-2 flex items-start gap-2 ${allFilesSelected ? 'bg-green-500/10' : someFilesSelected ? 'bg-green-500/5' : ''}`}>
                <input
                  type="checkbox"
                  checked={allFilesSelected}
                  ref={indeterminate => {
                    if (indeterminate) {
                      indeterminate.indeterminate = someFilesSelected && !allFilesSelected
                    }
                  }}
                  onChange={() => {
                    if (allFilesSelected) {
                      // Deselect all files in this book
                      const newSelected = new Set(selectedIndices)
                      group.file_indices.forEach(idx => newSelected.delete(idx))
                      onSelectBook(Array.from(newSelected))
                    } else {
                      // Select all files in this book - merge with existing selection
                      const newSelected = new Set(selectedIndices)
                      group.file_indices.forEach(idx => newSelected.add(idx))
                      onSelectBook(Array.from(newSelected))
                    }
                  }}
                  onClick={(e) => e.stopPropagation()}
                  className="mt-0.5 cursor-pointer accent-green-600"
                  title={allFilesSelected ? 'Deselect all files in this book' : 'Select all files in this book'}
                />
                <div
                  onClick={() => toggleGroup(group.author, group.name)}
                  className="cursor-pointer hover:bg-accent/50 rounded p-0.5 transition-colors"
                >
                  {isExpanded ? (
                    <ChevronDown className="h-4 w-4 flex-shrink-0" />
                  ) : (
                    <ChevronRight className="h-4 w-4 flex-shrink-0" />
                  )}
                </div>
                <FolderTree className={`h-4 w-4 mt-0.5 flex-shrink-0 ${allFilesSelected ? 'text-green-600' : 'text-orange-600'}`} />
                <div className="flex-1 min-w-0">
                  <div className="text-xs font-medium truncate">
                    {group.author || 'Unknown Author'} - {group.name}
                  </div>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-[10px] text-muted-foreground">
                      {group.file_count} files
                    </span>
                    {group.series && (
                      <span className="text-[10px] text-cyan-600">
                        {group.series}
                      </span>
                    )}
                    {someFilesSelected && (
                      <span className="text-[10px] text-green-600 font-medium">
                        {allFilesSelected ? '✓ All selected' : `${group.file_indices.filter(i => selectedIndices.has(i)).length} selected`}
                      </span>
                    )}
                  </div>
                </div>
              </div>

              {/* Expanded Files */}
              {isExpanded && (
                <div className="ml-6 border-l-2 border-border">
                  {group.file_indices.map((fileIdx) => {
                    const book = books[fileIdx]
                    if (!book) return null
                    const isSelected = selectedIndex === fileIdx
                    const isChecked = selectedIndices.has(fileIdx)
                    const filename = book.source_path?.split('/').pop() || 'Unknown file'

                    return (
                      <div
                        key={fileIdx}
                        className={`p-2 pl-4 transition-colors ${
                          isChecked
                            ? 'bg-green-500/10 border-l-2 border-l-green-500'
                            : isSelected
                            ? 'bg-primary/10 border-l-2 border-l-primary'
                            : 'hover:bg-accent/50 border-l-2 border-l-transparent'
                        }`}
                      >
                        <div className="flex items-start gap-2">
                          <input
                            type="checkbox"
                            checked={isChecked}
                            onChange={() => onToggleSelection(fileIdx)}
                            onClick={(e) => e.stopPropagation()}
                            className="mt-0.5 cursor-pointer accent-green-600"
                          />
                          <FileAudio
                            className={`h-3 w-3 mt-0.5 flex-shrink-0 cursor-pointer ${
                              isChecked ? 'text-green-600' : 'text-muted-foreground'
                            }`}
                            onClick={() => onSelect(fileIdx)}
                          />
                          <div
                            className="flex-1 min-w-0 cursor-pointer"
                            onClick={() => onSelect(fileIdx)}
                          >
                            <div className={`text-[11px] font-mono truncate ${isChecked ? 'font-semibold' : ''}`} title={filename}>
                              {filename}
                            </div>
                            {book.track_number && (
                              <div className="text-[10px] text-muted-foreground">
                                Track {book.track_number}
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )
        })}

        {/* Ungrouped Files */}
        {ungrouped.length > 0 && (
          <div className="bg-orange-50 dark:bg-orange-950/20">
            <div className="p-2 flex items-start gap-2">
              <AlertCircle className="h-4 w-4 mt-0.5 text-orange-600 flex-shrink-0" />
              <div className="flex-1">
                <div className="text-xs font-medium text-orange-600">
                  Ungrouped Files ({ungrouped.length})
                </div>
                <div className="text-[10px] text-muted-foreground">
                  Files without album/title metadata
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
