import { useState, useEffect, useRef } from 'react'
import { Toolbar } from './components/Toolbar'
import { StatisticsBar } from './components/StatisticsBar'
import { GroupedFileList } from './components/GroupedFileList'
import { MetadataEditor } from './components/MetadataEditor'
import { OutputPreviewSimple } from './components/OutputPreviewSimple'
import { OptionsPanel } from './components/OptionsPanel'
import { ExecutionPreview } from './components/ExecutionPreview'
import { ExecutionResults } from './components/ExecutionResults'
import { GetInitialDirectories, ScanDirectory } from '../wailsjs/go/main/App'
import { organizer } from '../wailsjs/go/models'

function App() {
  const [inputDir, setInputDir] = useState('')
  const [outputDir, setOutputDir] = useState('')
  const [books, setBooks] = useState<organizer.Metadata[]>([])
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)
  const [selectedIndices, setSelectedIndices] = useState<Set<number>>(new Set())
  const [loading, setLoading] = useState(false)
  const [currentView, setCurrentView] = useState<'editing' | 'preview' | 'results'>('editing')
  const [executionResults, setExecutionResults] = useState<{
    success: boolean
    filesProcessed: number
    errors: string[]
    movedFiles?: Array<{from: string, to: string}>
  } | null>(null)

  // Panel resizing state
  const [leftPanelWidth, setLeftPanelWidth] = useState(300)
  const [rightPanelWidth, setRightPanelWidth] = useState(350)
  const [isResizingLeft, setIsResizingLeft] = useState(false)
  const [isResizingRight, setIsResizingRight] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  // Load initial directories from CLI args
  useEffect(() => {
    GetInitialDirectories().then((dirs) => {
      if (dirs.input_dir) {
        setInputDir(dirs.input_dir)
        if (dirs.output_dir) {
          setOutputDir(dirs.output_dir)
        }
        // Auto-scan if directory provided
        scanDirectory(dirs.input_dir)
      }
    }).catch((err) => {
      console.error('Failed to get initial directories:', err)
    })
  }, [])

  const scanDirectory = async (dir: string, preserveSelection: boolean = false) => {
    console.log(`[App] Starting scan of directory: ${dir}, preserveSelection=${preserveSelection}`)
    // Save current selections
    const currentPaths = preserveSelection
      ? Array.from(selectedIndices).map(idx => books[idx]?.source_path).filter(Boolean)
      : []
    const currentSinglePath = selectedIndex !== null ? books[selectedIndex]?.source_path : null

    setLoading(true)
    ScanDirectory(dir)
      .then((result) => {
        console.log(`[App] Scan completed, found ${result?.length || 0} audiobooks`)
        setBooks(result || [])

        if (result && result.length > 0) {
          if (preserveSelection && currentPaths.length > 0) {
            // Restore all selected indices
            const newSelectedIndices = new Set<number>()
            currentPaths.forEach(path => {
              const idx = result.findIndex(book => book.source_path === path)
              if (idx !== -1) newSelectedIndices.add(idx)
            })

            if (newSelectedIndices.size > 0) {
              setSelectedIndices(newSelectedIndices)
              // Restore single selection if it still exists
              const newSingleIdx = currentSinglePath
                ? result.findIndex(book => book.source_path === currentSinglePath)
                : -1
              setSelectedIndex(newSingleIdx !== -1 ? newSingleIdx : Array.from(newSelectedIndices)[0])
            } else {
              // Fallback to first book
              setSelectedIndex(0)
              setSelectedIndices(new Set([0]))
            }
          } else {
            // Auto-select first book
            setSelectedIndex(0)
            setSelectedIndices(new Set([0]))
          }
        }
      })
      .catch((err) => {
        console.error('Failed to scan directory:', err)
        setBooks([])
      })
      .finally(() => {
        setLoading(false)
      })
  }

  const handleScan = () => {
    if (inputDir) {
      scanDirectory(inputDir)
    }
  }

  const selectedBook = selectedIndex !== null ? books[selectedIndex] : null

  // Auto-scan when input directory changes
  useEffect(() => {
    if (inputDir) {
      console.log('[App] Input directory changed, triggering scan:', inputDir)
      scanDirectory(inputDir, false)
    }
  }, [inputDir])

  // Handle mouse move for resizing
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!containerRef.current) return

      if (isResizingLeft) {
        const containerRect = containerRef.current.getBoundingClientRect()
        const newWidth = e.clientX - containerRect.left
        setLeftPanelWidth(Math.max(200, Math.min(newWidth, 600)))
      }

      if (isResizingRight) {
        const containerRect = containerRef.current.getBoundingClientRect()
        const newWidth = containerRect.right - e.clientX
        setRightPanelWidth(Math.max(250, Math.min(newWidth, 600)))
      }
    }

    const handleMouseUp = () => {
      setIsResizingLeft(false)
      setIsResizingRight(false)
    }

    if (isResizingLeft || isResizingRight) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
  }, [isResizingLeft, isResizingRight])

  return (
    <div className="flex flex-col h-screen bg-background">
      {/* Top Toolbar */}
      <Toolbar
        inputDir={inputDir}
        outputDir={outputDir}
        onInputDirChange={setInputDir}
        onOutputDirChange={setOutputDir}
        onScan={handleScan}
        loading={loading}
      />

      {/* Statistics Bar */}
      <StatisticsBar />

      {/* Editing View */}
      {currentView === 'editing' && (
        <>
          {/* Options Panel (always visible) */}
          <OptionsPanel
            outputDir={outputDir}
            selectedBook={selectedBook}
            onScanModeChange={handleScan}
            onFieldMappingChange={() => {
              if (inputDir) {
                scanDirectory(inputDir, true)
              }
            }}
            onLayoutChange={() => {
              // Force re-render of preview by updating a key or state
              setSelectedIndex(selectedIndex === null ? null : selectedIndex)
            }}
          />

          {/* Three-column layout: Input → Metadata → Output */}
          <div ref={containerRef} className="flex flex-1 overflow-hidden">
        {/* Left Column - Grouped Files */}
        <div className="border-r border-border overflow-y-auto bg-card" style={{ width: `${leftPanelWidth}px`, flexShrink: 0 }}>
          <GroupedFileList
            books={books}
            selectedIndex={selectedIndex}
            selectedIndices={selectedIndices}
            onSelect={setSelectedIndex}
            onToggleSelection={(index) => {
              const newSelected = new Set(selectedIndices)
              if (newSelected.has(index)) {
                newSelected.delete(index)
                // If we removed focus, move it to first remaining checked book
                if (selectedIndex === index) {
                  const remaining = Array.from(newSelected).sort((a, b) => a - b)
                  setSelectedIndex(remaining.length > 0 ? remaining[0] : null)
                }
              } else {
                newSelected.add(index)
                setSelectedIndex(index) // Focus the newly checked book
              }
              setSelectedIndices(newSelected)
            }}
            onSelectAll={() => {
              setSelectedIndices(new Set(books.map((_, i) => i)))
            }}
            onSelectNone={() => {
              setSelectedIndices(new Set())
            }}
            onSelectBook={(indices: number[]) => {
              setSelectedIndices(new Set(indices))
            }}
            loading={loading}
          />
        </div>

        {/* Left Resize Handle */}
        <div
          className="w-1 hover:w-1.5 bg-border hover:bg-primary cursor-col-resize flex-shrink-0 transition-all"
          onMouseDown={() => setIsResizingLeft(true)}
        />

        {/* Center Column - Metadata Editor */}
        <div className="flex-1 overflow-y-auto min-w-0 bg-background">
          <div className="sticky top-0 bg-background border-b border-border p-2 z-10">
            <div className="flex items-center justify-between">
              <div className="text-sm font-semibold">Metadata Editor</div>
              {selectedIndices.size > 1 && (() => {
                const sortedIndices = Array.from(selectedIndices).sort((a, b) => a - b)
                const currentPos = selectedIndex !== null ? sortedIndices.indexOf(selectedIndex) : -1
                const goPrev = () => {
                  if (currentPos > 0) setSelectedIndex(sortedIndices[currentPos - 1])
                  else setSelectedIndex(sortedIndices[sortedIndices.length - 1]) // wrap
                }
                const goNext = () => {
                  if (currentPos < sortedIndices.length - 1) setSelectedIndex(sortedIndices[currentPos + 1])
                  else setSelectedIndex(sortedIndices[0]) // wrap
                }
                return (
                  <div className="flex items-center gap-1 text-xs">
                    <button onClick={goPrev} className="px-1.5 py-0.5 rounded hover:bg-muted border border-border text-muted-foreground">←</button>
                    <span className="text-muted-foreground tabular-nums">{currentPos + 1} / {sortedIndices.length}</span>
                    <button onClick={goNext} className="px-1.5 py-0.5 rounded hover:bg-muted border border-border text-muted-foreground">→</button>
                  </div>
                )
              })()}
            </div>
            <div className="text-xs text-muted-foreground">Edit tags and information</div>
          </div>
          <MetadataEditor
            book={selectedBook}
            bookIndex={selectedIndex}
            books={books}
            selectedIndices={selectedIndices}
            outputDir={outputDir}
            onFieldMappingChange={() => {
              if (inputDir) {
                // Preserve selection by passing true
                scanDirectory(inputDir, true)
              }
            }}
          />
        </div>

        {/* Right Resize Handle */}
        <div
          className="w-1 hover:w-1.5 bg-border hover:bg-primary cursor-col-resize flex-shrink-0 transition-all"
          onMouseDown={() => setIsResizingRight(true)}
        />

        {/* Right Column - Output Preview */}
        <div className="border-l border-border overflow-y-auto bg-card" style={{ width: `${rightPanelWidth}px`, flexShrink: 0 }}>
          <div className="sticky top-0 bg-background border-b border-border p-2 z-10">
            <div className="text-sm font-semibold">Output Preview</div>
            <div className="text-xs text-muted-foreground">Organized path structure</div>
          </div>
          <OutputPreviewSimple
            book={selectedBook}
            outputDir={outputDir}
          />
        </div>
      </div>

      {/* Execute Button - Fixed at bottom (only in editing view) */}
      {selectedIndices.size > 0 && (
        <div className="border-t border-border bg-card p-3 flex justify-end">
          <button
            onClick={() => setCurrentView('preview')}
            className="px-6 py-2 rounded bg-primary text-primary-foreground hover:bg-primary/90 transition-colors font-medium"
          >
            Preview Organization ({selectedIndices.size} files) →
          </button>
        </div>
      )}
        </>
      )}

      {/* Results View */}
      {currentView === 'results' && executionResults && (
        <div className="flex-1">
          <ExecutionResults
            success={executionResults.success}
            filesProcessed={executionResults.filesProcessed}
            errors={executionResults.errors}
            movedFiles={executionResults.movedFiles}
            onBack={() => setCurrentView('editing')}
            onUndo={async () => {
              try {
                const { UndoLastOperation } = await import('../wailsjs/go/main/App')
                const result = await UndoLastOperation()
                console.log('Undo result:', result)

                // Show undo results
                setExecutionResults({
                  success: result.success || false,
                  filesProcessed: result.filesRestored || 0,
                  errors: result.errors || []
                })

                // Rescan to update file list
                if (inputDir) {
                  scanDirectory(inputDir, false)
                }
              } catch (err) {
                console.error('Undo failed:', err)
                setExecutionResults({
                  success: false,
                  filesProcessed: 0,
                  errors: [String(err)]
                })
              }
            }}
          />
        </div>
      )}

      {/* Preview/Execution View */}
      {currentView === 'preview' && (
        <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
          <ExecutionPreview
            books={books}
            selectedIndices={selectedIndices}
            outputDir={outputDir}
            onFieldMappingChange={() => scanDirectory(inputDir, true)}
            onExecute={async (_copyMode, operations) => {
              const selectedIndicesArray = Array.from(selectedIndices)
              try {
                const { PreviewChanges, ExecuteOrganize } = await import('../wailsjs/go/main/App')
                // Configure organizer with only selected files
                await PreviewChanges(inputDir, outputDir, selectedIndicesArray)
                // Execute (not a dry run)
                const summary = await ExecuteOrganize(false)
                setExecutionResults({
                  success: true,
                  filesProcessed: summary.Moves?.length || 0,
                  errors: [],
                  movedFiles: summary.Moves?.map(m => ({ from: m.from, to: m.to })) || operations,
                })
              } catch (err) {
                console.error('Organization failed:', err)
                setExecutionResults({
                  success: false,
                  filesProcessed: 0,
                  errors: [String(err)],
                  movedFiles: operations,
                })
              }
              setCurrentView('results')
            }}
            onCancel={() => setCurrentView('editing')}
          />
        </div>
      )}
    </div>
  )
}

export default App
