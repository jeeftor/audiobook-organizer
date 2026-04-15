import { FolderOpen, Moon, Sun } from 'lucide-react'
import { Button } from './ui/button'
import { SelectDirectory } from '../../wailsjs/go/main/App'
import { useTheme } from '../contexts/ThemeContext'

interface ToolbarProps {
  inputDir: string
  outputDir: string
  onInputDirChange: (dir: string) => void
  onOutputDirChange: (dir: string) => void
  onScan: () => void
  loading: boolean
}

export function Toolbar({
  inputDir,
  outputDir,
  onInputDirChange,
  onOutputDirChange,
  onScan,
  loading,
}: ToolbarProps) {
  const { theme, toggleTheme } = useTheme()
  const handleSelectInputDir = async () => {
    try {
      const dir = await SelectDirectory('')
      if (dir) {
        console.log('[Toolbar] Selected new directory:', dir)
        onInputDirChange(dir)
      }
    } catch (err) {
      console.error('Failed to select directory:', err)
    }
  }

  const handleSelectOutputDir = async () => {
    try {
      const dir = await SelectDirectory('')
      if (dir) {
        onOutputDirChange(dir)
      }
    } catch (err) {
      console.error('Failed to select directory:', err)
    }
  }

  return (
    <div className="border-b border-border bg-card">
      <div className="flex items-center gap-2 p-2">
        {/* Primary Actions */}
        <Button
          variant="outline"
          size="sm"
          onClick={handleSelectInputDir}
          disabled={loading}
          className="gap-2"
        >
          <FolderOpen className="h-4 w-4" />
          {loading ? 'Scanning...' : 'Open Folder'}
        </Button>

        {/* Directory Display */}
        <div className="flex-1 flex items-center gap-4 ml-4 text-xs">
          <div className="flex items-center gap-2">
            <span className="font-medium text-muted-foreground">Input:</span>
            <span className="truncate max-w-xs text-foreground" title={inputDir}>
              {inputDir || 'Not selected'}
            </span>
          </div>
          <div className="h-4 w-px bg-border" />
          <div className="flex items-center gap-2">
            <span className="font-medium text-muted-foreground">Output:</span>
            <button
              onClick={handleSelectOutputDir}
              className="truncate max-w-xs text-foreground hover:text-primary underline decoration-dotted cursor-pointer"
              title={outputDir || 'Click to select output directory'}
            >
              {outputDir || '..'}
            </button>
          </div>
        </div>

        {/* Theme Toggle — far right */}
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleTheme}
          title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          className="ml-auto flex-shrink-0"
        >
          {theme === 'dark' ? (
            <Sun className="h-4 w-4" />
          ) : (
            <Moon className="h-4 w-4" />
          )}
        </Button>
      </div>
    </div>
  )
}
