import { CheckCircle, XCircle, AlertCircle, ArrowLeft, Undo } from 'lucide-react'
import { useState } from 'react'

interface ExecutionResultsProps {
  success: boolean
  filesProcessed: number
  errors: string[]
  movedFiles?: Array<{from: string, to: string}>
  onBack: () => void
  onUndo?: () => void
}

export function ExecutionResults({ success, filesProcessed, errors, movedFiles, onBack, onUndo }: ExecutionResultsProps) {
  const [undoing, setUndoing] = useState(false)

  const handleUndo = async () => {
    if (!onUndo) return
    setUndoing(true)
    try {
      await onUndo()
    } finally {
      setUndoing(false)
    }
  }
  return (
    <div className="flex flex-col items-center justify-center h-full p-8">
      <div className="max-w-2xl w-full">
        {/* Success/Failure Icon */}
        <div className="flex justify-center mb-6">
          {success ? (
            <CheckCircle className="h-24 w-24 text-green-600" />
          ) : (
            <XCircle className="h-24 w-24 text-red-600" />
          )}
        </div>

        {/* Title */}
        <h1 className="text-3xl font-bold text-center mb-4">
          {success ? 'Organization Complete!' : 'Organization Failed'}
        </h1>

        {/* Summary */}
        <div className="bg-card border border-border rounded-lg p-6 mb-6">
          <div className="space-y-3">
            <div className="flex justify-between text-lg">
              <span className="text-muted-foreground">Files Processed:</span>
              <span className="font-semibold">{filesProcessed}</span>
            </div>
            {errors.length > 0 && (
              <div className="flex justify-between text-lg">
                <span className="text-muted-foreground">Errors:</span>
                <span className="font-semibold text-red-600">{errors.length}</span>
              </div>
            )}
          </div>
        </div>

        {/* Moved Files List */}
        {movedFiles && movedFiles.length > 0 && (
          <div className="bg-card border border-border rounded-lg mb-6 overflow-hidden">
            <div className="px-4 py-2 border-b border-border text-sm font-medium">Files {success ? 'moved to' : 'planned for'}</div>
            <div className="max-h-64 overflow-y-auto divide-y divide-border">
              {movedFiles.map((f, idx) => (
                <div key={idx} className="px-4 py-2 text-[11px] font-mono">
                  <div className="text-muted-foreground truncate">{f.from}</div>
                  <div className="text-green-600 truncate">→ {f.to}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Error Details */}
        {errors.length > 0 && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4 mb-6">
            <div className="flex items-start gap-2 mb-2">
              <AlertCircle className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
              <h3 className="font-semibold text-red-600">Errors Encountered</h3>
            </div>
            <ul className="space-y-1 ml-7 text-sm">
              {errors.map((error, idx) => (
                <li key={idx} className="text-red-700 dark:text-red-400">
                  {error}
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Success Message */}
        {success && (
          <p className="text-center text-muted-foreground mb-8">
            Your audiobooks have been successfully organized. You can now close this window or organize more files.
          </p>
        )}

        {/* Action Buttons */}
        <div className="flex justify-center gap-4">
          {success && onUndo && (
            <button
              onClick={handleUndo}
              disabled={undoing}
              className="flex items-center gap-2 px-6 py-3 rounded border-2 border-orange-600 text-orange-600 hover:bg-orange-600 hover:text-white transition-colors font-medium disabled:opacity-50"
            >
              <Undo className="h-5 w-5" />
              {undoing ? 'Undoing...' : 'Undo Organization'}
            </button>
          )}
          <button
            onClick={onBack}
            className="flex items-center gap-2 px-6 py-3 rounded bg-primary text-primary-foreground hover:bg-primary/90 transition-colors font-medium"
          >
            <ArrowLeft className="h-5 w-5" />
            Back to Editing
          </button>
        </div>
      </div>
    </div>
  )
}
