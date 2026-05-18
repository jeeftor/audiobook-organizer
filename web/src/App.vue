<template>
  <main class="app-shell">
    <header class="topbar">
      <div class="brand-mark"><AudioLines :size="22" /></div>
      <div>
        <h1>Audiobook Organizer</h1>
        <span class="app-subtitle">Local workflow console</span>
      </div>
      <div class="topbar-spacer"></div>
      <div class="status-dot" :class="{ online: health === 'ok' }"></div>
      <span class="server-status">{{ serverLabel }}</span>
    </header>

    <section class="workflow-switcher" aria-label="Workflow">
      <button
        v-for="workflow in workflows"
        :key="workflow.id"
        class="workflow-card"
        :class="{ active: activeWorkflow === workflow.id }"
        @click="selectWorkflow(workflow.id)"
      >
        <component :is="workflow.icon" :size="22" />
        <span>{{ workflow.label }}</span>
        <small>{{ workflow.description }}</small>
      </button>
    </section>

    <div class="workspace">
      <aside class="stage-rail" aria-label="Workflow stages">
        <button
          v-for="stage in stages"
          :key="stage.id"
          class="stage-item"
          :class="{ active: activeStage === stage.id, locked: isStageLocked(stage.id) }"
          :disabled="isStageLocked(stage.id)"
          @click="activeStage = stage.id"
        >
          <span class="stage-index">{{ stage.index }}</span>
          <span>
            <strong>{{ stage.label }}</strong>
            <small>{{ stage.description }}</small>
          </span>
        </button>
      </aside>

      <section class="stage-panel">
        <div class="stage-header">
          <span>{{ currentWorkflow.label }}</span>
          <h2>{{ currentStage.heading }}</h2>
          <p>{{ currentStage.copy }}</p>
        </div>

        <section v-if="activeStage === 'configure'" class="configure-grid">
          <div class="panel-section">
            <h3>{{ currentWorkflow.configureTitle }}</h3>
            <label>Source folder</label>
            <input v-model="sourceFolder" />
            <label v-if="activeWorkflow !== 'rename'">Output folder</label>
            <input v-if="activeWorkflow !== 'rename'" v-model="outputFolder" />
            <label>{{ currentWorkflow.modeLabel }}</label>
            <select v-model="scanMode">
              <option v-for="mode in scanModes" :key="mode.value" :value="mode.value">{{ mode.label }}</option>
            </select>
            <p class="hint">{{ currentWorkflow.configureHint }}</p>
          </div>

          <div v-if="activeWorkflow === 'rename'" class="panel-section">
            <h3>Rename Template</h3>
            <label>Template</label>
            <input v-model="renameTemplate" />
            <label class="check-row"><input v-model="renameRecursive" type="checkbox" /> Include subfolders</label>
            <label class="check-row"><input v-model="preservePath" type="checkbox" /> Preserve relative folders</label>
          </div>

          <div v-else class="panel-section">
            <h3>Organization Rules</h3>
            <label>Layout</label>
            <select v-model="layout">
              <option v-for="option in layouts" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <label class="check-row"><input v-model="useEmbeddedMetadata" type="checkbox" /> Use embedded metadata</label>
            <label class="check-row"><input v-model="removeEmpty" type="checkbox" /> Remove empty source folders after run</label>
          </div>

          <div v-if="activeWorkflow === 'abs'" class="panel-section">
            <h3>Audiobookshelf</h3>
            <label>Server URL</label>
            <input v-model="absUrl" placeholder="http://localhost:13378" />
            <label>Library ID</label>
            <input v-model="absLibrary" />
            <label>Path Mapping</label>
            <div class="deferred-state">
              <Server :size="18" />
              <span>Connection tests and library loading are tracked in the follow-up control wiring issue.</span>
            </div>
          </div>
        </section>

        <section v-else-if="activeStage === 'preview'" class="preview-layout">
          <div class="preview-empty">
            <Eye :size="30" />
            <h3>Dry-run preview first</h3>
            <p>{{ currentWorkflow.previewCopy }}</p>
            <button class="primary-action" @click="markPreviewReady"><Play :size="18" /> Mark Preview Reviewed</button>
          </div>
          <div class="preview-checklist">
            <h3>Review Gate</h3>
            <ul>
              <li>Filesystem-changing actions stay locked until a dry-run preview has been reviewed.</li>
              <li>Warnings and proposed paths belong in this stage, not on the first screen.</li>
              <li>Backend preview wiring remains scoped to #64.</li>
            </ul>
          </div>
        </section>

        <section v-else-if="activeStage === 'run'" class="run-layout">
          <div class="warning-panel">
            <AlertTriangle :size="28" />
            <div>
              <h3>{{ currentWorkflow.runTitle }}</h3>
              <p>{{ currentWorkflow.runCopy }}</p>
            </div>
          </div>
          <button class="danger-action" :disabled="!previewReady">
            <Play :size="18" /> {{ currentWorkflow.runAction }}
          </button>
        </section>

        <section v-else class="review-layout">
          <h3>Result Review</h3>
          <p>Run results, log paths, undo guidance, and follow-up warnings will appear here after an executed job.</p>
          <div class="result-grid">
            <span>Job status</span><strong>Waiting for run</strong>
            <span>Undo log</span><strong>Not created</strong>
            <span>Warnings</span><strong>None yet</strong>
          </div>
        </section>
      </section>

      <aside class="activity-panel">
        <h2>Activity</h2>
        <div class="event-list">
          <div v-for="event in events" :key="event.time + event.event" class="event-row" :class="event.level">
            <span>{{ event.time }}</span>
            <strong>{{ event.event }}</strong>
            <code>{{ event.detail }}</code>
          </div>
        </div>
      </aside>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  AlertTriangle,
  AudioLines,
  Eye,
  FilePenLine,
  FolderInput,
  Play,
  Server,
} from 'lucide-vue-next'
import { apiGet, type HealthResponse } from './api'

type WorkflowId = 'organize' | 'rename' | 'abs'
type StageId = 'configure' | 'preview' | 'run' | 'review'

type Option = {
  value: string
  label: string
}

type InitialConfigResponse = {
  initial?: {
    input_dir?: string
    output_dir?: string
  }
  organizer?: {
    base_dir?: string
    output_dir?: string
    layout?: string
    use_embedded_metadata?: boolean
    remove_empty?: boolean
  }
  rename?: {
    template?: string
    recursive?: boolean
    preserve_path?: boolean
  }
  abs?: {
    url?: string
    library_id?: string
  }
}

type OptionsResponse = {
  layouts?: Option[]
  scan_modes?: Option[]
}

type ActivityEvent = {
  time: string
  level: 'ok' | 'warn' | 'info'
  event: string
  detail: string
}

const workflows = [
  {
    id: 'organize' as const,
    label: 'Organize',
    description: 'Move or copy books into a clean library layout.',
    icon: FolderInput,
    configureTitle: 'Local Library Setup',
    configureHint: 'Choose the source, output, and metadata mode before creating a dry-run preview.',
    modeLabel: 'Metadata source',
    previewCopy: 'The next implementation pass will connect this stage to the organize preview endpoint.',
    runTitle: 'Run Organize',
    runCopy: 'This action changes files and stays locked until the preview stage has been reviewed.',
    runAction: 'Run Organize',
  },
  {
    id: 'rename' as const,
    label: 'Rename',
    description: 'Preview template-based filename changes in place.',
    icon: FilePenLine,
    configureTitle: 'Rename Setup',
    configureHint: 'Set the folder and filename template before creating rename candidates.',
    modeLabel: 'Metadata source',
    previewCopy: 'Rename candidates and conflicts belong here before any file operation is available.',
    runTitle: 'Run Rename',
    runCopy: 'Rename execution must stay behind candidate review and conflict checks.',
    runAction: 'Run Rename',
  },
  {
    id: 'abs' as const,
    label: 'Audiobookshelf',
    description: 'Use ABS metadata and path mapping as part of organize.',
    icon: Server,
    configureTitle: 'ABS-Assisted Setup',
    configureHint: 'ABS connection and mapping controls are shown only for this workflow.',
    modeLabel: 'Metadata source',
    previewCopy: 'ABS item loading, mapping validation, and organize preview should converge here.',
    runTitle: 'Run ABS-Assisted Organize',
    runCopy: 'Library scan triggers and cleanup actions stay staged behind mapping and preview review.',
    runAction: 'Run ABS Organize',
  },
]

const stages = [
  { id: 'configure' as const, index: '1', label: 'Configure & Scan', description: 'Choose workflow inputs' },
  { id: 'preview' as const, index: '2', label: 'Preview', description: 'Review dry-run output' },
  { id: 'run' as const, index: '3', label: 'Run', description: 'Execute after review' },
  { id: 'review' as const, index: '4', label: 'Review', description: 'Check logs and undo' },
]

const stageText: Record<StageId, { heading: string; copy: string }> = {
  configure: {
    heading: 'Configure and scan setup',
    copy: 'Pick the workflow-specific inputs before any preview or filesystem action is available.',
  },
  preview: {
    heading: 'Review a dry-run preview',
    copy: 'Preview is the required checkpoint between setup and mutating operations.',
  },
  run: {
    heading: 'Execute the reviewed plan',
    copy: 'Run actions remain locked until preview review is complete.',
  },
  review: {
    heading: 'Review results and recovery details',
    copy: 'Completed jobs should surface summaries, warnings, logs, and undo paths here.',
  },
}

const defaultLayouts: Option[] = [{ value: 'author-series-title', label: 'Author / Series / Title' }]
const defaultScanModes: Option[] = [
  { value: 'json', label: 'metadata.json' },
  { value: 'embedded-directory', label: 'Embedded metadata by directory' },
  { value: 'embedded-file', label: 'Embedded metadata by file' },
  { value: 'abs', label: 'Audiobookshelf metadata' },
]

const health = ref('offline')
const activeWorkflow = ref<WorkflowId>('organize')
const activeStage = ref<StageId>('configure')
const previewReady = ref(false)
const sourceFolder = ref('')
const outputFolder = ref('')
const scanMode = ref('json')
const layout = ref('author-series-title')
const useEmbeddedMetadata = ref(false)
const removeEmpty = ref(false)
const renameTemplate = ref('{author} - {series} {series_number} - {title}')
const renameRecursive = ref(true)
const preservePath = ref(true)
const absUrl = ref('')
const absLibrary = ref('main')
const layouts = ref<Option[]>(defaultLayouts)
const scanModes = ref<Option[]>(defaultScanModes)
const events = ref<ActivityEvent[]>([
  { time: 'Pending', level: 'info', event: 'Configure workflow', detail: 'Choose a workflow to begin.' },
])

const currentWorkflow = computed(() => workflows.find((workflow) => workflow.id === activeWorkflow.value) ?? workflows[0])
const currentStage = computed(() => stageText[activeStage.value])
const serverLabel = computed(() => (health.value === 'ok' ? 'localhost connected' : 'localhost pending'))

function selectWorkflow(workflow: WorkflowId) {
  activeWorkflow.value = workflow
  activeStage.value = 'configure'
  previewReady.value = false
  events.value = [
    {
      time: now(),
      level: 'info',
      event: `${currentWorkflow.value.label} selected`,
      detail: 'Configure inputs before preview.',
    },
  ]
}

function isStageLocked(stage: StageId) {
  return stage === 'run' && !previewReady.value
}

function markPreviewReady() {
  previewReady.value = true
  activeStage.value = 'run'
  events.value = [
    { time: now(), level: 'ok', event: 'Preview reviewed', detail: `${currentWorkflow.value.label} run stage unlocked.` },
    ...events.value,
  ]
}

function now() {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

onMounted(async () => {
  try {
    const response = await apiGet<HealthResponse>('/api/health')
    health.value = response.status
  } catch {
    health.value = 'offline'
  }

  try {
    const config = await apiGet<InitialConfigResponse>('/api/config/initial')
    sourceFolder.value = config.initial?.input_dir || config.organizer?.base_dir || ''
    outputFolder.value = config.initial?.output_dir || config.organizer?.output_dir || ''
    layout.value = config.organizer?.layout || layout.value
    useEmbeddedMetadata.value = config.organizer?.use_embedded_metadata ?? false
    removeEmpty.value = config.organizer?.remove_empty ?? false
    renameTemplate.value = config.rename?.template || renameTemplate.value
    renameRecursive.value = config.rename?.recursive ?? true
    preservePath.value = config.rename?.preserve_path ?? true
    absUrl.value = config.abs?.url ?? ''
    absLibrary.value = config.abs?.library_id || absLibrary.value
  } catch {
    events.value = [{ time: now(), level: 'warn', event: 'Config unavailable', detail: 'Using local defaults.' }, ...events.value]
  }

  try {
    const options = await apiGet<OptionsResponse>('/api/config/options')
    layouts.value = options.layouts?.length ? options.layouts : defaultLayouts
    scanModes.value = options.scan_modes?.length ? options.scan_modes : defaultScanModes
  } catch {
    events.value = [{ time: now(), level: 'warn', event: 'Options unavailable', detail: 'Using built-in option labels.' }, ...events.value]
  }
})
</script>
