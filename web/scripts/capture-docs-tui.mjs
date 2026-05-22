import { spawn } from 'node:child_process'
import { constants } from 'node:fs'
import { access, copyFile, mkdir, readdir, rm, stat, symlink, writeFile } from 'node:fs/promises'
import { homedir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const captureDir = join(repoRoot, 'output', 'docs-visuals', 'tui')
const sampleRoot = join(repoRoot, 'output', 'docs-tui-sample')
const tapeDir = join(repoRoot, 'docs', 'visuals', 'tui')
const browserShimDir = join(repoRoot, 'output', 'docs-vhs-browser-bin')
const containerImage = process.env.ABO_DOCS_TUI_VHS_IMAGE || 'ghcr.io/charmbracelet/vhs:v0.11.0'

const captures = [
  {
    name: 'organize-preview',
    gif: 'tui-organize-preview.gif',
    png: 'tui-organize-preview.png',
    tape: 'organize-preview.tape',
  },
  {
    name: 'rename-field-mapping',
    gif: 'tui-rename-field-mapping.gif',
    png: 'tui-rename-field-mapping.png',
    tape: 'rename-field-mapping.tape',
  },
]

async function main() {
  await assertCommand('ffmpeg', ['-version'], 'Install ffmpeg, then rerun "make docs-tui-captures".')
  const mode = resolveVHSMode()
  if (mode === 'native') {
    await assertCommand('vhs', ['--version'], 'Install Charmbracelet VHS, then rerun "make docs-tui-captures".')
  } else {
    await assertCommand('docker', ['version'], 'Start Docker or set ABO_DOCS_TUI_VHS_MODE=native on a Linux host.')
  }

  if (!process.env.ABO_DOCS_TUI_CAPTURES) {
    await rm(captureDir, { recursive: true, force: true })
  }
  await mkdir(captureDir, { recursive: true })
  const vhsEnv = mode === 'native' ? await createVHSEnv() : undefined

  const generated = []
  for (const capture of capturesToRun()) {
    await rm(join(captureDir, capture.gif), { force: true })
    await rm(join(captureDir, capture.png), { force: true })
    await createTUISampleLibrary()

    try {
      await runVHS(capture, { env: vhsEnv, mode })
      await extractFinalFrame(capture)
      generated.push(`  output/docs-visuals/tui/${capture.gif}`)
      generated.push(`  output/docs-visuals/tui/${capture.png}`)
    } finally {
      await removeTUISampleLibrary()
    }
  }

  console.log('Wrote local TUI docs captures:')
  for (const path of generated) {
    console.log(path)
  }
}

function resolveVHSMode() {
  const mode = process.env.ABO_DOCS_TUI_VHS_MODE || 'auto'
  if (!['auto', 'container', 'native'].includes(mode)) {
    throw new Error('ABO_DOCS_TUI_VHS_MODE must be one of: auto, container, native')
  }

  if (mode === 'container') {
    return 'container'
  }
  if (mode === 'native') {
    if (process.platform === 'darwin' && process.env.ABO_DOCS_TUI_ALLOW_MACOS_NATIVE_VHS !== '1') {
      throw new Error(
        'Native VHS is disabled on macOS because Rod can launch /Applications/Google Chrome.app and trigger crash dialogs. Use container mode or set ABO_DOCS_TUI_ALLOW_MACOS_NATIVE_VHS=1 to override.',
      )
    }
    return 'native'
  }

  return process.platform === 'darwin' ? 'container' : 'native'
}

function capturesToRun() {
  const selectedNames = selectedNamesFromEnv(
    'ABO_DOCS_TUI_CAPTURES',
    captures.map((capture) => capture.name),
  )
  return captures.filter((capture) => selectedNames.has(capture.name))
}

function selectedNamesFromEnv(envName, validNames) {
  const selected = process.env[envName]
  if (!selected) {
    return new Set(validNames)
  }

  const validNameSet = new Set(validNames)
  const names = selected.split(',').map((name) => name.trim()).filter(Boolean)
  const unknownNames = names.filter((name) => !validNameSet.has(name))
  if (unknownNames.length > 0) {
    throw new Error(
      `Unknown ${envName} value(s): ${unknownNames.join(', ')}. Valid values: ${validNames.join(', ')}`,
    )
  }
  return new Set(names)
}

async function createTUISampleLibrary() {
  await removeTUISampleLibrary()
  await mkdir(join(sampleRoot, 'embedded', 'source'), { recursive: true })
  await mkdir(join(sampleRoot, 'embedded', 'organized'), { recursive: true })
  await mkdir(join(sampleRoot, 'metadata-json', 'organized'), { recursive: true })

  await createEmbeddedSampleBooks()
  await createMetadataBook({
    directory: 'charles-dexter-ward',
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
  await createMetadataBook({
    directory: 'falstaffs-wedding',
    sourceFile: 'testdata/mp3flat/falstaffswedding1766version_1_kenrick_64kb.mp3',
    targetFile: '01-act-1.mp3',
    metadata: {
      title: 'Falstaffs Wedding',
      authors: ['William Kenrick'],
      series: ['Public Domain Drama #1'],
      narrators: ['LibriVox Volunteers'],
      publishedYear: '1766',
    },
  })
}

async function createEmbeddedSampleBooks() {
  for (const file of [
    'charlesdexterward_01_lovecraft_64kb.mp3',
    'falstaffswedding1766version_1_kenrick_64kb.mp3',
    'perouse_01_scott_64kb.mp3',
  ]) {
    await copyFile(
      join(repoRoot, 'testdata', 'mp3flat', file),
      join(sampleRoot, 'embedded', 'source', file),
    )
  }
}

async function removeTUISampleLibrary() {
  await rm(sampleRoot, { recursive: true, force: true })
}

async function createMetadataBook({ directory, sourceFile, targetFile, metadata }) {
  const bookDir = join(sampleRoot, 'metadata-json', 'source', directory)
  await mkdir(bookDir, { recursive: true })
  await writeFile(join(bookDir, 'metadata.json'), JSON.stringify(metadata, null, 2))
  await copyFile(join(repoRoot, sourceFile), join(bookDir, targetFile))
}

async function runVHS(capture, { env, mode }) {
  if (mode === 'container') {
    await runContainerVHS(capture)
  } else {
    await runCommand('vhs', ['-q', join(tapeDir, capture.tape)], {
      env,
    })
  }
  await waitForFile(join(captureDir, capture.gif))
}

async function runContainerVHS(capture) {
  const args = [
    'run',
    '--rm',
    '--volume',
    `${repoRoot}:/workspace`,
    '--workdir',
    '/workspace',
    '--env',
    'TERM=xterm-256color',
    '--env',
    'HOME=/tmp',
  ]

  const uid = typeof process.getuid === 'function' ? process.getuid() : undefined
  const gid = typeof process.getgid === 'function' ? process.getgid() : undefined
  if (uid !== undefined && gid !== undefined) {
    args.push('--user', `${uid}:${gid}`)
  }

  args.push(containerImage, '-q', join('docs', 'visuals', 'tui', capture.tape))

  try {
    await runCommand('docker', args)
  } catch (error) {
    throw new Error(
      `Unable to run VHS in Docker with image ${containerImage}. Pull or mirror the image, or set ABO_DOCS_TUI_VHS_IMAGE to an available VHS image.\n${error}`,
    )
  }
}

async function createVHSEnv() {
  const env = {
    ...process.env,
    TERM: process.env.TERM || 'xterm-256color',
  }
  const browserPath = await findPreferredBrowser()
  if (!browserPath) {
    return env
  }

  await rm(browserShimDir, { recursive: true, force: true })
  await mkdir(browserShimDir, { recursive: true })
  for (const name of ['google-chrome', 'chromium', 'chromium-browser']) {
    await symlink(browserPath, join(browserShimDir, name))
  }

  return {
    ...env,
    PATH: `${browserShimDir}:${env.PATH || ''}`,
  }
}

async function findPreferredBrowser() {
  const configuredPath = process.env.ABO_DOCS_BROWSER_EXECUTABLE_PATH || process.env.PUPPETEER_EXECUTABLE_PATH
  if (configuredPath) {
    return assertExecutableBrowser(configuredPath)
  }

  for (const root of [
    join(homedir(), '.cache', 'puppeteer', 'chrome-headless-shell'),
    join(homedir(), 'chrome-headless-shell'),
  ]) {
    const browserPath = await findChromeHeadlessShell(root)
    if (browserPath) {
      return browserPath
    }
  }

  return ''
}

async function assertExecutableBrowser(browserPath) {
  try {
    await access(browserPath, constants.X_OK)
  } catch (error) {
    throw new Error(`Configured docs browser is not executable: ${browserPath}\n${error}`)
  }
  return browserPath
}

async function findChromeHeadlessShell(root) {
  try {
    const platforms = await readdir(root, { withFileTypes: true })
    const platformNames = platforms
      .filter((entry) => entry.isDirectory())
      .map((entry) => entry.name)
      .sort()
      .reverse()

    for (const platformName of platformNames) {
      const platformRoot = join(root, platformName)
      const browserNames = await readdir(platformRoot, { withFileTypes: true })
      for (const browserName of browserNames.filter((entry) => entry.isDirectory()).map((entry) => entry.name).sort()) {
        const candidate = join(platformRoot, browserName, 'chrome-headless-shell')
        try {
          return await assertExecutableBrowser(candidate)
        } catch {
          // Keep searching; cached browser directories can be platform-specific.
        }
      }
    }
  } catch {
    return ''
  }

  return ''
}

async function extractFinalFrame(capture) {
  await runCommand('ffmpeg', [
    '-y',
    '-sseof',
    '-0.2',
    '-i',
    join(captureDir, capture.gif),
    '-frames:v',
    '1',
    join(captureDir, capture.png),
  ])
}

async function waitForFile(path) {
  const deadline = Date.now() + 5_000
  let lastError

  while (Date.now() < deadline) {
    try {
      await stat(path)
      return
    } catch (error) {
      lastError = error
      await new Promise((resolve) => setTimeout(resolve, 100))
    }
  }

  throw new Error(`Timed out waiting for VHS output ${path}.\n${lastError}`)
}

async function assertCommand(command, args, installHint) {
  try {
    await runCommand(command, args)
  } catch (error) {
    throw new Error(`Unable to run ${command}. ${installHint}\n${error}`)
  }
}

async function runCommand(command, args, options = {}) {
  const child = spawn(command, args, {
    cwd: repoRoot,
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
