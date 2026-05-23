import { spawn } from 'node:child_process'
import { mkdir, rm, stat } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'
import {
  createCLISampleLibrary,
  removeCLISampleLibrary,
  selectedNamesFromEnv,
} from './docs-cli-fixtures.mjs'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const captureDir = join(repoRoot, 'output', 'docs-visuals', 'cli')
const sampleRoot = join(repoRoot, 'output', 'docs-cli-sample')
const cliPath = resolve(repoRoot, process.env.ABO_DOCS_CLI_BINARY || 'bin/audiobook-organizer')

const captures = [
  {
    name: 'help',
    filename: 'cli-help.png',
    title: 'Command Overview',
    command: ['--help'],
  },
  {
    name: 'organize-dry-run',
    filename: 'cli-organize-dry-run.png',
    title: 'Dry-Run Organization Preview',
    command: [
      '--dir=output/docs-cli-sample/metadata-json/source',
      '--out=output/docs-cli-sample/metadata-json/organized',
      '--dry-run',
      '--verbose',
    ],
  },
  {
    name: 'metadata-inspect',
    filename: 'cli-metadata-inspect.png',
    title: 'Metadata Inspection',
    command: ['metadata', '--dir=output/docs-cli-sample/metadata-json/source', '--verbose'],
  },
  {
    name: 'rename-preview',
    filename: 'cli-rename-preview.png',
    title: 'Rename Preview',
    command: [
      'rename',
      '--dir=output/docs-cli-sample/rename/source',
      '--use-embedded-metadata',
      '--dry-run',
    ],
  },
]

async function main() {
  await assertCliBinary()
  await createCLISampleLibrary(repoRoot, sampleRoot)
  await rm(captureDir, { recursive: true, force: true })
  await mkdir(captureDir, { recursive: true })

  let browser
  try {
    browser = await launchBrowser()
    const page = await browser.newPage({
      viewport: { width: 1500, height: 900 },
      deviceScaleFactor: 1,
    })

    const generated = []
    for (const capture of capturesToRun()) {
      const result = await runCLI(capture.command)
      if (result.exitCode !== 0) {
        throw new Error(
          `CLI capture "${capture.name}" failed with exit code ${result.exitCode}.\n${result.output}`,
        )
      }

      const path = join(captureDir, capture.filename)
      await renderCapture(page, capture, result.output, path)
      generated.push(`  output/docs-visuals/cli/${capture.filename}`)
    }

    console.log('Wrote local CLI docs captures:')
    for (const path of generated) {
      console.log(path)
    }
  } finally {
    await browser?.close()
    await removeCLISampleLibrary(sampleRoot)
  }
}

function capturesToRun() {
  const selectedCaptureNames = selectedNamesFromEnv(
    'ABO_DOCS_CLI_CAPTURES',
    captures.map((capture) => capture.name),
  )
  return captures.filter((capture) => selectedCaptureNames.has(capture.name))
}

async function assertCliBinary() {
  try {
    const info = await stat(cliPath)
    if (!info.isFile()) {
      throw new Error(`${cliPath} is not a file`)
    }
  } catch (error) {
    throw new Error(
      `Unable to find the CLI binary at ${cliPath}. Run "make docs-cli-captures" from the repository root, or set ABO_DOCS_CLI_BINARY to an existing audiobook-organizer binary.\n${error}`,
    )
  }
}

async function runCLI(args) {
  const child = spawn(cliPath, args, {
    cwd: repoRoot,
    env: {
      ...process.env,
      NO_COLOR: '1',
      TERM: 'xterm-256color',
    },
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  let stdout = ''
  let stderr = ''
  child.stdout.on('data', (chunk) => {
    stdout += chunk.toString()
  })
  child.stderr.on('data', (chunk) => {
    stderr += chunk.toString()
  })

  const exitCode = await new Promise((resolveExit, rejectExit) => {
    const timeout = setTimeout(() => {
      child.kill('SIGKILL')
      rejectExit(new Error(`Timed out running CLI command: ${formatCommand(args)}`))
    }, 30_000)

    child.once('error', (error) => {
      clearTimeout(timeout)
      rejectExit(error)
    })
    child.once('exit', (code) => {
      clearTimeout(timeout)
      resolveExit(code ?? 1)
    })
  })

  return {
    exitCode,
    output: normalizeCLIOutput(`${stdout}${stderr}`),
  }
}

function normalizeCLIOutput(output) {
  return output
    .replace(/\x1B\[[0-?]*[ -/]*[@-~]/g, '')
    .replaceAll('\r\n', '\n')
    .replaceAll(`${repoRoot}/`, '')
    .replaceAll(repoRoot, '.')
    .replace(/(Duration:\s+).*/g, '$1<1s')
    .split('\n')
    .map((line) => line.replace(/\s+$/u, ''))
    .join('\n')
    .trimEnd()
}

async function renderCapture(page, capture, output, path) {
  await page.setContent(renderHTML(capture, output), { waitUntil: 'load' })
  const terminal = page.locator('.terminal')
  await terminal.screenshot({ path })
}

function renderHTML(capture, output) {
  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <style>
    :root {
      color-scheme: dark;
      font-family:
        ui-monospace,
        SFMono-Regular,
        Menlo,
        Monaco,
        Consolas,
        "Liberation Mono",
        "Apple Color Emoji",
        "Segoe UI Emoji",
        monospace;
      background: #f6f7f9;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      padding: 32px;
      background: #f6f7f9;
    }

    .terminal {
      width: 1400px;
      overflow: hidden;
      border: 1px solid #273043;
      border-radius: 8px;
      background: #111827;
      box-shadow: 0 18px 48px rgb(15 23 42 / 24%);
    }

    .chrome {
      display: flex;
      align-items: center;
      justify-content: space-between;
      min-height: 44px;
      padding: 0 18px;
      color: #cbd5e1;
      background: #1f2937;
      border-bottom: 1px solid #334155;
      font: 600 14px/1.2 system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      letter-spacing: 0;
    }

    .dots {
      display: flex;
      gap: 8px;
    }

    .dot {
      width: 12px;
      height: 12px;
      border-radius: 999px;
    }

    .red { background: #ef4444; }
    .yellow { background: #f59e0b; }
    .green { background: #22c55e; }

    pre {
      margin: 0;
      padding: 24px 28px 28px;
      white-space: pre-wrap;
      overflow-wrap: anywhere;
      color: #e5e7eb;
      background: #111827;
      font-size: 17px;
      line-height: 1.48;
      letter-spacing: 0;
    }

    .prompt {
      color: #67e8f9;
    }
  </style>
</head>
<body>
  <section class="terminal" aria-label="${escapeHTML(capture.title)}">
    <div class="chrome">
      <div class="dots" aria-hidden="true">
        <span class="dot red"></span>
        <span class="dot yellow"></span>
        <span class="dot green"></span>
      </div>
      <span>${escapeHTML(capture.title)}</span>
      <span>docs capture</span>
    </div>
    <pre><span class="prompt">$</span> ${escapeHTML(formatCommand(capture.command))}

${escapeHTML(output)}</pre>
  </section>
</body>
</html>`
}

function formatCommand(args) {
  return ['audiobook-organizer', ...args.map(shellEscape)].join(' ')
}

function shellEscape(value) {
  if (/^[A-Za-z0-9_./:=@-]+$/.test(value)) {
    return value
  }
  return `"${value.replaceAll('"', '\\"')}"`
}

function escapeHTML(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

async function launchBrowser() {
  try {
    return await chromium.launch({
      executablePath: process.env.ABO_DOCS_BROWSER_EXECUTABLE_PATH || process.env.PUPPETEER_EXECUTABLE_PATH || undefined,
    })
  } catch (error) {
    throw new Error(
      `Unable to launch Chromium for CLI docs captures. Run "cd web && npm run install:browsers" or set ABO_DOCS_BROWSER_EXECUTABLE_PATH to a local Chrome/Chrome Headless Shell binary.\n${error}`,
    )
  }
}

try {
  await main()
} catch (error) {
  console.error(error)
  process.exitCode = 1
}
