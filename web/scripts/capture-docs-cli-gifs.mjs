import { spawn } from 'node:child_process'
import { mkdir, rm } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  createCLISampleLibrary,
  removeCLISampleLibrary,
  selectedNamesFromEnv,
} from './docs-cli-fixtures.mjs'
import { optimizeGIF } from './docs-image-optimizer.mjs'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const captureDir = join(repoRoot, 'output', 'docs-visuals', 'cli')
const sampleRoot = join(repoRoot, 'output', 'docs-cli-sample')
const tapeDir = join(repoRoot, 'docs', 'visuals', 'cli')

const gifCaptures = [
  {
    name: 'organize-run',
    filename: 'cli-organize-run.gif',
    tape: 'organize-run.tape',
  },
  {
    name: 'metadata-inspect',
    filename: 'cli-metadata-inspect.gif',
    tape: 'metadata-inspect.tape',
  },
  {
    name: 'rename-preview',
    filename: 'cli-rename-preview.gif',
    tape: 'rename-preview.tape',
  },
]
const obsoleteGifFiles = ['cli-organize-dry-run.gif']

async function main() {
  await assertVHS()
  await mkdir(captureDir, { recursive: true })
  if (!process.env.ABO_DOCS_CLI_GIFS) {
    await removeObsoleteGIFs()
  }

  const generated = []
  for (const capture of gifsToRun()) {
    await rm(join(captureDir, capture.filename), { force: true })
    await createCLISampleLibrary(repoRoot, sampleRoot)
    try {
      await runVHS(capture)
      await optimizeGIF(join(captureDir, capture.filename))
      generated.push(`  output/docs-visuals/cli/${capture.filename}`)
    } finally {
      await removeCLISampleLibrary(sampleRoot)
    }
  }

  console.log('Wrote local CLI docs GIFs:')
  for (const path of generated) {
    console.log(path)
  }
}

async function removeObsoleteGIFs() {
  for (const filename of obsoleteGifFiles) {
    await rm(join(captureDir, filename), { force: true })
  }
}

function gifsToRun() {
  const selectedGifNames = selectedNamesFromEnv(
    'ABO_DOCS_CLI_GIFS',
    gifCaptures.map((capture) => capture.name),
  )
  return gifCaptures.filter((capture) => selectedGifNames.has(capture.name))
}

async function assertVHS() {
  try {
    await runCommand('vhs', ['--version'])
  } catch (error) {
    throw new Error(
      `Unable to run VHS for animated CLI GIF captures. Install Charmbracelet VHS, then rerun "make docs-cli-gifs".\n${error}`,
    )
  }
}

async function runVHS(capture) {
  const attempts = 3
  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      await runCommand('vhs', ['-q', join(tapeDir, capture.tape)], {
        cwd: repoRoot,
        env: {
          ...process.env,
          NO_COLOR: '1',
        },
      })
      return
    } catch (error) {
      if (attempt === attempts || !isRetryableVHSError(error)) {
        throw error
      }
      await delay(1000 * attempt)
    }
  }
}

function isRetryableVHSError(error) {
  const message = String(error?.message || error)
  return message.includes('could not open ttyd') || message.includes('net::ERR_EMPTY_RESPONSE')
}

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}

async function runCommand(command, args, options = {}) {
  const child = spawn(command, args, {
    cwd: options.cwd || repoRoot,
    env: options.env || process.env,
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  let output = ''
  child.stdout.on('data', (chunk) => {
    output += chunk.toString()
  })
  child.stderr.on('data', (chunk) => {
    output += chunk.toString()
  })

  const exitCode = await new Promise((resolveExit, rejectExit) => {
    child.once('error', rejectExit)
    child.once('exit', (code) => {
      resolveExit(code ?? 1)
    })
  })

  if (exitCode !== 0) {
    throw new Error(`${command} ${args.join(' ')} failed with exit code ${exitCode}.\n${output}`)
  }

  return output
}

try {
  await main()
} catch (error) {
  console.error(error)
  process.exitCode = 1
}
