import { spawn } from 'node:child_process'
import { rename, rm, stat } from 'node:fs/promises'
import { basename, dirname, extname, join } from 'node:path'

const missingOptionalCommands = new Set()

export async function createWebP(sourcePath, options = {}) {
  const outputPath = options.outputPath || replaceExtension(sourcePath, '.webp')
  const quality = String(options.quality || 82)

  await runCommand('cwebp', [
    '-quiet',
    '-q',
    quality,
    '-m',
    '6',
    '-mt',
    sourcePath,
    '-o',
    outputPath,
  ])

  await logSizeChange('webp', sourcePath, outputPath)
  return outputPath
}

export async function optimizeGIF(gifPath) {
  if (process.env.ABO_DOCS_GIF_OPTIMIZE === '0') {
    return
  }

  const tempPath = join(dirname(gifPath), `.${process.pid}-${Date.now()}-${basenameWithoutExtension(gifPath)}.optimized.gif`)
  const result = await runCommand('gifsicle', ['-O3', gifPath, '-o', tempPath], { optional: true })
  if (!result.ok) {
    return
  }

  const original = await stat(gifPath)
  const optimized = await stat(tempPath)
  if (optimized.size < original.size) {
    await rename(tempPath, gifPath)
    console.log(`Optimized GIF ${gifPath}: ${formatBytes(original.size)} -> ${formatBytes(optimized.size)}`)
    return
  }

  await rm(tempPath, { force: true })
  console.log(`Kept GIF ${gifPath}: optimizer output was not smaller`)
}

function replaceExtension(path, extension) {
  const currentExtension = extname(path)
  return currentExtension ? path.slice(0, -currentExtension.length) + extension : `${path}${extension}`
}

function basenameWithoutExtension(path) {
  const filename = basename(path)
  const extension = extname(filename)
  return extension ? filename.slice(0, -extension.length) : filename
}

async function logSizeChange(kind, sourcePath, outputPath) {
  const source = await stat(sourcePath)
  const output = await stat(outputPath)
  const percent = source.size === 0 ? 0 : Math.round((1 - output.size / source.size) * 100)
  console.log(
    `Generated ${kind.toUpperCase()} ${outputPath}: ${formatBytes(source.size)} -> ${formatBytes(output.size)} (${percent}% smaller)`,
  )
}

async function runCommand(command, args, options = {}) {
  const child = spawn(command, args, {
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
  }).catch((error) => {
    if (options.optional && error?.code === 'ENOENT') {
      warnMissingOptionalCommand(command)
      return null
    }
    if (error?.code === 'ENOENT') {
      throw new Error(`Unable to run ${command}. Install the WebP tools package so generated docs images can be optimized.`)
    }
    throw error
  })

  if (exitCode === null) {
    return { ok: false }
  }
  if (exitCode !== 0) {
    throw new Error(`${command} ${args.join(' ')} failed with exit code ${exitCode}.\n${output}`)
  }

  return { ok: true, output }
}

function warnMissingOptionalCommand(command) {
  if (missingOptionalCommands.has(command)) {
    return
  }
  missingOptionalCommands.add(command)
  console.warn(`Skipped optional docs GIF optimization because ${command} is not installed.`)
}

function formatBytes(size) {
  if (size < 1024) {
    return `${size} B`
  }
  return `${Math.round(size / 1024)} KiB`
}
