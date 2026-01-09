import { useState, useEffect } from 'react'
import { DirectoryPicker } from './components/DirectoryPicker'
import { BookList } from './components/BookListNew'
import { PreviewChanges } from './components/PreviewChanges'
import { Card, CardContent, CardHeader, CardTitle } from './components/ui/card'
import { Button } from './components/ui/button'
import { GetInitialDirectories } from '../wailsjs/go/main/App'

type Step = 'directory' | 'books' | 'preview' | 'complete'

function App() {
  const [step, setStep] = useState<Step>('directory')
  const [inputDir, setInputDir] = useState('')
  const [outputDir, setOutputDir] = useState('')
  const [selectedIndices, setSelectedIndices] = useState<number[]>([])

  // Check for pre-set directories from CLI args on startup
  useEffect(() => {
    GetInitialDirectories().then((dirs) => {
      if (dirs.input_dir) {
        setInputDir(dirs.input_dir)
        if (dirs.output_dir) {
          setOutputDir(dirs.output_dir)
        }
        // Auto-advance to books step if input directory is set
        setStep('books')
      }
    }).catch((err) => {
      console.error('Failed to get initial directories:', err)
    })
  }, [])

  const handleDirectoryNext = (input: string, output: string) => {
    setInputDir(input)
    setOutputDir(output)
    setStep('books')
  }

  const handleBooksNext = (indices: number[]) => {
    setSelectedIndices(indices)
    setStep('preview')
  }

  const handleComplete = () => {
    setStep('complete')
  }

  const handleReset = () => {
    setStep('directory')
    setInputDir('')
    setOutputDir('')
    setSelectedIndices([])
  }

  return (
    <div className="min-h-screen bg-background">
      {step === 'directory' && (
        <DirectoryPicker onNext={handleDirectoryNext} />
      )}

      {step === 'books' && (
        <BookList
          inputDir={inputDir}
          outputDir={outputDir}
          onNext={handleBooksNext}
          onBack={() => setStep('directory')}
        />
      )}

      {step === 'preview' && (
        <PreviewChanges
          inputDir={inputDir}
          outputDir={outputDir}
          selectedIndices={selectedIndices}
          onBack={() => setStep('books')}
          onComplete={handleComplete}
        />
      )}

      {step === 'complete' && (
        <div className="flex items-center justify-center min-h-screen p-8">
          <Card className="w-full max-w-2xl">
            <CardHeader>
              <CardTitle className="text-center text-2xl">✅ Organization Complete!</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6 text-center">
              <p className="text-muted-foreground">
                Your audiobooks have been successfully organized.
              </p>
              <div className="flex justify-center gap-4">
                <Button onClick={handleReset} size="lg">
                  Organize More Files
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

export default App
