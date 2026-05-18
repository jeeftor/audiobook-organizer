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
            <input v-model="sourceFolder" aria-label="Source folder" />
            <label v-if="activeWorkflow !== 'rename'">Output folder</label>
            <input v-if="activeWorkflow !== 'rename'" v-model="outputFolder" aria-label="Output folder" />
            <label>{{ currentWorkflow.modeLabel }}</label>
            <select
              v-model="scanMode"
              aria-label="Metadata source"
              :disabled="optionsLoading && workflowScanModes.length === 0"
            >
              <option v-if="optionsLoading && workflowScanModes.length === 0" value="" disabled>Loading options</option>
              <option v-for="mode in workflowScanModes" :key="mode.value" :value="mode.value">{{ mode.label }}</option>
            </select>
            <p class="hint">{{ currentWorkflow.configureHint }}</p>
          </div>

          <div v-if="activeWorkflow === 'rename'" class="panel-section">
            <h3>Rename Template</h3>
            <label>Template</label>
            <input v-model="renameTemplate" aria-label="Rename template" />
            <label class="check-row"><input v-model="renameRecursive" type="checkbox" /> Include subfolders</label>
            <label class="check-row"><input v-model="preservePath" type="checkbox" /> Preserve relative folders</label>
          </div>

          <div v-else class="panel-section">
            <h3>Organization Rules</h3>
            <label>Layout</label>
            <select v-model="layout" aria-label="Layout" :disabled="optionsLoading && layoutOptions.length === 0">
              <option v-if="optionsLoading && layoutOptions.length === 0" value="" disabled>Loading options</option>
              <option v-for="option in layoutOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
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
              <span>{{ absSetupState }}</span>
            </div>
          </div>
        </section>

        <section v-else-if="activeStage === 'preview'" class="preview-layout">
          <div class="preview-empty">
            <Eye :size="30" />
            <h3>{{ previewHeading }}</h3>
            <p>{{ currentWorkflow.previewCopy }}</p>
            <template v-if="activeWorkflow === 'organize'">
              <button
                class="primary-action"
                :disabled="organizePreviewStatus === 'loading'"
                @click="createOrganizePreview"
              >
                <Play :size="18" /> {{ organizePreviewActionLabel }}
              </button>
              <button
                v-if="organizePreviewStatus === 'success'"
                class="secondary-action"
                @click="reviewOrganizePreview"
              >
                Review Preview & Continue
              </button>
              <p v-if="organizePreviewError" class="inline-alert">{{ organizePreviewError }}</p>
            </template>
            <template v-else-if="activeWorkflow === 'rename'">
              <button class="primary-action" :disabled="renamePreviewStatus === 'loading'" @click="createRenamePreview">
                <Play :size="18" /> {{ renamePreviewActionLabel }}
              </button>
              <button v-if="renamePreviewStatus === 'success'" class="secondary-action" @click="reviewRenamePreview">
                Review Candidates & Continue
              </button>
              <p v-if="renamePreviewError" class="inline-alert">{{ renamePreviewError }}</p>
            </template>
            <button v-else class="primary-action" @click="markPreviewReady">
              <Play :size="18" /> Mark Preview Reviewed
            </button>
          </div>
          <div v-if="activeWorkflow === 'organize'" class="preview-checklist">
            <h3>Preview Summary</h3>
            <p v-if="!organizePreview">No organize preview has run.</p>
            <template v-else>
              <div class="result-grid compact">
                <span>Metadata found</span><strong>{{ organizePreview.summary.MetadataFound.length }}</strong>
                <span>Planned moves</span><strong>{{ organizePreview.summary.Moves.length }}</strong>
                <span>Warnings</span><strong>{{ organizePreview.summary.MetadataMissing.length }}</strong>
                <span>Log path</span><strong>{{ organizePreview.log_path || 'Not created during dry-run' }}</strong>
              </div>
              <ul v-if="organizePreview.summary.MetadataMissing.length > 0" class="warning-list">
                <li v-for="missing in organizePreview.summary.MetadataMissing.slice(0, 4)" :key="missing">{{ missing }}</li>
              </ul>
              <div v-if="organizePreview.summary.Moves.length > 0" class="move-list">
                <div v-for="move in organizePreview.summary.Moves.slice(0, 4)" :key="move.from + move.to">
                  <span>{{ move.from }}</span>
                  <strong>{{ move.to }}</strong>
                </div>
              </div>
            </template>
          </div>
          <div v-else-if="activeWorkflow === 'rename'" class="preview-checklist">
            <h3>Rename Preview Summary</h3>
            <p v-if="!renamePreview">No rename preview has run.</p>
            <template v-else>
              <div class="result-grid compact">
                <span>Files scanned</span><strong>{{ renamePreview.summary.FilesScanned }}</strong>
                <span>Candidates</span><strong>{{ renamePreview.candidates.length }}</strong>
                <span>Conflicts</span><strong>{{ renamePreview.summary.ConflictsFound }}</strong>
                <span>Skipped</span><strong>{{ renamePreview.summary.FilesSkipped }}</strong>
                <span>Errors</span><strong>{{ renamePreview.summary.Errors.length }}</strong>
              </div>
              <ul v-if="renamePreview.summary.Errors.length > 0" class="warning-list">
                <li v-for="error in renamePreview.summary.Errors.slice(0, 4)" :key="error">{{ error }}</li>
              </ul>
              <div v-if="renamePreview.candidates.length > 0" class="move-list">
                <div
                  v-for="candidate in renamePreview.candidates.slice(0, 5)"
                  :key="candidate.CurrentPath + candidate.ProposedPath"
                  :class="{ warning: candidate.IsConflict || candidate.IsNoOp || !!candidate.Error }"
                >
                  <span>{{ candidate.CurrentPath }}</span>
                  <strong>{{ candidate.ProposedPath }}</strong>
                  <em v-if="candidate.IsConflict">Conflict</em>
                  <em v-else-if="candidate.IsNoOp">Skipped: unchanged</em>
                  <em v-else-if="candidate.Error">{{ candidate.Error }}</em>
                </div>
              </div>
            </template>
          </div>
          <div v-else class="preview-checklist">
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
          <p v-if="activeWorkflow === 'organize' && organizeRunError" class="inline-alert">{{ organizeRunError }}</p>
          <p v-if="activeWorkflow === 'rename'" class="inline-alert">
            Rename execution is deferred until a backend run endpoint is implemented. This workflow can only preview
            candidates right now.
          </p>
          <button
            class="danger-action"
            :disabled="isRunActionDisabled"
            @click="activeWorkflow === 'organize' ? runOrganize() : undefined"
          >
            <Play :size="18" /> {{ currentWorkflow.runAction }}
          </button>
        </section>

        <section v-else-if="activeWorkflow === 'organize' && organizeRun" class="review-layout">
          <h3>Organize Run Complete</h3>
          <p>The reviewed organize plan finished with backend results.</p>
          <div class="result-grid">
            <span>Job status</span><strong>Complete</strong>
            <span>Files organized</span><strong>{{ organizeRun.summary.Moves.length }}</strong>
            <span>Undo log</span><strong>{{ organizeRun.log_path || 'Not reported' }}</strong>
            <span>Warnings</span><strong>{{ organizeRun.summary.MetadataMissing.length }}</strong>
          </div>
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
import { computed, onMounted, ref, watch } from 'vue'
import {
  AlertTriangle,
  AudioLines,
  Eye,
  FilePenLine,
  FolderInput,
  Play,
  Server,
} from 'lucide-vue-next'
import {
  apiGet,
  apiPost,
  type FieldMapping,
  type HealthResponse,
  type Option,
  type OrganizerConfig,
  type OrganizePreviewResponse,
  type OrganizeRunResponse,
  type OptionsResponse,
  type RenameConfig,
  type RenamePreviewResponse,
  type WebConfig,
} from './api'

type WorkflowId = 'organize' | 'rename' | 'abs'
type StageId = 'configure' | 'preview' | 'run' | 'review'
type LoadState = 'loading' | 'ready' | 'fallback'
type CredentialState = 'empty' | 'redacted'
type RequestState = 'idle' | 'loading' | 'success' | 'error'

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
    previewCopy: 'Create a backend dry-run preview before the organize run stage can unlock.',
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
    previewCopy: 'Create real rename candidates from the backend before considering any filesystem action.',
    runTitle: 'Rename Execution Deferred',
    runCopy: 'The web UI does not expose a rename execution endpoint yet. Candidate review stays available here.',
    runAction: 'Rename Execution Deferred',
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
const defaultFieldMapping: FieldMapping = {
  title_field: 'title',
  series_field: 'series',
  author_fields: ['authors'],
  track_field: 'track',
}

const health = ref('offline')
const configState = ref<LoadState>('loading')
const optionsState = ref<LoadState>('loading')
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
const absCredentialState = ref<CredentialState>('empty')
const organizerDefaults = ref<OrganizerConfig | null>(null)
const renameDefaults = ref<RenameConfig | null>(null)
const layouts = ref<Option[]>([])
const scanModes = ref<Option[]>([])
const organizePreview = ref<OrganizePreviewResponse | null>(null)
const organizeRun = ref<OrganizeRunResponse | null>(null)
const renamePreview = ref<RenamePreviewResponse | null>(null)
const organizePreviewStatus = ref<RequestState>('idle')
const organizeRunStatus = ref<RequestState>('idle')
const renamePreviewStatus = ref<RequestState>('idle')
const organizePreviewError = ref('')
const organizeRunError = ref('')
const renamePreviewError = ref('')
const events = ref<ActivityEvent[]>([
  { time: 'Pending', level: 'info', event: 'Startup checks', detail: 'Loading server health, config, and options.' },
])

const currentWorkflow = computed(() => workflows.find((workflow) => workflow.id === activeWorkflow.value) ?? workflows[0])
const currentStage = computed(() => stageText[activeStage.value])
const optionsLoading = computed(() => optionsState.value === 'loading')
const layoutOptions = computed(() => {
  if (layouts.value.length > 0) {
    return layouts.value
  }
  return optionsState.value === 'fallback' ? defaultLayouts : []
})
const scanModeOptions = computed(() => {
  if (scanModes.value.length > 0) {
    return scanModes.value
  }
  return optionsState.value === 'fallback' ? defaultScanModes : []
})
const workflowScanModes = computed(() => {
  return scanModeOptions.value.filter((mode) => activeWorkflow.value === 'abs' || mode.value !== 'abs')
})
const serverLabel = computed(() =>
  [
    health.value === 'ok' ? 'localhost connected' : 'localhost pending',
    stateLabel('config', configState.value),
    stateLabel('options', optionsState.value),
  ].join(' · '),
)
const absSetupState = computed(() => {
  if (absCredentialState.value === 'redacted') {
    return 'Saved ABS credentials are redacted and will need a fresh token when connection controls are wired.'
  }
  return 'Connection tests and library loading are tracked in the follow-up control wiring issue.'
})
const previewHeading = computed(() => {
  if (activeWorkflow.value === 'rename') {
    if (renamePreviewStatus.value === 'success') {
      return 'Rename preview ready'
    }
    if (renamePreviewStatus.value === 'error') {
      return 'Rename preview needs attention'
    }
    return 'Create a rename preview'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStatus.value === 'success') {
    return 'Organize preview ready'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStatus.value === 'error') {
    return 'Preview needs attention'
  }
  return activeWorkflow.value === 'organize' ? 'Create an organize preview' : 'Dry-run preview first'
})
const organizePreviewActionLabel = computed(() =>
  organizePreviewStatus.value === 'loading' ? 'Creating Preview' : 'Create Dry-run Preview',
)
const renamePreviewActionLabel = computed(() =>
  renamePreviewStatus.value === 'loading' ? 'Creating Preview' : 'Create Rename Preview',
)
const isRunActionDisabled = computed(() => {
  if (activeWorkflow.value === 'rename') {
    return true
  }
  if (activeWorkflow.value === 'abs') {
    return !previewReady.value
  }
  return !previewReady.value || organizeRunStatus.value === 'loading'
})

function selectWorkflow(workflow: WorkflowId) {
  activeWorkflow.value = workflow
  activeStage.value = 'configure'
  previewReady.value = false
  if (workflow === 'organize') {
    resetOrganizeResults()
  } else if (workflow === 'rename') {
    resetRenameResults()
  }
  ensureScanModeFitsWorkflow()
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
  if (stage !== 'run') {
    return false
  }
  if (activeWorkflow.value === 'organize') {
    return !previewReady.value || organizeRunStatus.value === 'loading'
  }
  return !previewReady.value
}

function markPreviewReady() {
  previewReady.value = true
  activeStage.value = 'run'
  events.value = [
    { time: now(), level: 'ok', event: 'Preview reviewed', detail: `${currentWorkflow.value.label} run stage unlocked.` },
    ...events.value,
  ]
}

function stateLabel(name: string, state: LoadState) {
  if (state === 'ready') {
    return `${name} ready`
  }
  if (state === 'fallback') {
    return `${name} fallback`
  }
  return `${name} loading`
}

function addEvent(event: ActivityEvent) {
  events.value = [event, ...events.value]
}

async function createOrganizePreview() {
  organizePreviewStatus.value = 'loading'
  organizePreviewError.value = ''
  organizeRun.value = null
  organizeRunError.value = ''
  previewReady.value = false

  try {
    if (!sourceFolder.value.trim() || !outputFolder.value.trim()) {
      throw new Error('Source and output folders are required for organize preview.')
    }
    const response = normalizeOrganizeResponse(
      await apiPost<OrganizePreviewResponse>('/api/organize/preview', {
        config: buildOrganizerConfig(true),
      }),
    )
    organizePreview.value = response
    organizePreviewStatus.value = 'success'
    addEvent({
      time: now(),
      level: 'ok',
      event: 'Organize preview ready',
      detail: `${response.summary.Moves.length} planned move(s), ${response.summary.MetadataMissing.length} warning(s).`,
    })
  } catch (error) {
    organizePreview.value = null
    organizePreviewStatus.value = 'error'
    organizePreviewError.value = error instanceof Error ? error.message : 'Preview failed.'
    addEvent({ time: now(), level: 'warn', event: 'Organize preview failed', detail: organizePreviewError.value })
  }
}

function reviewOrganizePreview() {
  if (organizePreviewStatus.value !== 'success') {
    return
  }
  previewReady.value = true
  activeStage.value = 'run'
  addEvent({ time: now(), level: 'ok', event: 'Preview reviewed', detail: 'Organize run stage unlocked.' })
}

async function createRenamePreview() {
  renamePreviewStatus.value = 'loading'
  renamePreviewError.value = ''
  previewReady.value = false

  try {
    if (!sourceFolder.value.trim()) {
      throw new Error('Source folder is required for rename preview.')
    }
    if (!renameTemplate.value.trim()) {
      throw new Error('Rename template is required for preview.')
    }
    const response = normalizeRenameResponse(
      await apiPost<RenamePreviewResponse>('/api/rename/preview', {
        config: buildRenameConfig(),
      }),
    )
    renamePreview.value = response
    renamePreviewStatus.value = 'success'
    addEvent({
      time: now(),
      level: 'ok',
      event: 'Rename preview ready',
      detail: `${response.candidates.length} candidate(s), ${response.summary.ConflictsFound} conflict(s).`,
    })
  } catch (error) {
    renamePreview.value = null
    renamePreviewStatus.value = 'error'
    renamePreviewError.value = error instanceof Error ? error.message : 'Rename preview failed.'
    addEvent({ time: now(), level: 'warn', event: 'Rename preview failed', detail: renamePreviewError.value })
  }
}

function reviewRenamePreview() {
  if (renamePreviewStatus.value !== 'success') {
    return
  }
  previewReady.value = true
  activeStage.value = 'run'
  addEvent({
    time: now(),
    level: 'ok',
    event: 'Rename candidates reviewed',
    detail: 'Rename execution remains deferred until the backend supports it.',
  })
}

async function runOrganize() {
  if (organizePreviewStatus.value !== 'success' || !previewReady.value || organizeRunStatus.value === 'loading') {
    return
  }
  if (!window.confirm('Run Organize will change files using the reviewed preview. Continue?')) {
    return
  }

  organizeRunStatus.value = 'loading'
  organizeRunError.value = ''
  try {
    const response = normalizeOrganizeResponse(
      await apiPost<OrganizeRunResponse>('/api/organize/run', {
        config: buildOrganizerConfig(false),
      }),
    )
    organizeRun.value = response
    organizeRunStatus.value = 'success'
    activeStage.value = 'review'
    addEvent({
      time: now(),
      level: 'ok',
      event: 'Organize run complete',
      detail: `${response.summary.Moves.length} file operation(s).`,
    })
  } catch (error) {
    organizeRunStatus.value = 'error'
    organizeRunError.value = error instanceof Error ? error.message : 'Organize run failed.'
    addEvent({ time: now(), level: 'warn', event: 'Organize run failed', detail: organizeRunError.value })
  }
}

function buildOrganizerConfig(dryRun: boolean): OrganizerConfig {
  const defaults = organizerDefaults.value
  return {
    base_dir: sourceFolder.value.trim(),
    output_dir: outputFolder.value.trim(),
    replace_space: defaults?.replace_space ?? '',
    dry_run: dryRun,
    remove_empty: removeEmpty.value,
    use_embedded_metadata: shouldUseEmbeddedMetadata(),
    flat: shouldUseFlatMode(),
    skip_errors: defaults?.skip_errors ?? false,
    layout: layout.value,
    author_format: defaults?.author_format || 'first-last',
    field_mapping: defaults?.field_mapping ?? defaultFieldMapping,
    allowed_source_paths: defaults?.allowed_source_paths,
  }
}

function buildRenameConfig(): RenameConfig {
  const defaults = renameDefaults.value
  return {
    base_dir: sourceFolder.value.trim(),
    template: renameTemplate.value.trim(),
    dry_run: true,
    author_format: defaults?.author_format || 'first-last',
    recursive: renameRecursive.value,
    field_mapping: defaults?.field_mapping ?? defaultFieldMapping,
    replace_space: defaults?.replace_space ?? '',
    strict_mode: defaults?.strict_mode ?? false,
    preserve_path: preservePath.value,
    use_embedded_metadata: shouldUseEmbeddedMetadata(),
  }
}

function shouldUseEmbeddedMetadata() {
  return useEmbeddedMetadata.value || scanMode.value === 'embedded-directory' || scanMode.value === 'embedded-file'
}

function shouldUseFlatMode() {
  if (scanMode.value === 'embedded-file') {
    return true
  }
  return organizerDefaults.value?.flat ?? false
}

function resetOrganizeResults() {
  organizePreview.value = null
  organizeRun.value = null
  organizePreviewStatus.value = 'idle'
  organizeRunStatus.value = 'idle'
  organizePreviewError.value = ''
  organizeRunError.value = ''
}

function resetRenameResults() {
  renamePreview.value = null
  renamePreviewStatus.value = 'idle'
  renamePreviewError.value = ''
}

function normalizeOrganizeResponse<T extends OrganizePreviewResponse | OrganizeRunResponse>(response: T): T {
  return {
    ...response,
    summary: {
      MetadataFound: response.summary.MetadataFound ?? [],
      MetadataMissing: response.summary.MetadataMissing ?? [],
      Moves: response.summary.Moves ?? [],
      EmptyDirsRemoved: response.summary.EmptyDirsRemoved ?? [],
    },
  }
}

function normalizeRenameResponse(response: RenamePreviewResponse): RenamePreviewResponse {
  return {
    ...response,
    candidates: response.candidates ?? [],
    summary: {
      FilesScanned: response.summary.FilesScanned ?? 0,
      FilesRenamed: response.summary.FilesRenamed ?? 0,
      FilesSkipped: response.summary.FilesSkipped ?? 0,
      ConflictsFound: response.summary.ConflictsFound ?? 0,
      Errors: response.summary.Errors ?? [],
    },
  }
}

function ensureScanModeFitsWorkflow() {
  if (workflowScanModes.value.some((mode) => mode.value === scanMode.value)) {
    return
  }
  scanMode.value = workflowScanModes.value[0]?.value || 'json'
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
    const config = await apiGet<WebConfig>('/api/config/initial')
    organizerDefaults.value = config.organizer
    renameDefaults.value = config.rename
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
    absCredentialState.value = config.abs?.token === 'redacted' ? 'redacted' : 'empty'
    configState.value = 'ready'
    addEvent({ time: now(), level: 'ok', event: 'Config loaded', detail: 'Startup config is ready.' })
  } catch {
    configState.value = 'fallback'
    addEvent({ time: now(), level: 'warn', event: 'Config unavailable', detail: 'Using local defaults.' })
  }

  try {
    const options = await apiGet<OptionsResponse>('/api/config/options')
    layouts.value = Array.isArray(options.layouts) ? options.layouts : []
    scanModes.value = Array.isArray(options.scan_modes) ? options.scan_modes : []
    optionsState.value = 'ready'
    ensureScanModeFitsWorkflow()
    addEvent({ time: now(), level: 'ok', event: 'Options loaded', detail: 'Layout and scan mode options are ready.' })
  } catch {
    optionsState.value = 'fallback'
    ensureScanModeFitsWorkflow()
    addEvent({ time: now(), level: 'warn', event: 'Options unavailable', detail: 'Using built-in option labels.' })
  }
})

watch([sourceFolder, outputFolder, scanMode, layout, useEmbeddedMetadata, removeEmpty], () => {
  if (activeWorkflow.value !== 'organize') {
    return
  }
  previewReady.value = false
  resetOrganizeResults()
})

watch([sourceFolder, scanMode, useEmbeddedMetadata, renameTemplate, renameRecursive, preservePath], () => {
  if (activeWorkflow.value !== 'rename') {
    return
  }
  previewReady.value = false
  resetRenameResults()
})
</script>
