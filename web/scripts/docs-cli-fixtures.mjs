import { copyFile, mkdir, rm, writeFile } from 'node:fs/promises'
import { join } from 'node:path'

export const cliCaptureNames = [
  'help',
  'organize-dry-run',
  'metadata-inspect',
  'rename-preview',
]

export function selectedNamesFromEnv(envName, validNames = cliCaptureNames) {
  const selectedNames = new Set(
    (process.env[envName] || '')
      .split(',')
      .map((name) => name.trim())
      .filter(Boolean),
  )

  if (selectedNames.size === 0) {
    return new Set(validNames)
  }

  const knownNames = new Set(validNames)
  const unknownNames = [...selectedNames].filter((name) => !knownNames.has(name))
  if (unknownNames.length > 0) {
    throw new Error(
      `Unknown CLI capture name(s): ${unknownNames.join(', ')}. Valid names: ${validNames.join(', ')}`,
    )
  }

  return selectedNames
}

export async function createCLISampleLibrary(repoRoot, sampleRoot) {
  await rm(sampleRoot, { recursive: true, force: true })
  await mkdir(join(sampleRoot, 'metadata-json', 'organized'), { recursive: true })
  await mkdir(join(sampleRoot, 'rename', 'source'), { recursive: true })

  await createMetadataBook(repoRoot, sampleRoot, {
    sourceFile: 'testdata/mp3flat/charlesdexterward_01_lovecraft_64kb.mp3',
    targetFile: '01-chapter-1.mp3',
    metadata: {
      title: 'The Case of Charles Dexter Ward',
      authors: ['H. P. Lovecraft'],
      series: ['LibriVox Horror Classics #1'],
      narrators: ['LibriVox Volunteers'],
      publishedYear: '1927',
    },
  })

  await copyFile(
    join(repoRoot, 'testdata/mp3flat/falstaffswedding1766version_1_kenrick_64kb.mp3'),
    join(sampleRoot, 'rename', 'source', 'falstaffswedding1766version_1_kenrick_64kb.mp3'),
  )
}

export async function removeCLISampleLibrary(sampleRoot) {
  await rm(sampleRoot, { recursive: true, force: true })
}

async function createMetadataBook(repoRoot, sampleRoot, { sourceFile, targetFile, metadata }) {
  const bookDir = join(sampleRoot, 'metadata-json', 'source')
  await mkdir(bookDir, { recursive: true })
  await writeFile(join(bookDir, 'metadata.json'), JSON.stringify(metadata, null, 2))
  await copyFile(join(repoRoot, sourceFile), join(bookDir, targetFile))
}
