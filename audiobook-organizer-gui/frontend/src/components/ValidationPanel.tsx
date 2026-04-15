import { useState, useEffect } from 'react'
import { AlertCircle, AlertTriangle, ChevronDown, ChevronUp, CheckCircle } from 'lucide-react'
// @ts-ignore — GetValidationWarnings will exist once the backend is regenerated
import { GetValidationWarnings } from '../../wailsjs/go/main/App'

interface ValidationWarning {
  book_index: number
  book_title: string
  type: string       // "missing_author" | "missing_title" | "duplicate_path"
  message: string
  severity: string   // "error" | "warning"
}

interface ValidationPanelProps {
  scanVersion: number
}

export function ValidationPanel({ scanVersion }: ValidationPanelProps) {
  const [warnings, setWarnings] = useState<ValidationWarning[]>([])
  const [loading, setLoading] = useState(false)
  const [expanded, setExpanded] = useState(false)

  useEffect(() => {
    setLoading(true)
    GetValidationWarnings()
      .then((result: ValidationWarning[]) => {
        const list = result || []
        setWarnings(list)
        // Auto-expand if there are 3 or fewer items
        setExpanded(list.length > 0 && list.length <= 3)
      })
      .catch((err: unknown) => {
        console.error('Failed to get validation warnings:', err)
        setWarnings([])
      })
      .finally(() => {
        setLoading(false)
      })
  }, [scanVersion])

  if (loading) {
    return (
      <div className="px-4 py-2 text-xs text-muted-foreground animate-pulse">
        Checking for issues...
      </div>
    )
  }

  if (warnings.length === 0) {
    return (
      <div className="px-4 py-2 flex items-center gap-2 text-xs text-green-600 dark:text-green-500">
        <CheckCircle className="h-3.5 w-3.5 flex-shrink-0" />
        <span>No issues detected</span>
      </div>
    )
  }

  const errors = warnings.filter(w => w.severity === 'error')
  const warningItems = warnings.filter(w => w.severity === 'warning')

  // Build summary label
  const parts: string[] = []
  if (errors.length > 0) parts.push(`${errors.length} Error${errors.length !== 1 ? 's' : ''}`)
  if (warningItems.length > 0) parts.push(`${warningItems.length} Warning${warningItems.length !== 1 ? 's' : ''}`)
  const summaryLabel = parts.join(', ')

  // Errors first, then warnings
  const sorted = [...errors, ...warningItems]

  return (
    <div className="text-xs">
      {/* Summary header — always visible, click to toggle */}
      <button
        onClick={() => setExpanded(v => !v)}
        className="w-full flex items-center justify-between px-4 py-2 hover:bg-muted/50 transition-colors text-left"
      >
        <div className="flex items-center gap-2">
          {errors.length > 0 ? (
            <AlertCircle className="h-3.5 w-3.5 flex-shrink-0 text-destructive" />
          ) : (
            <AlertTriangle className="h-3.5 w-3.5 flex-shrink-0 text-amber-500" />
          )}
          <span className={errors.length > 0 ? 'text-destructive font-medium' : 'text-amber-600 dark:text-amber-500 font-medium'}>
            {summaryLabel}
          </span>
        </div>
        {expanded
          ? <ChevronUp className="h-3.5 w-3.5 text-muted-foreground" />
          : <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
        }
      </button>

      {/* Warning list */}
      {expanded && (
        <ul className="px-4 pb-2 space-y-1.5">
          {sorted.map((w, i) => (
            <li key={i} className="flex items-start gap-2">
              {w.severity === 'error' ? (
                <AlertCircle className="h-3.5 w-3.5 flex-shrink-0 mt-0.5 text-destructive" />
              ) : (
                <AlertTriangle className="h-3.5 w-3.5 flex-shrink-0 mt-0.5 text-amber-500" />
              )}
              <div className="min-w-0">
                <span
                  className="font-medium truncate block max-w-[180px]"
                  title={w.book_title}
                >
                  {w.book_title || 'Unknown'}
                </span>
                <span className="text-muted-foreground leading-snug">{w.message}</span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
