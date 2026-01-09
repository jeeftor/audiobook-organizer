import { useState } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { SelectDirectory } from '../../wailsjs/go/main/App'

interface DirectoryPickerProps {
  onNext: (inputDir: string, outputDir: string) => void
}

export function DirectoryPicker({ onNext }: DirectoryPickerProps) {
  const [inputDir, setInputDir] = useState('')
  const [outputDir, setOutputDir] = useState('')
  const [error, setError] = useState('')

  const selectInputDirectory = async () => {
    try {
      const dir = await SelectDirectory("Select Input Directory")
      if (dir) {
        setInputDir(dir)
        setError('')
      }
    } catch (err) {
      setError(`Failed to select directory: ${err}`)
    }
  }

  const selectOutputDirectory = async () => {
    try {
      const dir = await SelectDirectory("Select Output Directory")
      if (dir) {
        setOutputDir(dir)
        setError('')
      }
    } catch (err) {
      setError(`Failed to select directory: ${err}`)
    }
  }

  const handleNext = () => {
    if (!inputDir) {
      setError('Please select an input directory')
      return
    }
    if (!outputDir) {
      setError('Please select an output directory')
      return
    }
    onNext(inputDir, outputDir)
  }

  return (
    <div className="flex items-center justify-center min-h-screen p-8">
      <Card className="w-full max-w-2xl">
        <CardHeader>
          <CardTitle>Select Directories</CardTitle>
          <CardDescription>
            Choose the input directory containing your audiobooks and the output directory for organized files
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <label className="text-sm font-medium">Input Directory</label>
            <div className="flex gap-2">
              <Input
                value={inputDir}
                readOnly
                placeholder="No directory selected"
                className="flex-1"
              />
              <Button onClick={selectInputDirectory}>Browse...</Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Directory containing audiobooks to organize
            </p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Output Directory</label>
            <div className="flex gap-2">
              <Input
                value={outputDir}
                readOnly
                placeholder="No directory selected"
                className="flex-1"
              />
              <Button onClick={selectOutputDirectory}>Browse...</Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Directory where organized audiobooks will be placed
            </p>
          </div>

          {error && (
            <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
              {error}
            </div>
          )}

          <div className="flex justify-end pt-4">
            <Button
              onClick={handleNext}
              disabled={!inputDir || !outputDir}
              size="lg"
            >
              Next: Scan for Audiobooks
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
