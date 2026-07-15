import { spawn, type ChildProcessWithoutNullStreams } from 'node:child_process'
import { once } from 'node:events'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

export type TestServer = {
  origin: string
  token: string
  url: string
  stop: () => Promise<void>
}

const repoRoot = new URL('../../..', import.meta.url).pathname
const serverURLPattern = /http:\/\/127\.0\.0\.1:(\d+)\/\?token=([a-f0-9]+)/

export async function startTestServer(): Promise<TestServer> {
  const child = spawn('go', ['run', '.', 'web', '--host', '127.0.0.1', '--port', '0', '--no-open'], {
    cwd: repoRoot,
    env: {
      ...process.env,
      GOCACHE: join(tmpdir(), 'audiobook-organizer-go-build'),
    },
    stdio: ['ignore', 'pipe', 'pipe'],
  })

  let output = ''
  const startup = new Promise<TestServer>((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error(`Timed out waiting for web server URL.\n${output}`))
    }, 45_000)

    child.once('error', reject)
    child.once('exit', (code, signal) => {
      reject(new Error(`Web server exited before startup: code=${code} signal=${signal}\n${output}`))
    })

    child.stdout.on('data', (chunk: Buffer) => {
      output += chunk.toString()
      const match = output.match(serverURLPattern)
      if (!match) {
        return
      }
      clearTimeout(timeout)
      const port = match[1]
      const token = match[2]
      const origin = `http://127.0.0.1:${port}`
      resolve({
        origin,
        token,
        url: `${origin}/?token=${token}`,
        stop: () => stopServer(child),
      })
    })

    child.stderr.on('data', (chunk: Buffer) => {
      output += chunk.toString()
    })
  })

  return startup
}

async function stopServer(child: ChildProcessWithoutNullStreams): Promise<void> {
  if (child.exitCode !== null || child.killed) {
    return
  }

  child.kill('SIGTERM')
  await Promise.race([
    once(child, 'exit'),
    new Promise<void>((resolve) => {
      setTimeout(() => {
        if (child.exitCode === null) {
          child.kill('SIGKILL')
        }
        resolve()
      }, 5_000)
    }),
  ])
}
