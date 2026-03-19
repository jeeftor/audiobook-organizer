interface ColoredPathProps {
  parts: string[]
  author: string
  series: string | undefined
  title: string
  filename: string
}

export function ColoredPath({ parts, author, series, title, filename }: ColoredPathProps) {
  return (
    <span className="inline-flex flex-wrap items-center gap-0.5 leading-relaxed">
      {parts.map((part, idx) => {
        // Base output dir — plain muted text, no badge
        if (idx === 0) {
          return (
            <span key={idx} className="text-muted-foreground">
              {part}
              <span className="mx-0.5">/</span>
            </span>
          )
        }

        let textColor = 'text-foreground'
        let bgColor = 'bg-transparent'
        let borderColor = 'border-transparent'

        if (part === author) {
          textColor = 'text-orange-700 dark:text-orange-400'
          bgColor = 'bg-orange-50 dark:bg-orange-950/40'
          borderColor = 'border-orange-300 dark:border-orange-700'
        } else if (series && part === series) {
          textColor = 'text-cyan-700 dark:text-cyan-400'
          bgColor = 'bg-cyan-50 dark:bg-cyan-950/40'
          borderColor = 'border-cyan-300 dark:border-cyan-700'
        } else if (part === title) {
          textColor = 'text-green-700 dark:text-green-400'
          bgColor = 'bg-green-50 dark:bg-green-950/40'
          borderColor = 'border-green-300 dark:border-green-700'
        } else if (part === filename) {
          textColor = 'text-blue-700 dark:text-blue-400'
          bgColor = 'bg-blue-50 dark:bg-blue-950/40'
          borderColor = 'border-blue-300 dark:border-blue-700'
        }

        const isLast = idx === parts.length - 1
        return (
          <span key={idx} className="inline-flex items-center gap-0.5">
            <span className={`inline-block px-1 py-0.5 rounded border ${textColor} ${bgColor} ${borderColor} font-mono`}>
              {part}
            </span>
            {!isLast && <span className="text-muted-foreground">/</span>}
          </span>
        )
      })}
    </span>
  )
}
