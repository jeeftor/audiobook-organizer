import { useEffect, useState } from 'react'
import { GetScanStatistics } from '../../wailsjs/go/main/App'
import { FileAudio, FolderTree, AlertCircle } from 'lucide-react'

interface ScanStats {
  total_files: number
  total_audiobooks: number
  missing_metadata: number
}

export function StatisticsBar() {
  const [stats, setStats] = useState<ScanStats | null>(null)

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    try {
      const result = await GetScanStatistics()
      setStats(result as ScanStats)
    } catch (err) {
      console.error('Failed to load statistics:', err)
    }
  }

  // Expose refresh function globally so parent can call it after scan
  useEffect(() => {
    (window as any).refreshStats = loadStats
  }, [])

  if (!stats || stats.total_files === 0) {
    return null
  }

  return (
    <div className="border-b border-border bg-muted/30 px-4 py-2">
      <div className="flex items-center gap-6 text-sm">
        <div className="flex items-center gap-2">
          <FileAudio className="h-4 w-4 text-blue-600" />
          <span className="font-medium">{stats.total_files}</span>
          <span className="text-muted-foreground">files found</span>
        </div>

        <div className="h-4 w-px bg-border" />

        <div className="flex items-center gap-2">
          <FolderTree className="h-4 w-4 text-green-600" />
          <span className="font-medium">{stats.total_audiobooks}</span>
          <span className="text-muted-foreground">audiobooks detected</span>
        </div>

        {stats.missing_metadata > 0 && (
          <>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2 text-orange-600">
              <AlertCircle className="h-4 w-4" />
              <span className="font-medium">{stats.missing_metadata}</span>
              <span>files missing metadata</span>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
