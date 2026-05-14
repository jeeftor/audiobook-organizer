<template>
  <main class="app-shell">
    <header class="topbar">
      <button class="icon-button" aria-label="Menu"><Menu :size="20" /></button>
      <div class="brand-mark"><AudioLines :size="22" /></div>
      <h1>Audiobook Organizer</h1>
      <div class="topbar-spacer"></div>
      <div class="status-dot" :class="{ online: health === 'ok' }"></div>
      <span class="server-status">{{ serverLabel }}</span>
      <button class="icon-button" aria-label="Settings"><Settings :size="19" /></button>
    </header>

    <div class="workspace">
      <aside class="rail" aria-label="Primary">
        <button v-for="item in navigation" :key="item.label" class="rail-item" :class="{ active: item.active }">
          <component :is="item.icon" :size="20" />
          <span>{{ item.label }}</span>
        </button>
        <div class="rail-fill"></div>
        <button class="rail-item"><Moon :size="20" /><span>Theme</span></button>
        <button class="rail-item"><Settings :size="20" /><span>Settings</span></button>
      </aside>

      <section class="settings-pane">
        <section class="panel-section">
          <h2>Source & Output</h2>
          <label>Source folder</label>
          <div class="path-input">
            <input v-model="sourceFolder" />
            <button aria-label="Browse source"><Folder :size="18" /></button>
          </div>
          <label>Output folder</label>
          <div class="path-input">
            <input v-model="outputFolder" />
            <button aria-label="Browse output"><Folder :size="18" /></button>
          </div>
          <label class="check-row"><input type="checkbox" checked /> Include subfolders</label>
        </section>

        <section class="panel-section">
          <h2>Scan Mode</h2>
          <label>Mode</label>
          <select v-model="scanMode">
            <option>Smart Scan (Recommended)</option>
            <option>metadata.json</option>
            <option>Embedded metadata</option>
            <option>Audiobookshelf metadata</option>
          </select>
          <p class="hint">Automatically detect audiobook files and metadata.</p>
          <label>File types</label>
          <div class="chip-row">
            <span>m4b</span><span>mp3</span><span>m4a</span><span>flac</span><span>wav</span>
          </div>
          <label class="check-row"><input type="checkbox" checked /> Ignore small files <input class="small-number" value="20" /> MB</label>
          <label class="check-row"><input type="checkbox" checked /> Skip already organized files</label>
          <label class="check-row"><input type="checkbox" /> Use existing metadata if available</label>
        </section>

        <section class="panel-section">
          <h2>Audiobookshelf Connection</h2>
          <label>Server URL</label>
          <div class="connected-input">
            <input v-model="absUrl" />
            <CheckCircle2 :size="18" />
          </div>
          <label>Library</label>
          <select v-model="absLibrary">
            <option>Audiobooks</option>
            <option>Archive</option>
          </select>
          <label>Path Mapping</label>
          <div class="mapping-state">
            <CheckCircle2 :size="18" />
            <div>
              <strong>Path mapping looks good</strong>
              <span>All tested paths are accessible.</span>
            </div>
            <button>Test Mapping</button>
          </div>
          <button class="primary-action" @click="simulateScan"><Play :size="18" /> Scan Library</button>
        </section>
      </section>

      <section class="table-pane">
        <div class="table-toolbar">
          <div class="search"><input placeholder="Search title, author, series..." /><Search :size="18" /></div>
          <button><Filter :size="17" /> Filters</button>
          <button><Columns3 :size="17" /> Columns</button>
          <span class="selection-count">20 of 20 selected</span>
        </div>
        <table>
          <thead>
            <tr>
              <th><input type="checkbox" checked /></th>
              <th>Title</th>
              <th>Author</th>
              <th>Series</th>
              <th>Source</th>
              <th>Current path</th>
              <th>Proposed path</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in visibleRows" :key="row.id" :class="{ selected: row.id === selected.id }" @click="selected = row">
              <td><input type="checkbox" checked /></td>
              <td class="strong">{{ row.title }}</td>
              <td>{{ row.author }}</td>
              <td>{{ row.series }}</td>
              <td>{{ row.source }}</td>
              <td>{{ row.currentPath }}</td>
              <td>{{ row.proposedPath }}</td>
              <td><span class="row-status" :class="row.status"></span>{{ row.statusText }}</td>
            </tr>
          </tbody>
        </table>
        <footer class="table-footer">
          <span>Showing 1 to {{ visibleRows.length }} of 20 items</span>
          <div class="pagination"><button disabled><ArrowLeft :size="16" /></button><button class="active-page">1</button><button>2</button><button><ArrowRight :size="16" /></button></div>
          <label>Rows per page <select><option>10</option><option>20</option></select></label>
        </footer>
      </section>

      <aside class="inspector">
        <h2>Selected: {{ selected.title }}</h2>
        <nav class="tabs"><button class="active">Metadata</button><button>Field Mapping</button><button>ABS</button></nav>
        <div class="form-grid">
          <label>Title *</label><input v-model="selected.title" />
          <label>Author *</label><input v-model="selected.author" />
          <label>Series</label><input :value="selected.series === '-' ? '' : selected.series" placeholder="Series name" />
          <label>Series Index</label><input value="1" />
          <label>Publish Date</label><input value="2021-05-04" />
          <label>Genres</label><select><option>Science Fiction, Adventure</option></select>
          <label>Language</label><select><option>English</option></select>
          <label>Narrator</label><input value="Ray Porter" />
          <label>Publisher</label><input value="Ballantine Books" />
          <label>ISBN</label><input value="9780593135218" />
          <label>Duration</label><input value="16h 10m" />
          <label>Description</label><textarea value="Ryland Grace is the sole survivor on a last-chance mission and if he fails, humanity and Earth itself will perish." />
        </div>
        <div class="inspector-actions"><button><RefreshCcw :size="17" /> Refresh Metadata</button><button class="save">Save Metadata</button></div>
      </aside>
    </div>

    <footer class="job-console">
      <div class="console-header"><span>Job Console</span><button><Trash2 :size="16" /> Clear</button></div>
      <div class="console-body">
        <div class="event-list">
          <div v-for="event in events" :key="event.time + event.event" class="event-row" :class="event.level">
            <span>{{ event.time }}</span><strong>{{ event.event }}</strong><code>{{ event.detail }}</code>
          </div>
        </div>
        <div class="job-summary">
          <span>Current Job</span><strong>Scan Library</strong>
          <span>Status</span><strong class="complete">Completed</strong>
          <span>Duration</span><strong>1.42s</strong>
          <span>Items Found</span><strong>20</strong>
          <span>Warnings</span><strong class="warning">2</strong>
          <button><FileText :size="16" /> View Full Log</button>
        </div>
      </div>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  ArrowLeft,
  ArrowRight,
  AudioLines,
  BookOpen,
  CheckCircle2,
  Columns3,
  Eye,
  FileText,
  Filter,
  Folder,
  ListChecks,
  Menu,
  Moon,
  Pencil,
  Play,
  RefreshCcw,
  Search,
  Server,
  Settings,
  Trash2,
} from 'lucide-vue-next'
import { apiGet, type HealthResponse } from './api'
import { jobEvents, rows, type BookRow } from './sampleData'

const health = ref('offline')
const sourceFolder = ref('/Volumes/Media/Audiobooks/Unsorted')
const outputFolder = ref('/Volumes/Media/Audiobooks/Organized')
const scanMode = ref('Smart Scan (Recommended)')
const absUrl = ref('http://localhost:13378')
const absLibrary = ref('Audiobooks')
const selected = ref<BookRow>(rows[0])
const events = ref(jobEvents)

const visibleRows = computed(() => rows)
const serverLabel = computed(() => (health.value === 'ok' ? 'localhost connected' : 'localhost pending'))

const navigation = [
  { label: 'Library', icon: BookOpen, active: true },
  { label: 'Scan', icon: Search, active: false },
  { label: 'Preview', icon: Eye, active: false },
  { label: 'Rename', icon: Pencil, active: false },
  { label: 'Audiobookshelf', icon: Server, active: false },
  { label: 'Jobs', icon: ListChecks, active: false },
]

function simulateScan() {
  events.value = [
    { time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }), level: 'info', event: 'Scan requested', detail: sourceFolder.value },
    ...jobEvents,
  ]
}

onMounted(async () => {
  try {
    const response = await apiGet<HealthResponse>('/api/health')
    health.value = response.status
  } catch {
    health.value = 'offline'
  }
})
</script>
