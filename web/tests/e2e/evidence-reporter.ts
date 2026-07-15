import type { FullConfig, FullResult, Reporter, TestCase, TestResult } from '@playwright/test/reporter'
import { mkdirSync, writeFileSync } from 'node:fs'
import { dirname, relative } from 'node:path'

type EvidenceRecord = {
  test: string
  project: string
  outcome: string
  duration_ms: number
  evidence: string[]
}

const summaryPath = 'test-results/ui-e2e-summary.json'

class EvidenceReporter implements Reporter {
  private readonly records = new Map<string, EvidenceRecord>()

  onBegin(_config: FullConfig): void {
    mkdirSync(dirname(summaryPath), { recursive: true })
  }

  onTestEnd(test: TestCase, result: TestResult): void {
    const evidence = result.attachments
      .filter((attachment) => attachment.path)
      .map((attachment) => relative(process.cwd(), attachment.path!))

    const project = test.parent.project()?.name ?? 'unknown'
    const title = test.titlePath().filter(Boolean).join(' > ')

    this.records.set(`${project}:${title}`, {
      test: title,
      project,
      outcome: result.status,
      duration_ms: result.duration,
      evidence,
    })
  }

  onEnd(_result: FullResult): void {
    writeFileSync(summaryPath, `${JSON.stringify({ generated_at: new Date().toISOString(), tests: [...this.records.values()] }, null, 2)}\n`)
  }
}

export default EvidenceReporter
