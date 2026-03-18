import { organizer } from '../../wailsjs/go/models'

export function formatAuthor(name: string, authorFormat: string): string {
  if (authorFormat === 'preserve' || authorFormat === 'first-last') return name
  const parts = name.split(' ')
  if (parts.length < 2) return name
  if (authorFormat === 'last-first') {
    const lastName = parts[parts.length - 1]
    const firstName = parts.slice(0, -1).join(' ')
    return `${lastName}, ${firstName}`
  }
  return name
}

export interface PathParts {
  parts: string[]
  author: string
  series: string | undefined
  title: string
  filename: string
}

export function buildOutputParts(
  book: organizer.Metadata,
  outputDir: string,
  layout: string,
  authorFormat: string
): PathParts {
  const filename = book.source_path?.split('/').pop() || 'unknown'
  const rawAuthor = book.authors?.[0] || 'Unknown'
  const author = formatAuthor(rawAuthor, authorFormat)
  const series = book.series?.[0]
  const title = book.title || book.album || 'Unknown'
  const parts = [outputDir || '/output']
  switch (layout) {
    case 'author-series-title':
      parts.push(author)
      if (series) parts.push(series)
      parts.push(title)
      break
    case 'author-title':
      parts.push(author, title)
      break
    case 'series-title':
      if (series) parts.push(series)
      parts.push(title)
      break
    case 'author-only':
      parts.push(author)
      break
    default:
      parts.push(author)
      if (series) parts.push(series)
      parts.push(title)
  }
  parts.push(filename)
  return { parts, author, series, title, filename }
}
