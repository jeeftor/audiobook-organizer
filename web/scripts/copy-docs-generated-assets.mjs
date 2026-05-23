import { cp, mkdir, stat } from 'node:fs/promises'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webRoot = resolve(scriptDir, '..')
const repoRoot = resolve(webRoot, '..')
const generatedVisualsRoot = join(repoRoot, 'output', 'docs-visuals')
const starlightGeneratedRoot = join(repoRoot, 'output', 'docs-starlight', 'assets', 'generated')
const publicGeneratedRoot = join(webRoot, 'public', 'assets', 'generated')

try {
  const info = await stat(generatedVisualsRoot)
  if (info.isDirectory()) {
    for (const destination of [starlightGeneratedRoot, publicGeneratedRoot]) {
      await mkdir(destination, { recursive: true })
      await cp(generatedVisualsRoot, destination, { recursive: true })
    }
    console.log(`Copied docs visuals to ${relative(repoRoot, starlightGeneratedRoot)}`)
    console.log(`Copied docs visuals to ${relative(repoRoot, publicGeneratedRoot)}`)
  }
} catch {
  console.log('Skipped generated docs visuals copy because output/docs-visuals does not exist.')
}
