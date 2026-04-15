import { useState, useEffect } from 'react'
import { Music } from 'lucide-react'
// @ts-ignore — GetCoverArt will exist once the backend is regenerated
import { GetCoverArt } from '../../wailsjs/go/main/App'

interface CoverArtProps {
  bookIdx: number | null
}

export function CoverArt({ bookIdx }: CoverArtProps) {
  const [dataUrl, setDataUrl] = useState<string>('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (bookIdx === null || bookIdx < 0) {
      setDataUrl('')
      return
    }

    setLoading(true)
    setDataUrl('')

    GetCoverArt(bookIdx)
      .then((result: string) => {
        setDataUrl(result || '')
      })
      .catch((err: unknown) => {
        console.error('Failed to get cover art:', err)
        setDataUrl('')
      })
      .finally(() => {
        setLoading(false)
      })
  }, [bookIdx])

  if (loading) {
    return (
      <div className="w-full aspect-square max-w-[200px] mx-auto rounded-lg bg-muted animate-pulse flex items-center justify-center">
        <Music className="h-8 w-8 text-muted-foreground/40" />
      </div>
    )
  }

  if (dataUrl) {
    return (
      <div className="w-full max-w-[200px] mx-auto">
        <img
          src={dataUrl}
          alt="Cover art"
          className="w-full aspect-square object-cover rounded-lg shadow-sm border border-border"
        />
      </div>
    )
  }

  return (
    <div className="w-full aspect-square max-w-[200px] mx-auto rounded-lg bg-muted border border-border flex flex-col items-center justify-center gap-2">
      <Music className="h-8 w-8 text-muted-foreground/50" />
      <span className="text-xs text-muted-foreground">No Cover Art</span>
    </div>
  )
}
