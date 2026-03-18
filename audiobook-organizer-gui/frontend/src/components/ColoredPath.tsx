interface ColoredPathProps {
  parts: string[]
  author: string
  series: string | undefined
  title: string
  filename: string
}

export function ColoredPath({ parts, author, series, title, filename }: ColoredPathProps) {
  return (
    <>
      {parts.map((part, idx) => {
        let color = 'text-foreground'
        if (idx === 0) color = 'text-muted-foreground'
        else if (part === author) color = 'text-orange-600'
        else if (part === series) color = 'text-cyan-600'
        else if (part === title) color = 'text-green-600'
        else if (part === filename) color = 'text-blue-600'
        return (
          <span key={idx}>
            <span className={color}>{part}</span>
            {idx < parts.length - 1 && <span className="text-muted-foreground">/</span>}
          </span>
        )
      })}
    </>
  )
}
