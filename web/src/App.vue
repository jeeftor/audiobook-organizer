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

    <p v-if="!hasWebSessionToken" class="inline-alert session-token-alert" role="alert">
      This web session link is missing its token. Reopen the complete startup URL.
    </p>

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

        <section v-if="activeStage === 'configure'" class="setup-preview-grid">
          <div class="setup-controls">
            <div class="panel-section">
              <h3>{{ currentWorkflow.configureTitle }}</h3>
              <label>Source folder</label>
              <div
                class="path-control"
                :class="{ dragging: activePathDropTarget === 'source' }"
                data-path-field="source"
                @dragenter.prevent="activePathDropTarget = 'source'"
                @dragover.prevent="activePathDropTarget = 'source'"
                @dragleave.prevent="activePathDropTarget = null"
                @drop.prevent="handlePathDrop('source', $event)"
              >
                <input v-model="sourceFolder" aria-label="Source folder" @input="clearPathMessage('source')" />
                <button
                  class="icon-button"
                  type="button"
                  aria-label="Choose source folder"
                  title="Choose source folder"
                  @click="openPathPicker('source')"
                >
                  <FolderOpen :size="16" />
                </button>
                <input
                  ref="sourceFolderPicker"
                  class="file-picker"
                  type="file"
                  aria-label="Source folder directory picker"
                  tabindex="-1"
                  webkitdirectory
                  directory
                  multiple
                  @change="handlePathPickerChange('source', $event)"
                />
              </div>
              <p v-if="sourcePathMessage" class="hint path-message">{{ sourcePathMessage }}</p>
              <label v-if="activeWorkflow !== 'rename'">Output folder</label>
              <div
                v-if="activeWorkflow !== 'rename'"
                class="path-control"
                :class="{ dragging: activePathDropTarget === 'output' }"
                data-path-field="output"
                @dragenter.prevent="activePathDropTarget = 'output'"
                @dragover.prevent="activePathDropTarget = 'output'"
                @dragleave.prevent="activePathDropTarget = null"
                @drop.prevent="handlePathDrop('output', $event)"
              >
                <input v-model="outputFolder" aria-label="Output folder" @input="clearPathMessage('output')" />
                <button
                  class="icon-button"
                  type="button"
                  aria-label="Choose output folder"
                  title="Choose output folder"
                  @click="openPathPicker('output')"
                >
                  <FolderOpen :size="16" />
                </button>
                <input
                  ref="outputFolderPicker"
                  class="file-picker"
                  type="file"
                  aria-label="Output folder directory picker"
                  tabindex="-1"
                  webkitdirectory
                  directory
                  multiple
                  @change="handlePathPickerChange('output', $event)"
                />
              </div>
              <p v-if="activeWorkflow !== 'rename' && outputPathMessage" class="hint path-message">
                {{ outputPathMessage }}
              </p>
              <label>{{ currentWorkflow.modeLabel }}</label>
              <div class="metadata-source-control" role="radiogroup" aria-label="Metadata source">
                <button
                  v-for="mode in workflowScanModes"
                  :key="mode.value"
                  class="metadata-source-option"
                  :class="{ active: scanMode === mode.value }"
                  role="radio"
                  type="button"
                  :aria-checked="scanMode === mode.value"
                  :disabled="optionsLoading && workflowScanModes.length === 0"
                  @click="selectScanMode(mode.value)"
                >
                  <FilePenLine v-if="mode.value === 'json'" :size="18" />
                  <FolderInput v-else-if="mode.value === 'embedded-directory'" :size="18" />
                  <AudioLines v-else :size="18" />
                  <span>{{ mode.label }}</span>
                </button>
                <div v-if="optionsLoading && workflowScanModes.length === 0" class="metadata-source-loading">
                  Loading options
                </div>
              </div>
              <p class="hint">{{ currentWorkflow.configureHint }}</p>
            </div>

            <div v-if="activeWorkflow !== 'abs'" class="panel-section field-mapping-panel">
              <h3>Metadata Field Mapping</h3>
              <p class="hint">Choose a source preset, then adjust field names to match your metadata.</p>
              <label>Mapping preset</label>
              <select
                :value="activeFieldMappingPreset"
                aria-label="Field mapping preset"
                @change="applyFieldMappingPreset(($event.target as HTMLSelectElement).value)"
              >
                <option value="custom">Custom mapping</option>
                <option v-for="preset in fieldMappingPresets" :key="preset.value" :value="preset.value">
                  {{ preset.label }}
                </option>
              </select>
              <label>Title field</label>
              <input v-model="activeFieldMapping.title_field" aria-label="Title field mapping" list="field-mapping-fields" />
              <label>Author field or fields</label>
              <input
                :value="activeFieldMapping.author_fields?.join(', ') ?? ''"
                aria-label="Author field mapping"
                list="field-mapping-fields"
                @input="updateAuthorFieldMapping(($event.target as HTMLInputElement).value)"
              />
              <p class="hint">Separate multiple author fields with commas.</p>
              <label>Series field</label>
              <input v-model="activeFieldMapping.series_field" aria-label="Series field mapping" list="field-mapping-fields" />
              <label>Track field</label>
              <input v-model="activeFieldMapping.track_field" aria-label="Track field mapping" list="field-mapping-fields" />
              <label>Disc field</label>
              <input v-model="activeFieldMapping.disc_field" aria-label="Disc field mapping" list="field-mapping-fields" />
              <datalist id="field-mapping-fields">
                <option v-for="field in fieldMappingFieldNames" :key="field" :value="field" />
              </datalist>
            </div>

            <div v-if="activeWorkflow === 'rename'" class="panel-section">
              <h3>Rename Template</h3>
              <TemplateBuilder
                v-model="renameTemplate"
                label="Filename template"
                placeholder="{author} - {series} {series_number} - {title}"
                :fields="renameTemplateFields"
                empty-text="Select fields to build a filename template."
              />
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
              <TemplateBuilder
                v-if="layout === customLayoutValue"
                v-model="layoutTemplate"
                label="Custom layout template"
                placeholder="{author}/{series}/{series-count} - {title}"
                :fields="layoutTemplateFields"
                empty-text="Select fields to build a custom path."
                hint="Use slashes to create folders. Metadata values are sanitized inside each folder segment."
              />
              <div v-else class="field-color-legend" aria-label="Preview color legend">
                <span class="legend-token author">Author</span>
                <span class="legend-token series">Series</span>
                <span class="legend-token title">Title</span>
                <span class="legend-token other">Other</span>
              </div>
              <label class="check-row"><input v-model="removeEmpty" type="checkbox" /> Remove empty source folders after run</label>
            </div>

          <div v-if="showABSSetup" class="panel-section">
            <h3>Audiobookshelf</h3>
            <label>Server URL</label>
            <input v-model="absUrl" aria-label="ABS server URL" placeholder="http://localhost:13378" />
            <label>API Token</label>
            <input v-model="absToken" aria-label="ABS API token" autocomplete="off" type="password" />
            <p v-if="absCredentialState === 'redacted' && !absToken" class="hint">
              Saved ABS credentials are redacted. Enter a fresh token before sending requests.
            </p>
            <div class="action-row">
              <button class="secondary-action" :disabled="absLibrariesStatus === 'loading'" @click="loadABSLibraries">
                <Server :size="18" /> {{ absLibrariesActionLabel }}
              </button>
            </div>
            <p v-if="absLibrariesError" class="inline-alert">{{ absLibrariesError }}</p>
            <template v-if="absLibrariesStatus === 'success' && absLibraries.length > 0">
              <label>Library</label>
              <select v-model="absLibrary" aria-label="ABS library">
                <option value="" disabled>Select a library</option>
                <option
                  v-for="library in absLibraries"
                  :key="library.id"
                  :value="library.id"
                >
                  {{ library.name || library.id }} ({{ library.id }})
                </option>
              </select>
            </template>
            <p v-else-if="absLibrariesStatus === 'success'" class="hint">
              The ABS server responded, but no libraries were returned.
            </p>
            <p v-else class="hint">Test the ABS URL and token before choosing a library.</p>
            <div v-if="absLibraries.length > 0" class="library-list">
              <div
                v-for="library in absLibraries"
                :key="library.id"
                class="library-option"
                :class="{ selected: absLibrary === library.id }"
              >
                <strong>{{ library.name || library.id }}</strong>
                <span>{{ library.id }} · {{ library.mediaType || 'library' }}</span>
              </div>
            </div>
            <label>Custom Header</label>
            <div class="split-row">
              <input v-model="absHeaderName" aria-label="ABS header name" placeholder="Header name" />
              <input
                v-model="absHeaderValue"
                aria-label="ABS header value"
                autocomplete="off"
                placeholder="Header value"
                type="password"
              />
            </div>
            <label>SQLite Database Path</label>
            <input v-model="absSQLitePath" aria-label="ABS SQLite database path" />
            <p class="hint">Leave SQLite empty to validate manual path mappings.</p>
            <label>Path Mapping</label>
            <div class="mapping-list">
              <div v-for="(mapping, index) in absPathMappings" :key="index" class="mapping-row">
                <input v-model="mapping.abs_prefix" aria-label="ABS path prefix" placeholder="/audiobooks" />
                <input v-model="mapping.local_prefix" aria-label="Local path prefix" placeholder="/host/audiobooks" />
                <button
                  class="icon-button"
                  type="button"
                  :disabled="absPathMappings.length === 1"
                  aria-label="Remove path mapping"
                  @click="removeABSPathMapping(index)"
                >
                  <Trash2 :size="16" />
                </button>
              </div>
            </div>
            <div class="action-row">
              <button class="secondary-action" type="button" @click="addABSPathMapping">
                <Plus :size="18" /> Add Mapping
              </button>
              <button
                class="primary-action"
                :disabled="!absLibrarySelectionReady || absPathStatus === 'loading'"
                type="button"
                @click="testABSPathMappings"
              >
                <Server :size="18" /> {{ absPathActionLabel }}
              </button>
            </div>
            <p v-if="absPathError" class="inline-alert">{{ absPathError }}</p>
            <div class="deferred-state" :class="{ ready: absSetupReady }">
              <Server :size="18" />
              <span>{{ absSetupState }}</span>
            </div>
            <div v-if="absResolvedMappings.length > 0" class="move-list">
              <div v-for="mapping in absResolvedMappings" :key="mapping.abs_prefix + mapping.local_prefix">
                <span>{{ mapping.abs_prefix }}</span>
                <strong>{{ mapping.local_prefix }}</strong>
              </div>
            </div>
          </div>

            <div class="configure-actions">
              <p v-if="pathValidationError" class="inline-alert configure-path-alert">{{ pathValidationError }}</p>
            </div>
          </div>

          <div class="setup-preview-column">
            <div v-if="activeWorkflow === 'organize'" class="preview-checklist setup-preview-window" :class="{ stale: activePreviewStale }">
              <div class="preview-window-header">
                <div>
                  <h3>{{ previewHeading }}</h3>
                  <p>{{ setupPreviewCopy }}</p>
                </div>
                <button
                  v-if="organizePreviewStatus === 'success'"
                  class="primary-action compact-action"
                  type="button"
                  :disabled="!canOpenOrganizeReview"
                  @click="openOrganizeReview"
                >
                  Review & Run
                </button>
              </div>
              <p v-if="organizePreviewError" class="inline-alert">{{ organizePreviewError }}</p>
              <p v-if="!organizePreview">No organize preview has run.</p>
              <template v-else>
                <div class="result-grid compact">
                  <span>Metadata found</span><strong>{{ organizePreview.summary.MetadataFound.length }}</strong>
                  <span>Planned moves</span><strong>{{ organizePreview.summary.Moves.length }}</strong>
                  <span>Selected moves</span><strong>{{ selectedOrganizeMoveCount }}</strong>
                  <span>Warnings</span><strong>{{ organizePreview.summary.MetadataMissing.length }}</strong>
                  <template v-if="organizePreview.log_path">
                    <span>Log path</span><strong>{{ organizePreview.log_path }}</strong>
                  </template>
                </div>
                <ul v-if="organizePreview.summary.MetadataMissing.length > 0" class="warning-list">
                  <li v-for="missing in organizePreview.summary.MetadataMissing.slice(0, 4)" :key="missing">
                    {{ displayOrganizeSourcePath(missing) }}
                  </li>
                </ul>
                <div v-if="organizePreview.summary.Moves.length > 0" class="move-list">
                  <div
                    v-for="move in organizePreview.summary.Moves.slice(0, 5)"
                    :key="move.from + move.to"
                  >
                    <span>{{ displayOrganizeSourcePath(move.from) }}</span>
                    <strong class="colored-path">
                      <template v-for="(part, index) in coloredOrganizeTargetParts(move.to)" :key="index">
                        <span :class="part.kind">{{ part.value }}</span>
                      </template>
                    </strong>
                  </div>
                </div>
              </template>
            </div>
            <div v-else-if="activeWorkflow === 'rename'" class="preview-checklist setup-preview-window" :class="{ stale: activePreviewStale }">
              <div class="preview-window-header">
                <div>
                  <h3>{{ previewHeading }}</h3>
                  <p>{{ setupPreviewCopy }}</p>
                </div>
                <button
                  v-if="renamePreviewStatus === 'success'"
                  class="primary-action compact-action"
                  type="button"
                  :disabled="!canOpenRenameReview"
                  @click="openRenameReview"
                >
                  Review & Run
                </button>
              </div>
              <p v-if="renamePreviewError" class="inline-alert">{{ renamePreviewError }}</p>
              <p v-if="!renamePreview">No rename preview has run.</p>
              <template v-else>
                <div class="result-grid compact">
                  <span>Files scanned</span><strong>{{ renamePreview.summary.FilesScanned }}</strong>
                  <span>Candidates</span><strong>{{ renamePreview.candidates.length }}</strong>
                  <span>Selected files</span><strong>{{ selectedRenameCandidateCount }}</strong>
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
                    <span>{{ displayRenameSourcePath(candidate.CurrentPath) }}</span>
                    <strong class="colored-path">
                      <template v-for="(part, index) in coloredRenameTargetParts(candidate.ProposedPath)" :key="index">
                        <span :class="part.kind">{{ part.value }}</span>
                      </template>
                    </strong>
                    <em v-if="candidate.IsConflict">Conflict</em>
                    <em v-else-if="candidate.IsNoOp">Skipped: unchanged</em>
                    <em v-else-if="candidate.Error">{{ candidate.Error }}</em>
                  </div>
                </div>
              </template>
            </div>
            <div v-else class="preview-checklist">
              <h3>ABS Setup Summary</h3>
              <div class="result-grid compact">
                <span>Libraries</span><strong>{{ absLibraries.length }}</strong>
                <span>Mappings</span><strong>{{ absResolvedMappings.length }}</strong>
                <span>Status</span><strong>{{ absSetupReady ? 'Ready for ABS operations' : 'Setup incomplete' }}</strong>
              </div>
            </div>
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
          <template v-if="activeWorkflow === 'abs'">
            <div class="operation-grid">
              <div class="operation-card">
                <h3>Review ABS Data</h3>
                <p>Load real metadata and library state before maintenance actions.</p>
                <div class="operation-actions">
                  <button
                    class="primary-action"
                    :disabled="!absSetupReady || absItemsStatus === 'loading'"
                    @click="loadABSItems"
                  >
                    <Server :size="18" /> {{ absItemsActionLabel }}
                  </button>
                  <button
                    class="secondary-action"
                    :disabled="!absSetupReady || absLibraryStateStatus === 'loading'"
                    @click="loadABSLibraryState"
                  >
                    <Eye :size="18" /> {{ absLibraryStateActionLabel }}
                  </button>
                </div>
                <div class="deferred-state" :class="{ ready: absSetupReady }">
                  <Server :size="18" />
                  <span>{{ absSetupReady ? 'Ready for ABS operations' : 'ABS setup incomplete' }}</span>
                </div>
                <p v-if="!absSetupReady" class="inline-alert">ABS setup must load libraries and validate paths first.</p>
                <p v-if="absItemsError" class="inline-alert">{{ absItemsError }}</p>
                <p v-if="absLibraryStateError" class="inline-alert">{{ absLibraryStateError }}</p>
              </div>
              <div class="operation-card">
                <h3>Library Scan</h3>
                <p>Trigger a real Audiobookshelf scan for the configured library.</p>
                <button
                  class="primary-action"
                  :disabled="!absSetupReady || absScanStatus === 'loading'"
                  @click="triggerABSScan"
                >
                  <Play :size="18" /> {{ absScanActionLabel }}
                </button>
                <p v-if="absScanStatus === 'success'" class="success-note">
                  Scan triggered for {{ absScanResult?.library_id }}.
                </p>
                <p v-if="absScanError" class="inline-alert">{{ absScanError }}</p>
              </div>
              <div class="operation-card danger-card">
                <h3>Clean Missing Items</h3>
                <p>Remove missing or invalid item records reported by Audiobookshelf for this library.</p>
                <label class="check-row">
                  <input v-model="absCleanConfirmed" type="checkbox" />
                  I understand this removes ABS missing item records
                </label>
                <button
                  class="danger-action"
                  :disabled="!absSetupReady || !absCleanConfirmed || absCleanStatus === 'loading'"
                  @click="cleanABSMissing"
                >
                  <Trash2 :size="18" /> {{ absCleanActionLabel }}
                </button>
                <p v-if="absCleanStatus === 'success'" class="success-note">
                  Cleanup completed for {{ absCleanResult?.library_id }}.
                </p>
                <p v-if="absCleanError" class="inline-alert">{{ absCleanError }}</p>
              </div>
            </div>
            <div v-if="hasABSReviewResults || absReviewErrors.length > 0" class="review-layout">
              <h3>{{ absReviewHeading }}</h3>
              <p>{{ absReviewCopy }}</p>
              <div v-if="hasABSReviewResults" class="result-grid">
                <span>Metadata items</span><strong>{{ absItems?.items.length ?? 0 }}</strong>
                <span>Library state items</span><strong>{{ absLibraryState?.items.length ?? 0 }}</strong>
                <span>Missing / invalid</span><strong>{{ absMissingCount }} / {{ absInvalidCount }}</strong>
                <template v-if="absScanResult">
                  <span>Last scan</span><strong>{{ absScanResult.triggered ? absScanResult.library_id : 'Not triggered' }}</strong>
                </template>
                <template v-if="absCleanResult">
                  <span>Last cleanup</span><strong>{{ absCleanResult.cleaned ? absCleanResult.library_id : 'Not cleaned' }}</strong>
                </template>
              </div>
              <div v-if="absItems" class="move-list operation-list">
                <div v-for="item in absItems.items.slice(0, 5)" :key="item.source_path + item.title">
                  <span>{{ item.title || 'Untitled ABS item' }}</span>
                  <strong>{{ item.source_path }}</strong>
                </div>
              </div>
              <div v-if="absLibraryState" class="move-list operation-list">
                <div
                  v-for="item in absLibraryState.items.slice(0, 5)"
                  :key="item.id"
                  :class="{ warning: item.is_missing || item.is_invalid }"
                >
                  <span>{{ item.title || item.id }}</span>
                  <strong>{{ item.path }}</strong>
                  <em v-if="item.is_missing">Missing</em>
                  <em v-else-if="item.is_invalid">Invalid</em>
                </div>
              </div>
              <div v-if="absReviewWarnings.length > 0" class="review-details">
                <h4>Warnings</h4>
                <ul class="warning-list">
                  <li v-for="warning in absReviewWarnings" :key="warning">{{ warning }}</li>
                </ul>
              </div>
              <div v-if="absReviewErrors.length > 0" class="review-details">
                <h4>Errors</h4>
                <ul class="error-list">
                  <li v-for="error in absReviewErrors" :key="error">{{ error }}</li>
                </ul>
              </div>
            </div>
          </template>
          <template v-else>
            <div v-if="activeWorkflow === 'organize' && organizePreview" class="preview-checklist reviewed-plan">
              <h3>Reviewed Organize Plan</h3>
              <div class="result-grid compact">
                <span>Metadata found</span><strong>{{ organizePreview.summary.MetadataFound.length }}</strong>
                <span>Planned moves</span><strong>{{ organizePreview.summary.Moves.length }}</strong>
                <span>Selected moves</span><strong>{{ selectedOrganizeMoveCount }}</strong>
                <span>Warnings</span><strong>{{ organizePreview.summary.MetadataMissing.length }}</strong>
              </div>
              <ul v-if="organizePreview.summary.MetadataMissing.length > 0" class="warning-list">
                <li v-for="missing in organizePreview.summary.MetadataMissing.slice(0, 4)" :key="missing">
                  {{ displayOrganizeSourcePath(missing) }}
                </li>
              </ul>
              <div v-if="organizePreview.summary.Moves.length > 0" class="selection-toolbar">
                <button class="secondary-action compact-action" type="button" @click="selectAllOrganizeMoves">
                  Select All
                </button>
                <button class="secondary-action compact-action" type="button" @click="clearOrganizeMoveSelection">
                  Select None
                </button>
              </div>
              <div v-if="organizePreview.summary.Moves.length > 0" class="move-list selectable-list">
                <label
                  v-for="move in organizePreview.summary.Moves"
                  :key="move.from + move.to"
                  class="selection-row"
                >
                  <input
                    type="checkbox"
                    :checked="isOrganizeMoveSelected(move.from)"
                    :aria-label="`Select move ${move.from}`"
                    @change="toggleOrganizeMove(move.from)"
                  />
                  <span>
                    <span>{{ displayOrganizeSourcePath(move.from) }}</span>
                    <strong class="colored-path">
                      <template v-for="(part, index) in coloredOrganizeTargetParts(move.to)" :key="index">
                        <span :class="part.kind">{{ part.value }}</span>
                      </template>
                    </strong>
                  </span>
                </label>
              </div>
            </div>
            <div v-if="activeWorkflow === 'rename' && renamePreview" class="preview-checklist reviewed-plan">
              <h3>Reviewed Rename Plan</h3>
              <div class="result-grid compact">
                <span>Files scanned</span><strong>{{ renamePreview.summary.FilesScanned }}</strong>
                <span>Candidates</span><strong>{{ renamePreview.candidates.length }}</strong>
                <span>Selected files</span><strong>{{ selectedRenameCandidateCount }}</strong>
                <span>Conflicts</span><strong>{{ renamePreview.summary.ConflictsFound }}</strong>
                <span>Skipped</span><strong>{{ renamePreview.summary.FilesSkipped }}</strong>
                <span>Errors</span><strong>{{ renamePreview.summary.Errors.length }}</strong>
              </div>
              <ul v-if="renamePreview.summary.Errors.length > 0" class="warning-list">
                <li v-for="error in renamePreview.summary.Errors.slice(0, 4)" :key="error">{{ error }}</li>
              </ul>
              <div v-if="renamePreview.candidates.length > 0" class="selection-toolbar">
                <button class="secondary-action compact-action" type="button" @click="selectAllRenameCandidates">
                  Select Actionable
                </button>
                <button class="secondary-action compact-action" type="button" @click="clearRenameCandidateSelection">
                  Select None
                </button>
              </div>
              <div v-if="renamePreview.candidates.length > 0" class="move-list selectable-list">
                <label
                  v-for="candidate in renamePreview.candidates"
                  :key="candidate.CurrentPath + candidate.ProposedPath"
                  class="selection-row"
                  :class="{ warning: candidate.IsConflict || candidate.IsNoOp || !!candidate.Error }"
                >
                  <input
                    type="checkbox"
                    :checked="isRenameCandidateSelected(candidate.CurrentPath)"
                    :disabled="!isRenameCandidateSelectable(candidate)"
                    :aria-label="`Select rename candidate ${candidate.CurrentPath}`"
                    @change="toggleRenameCandidate(candidate.CurrentPath)"
                  />
                  <span>
                    <span>{{ displayRenameSourcePath(candidate.CurrentPath) }}</span>
                    <strong class="colored-path">
                      <template v-for="(part, index) in coloredRenameTargetParts(candidate.ProposedPath)" :key="index">
                        <span :class="part.kind">{{ part.value }}</span>
                      </template>
                    </strong>
                    <em v-if="candidate.IsConflict">Conflict</em>
                    <em v-else-if="candidate.IsNoOp">Skipped: unchanged</em>
                    <em v-else-if="candidate.Error">{{ candidate.Error }}</em>
                  </span>
                </label>
              </div>
            </div>
            <p v-if="activeWorkflow === 'organize' && organizeRunError" class="inline-alert">{{ organizeRunError }}</p>
            <p v-if="activeWorkflow === 'rename' && renameRunError" class="inline-alert">{{ renameRunError }}</p>
            <button
              class="danger-action"
              :disabled="isRunActionDisabled"
              @click="activeWorkflow === 'organize' ? runOrganize() : runRename()"
            >
              <Play :size="18" /> {{ runActionLabel }}
            </button>
            <div v-if="activeWorkflow === 'organize' && (organizeRun || organizeReviewErrors.length > 0)" class="review-layout">
              <h3>{{ organizeReviewHeading }}</h3>
              <p>{{ organizeReviewCopy }}</p>
              <div v-if="organizeReviewSummary" class="result-grid">
                <span>Job status</span><strong>{{ organizeRun ? 'Complete' : 'Preview complete' }}</strong>
                <span>{{ organizeRun ? 'Files organized' : 'Planned moves' }}</span><strong>{{ organizeReviewSummary.summary.Moves.length }}</strong>
                <span>Metadata found</span><strong>{{ organizeReviewSummary.summary.MetadataFound.length }}</strong>
                <span>Warnings</span><strong>{{ organizeReviewSummary.summary.MetadataMissing.length }}</strong>
                <template v-if="organizeRun?.log_path">
                  <span>Undo log</span><strong>{{ organizeRun.log_path }}</strong>
                </template>
              </div>
              <div v-if="organizeRun?.log_path" class="recovery-note">
                Undo details are available in the backend log at {{ organizeRun.log_path }}.
              </div>
              <div v-if="organizeRun?.summary.Moves.length" class="move-list operation-list">
                <div
                  v-for="move in organizeRun.summary.Moves"
                  :key="move.from + move.to"
                >
                  <span>{{ displayOrganizeSourcePath(move.from) }}</span>
                  <strong class="colored-path">
                    <template v-for="(part, index) in coloredOrganizeTargetParts(move.to)" :key="index">
                      <span :class="part.kind">{{ part.value }}</span>
                    </template>
                  </strong>
                </div>
              </div>
              <div v-if="organizeReviewWarnings.length > 0" class="review-details">
                <h4>Warnings</h4>
                <ul class="warning-list">
                  <li v-for="warning in organizeReviewWarnings" :key="warning">{{ warning }}</li>
                </ul>
              </div>
              <div v-if="organizeReviewErrors.length > 0" class="review-details">
                <h4>Errors</h4>
                <ul class="error-list">
                  <li v-for="error in organizeReviewErrors" :key="error">{{ error }}</li>
                </ul>
              </div>
            </div>
            <div v-if="activeWorkflow === 'rename' && (renameRun || renameReviewErrors.length > 0)" class="review-layout">
              <h3>{{ renameReviewHeading }}</h3>
              <p>{{ renameReviewCopy }}</p>
              <div v-if="renameReviewSummary" class="result-grid">
                <span>Job status</span><strong>{{ renameRun ? 'Complete' : 'Preview complete' }}</strong>
                <span>Files scanned</span><strong>{{ renameReviewSummary.summary.FilesScanned }}</strong>
                <span>{{ renameRun ? 'Files renamed' : 'Candidates' }}</span>
                <strong>
                  {{ renameRun ? renameReviewSummary.summary.FilesRenamed : renameReviewSummary.candidates.length }}
                </strong>
                <span>Conflicts</span><strong>{{ renameReviewSummary.summary.ConflictsFound }}</strong>
                <span>Skipped</span><strong>{{ renameReviewSummary.summary.FilesSkipped }}</strong>
                <span>Errors</span><strong>{{ renameReviewSummary.summary.Errors.length }}</strong>
                <template v-if="renameRun?.log_path">
                  <span>Undo log</span><strong>{{ renameRun.log_path }}</strong>
                </template>
              </div>
              <div v-if="renameRun?.log_path" class="recovery-note">
                Undo details are available in the backend log at {{ renameRun.log_path }}.
              </div>
              <div v-if="renameRun?.candidates.length" class="move-list operation-list">
                <div
                  v-for="candidate in renameRun.candidates"
                  :key="candidate.CurrentPath + candidate.ProposedPath"
                  :class="{ warning: candidate.IsConflict || candidate.IsNoOp || !!candidate.Error }"
                >
                  <span>{{ displayRenameSourcePath(candidate.CurrentPath) }}</span>
                  <strong class="colored-path">
                    <template v-for="(part, index) in coloredRenameTargetParts(candidate.ProposedPath)" :key="index">
                      <span :class="part.kind">{{ part.value }}</span>
                    </template>
                  </strong>
                  <em v-if="candidate.IsConflict">Conflict</em>
                  <em v-else-if="candidate.IsNoOp">Skipped: unchanged</em>
                  <em v-else-if="candidate.Error">{{ candidate.Error }}</em>
                </div>
              </div>
              <div v-if="renameReviewWarnings.length > 0" class="review-details">
                <h4>Warnings</h4>
                <ul class="warning-list">
                  <li v-for="warning in renameReviewWarnings" :key="warning">{{ warning }}</li>
                </ul>
              </div>
              <div v-if="renameReviewErrors.length > 0" class="review-details">
                <h4>Errors</h4>
                <ul class="error-list">
                  <li v-for="error in renameReviewErrors" :key="error">{{ error }}</li>
                </ul>
              </div>
            </div>
          </template>
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
  FolderOpen,
  FolderInput,
  Play,
  Plus,
  Server,
  Trash2,
} from 'lucide-vue-next'
import TemplateBuilder, { type TemplateField } from './components/TemplateBuilder.vue'
import {
  apiGet,
  apiPost,
  hasWebSessionToken,
  type ABSCleanMissingResponse,
  type ABSConfig,
  type ABSItemsResponse,
  type ABSLibrariesResponse,
  type ABSLibrary,
  type ABSLibraryStateResponse,
  type ABSPathMappingResponse,
  type ABSScanTriggerResponse,
  type FieldMapping,
  type HealthResponse,
  type Option,
  type OrganizerConfig,
  type OrganizePreviewResponse,
  type OrganizeRunResponse,
  type OptionsResponse,
  type PathValidationItem,
  type PathValidationResponse,
  type RenameCandidate,
  type RenameConfig,
  type RenamePreviewResponse,
  type RenameRunResponse,
  type WebConfig,
} from './api'

type WorkflowId = 'organize' | 'rename' | 'abs'
type StageId = 'configure' | 'run'
type PathFieldId = 'source' | 'output'
type LoadState = 'loading' | 'ready' | 'fallback'
type CredentialState = 'empty' | 'redacted'
type RequestState = 'idle' | 'loading' | 'success' | 'error'
type EditablePathMapping = {
  abs_prefix: string
  local_prefix: string
}
type ColoredTextPart = {
  kind: TemplateField['kind'] | 'text' | 'separator'
  value: string
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
    previewCopy: 'Create a backend dry-run preview before the organize run stage can unlock.',
    runTitle: 'Run Organize',
    runCopy: 'This action changes files and stays locked until a current preview is available.',
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
    runTitle: 'Run Rename',
    runCopy: 'This action renames files in place and stays locked until current candidates are available.',
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
    previewCopy: 'Load real ABS metadata and library state through the local backend before maintenance actions.',
    runTitle: 'ABS Maintenance Actions',
    runCopy: 'Library scan triggers and cleanup actions require completed ABS setup.',
    runAction: 'Run ABS Action',
  },
]

const stages = [
  { id: 'configure' as const, index: '1', label: 'Setup & Preview', description: 'Tweak inputs and preview' },
  { id: 'run' as const, index: '2', label: 'Review & Run', description: 'Select, execute, inspect' },
]

const stageText: Record<StageId, { heading: string; copy: string }> = {
  configure: {
    heading: 'Setup and preview',
    copy: 'Adjust workflow inputs and watch the dry-run preview refresh automatically.',
  },
  run: {
    heading: 'Review and run',
    copy: 'Select planned changes, execute the filesystem action, and inspect backend results.',
  },
}

const customLayoutValue = 'custom'
const defaultCustomLayoutTemplate = '{author}/{series|Standalone}/{Vol series_number:02 - }{title}{ [narrator]}'
const customLayoutOption: Option = { value: customLayoutValue, label: 'Custom' }
const defaultLayouts: Option[] = [{ value: 'author-series-title', label: 'Author / Series / Title' }]
const layoutTemplateFields: TemplateField[] = [
  { value: 'author', label: 'Author', kind: 'author' },
  { value: 'authors', label: 'Authors', kind: 'author' },
  { value: 'title', label: 'Title', kind: 'title' },
  { value: 'series', label: 'Series', kind: 'series' },
  { value: 'series-count', label: 'Series #', kind: 'series' },
  { value: 'narrator', label: 'Narrator', kind: 'other' },
  { value: 'track', label: 'Track', kind: 'other' },
  { value: 'year', label: 'Year', kind: 'other' },
]
const renameTemplateFields: TemplateField[] = [
  { value: 'author', label: 'Author', kind: 'author' },
  { value: 'title', label: 'Title', kind: 'title' },
  { value: 'series', label: 'Series', kind: 'series' },
  { value: 'series_number', label: 'Series #', kind: 'series' },
  { value: 'narrator', label: 'Narrator', kind: 'other' },
  { value: 'track', label: 'Track', kind: 'other' },
  { value: 'year', label: 'Year', kind: 'other' },
]
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
const fieldMappingFieldNames = ['title', 'authors', 'author', 'artist', 'album_artist', 'series', 'album', 'track', 'track_number', 'disc', 'discnumber']

const health = ref('offline')
const configState = ref<LoadState>('loading')
const optionsState = ref<LoadState>('loading')
const activeWorkflow = ref<WorkflowId>('organize')
const activeStage = ref<StageId>('configure')
const bootstrapComplete = ref(false)
const sourceFolder = ref('')
const outputFolder = ref('')
const sourceFolderPicker = ref<HTMLInputElement | null>(null)
const outputFolderPicker = ref<HTMLInputElement | null>(null)
const sourcePathMessage = ref('')
const outputPathMessage = ref('')
const activePathDropTarget = ref<PathFieldId | null>(null)
const scanMode = ref('json')
const scanModeUserSelected = ref(false)
const layout = ref('author-series-title')
const layoutTemplate = ref('')
const removeEmpty = ref(false)
const renameTemplate = ref('{author} - {series} {series_number} - {title}')
const renameRecursive = ref(true)
const preservePath = ref(true)
const absUrl = ref('')
const absToken = ref('')
const absLibrary = ref('')
const absCredentialState = ref<CredentialState>('empty')
const absHeaderName = ref('')
const absHeaderValue = ref('')
const absSQLitePath = ref('')
const absPathMappings = ref<EditablePathMapping[]>([{ abs_prefix: '/audiobooks', local_prefix: '' }])
const organizerDefaults = ref<OrganizerConfig | null>(null)
const renameDefaults = ref<RenameConfig | null>(null)
const organizeFieldMapping = ref<FieldMapping>({ ...defaultFieldMapping })
const renameFieldMapping = ref<FieldMapping>({ ...defaultFieldMapping })
const layouts = ref<Option[]>([])
const scanModes = ref<Option[]>([])
const fieldMappings = ref<Record<string, FieldMapping>>({})
const absLibraries = ref<ABSLibrary[]>([])
const absResolvedMappings = ref<EditablePathMapping[]>([])
const absItems = ref<ABSItemsResponse | null>(null)
const absLibraryState = ref<ABSLibraryStateResponse | null>(null)
const absScanResult = ref<ABSScanTriggerResponse | null>(null)
const absCleanResult = ref<ABSCleanMissingResponse | null>(null)
const organizePreview = ref<OrganizePreviewResponse | null>(null)
const organizeRun = ref<OrganizeRunResponse | null>(null)
const renamePreview = ref<RenamePreviewResponse | null>(null)
const renameRun = ref<RenameRunResponse | null>(null)
const selectedOrganizeSources = ref<string[]>([])
const selectedRenamePaths = ref<string[]>([])
const organizePreviewStale = ref(false)
const renamePreviewStale = ref(false)
const organizePreviewStatus = ref<RequestState>('idle')
const organizeRunStatus = ref<RequestState>('idle')
const renamePreviewStatus = ref<RequestState>('idle')
const renameRunStatus = ref<RequestState>('idle')
const pathValidationStatus = ref<RequestState>('idle')
const absLibrariesStatus = ref<RequestState>('idle')
const absPathStatus = ref<RequestState>('idle')
const absItemsStatus = ref<RequestState>('idle')
const absLibraryStateStatus = ref<RequestState>('idle')
const absScanStatus = ref<RequestState>('idle')
const absCleanStatus = ref<RequestState>('idle')
const organizePreviewError = ref('')
const organizeRunError = ref('')
const renamePreviewError = ref('')
const renameRunError = ref('')
const pathValidationError = ref('')
const absLibrariesError = ref('')
const absPathError = ref('')
const absItemsError = ref('')
const absLibraryStateError = ref('')
const absScanError = ref('')
const absCleanError = ref('')
const absCleanConfirmed = ref(false)
const events = ref<ActivityEvent[]>([
  { time: now(), level: 'info', event: 'Local UI ready', detail: 'No workflow request has run yet.' },
])
let autoPreviewTimer: number | null = null

const currentWorkflow = computed(() => workflows.find((workflow) => workflow.id === activeWorkflow.value) ?? workflows[0])
const currentStage = computed(() => stageText[activeStage.value])
const optionsLoading = computed(() => optionsState.value === 'loading')
const layoutOptions = computed(() => {
  const appendCustom = (options: Option[]) =>
    options.some((option) => option.value === customLayoutValue) ? options : [...options, customLayoutOption]
  if (layouts.value.length > 0) {
    return appendCustom(layouts.value)
  }
  return optionsState.value === 'fallback' ? appendCustom(defaultLayouts) : []
})
const scanModeOptions = computed(() => {
  if (scanModes.value.length > 0) {
    return scanModes.value
  }
  return optionsState.value === 'fallback' ? defaultScanModes : []
})
const workflowScanModes = computed(() => {
  return scanModeOptions.value.filter((mode) => activeWorkflow.value !== 'rename' || mode.value !== 'abs')
})
const showABSSetup = computed(() => activeWorkflow.value === 'abs' || (activeWorkflow.value === 'organize' && scanMode.value === 'abs'))
const selectedMetadataSourceLabel = computed(() => {
  return workflowScanModes.value.find((mode) => mode.value === scanMode.value)?.label ?? scanMode.value
})
const selectedLayoutLabel = computed(() => {
  return layoutOptions.value.find((option) => option.value === layout.value)?.label ?? layout.value
})
const fieldMappingPresets = computed(() =>
  Object.entries(fieldMappings.value).map(([value, mapping]) => ({
    value,
    label: value === 'default' ? 'metadata.json default' : `${value.toUpperCase()} preset`,
    mapping,
  })),
)
const activeFieldMapping = computed(() =>
  activeWorkflow.value === 'rename' ? renameFieldMapping.value : organizeFieldMapping.value,
)
const activeFieldMappingPreset = computed(() => {
  const current = activeFieldMapping.value
  return (
    fieldMappingPresets.value.find(({ mapping }) => fieldMappingsEqual(mapping, current))?.value ?? 'custom'
  )
})
const serverLabel = computed(() =>
  [
    health.value === 'ok' ? 'localhost connected' : 'localhost pending',
    stateLabel('config', configState.value),
    stateLabel('options', optionsState.value),
  ].join(' · '),
)
const absSetupState = computed(() => {
  if (absSetupReady.value) {
    return 'ABS libraries loaded and path mappings validated.'
  }
  if (absCredentialState.value === 'redacted' && !absToken.value) {
    return 'Saved ABS credentials are redacted. Enter a fresh token before sending requests.'
  }
  return 'Load libraries and validate path mappings to complete ABS setup.'
})
const absLibrarySelectionReady = computed(() =>
  absLibrariesStatus.value === 'success' && absLibraries.value.some((library) => library.id === absLibrary.value),
)
const absSetupReady = computed(() => absLibrarySelectionReady.value && absPathStatus.value === 'success')
const activePreviewStale = computed(() => {
  if (activeWorkflow.value === 'rename') {
    return renamePreviewStale.value
  }
  if (activeWorkflow.value === 'organize') {
    return organizePreviewStale.value
  }
  return false
})
const canOpenOrganizeReview = computed(
  () => organizePreviewStatus.value === 'success' && !organizePreviewStale.value && !!organizePreview.value,
)
const canOpenRenameReview = computed(
  () => renamePreviewStatus.value === 'success' && !renamePreviewStale.value && !!renamePreview.value,
)
const previewHeading = computed(() => {
  if (activeWorkflow.value === 'rename') {
    if (renamePreviewStale.value) {
      return 'Rename preview stale'
    }
    if (renamePreviewStatus.value === 'loading') {
      return 'Creating rename preview'
    }
    if (renamePreviewStatus.value === 'success') {
      return 'Rename preview ready'
    }
    if (renamePreviewStatus.value === 'error') {
      return 'Rename preview needs attention'
    }
    return 'Waiting for rename inputs'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStale.value) {
    return 'Organize preview stale'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStatus.value === 'loading') {
    return 'Creating organize preview'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStatus.value === 'success') {
    return 'Organize preview ready'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStatus.value === 'error') {
    return 'Preview needs attention'
  }
  if (activeWorkflow.value === 'abs') {
    if (absItemsStatus.value === 'success' || absLibraryStateStatus.value === 'success') {
      return 'ABS operation data ready'
    }
    if (absItemsStatus.value === 'error' || absLibraryStateStatus.value === 'error') {
      return 'ABS operation needs attention'
    }
    return 'Load ABS operation data'
  }
  return activeWorkflow.value === 'organize' ? 'Waiting for organize inputs' : 'Dry-run preview first'
})
const setupPreviewCopy = computed(() => {
  if (activeWorkflow.value === 'abs') {
    return 'Complete ABS setup, then review operation data before maintenance actions.'
  }
  if (activePreviewStale.value) {
    return 'Inputs changed after the last backend preview. The plan will refresh automatically.'
  }
  if (activeWorkflow.value === 'organize' && organizePreviewStatus.value === 'idle') {
    return 'Enter a valid source and output folder to create the first dry-run preview.'
  }
  if (activeWorkflow.value === 'rename' && renamePreviewStatus.value === 'idle') {
    return 'Enter a valid source folder and template to create rename candidates.'
  }
  return currentWorkflow.value.previewCopy
})
const absLibrariesActionLabel = computed(() =>
  absLibrariesStatus.value === 'loading' ? 'Testing Connection' : 'Test Connection',
)
const absPathActionLabel = computed(() =>
  absPathStatus.value === 'loading' ? 'Validating Paths' : 'Validate Paths',
)
const absItemsActionLabel = computed(() =>
  absItemsStatus.value === 'loading' ? 'Loading Items' : 'Load ABS Items',
)
const absLibraryStateActionLabel = computed(() =>
  absLibraryStateStatus.value === 'loading' ? 'Checking State' : 'Check Library State',
)
const absScanActionLabel = computed(() =>
  absScanStatus.value === 'loading' ? 'Triggering Scan' : 'Trigger Scan',
)
const absCleanActionLabel = computed(() =>
  absCleanStatus.value === 'loading' ? 'Cleaning Missing Items' : 'Clean Missing Items',
)
const runActionLabel = computed(() => {
  if (activeWorkflow.value === 'rename' && renameRunStatus.value === 'loading') {
    return 'Renaming Files'
  }
  if (activeWorkflow.value === 'organize' && organizeRunStatus.value === 'loading') {
    return 'Running Organize'
  }
  if (activeWorkflow.value === 'rename') {
    return selectedRenameCandidateCount.value === 1 ? 'Run 1 Selected File' : `Run ${selectedRenameCandidateCount.value} Selected Files`
  }
  if (activeWorkflow.value === 'organize') {
    return selectedOrganizeMoveCount.value === 1 ? 'Run 1 Selected Move' : `Run ${selectedOrganizeMoveCount.value} Selected Moves`
  }
  return currentWorkflow.value.runAction
})
const absMissingCount = computed(() => absLibraryState.value?.items.filter((item) => item.is_missing).length ?? 0)
const absInvalidCount = computed(() => absLibraryState.value?.items.filter((item) => item.is_invalid).length ?? 0)
const selectedOrganizeMoveCount = computed(
  () => organizePreview.value?.summary.Moves.filter((move) => isOrganizeMoveSelected(move.from)).length ?? 0,
)
const selectedRenameCandidateCount = computed(
  () => renamePreview.value?.candidates.filter((candidate) => isRenameCandidateSelected(candidate.CurrentPath)).length ?? 0,
)
const organizeReviewSummary = computed(() => organizeRun.value ?? organizePreview.value)
const organizeReviewWarnings = computed(() => organizeReviewSummary.value?.summary.MetadataMissing ?? [])
const organizeReviewErrors = computed(() => {
  const errors: string[] = []
  if (organizePreviewStatus.value === 'error' && organizePreviewError.value) {
    errors.push(`Organize preview: ${organizePreviewError.value}`)
  }
  if (organizeRunStatus.value === 'error' && organizeRunError.value) {
    errors.push(`Organize run: ${organizeRunError.value}`)
  }
  return errors
})
const organizeReviewHeading = computed(() => {
  if (organizeRun.value) {
    return 'Organize Run Complete'
  }
  if (organizePreview.value) {
    return 'Organize Preview Results'
  }
  if (organizeReviewErrors.value.length > 0) {
    return 'Organize Results Need Attention'
  }
  return 'Organize Results'
})
const organizeReviewCopy = computed(() => {
  if (organizeRun.value) {
	if (scanMode.value === 'abs') {
		return 'The ABS-backed organize run finished. Open Audiobookshelf to trigger a library scan, inspect the refreshed state, and clean only confirmed missing old paths.'
	}
    return 'The reviewed organize plan finished with backend results.'
  }
  if (organizePreview.value) {
    return 'The latest backend organize preview is available for inspection before execution.'
  }
  if (organizeReviewErrors.value.length > 0) {
    return 'The backend reported an error. Details remain available here while you adjust inputs or retry.'
  }
  return 'Completed organize runs will appear here after you run a reviewed preview.'
})
const renameReviewSummary = computed(() => renameRun.value ?? renamePreview.value)
const renameReviewWarnings = computed(() => {
  const warnings =
    renameReviewSummary.value?.candidates
      .filter((candidate) => candidate.IsConflict || candidate.IsNoOp)
      .map((candidate) => {
        if (candidate.IsConflict) {
          return `Conflict: ${candidate.CurrentPath} -> ${candidate.ProposedPath}`
        }
        return `Skipped unchanged: ${candidate.CurrentPath}`
      }) ?? []
  return warnings
})
const renameReviewErrors = computed(() => {
  const errors: string[] = []
  if (renamePreviewStatus.value === 'error' && renamePreviewError.value) {
    errors.push(`Rename preview: ${renamePreviewError.value}`)
  }
  if (renameRunStatus.value === 'error' && renameRunError.value) {
    errors.push(`Rename run: ${renameRunError.value}`)
  }
  if (renameReviewSummary.value) {
    errors.push(...renameReviewSummary.value.summary.Errors)
    errors.push(
      ...renameReviewSummary.value.candidates
        .filter((candidate) => candidate.Error)
        .map((candidate) => `${candidate.CurrentPath}: ${candidate.Error}`),
    )
  }
  return errors
})
const renameReviewHeading = computed(() => {
  if (renameRun.value) {
    return 'Rename Run Complete'
  }
  if (renamePreview.value) {
    return 'Rename Preview Results'
  }
  if (renameReviewErrors.value.length > 0) {
    return 'Rename Results Need Attention'
  }
  return 'Rename Results'
})
const renameReviewCopy = computed(() => {
  if (renameRun.value) {
    return 'The reviewed rename plan finished with backend results.'
  }
  if (renamePreview.value) {
    return 'The latest backend rename preview is available for inspection before execution.'
  }
  if (renameReviewErrors.value.length > 0) {
    return 'The backend reported an error. Details remain available here while you adjust inputs or retry.'
  }
  return 'Completed rename previews will appear here after you request candidates from the backend.'
})
const hasABSReviewResults = computed(
  () => !!absItems.value || !!absLibraryState.value || !!absScanResult.value || !!absCleanResult.value,
)
const absReviewWarnings = computed(() => {
  const warnings: string[] = []
  if (absLibraryState.value) {
    warnings.push(
      ...absLibraryState.value.items
        .filter((item) => item.is_missing || item.is_invalid)
        .map((item) => {
          const states = [item.is_missing ? 'missing' : '', item.is_invalid ? 'invalid' : ''].filter(Boolean).join(', ')
          return `${item.title || item.id}: ${states} at ${item.path}`
        }),
    )
  }
  if (absScanResult.value && !absScanResult.value.triggered) {
    warnings.push(`ABS scan was not triggered for ${absScanResult.value.library_id}.`)
  }
  if (absCleanResult.value && !absCleanResult.value.cleaned) {
    warnings.push(`ABS cleanup did not report changes for ${absCleanResult.value.library_id}.`)
  }
  return warnings
})
const absReviewErrors = computed(() =>
  [
    absLibrariesError.value && `ABS libraries: ${absLibrariesError.value}`,
    absPathError.value && `ABS path validation: ${absPathError.value}`,
    absItemsError.value && `ABS items: ${absItemsError.value}`,
    absLibraryStateError.value && `ABS library state: ${absLibraryStateError.value}`,
    absScanError.value && `ABS scan: ${absScanError.value}`,
    absCleanError.value && `ABS cleanup: ${absCleanError.value}`,
  ].filter((error): error is string => Boolean(error)),
)
const absReviewHeading = computed(() => {
  if (hasABSReviewResults.value) {
    return 'ABS Operation Results'
  }
  if (absReviewErrors.value.length > 0) {
    return 'ABS Results Need Attention'
  }
  return 'ABS Results'
})
const absReviewCopy = computed(() => {
  if (hasABSReviewResults.value) {
    return 'Completed ABS backend actions are summarized here.'
  }
  if (absReviewErrors.value.length > 0) {
    return 'The backend reported an error. Details remain available here while you adjust inputs or retry.'
  }
  return 'Completed ABS backend actions will appear here after you load items, check state, trigger a scan, or clean missing records.'
})
const isRunActionDisabled = computed(() => {
  if (activeWorkflow.value === 'rename') {
    return !canOpenRenameReview.value || renameRunStatus.value === 'loading' || selectedRenameCandidateCount.value === 0
  }
  if (activeWorkflow.value === 'abs') {
    return !absSetupReady.value
  }
  return !canOpenOrganizeReview.value || organizeRunStatus.value === 'loading' || selectedOrganizeMoveCount.value === 0
})

function selectWorkflow(workflow: WorkflowId) {
  activeWorkflow.value = workflow
  activeStage.value = 'configure'
  ensureScanModeFitsWorkflow()
  scheduleActivePreviewRefresh()
  addEvent({
    time: now(),
    level: 'info',
    event: `Local navigation: ${currentWorkflow.value.label} selected`,
    detail: 'Configure inputs before preview.',
  })
}

function selectScanMode(mode: string) {
  scanModeUserSelected.value = true
  scanMode.value = mode
}

function isStageLocked(stage: StageId) {
  if (stage === 'configure') {
    return false
  }
  if (activeWorkflow.value === 'organize') {
    return !canOpenOrganizeReview.value || organizeRunStatus.value === 'loading'
  }
  if (activeWorkflow.value === 'abs') {
    return !absSetupReady.value || absScanStatus.value === 'loading' || absCleanStatus.value === 'loading'
  }
  return !canOpenRenameReview.value || renameRunStatus.value === 'loading'
}

function openOrganizeReview() {
  if (!canOpenOrganizeReview.value) {
    return
  }
  activeStage.value = 'run'
}

function openRenameReview() {
  if (!canOpenRenameReview.value) {
    return
  }
  activeStage.value = 'run'
}

function isOrganizeMoveSelected(sourcePath: string) {
  return selectedOrganizeSources.value.includes(sourcePath)
}

function toggleOrganizeMove(sourcePath: string) {
  if (isOrganizeMoveSelected(sourcePath)) {
    selectedOrganizeSources.value = selectedOrganizeSources.value.filter((path) => path !== sourcePath)
    return
  }
  selectedOrganizeSources.value = [...selectedOrganizeSources.value, sourcePath]
}

function selectAllOrganizeMoves() {
  selectedOrganizeSources.value = organizePreview.value?.summary.Moves.map((move) => move.from) ?? []
}

function clearOrganizeMoveSelection() {
  selectedOrganizeSources.value = []
}

function isRenameCandidateSelectable(candidate: RenameCandidate) {
  return !candidate.Error && !candidate.IsNoOp
}

function isRenameCandidateSelected(currentPath: string) {
  return selectedRenamePaths.value.includes(currentPath)
}

function toggleRenameCandidate(currentPath: string) {
  if (isRenameCandidateSelected(currentPath)) {
    selectedRenamePaths.value = selectedRenamePaths.value.filter((path) => path !== currentPath)
    return
  }
  selectedRenamePaths.value = [...selectedRenamePaths.value, currentPath]
}

function selectAllRenameCandidates() {
  selectedRenamePaths.value =
    renamePreview.value?.candidates
      .filter((candidate) => isRenameCandidateSelectable(candidate))
      .map((candidate) => candidate.CurrentPath) ?? []
}

function clearRenameCandidateSelection() {
  selectedRenamePaths.value = []
}

function displayOrganizeSourcePath(path: string): string {
  return displayLocalPath(path, sourceFolder.value)
}

function displayRenameSourcePath(path: string): string {
  return displayLocalPath(path, sourceFolder.value)
}

function coloredOrganizeTargetParts(path: string): ColoredTextPart[] {
  return coloredPathParts(path, outputFolder.value, activeLayoutSegmentKinds())
}

function coloredRenameTargetParts(path: string): ColoredTextPart[] {
  return coloredPathParts(path, sourceFolder.value, templateSegmentKinds(renameTemplate.value, renameTemplateFields))
}

function coloredPathParts(path: string, basePath: string, segmentKinds: TemplateField['kind'][]): ColoredTextPart[] {
  if (segmentKinds.length === 0) {
    return [{ kind: 'text', value: displayLocalPath(path, basePath) }]
  }

  const normalizedPath = path.replaceAll('\\', '/')
  const normalizedBase = trimTrailingPathSeparators(basePath.trim()).replaceAll('\\', '/')
  const baseEndIndex = findBasePathEndIndex(normalizedPath, normalizedBase)
  if (baseEndIndex === normalizedPath.length) {
    return [{ kind: 'text', value: path }]
  }
  if (baseEndIndex >= 0) {
    const displayBase = displayBasePath(path.slice(0, baseEndIndex), basePath)
    return [
      { kind: 'text', value: displayBase },
      { kind: 'separator', value: path[baseEndIndex] ?? '/' },
      ...coloredRelativePathParts(path.slice(baseEndIndex + 1), segmentKinds),
    ]
  }

  return coloredRelativePathParts(path, segmentKinds)
}

function displayLocalPath(path: string, basePath: string): string {
  const normalizedPath = path.replaceAll('\\', '/')
  const normalizedBase = trimTrailingPathSeparators(basePath.trim()).replaceAll('\\', '/')
  const baseEndIndex = findBasePathEndIndex(normalizedPath, normalizedBase)
  if (baseEndIndex < 0) {
    return path
  }
  const suffix = path.slice(baseEndIndex)
  return `${displayBasePath(path.slice(0, baseEndIndex), basePath)}${suffix}`
}

function displayBasePath(absoluteBase: string, configuredBase: string): string {
  const trimmed = trimTrailingPathSeparators(configuredBase.trim())
  if (!trimmed || isAbsolutePath(trimmed)) {
    return absoluteBase
  }
  return trimmed
}

function isAbsolutePath(path: string): boolean {
  return path.startsWith('/') || /^[A-Za-z]:[\\/]/.test(path)
}

function findBasePathEndIndex(normalizedPath: string, normalizedBase: string): number {
  const candidates = basePathMatchCandidates(normalizedBase)
  for (const candidate of candidates) {
    if (normalizedPath === candidate) {
      return normalizedPath.length
    }
    if (normalizedPath.startsWith(`${candidate}/`)) {
      return candidate.length
    }
    const embeddedIndex = normalizedPath.indexOf(`/${candidate}/`)
    if (embeddedIndex >= 0) {
      return embeddedIndex + 1 + candidate.length
    }
  }
  return -1
}

function basePathMatchCandidates(normalizedBase: string): string[] {
  const candidates = new Set<string>()
  const cleanedBase = trimLeadingCurrentDirectory(normalizedBase)
  if (cleanedBase) {
    candidates.add(cleanedBase)
  }
  const segments = cleanedBase.split('/').filter(Boolean)
  for (let index = 1; index < segments.length; index += 1) {
    candidates.add(segments.slice(index).join('/'))
  }
  return [...candidates].sort((left, right) => right.length - left.length)
}

function trimLeadingCurrentDirectory(path: string): string {
  let trimmed = path
  while (trimmed.startsWith('./')) {
    trimmed = trimmed.slice(2)
  }
  return trimmed
}

function coloredRelativePathParts(path: string, segmentKinds: TemplateField['kind'][]): ColoredTextPart[] {
  const parts: ColoredTextPart[] = []
  const segments = path.split(/([/\\])/)
  let segmentIndex = 0
  for (const segment of segments) {
    if (!segment) {
      continue
    }
    if (segment === '/' || segment === '\\') {
      parts.push({ kind: 'separator', value: segment })
      continue
    }
    parts.push({ kind: segmentKinds[segmentIndex] ?? 'text', value: segment })
    segmentIndex += 1
  }
  return parts.length > 0 ? parts : [{ kind: 'text', value: path }]
}

function activeLayoutSegmentKinds(): TemplateField['kind'][] {
  if (layout.value === customLayoutValue) {
    return templateSegmentKinds(layoutTemplate.value, layoutTemplateFields)
  }
  switch (layout.value) {
    case 'author-only':
      return ['author']
    case 'author-title':
      return ['author', 'title']
    case 'author-series':
      return ['author', 'series']
    case 'author-series-title':
      return ['author', 'series', 'title']
    case 'author-series-title-number':
      return ['author', 'series', 'title']
    case 'series-title':
      return ['series', 'title']
    case 'series-title-number':
      return ['series', 'title']
    default:
      return []
  }
}

function templateSegmentKinds(template: string, fields: TemplateField[]): TemplateField['kind'][] {
  const fieldKinds = new Map(fields.map((field) => [field.value.toLowerCase(), field.kind]))
  return template
    .split('/')
    .map((segment) => templateTokenKinds(segment, fieldKinds))
    .filter((kind): kind is TemplateField['kind'] => Boolean(kind))
}

function templateTokenKinds(segment: string, fieldKinds: Map<string, TemplateField['kind']>): TemplateField['kind'] | '' {
  const kinds = [...segment.matchAll(/\{([^{}]+)\}/g)]
    .map((match) => match[1].split('|')[0]?.trim().toLowerCase() ?? '')
    .map((field) => fieldKinds.get(field) ?? 'other')
  const uniqueKinds = [...new Set(kinds)]
  if (uniqueKinds.length === 0) {
    return ''
  }
  if (uniqueKinds.length === 1) {
    return uniqueKinds[0]
  }
  if (uniqueKinds.includes('title')) {
    return 'title'
  }
  return 'other'
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

function addRequestStart(label: string, detail: string) {
  addEvent({ time: now(), level: 'info', event: `Request started: ${label}`, detail })
}

function addRequestSuccess(label: string, detail: string) {
  addEvent({ time: now(), level: 'ok', event: `Request succeeded: ${label}`, detail })
}

function addRequestError(label: string, detail: string) {
  addEvent({ time: now(), level: 'warn', event: `Request failed: ${label}`, detail })
}

function addActionError(label: string, detail: string, requestStarted: boolean) {
  if (requestStarted) {
    addRequestError(label, detail)
    return
  }
  addEvent({ time: now(), level: 'warn', event: `Local validation failed: ${label}`, detail })
}

function openPathPicker(field: PathFieldId) {
  clearPathMessage(field)
  const picker = field === 'source' ? sourceFolderPicker.value : outputFolderPicker.value
  if (!picker) {
    setPathMessage(field, 'Folder selection is unavailable here. Type or paste the folder path instead.')
    return
  }
  picker.value = ''
  picker.click()
}

function handlePathPickerChange(field: PathFieldId, event: Event) {
  const input = event.target as HTMLInputElement
  applyPathFiles(field, input.files, 'selected')
  input.value = ''
}

function handlePathDrop(field: PathFieldId, event: DragEvent) {
  activePathDropTarget.value = null
  applyPathFiles(field, event.dataTransfer?.files ?? null, 'dropped')
}

function applyPathFiles(field: PathFieldId, files: FileList | null, action: 'selected' | 'dropped') {
  if (!files || files.length === 0) {
    setPathMessage(field, 'No folder files were available. Type or paste the folder path instead.')
    return
  }

  const path = extractLocalDirectoryPath(files[0])
  if (!path) {
    const actionLabel = action === 'selected' ? 'selected' : 'dropped'
    setPathMessage(
      field,
      `Folder ${actionLabel}, but this browser did not expose a local path. Type or paste the folder path instead.`,
    )
    return
  }

  setPathValue(field, path)
  setPathMessage(field, `${pathLabel(field)} set from ${action} folder.`)
}

function extractLocalDirectoryPath(file: File): string {
  const filePath = (file as File & { path?: string }).path
  if (!filePath) {
    return ''
  }
  const relativeSegments = file.webkitRelativePath.split('/').filter(Boolean)
  let path = filePath
  const levelsToRemove = relativeSegments.length > 0 ? relativeSegments.length : 1
  for (let index = 0; index < levelsToRemove; index += 1) {
    path = parentPath(path)
  }
  return path
}

function parentPath(path: string): string {
  const trimmed = trimTrailingPathSeparators(path)
  const index = Math.max(trimmed.lastIndexOf('/'), trimmed.lastIndexOf('\\'))
  return index > 0 ? trimmed.slice(0, index) : trimmed
}

function trimTrailingPathSeparators(path: string): string {
  let end = path.length
  while (end > 0 && isPathSeparator(path[end - 1])) {
    end -= 1
  }
  return path.slice(0, end)
}

function isPathSeparator(character: string): boolean {
  return character === '/' || character === '\\'
}

function setPathValue(field: PathFieldId, value: string) {
  if (field === 'source') {
    sourceFolder.value = value
    return
  }
  outputFolder.value = value
}

function clearPathMessage(field: PathFieldId) {
  setPathMessage(field, '')
  pathValidationError.value = ''
  if (pathValidationStatus.value !== 'loading') {
    pathValidationStatus.value = 'idle'
  }
}

function setPathMessage(field: PathFieldId, message: string) {
  if (field === 'source') {
    sourcePathMessage.value = message
    return
  }
  outputPathMessage.value = message
}

function pathLabel(field: PathFieldId): string {
  return field === 'source' ? 'Source folder' : 'Output folder'
}

async function validateConfigurePaths(): Promise<boolean> {
  pathValidationStatus.value = 'loading'
  pathValidationError.value = ''
  clearPathMessage('source')
  clearPathMessage('output')

  try {
    addRequestStart('Path validation', 'POST /api/paths/validate')
    const response = await apiPost<PathValidationResponse>('/api/paths/validate', {
      paths: buildPathValidationItems(),
    })
    const invalid = response.results.filter((result) => !result.valid)
    for (const result of response.results) {
      if (result.id === 'source') {
        sourcePathMessage.value = result.valid ? 'Source folder is ready.' : result.error || 'Source folder is invalid.'
      }
      if (result.id === 'output') {
        outputPathMessage.value = result.valid ? 'Output folder is ready.' : result.error || 'Output folder is invalid.'
      }
    }
    if (invalid.length > 0) {
      pathValidationStatus.value = 'error'
      pathValidationError.value = invalid.map((result) => result.error || `${result.id} path is invalid.`).join(' ')
      addActionError('Path validation', pathValidationError.value, false)
      return false
    }
    pathValidationStatus.value = 'success'
    addRequestSuccess('Path validation', pathValidationSuccessMessage())
    return true
  } catch (error) {
    pathValidationStatus.value = 'error'
    pathValidationError.value = error instanceof Error ? error.message : 'Path validation failed.'
    addActionError('Path validation', pathValidationError.value, true)
    return false
  }
}

function buildPathValidationItems(): PathValidationItem[] {
  const paths: PathValidationItem[] = [
    { id: 'source', path: sourceFolder.value.trim(), kind: 'existing-directory' },
  ]
  if (activeWorkflow.value !== 'rename') {
    paths.push({ id: 'output', path: outputFolder.value.trim(), kind: 'output-directory' })
  }
  return paths
}

function pathValidationSuccessMessage(): string {
  return activeWorkflow.value === 'rename' ? 'Source path is ready.' : 'Source and output paths are ready.'
}

async function createOrganizePreview() {
  organizePreviewStatus.value = 'loading'
  organizePreviewError.value = ''
  organizeRun.value = null
  organizeRunStatus.value = 'idle'
  organizeRunError.value = ''
  organizePreviewStale.value = false
  let requestStarted = false

  try {
    if (!sourceFolder.value.trim() || !outputFolder.value.trim()) {
      throw new Error('Source and output folders are required for organize preview.')
    }
    if (scanMode.value === 'abs') {
      assertABSSetupReady()
    }
    addRequestStart('Organize preview', 'POST /api/organize/preview')
    requestStarted = true
    const response = normalizeOrganizeResponse(
      await apiPost<OrganizePreviewResponse>('/api/organize/preview', {
        config: buildOrganizerConfig(true),
      }),
    )
    if (shouldDefaultToEmbeddedFileMetadata(response)) {
      organizePreviewStatus.value = 'idle'
      organizePreviewStale.value = false
      scanMode.value = 'embedded-file'
      addEvent({
        time: now(),
        level: 'info',
        event: 'Local default: Embedded metadata by file',
        detail: 'No metadata.json records were found, so the preview will retry with file metadata.',
      })
      scheduleActivePreviewRefresh()
      return
    }
    organizePreview.value = response
    selectedOrganizeSources.value = response.summary.Moves.map((move) => move.from)
    organizePreviewStatus.value = 'success'
    organizePreviewStale.value = false
    addRequestSuccess(
      'Organize preview',
      `${response.summary.Moves.length} planned move(s), ${response.summary.MetadataMissing.length} warning(s).`,
    )
  } catch (error) {
    organizePreview.value = null
    selectedOrganizeSources.value = []
    organizePreviewStatus.value = 'error'
    organizePreviewStale.value = false
    organizePreviewError.value = error instanceof Error ? error.message : 'Preview failed.'
    addActionError('Organize preview', organizePreviewError.value, requestStarted)
  }
}

function shouldDefaultToEmbeddedFileMetadata(response: OrganizePreviewResponse): boolean {
  return (
    activeWorkflow.value === 'organize' &&
    scanMode.value === 'json' &&
    !scanModeUserSelected.value &&
    response.summary.MetadataFound.length === 0 &&
    response.summary.MetadataMissing.length > 0 &&
    workflowScanModes.value.some((mode) => mode.value === 'embedded-file')
  )
}

async function createRenamePreview() {
  renamePreviewStatus.value = 'loading'
  renamePreviewError.value = ''
  renameRun.value = null
  renameRunStatus.value = 'idle'
  renameRunError.value = ''
  renamePreviewStale.value = false
  let requestStarted = false

  try {
    if (!sourceFolder.value.trim()) {
      throw new Error('Source folder is required for rename preview.')
    }
    if (!renameTemplate.value.trim()) {
      throw new Error('Rename template is required for preview.')
    }
    addRequestStart('Rename preview', 'POST /api/rename/preview')
    requestStarted = true
    const response = normalizeRenameResponse(
      await apiPost<RenamePreviewResponse>('/api/rename/preview', {
        config: buildRenameConfig(true),
      }),
    )
    renamePreview.value = response
    selectedRenamePaths.value = response.candidates
      .filter((candidate) => isRenameCandidateSelectable(candidate))
      .map((candidate) => candidate.CurrentPath)
    renamePreviewStatus.value = 'success'
    renamePreviewStale.value = false
    addRequestSuccess(
      'Rename preview',
      `${response.candidates.length} candidate(s), ${response.summary.ConflictsFound} conflict(s).`,
    )
  } catch (error) {
    renamePreview.value = null
    selectedRenamePaths.value = []
    renamePreviewStatus.value = 'error'
    renamePreviewStale.value = false
    renamePreviewError.value = error instanceof Error ? error.message : 'Rename preview failed.'
    addActionError('Rename preview', renamePreviewError.value, requestStarted)
  }
}

async function loadABSLibraries() {
  absLibrariesStatus.value = 'loading'
  absLibrariesError.value = ''
  absLibraries.value = []
  let requestStarted = false

  try {
    addRequestStart('ABS libraries', 'POST /api/abs/libraries')
    requestStarted = true
    const response = await apiPost<ABSLibrariesResponse>('/api/abs/libraries', buildABSConfig())
    absLibraries.value = Array.isArray(response.libraries) ? response.libraries : []
    if (!absLibraries.value.some((library) => library.id === absLibrary.value)) {
      absLibrary.value = absLibraries.value[0]?.id ?? ''
    }
    absLibrariesStatus.value = 'success'
    addRequestSuccess('ABS libraries', `${absLibraries.value.length} library/libraries returned.`)
  } catch (error) {
    absLibrary.value = ''
    absLibrariesStatus.value = 'error'
    absLibrariesError.value = error instanceof Error ? error.message : 'ABS library request failed.'
    addActionError('ABS libraries', absLibrariesError.value, requestStarted)
  }
}

async function testABSPathMappings() {
  absPathStatus.value = 'loading'
  absPathError.value = ''
  absResolvedMappings.value = []
  let requestStarted = false

  try {
    addRequestStart('ABS path validation', 'POST /api/abs/test-paths')
    requestStarted = true
    const response = await apiPost<ABSPathMappingResponse>('/api/abs/test-paths', {
      input_dir: sourceFolder.value.trim(),
      config: buildABSConfig(),
    })
    absResolvedMappings.value = response.mappings ?? []
    absPathStatus.value = 'success'
    addRequestSuccess('ABS path validation', `${absResolvedMappings.value.length} path mapping(s) resolved.`)
    if (activeWorkflow.value === 'organize' && scanMode.value === 'abs') {
      scheduleActivePreviewRefresh()
    }
  } catch (error) {
    absPathStatus.value = 'error'
    absPathError.value = error instanceof Error ? error.message : 'ABS path validation failed.'
    addActionError('ABS path validation', absPathError.value, requestStarted)
  }
}

async function loadABSItems() {
  absItemsStatus.value = 'loading'
  absItemsError.value = ''
  absItems.value = null
  let requestStarted = false

  try {
    assertABSSetupReady()
    addRequestStart('ABS items', 'POST /api/abs/items')
    requestStarted = true
    const response = await apiPost<ABSItemsResponse>('/api/abs/items', {
      config: buildABSConfig(),
    })
    absItems.value = { items: response.items ?? [] }
    absItemsStatus.value = 'success'
    addRequestSuccess('ABS items', `${absItems.value.items.length} metadata item(s) returned.`)
  } catch (error) {
    absItemsStatus.value = 'error'
    absItemsError.value = error instanceof Error ? error.message : 'ABS item loading failed.'
    addActionError('ABS items', absItemsError.value, requestStarted)
  }
}

async function loadABSLibraryState() {
  absLibraryStateStatus.value = 'loading'
  absLibraryStateError.value = ''
  absLibraryState.value = null
  let requestStarted = false

  try {
    assertABSSetupReady()
    addRequestStart('ABS library state', 'POST /api/abs/library-state')
    requestStarted = true
    const response = await apiPost<ABSLibraryStateResponse>('/api/abs/library-state', {
      config: buildABSConfig(),
    })
    absLibraryState.value = { ...response, items: response.items ?? [] }
    absLibraryStateStatus.value = 'success'
    addRequestSuccess(
      'ABS library state',
      `${absLibraryState.value.items.length} item(s), ${absMissingCount.value} missing, ${absInvalidCount.value} invalid.`,
    )
  } catch (error) {
    absLibraryStateStatus.value = 'error'
    absLibraryStateError.value = error instanceof Error ? error.message : 'ABS library state request failed.'
    addActionError('ABS library state', absLibraryStateError.value, requestStarted)
  }
}

async function triggerABSScan() {
  absScanStatus.value = 'loading'
  absScanError.value = ''
  absScanResult.value = null
  let requestStarted = false

  try {
    assertABSSetupReady()
    addRequestStart('ABS scan trigger', 'POST /api/abs/scan-trigger')
    requestStarted = true
    const response = await apiPost<ABSScanTriggerResponse>('/api/abs/scan-trigger', {
      config: buildABSConfig(),
    })
    absScanResult.value = response
    absScanStatus.value = 'success'
    addRequestSuccess(
      'ABS scan trigger',
      response.triggered ? `Library ${response.library_id} accepted the scan request.` : 'Backend did not report a scan trigger.',
    )
  } catch (error) {
    absScanStatus.value = 'error'
    absScanError.value = error instanceof Error ? error.message : 'ABS scan trigger failed.'
    addActionError('ABS scan trigger', absScanError.value, requestStarted)
  }
}

async function cleanABSMissing() {
  absCleanStatus.value = 'loading'
  absCleanError.value = ''
  absCleanResult.value = null
  let requestStarted = false

  try {
    assertABSSetupReady()
    if (!absCleanConfirmed.value) {
      throw new Error('Confirm missing-item cleanup before running this destructive action.')
    }
    if (!window.confirm('Clean missing ABS item records for this library? This cannot be undone from Audiobook Organizer.')) {
      absCleanStatus.value = 'idle'
      return
    }
    addRequestStart('ABS missing item cleanup', 'POST /api/abs/clean-missing')
    requestStarted = true
    const response = await apiPost<ABSCleanMissingResponse>('/api/abs/clean-missing', {
      config: buildABSConfig(),
    })
    absCleanResult.value = response
    absCleanStatus.value = 'success'
    absCleanConfirmed.value = false
    addRequestSuccess(
      'ABS missing item cleanup',
      response.cleaned ? `Library ${response.library_id} cleanup completed.` : 'Backend did not report cleanup.',
    )
  } catch (error) {
    absCleanStatus.value = 'error'
    absCleanError.value = error instanceof Error ? error.message : 'ABS missing-item cleanup failed.'
    addActionError('ABS missing item cleanup', absCleanError.value, requestStarted)
  }
}

function addABSPathMapping() {
  absPathMappings.value = [...absPathMappings.value, { abs_prefix: '', local_prefix: '' }]
}

function removeABSPathMapping(index: number) {
  if (absPathMappings.value.length === 1) {
    return
  }
  absPathMappings.value = absPathMappings.value.filter((_, mappingIndex) => mappingIndex !== index)
  resetABSPathResults()
}

async function runOrganize() {
  if (!canOpenOrganizeReview.value || organizeRunStatus.value === 'loading') {
    return
  }
  if (selectedOrganizeMoveCount.value === 0) {
    organizeRunError.value = 'Select at least one planned move before running organize.'
    addActionError('Organize run', organizeRunError.value, false)
    return
  }
  if (
    !window.confirm(
      `Run Organize will change files for ${selectedOrganizeMoveCount.value} selected move(s). Continue?`,
    )
  ) {
    return
  }

  organizeRunStatus.value = 'loading'
  organizeRunError.value = ''
  let requestStarted = false
  try {
    addRequestStart('Organize run', 'POST /api/organize/run')
    requestStarted = true
    const response = normalizeOrganizeResponse(
      await apiPost<OrganizeRunResponse>('/api/organize/run', {
        config: buildOrganizerConfig(false, selectedOrganizeSources.value),
      }),
    )
    organizeRun.value = response
    organizeRunStatus.value = 'success'
    activeStage.value = 'run'
    addRequestSuccess('Organize run', `${response.summary.Moves.length} file operation(s).`)
  } catch (error) {
    organizeRunStatus.value = 'error'
    organizeRunError.value = error instanceof Error ? error.message : 'Organize run failed.'
    addActionError('Organize run', organizeRunError.value, requestStarted)
  }
}

async function runRename() {
  if (!canOpenRenameReview.value || renameRunStatus.value === 'loading') {
    return
  }
  if (selectedRenameCandidateCount.value === 0) {
    renameRunError.value = 'Select at least one rename candidate before running rename.'
    addActionError('Rename run', renameRunError.value, false)
    return
  }
  if (
    !window.confirm(
      `Run Rename will change ${selectedRenameCandidateCount.value} selected file(s). Continue?`,
    )
  ) {
    return
  }

  renameRunStatus.value = 'loading'
  renameRunError.value = ''
  let requestStarted = false
  try {
    addRequestStart('Rename run', 'POST /api/rename/run')
    requestStarted = true
    const response = normalizeRenameResponse(
      await apiPost<RenameRunResponse>('/api/rename/run', {
        config: buildRenameConfig(false, selectedRenamePaths.value),
      }),
    )
    renameRun.value = response
    renameRunStatus.value = 'success'
    activeStage.value = 'run'
    addRequestSuccess('Rename run', `${response.summary.FilesRenamed} file(s) renamed.`)
  } catch (error) {
    renameRunStatus.value = 'error'
    renameRunError.value = error instanceof Error ? error.message : 'Rename run failed.'
    addActionError('Rename run', renameRunError.value, requestStarted)
  }
}

function buildOrganizerConfig(dryRun: boolean, selectedSourcePaths?: string[]): OrganizerConfig {
  const defaults = organizerDefaults.value
  const customLayoutSelected = layout.value === customLayoutValue
  const selectedLayout = customLayoutSelected ? defaults?.layout || defaultLayouts[0].value : layout.value
  return {
    base_dir: sourceFolder.value.trim(),
    output_dir: outputFolder.value.trim(),
    replace_space: defaults?.replace_space ?? '',
    dry_run: dryRun,
    remove_empty: removeEmpty.value,
    use_embedded_metadata: shouldUseEmbeddedMetadata(),
    flat: shouldUseFlatMode(),
    skip_errors: defaults?.skip_errors ?? false,
    layout: selectedLayout,
    layout_template: customLayoutSelected ? layoutTemplate.value.trim() : '',
    author_format: defaults?.author_format || 'first-last',
    field_mapping: cloneFieldMapping(organizeFieldMapping.value),
    allowed_source_paths: selectedSourcePaths ?? defaults?.allowed_source_paths,
    metadata_source: scanMode.value,
    abs: scanMode.value === 'abs' ? buildABSConfig() : undefined,
  }
}

function buildRenameConfig(dryRun: boolean, selectedCurrentPaths?: string[]): RenameConfig {
  const defaults = renameDefaults.value
  return {
    base_dir: sourceFolder.value.trim(),
    template: renameTemplate.value.trim(),
    dry_run: dryRun,
    author_format: defaults?.author_format || 'first-last',
    recursive: renameRecursive.value,
    field_mapping: cloneFieldMapping(renameFieldMapping.value),
    replace_space: defaults?.replace_space ?? '',
    strict_mode: defaults?.strict_mode ?? false,
    preserve_path: preservePath.value,
    use_embedded_metadata: shouldUseEmbeddedMetadata(),
    allowed_current_paths: selectedCurrentPaths ?? defaults?.allowed_current_paths,
  }
}

function applyFieldMappingPreset(preset: string) {
  const mapping = fieldMappings.value[preset]
  if (!mapping) {
    return
  }
  if (activeWorkflow.value === 'rename') {
    renameFieldMapping.value = cloneFieldMapping(mapping)
    return
  }
  organizeFieldMapping.value = cloneFieldMapping(mapping)
}

function updateAuthorFieldMapping(value: string) {
  activeFieldMapping.value.author_fields = value
    .split(',')
    .map((field) => field.trim())
    .filter(Boolean)
}

function cloneFieldMapping(mapping: FieldMapping): FieldMapping {
  return { ...mapping, author_fields: [...(mapping.author_fields ?? [])] }
}

function fieldMappingsEqual(left: FieldMapping, right: FieldMapping): boolean {
  return (
    left.title_field === right.title_field &&
    left.series_field === right.series_field &&
    left.track_field === right.track_field &&
    left.disc_field === right.disc_field &&
    (left.author_fields ?? []).join('\u0000') === (right.author_fields ?? []).join('\u0000')
  )
}

function buildABSConfig(): ABSConfig {
  const headers =
    absHeaderName.value.trim() && absHeaderValue.value
      ? [{ name: absHeaderName.value.trim(), value: absHeaderValue.value }]
      : undefined
  return {
    url: absUrl.value.trim(),
    token: absToken.value,
    library_id: absLibrary.value.trim(),
    sqlite_path: absSQLitePath.value.trim() || undefined,
    path_mappings: absPathMappings.value
      .map((mapping) => ({
        abs_prefix: mapping.abs_prefix.trim(),
        local_prefix: mapping.local_prefix.trim(),
      }))
      .filter((mapping) => mapping.abs_prefix || mapping.local_prefix),
    all_libraries: false,
    headers,
  }
}

function assertABSSetupReady() {
  if (!absSetupReady.value) {
    throw new Error('ABS setup must load libraries and validate paths first.')
  }
}

function shouldUseEmbeddedMetadata() {
  return scanMode.value === 'embedded-directory' || scanMode.value === 'embedded-file'
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
  selectedOrganizeSources.value = []
  organizePreviewStale.value = false
  organizePreviewStatus.value = 'idle'
  organizeRunStatus.value = 'idle'
  organizePreviewError.value = ''
  organizeRunError.value = ''
}

function resetRenameResults() {
  renamePreview.value = null
  renameRun.value = null
  selectedRenamePaths.value = []
  renamePreviewStale.value = false
  renamePreviewStatus.value = 'idle'
  renameRunStatus.value = 'idle'
  renamePreviewError.value = ''
  renameRunError.value = ''
}

function scheduleActivePreviewRefresh() {
  if (!bootstrapComplete.value) {
    return
  }
  if (autoPreviewTimer) {
    window.clearTimeout(autoPreviewTimer)
    autoPreviewTimer = null
  }
  if (!canAutoPreviewActiveWorkflow()) {
    return
  }
  autoPreviewTimer = window.setTimeout(() => {
    autoPreviewTimer = null
    void refreshActivePreview()
  }, 550)
}

function canAutoPreviewActiveWorkflow() {
  if (pathValidationStatus.value === 'loading') {
    return false
  }
  if (activeWorkflow.value === 'organize') {
    return (
      !!sourceFolder.value.trim() &&
      !!outputFolder.value.trim() &&
      organizePreviewStatus.value !== 'loading'
    )
  }
  if (activeWorkflow.value === 'rename') {
    return (
      !!sourceFolder.value.trim() &&
      !!renameTemplate.value.trim() &&
      renamePreviewStatus.value !== 'loading'
    )
  }
  return false
}

async function refreshActivePreview() {
  if (!canAutoPreviewActiveWorkflow()) {
    return
  }
  if (!(await validateConfigurePaths())) {
    return
  }
  if (activeWorkflow.value === 'rename') {
    await createRenamePreview()
    return
  }
  if (activeWorkflow.value === 'organize') {
    await createOrganizePreview()
  }
}

function markOrganizePreviewStale() {
  organizeRun.value = null
  organizeRunStatus.value = 'idle'
  organizeRunError.value = ''
  if (organizePreview.value && organizePreviewStatus.value === 'success') {
    organizePreviewStale.value = true
    return
  }
  resetOrganizeResults()
}

function markRenamePreviewStale() {
  renameRun.value = null
  renameRunStatus.value = 'idle'
  renameRunError.value = ''
  if (renamePreview.value && renamePreviewStatus.value === 'success') {
    renamePreviewStale.value = true
    return
  }
  resetRenameResults()
}

function resetABSConnectionResults() {
  absLibraries.value = []
  absLibrary.value = ''
  absLibrariesStatus.value = 'idle'
  absLibrariesError.value = ''
  resetABSOperationResults()
  resetABSPathResults()
}

function resetABSPathResults() {
  absResolvedMappings.value = []
  absPathStatus.value = 'idle'
  absPathError.value = ''
  resetABSOperationResults()
}

function resetABSOperationResults() {
  absItems.value = null
  absLibraryState.value = null
  absScanResult.value = null
  absCleanResult.value = null
  absItemsStatus.value = 'idle'
  absLibraryStateStatus.value = 'idle'
  absScanStatus.value = 'idle'
  absCleanStatus.value = 'idle'
  absItemsError.value = ''
  absLibraryStateError.value = ''
  absScanError.value = ''
  absCleanError.value = ''
  absCleanConfirmed.value = false
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

function normalizeRenameResponse<T extends RenamePreviewResponse | RenameRunResponse>(response: T): T {
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

function setInitialScanMode(config: WebConfig) {
  if (config.organizer?.flat) {
    scanMode.value = 'embedded-file'
    return
  }
  if (config.organizer?.use_embedded_metadata || config.rename?.use_embedded_metadata) {
    scanMode.value = 'embedded-directory'
    return
  }
  scanMode.value = 'json'
}

function now() {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

onMounted(async () => {
  if (!hasWebSessionToken) {
    bootstrapComplete.value = true
    return
  }

  addRequestStart('Health check', 'GET /api/health')
  try {
    const response = await apiGet<HealthResponse>('/api/health')
    health.value = response.status
    addRequestSuccess('Health check', `Server reported ${response.status}.`)
  } catch {
    health.value = 'offline'
    addRequestError('Health check', 'Server health request failed.')
  }

  addRequestStart('Initial config', 'GET /api/config/initial')
  try {
    const config = await apiGet<WebConfig>('/api/config/initial')
    organizerDefaults.value = config.organizer
    renameDefaults.value = config.rename
    organizeFieldMapping.value = cloneFieldMapping(config.organizer?.field_mapping ?? defaultFieldMapping)
    renameFieldMapping.value = cloneFieldMapping(config.rename?.field_mapping ?? defaultFieldMapping)
    sourceFolder.value = config.initial?.input_dir || config.organizer?.base_dir || ''
    outputFolder.value = config.initial?.output_dir || config.organizer?.output_dir || ''
    layoutTemplate.value = config.organizer?.layout_template || ''
    layout.value = layoutTemplate.value ? customLayoutValue : config.organizer?.layout || layout.value
    removeEmpty.value = config.organizer?.remove_empty ?? false
    setInitialScanMode(config)
    renameTemplate.value = config.rename?.template || renameTemplate.value
    renameRecursive.value = config.rename?.recursive ?? true
    preservePath.value = config.rename?.preserve_path ?? true
    absUrl.value = config.abs?.url ?? ''
    absLibrary.value = config.abs?.library_id || absLibrary.value
    absCredentialState.value = config.abs?.token === 'redacted' ? 'redacted' : 'empty'
    absSQLitePath.value = config.abs?.sqlite_path ?? ''
    absPathMappings.value =
      config.abs?.path_mappings && config.abs.path_mappings.length > 0
        ? config.abs.path_mappings.map((mapping) => ({
            abs_prefix: mapping.abs_prefix,
            local_prefix: mapping.local_prefix,
          }))
        : absPathMappings.value
    configState.value = 'ready'
    addRequestSuccess('Initial config', 'Startup config is ready.')
  } catch {
    configState.value = 'fallback'
    addRequestError('Initial config', 'Config unavailable. Using local defaults.')
  }

  addRequestStart('Config options', 'GET /api/config/options')
  try {
    const options = await apiGet<OptionsResponse>('/api/config/options')
    layouts.value = Array.isArray(options.layouts) ? options.layouts : []
    scanModes.value = Array.isArray(options.scan_modes) ? options.scan_modes : []
    fieldMappings.value = options.field_mappings ?? {}
    optionsState.value = 'ready'
    ensureScanModeFitsWorkflow()
    addRequestSuccess('Config options', 'Layout and scan mode options are ready.')
  } catch {
    optionsState.value = 'fallback'
    ensureScanModeFitsWorkflow()
    addRequestError('Config options', 'Options unavailable. Using built-in option labels.')
  }
  bootstrapComplete.value = true
  scheduleActivePreviewRefresh()
})

watch(layout, () => {
  if (layout.value === customLayoutValue && !layoutTemplate.value.trim()) {
    layoutTemplate.value = defaultCustomLayoutTemplate
  }
})

watch([sourceFolder, outputFolder, scanMode, layout, layoutTemplate, removeEmpty, organizeFieldMapping], () => {
  if (activeWorkflow.value !== 'organize') {
    return
  }
  markOrganizePreviewStale()
  scheduleActivePreviewRefresh()
}, { deep: true })

watch([sourceFolder, scanMode, renameTemplate, renameRecursive, preservePath, renameFieldMapping], () => {
  if (activeWorkflow.value !== 'rename') {
    return
  }
  markRenamePreviewStale()
  scheduleActivePreviewRefresh()
}, { deep: true })

watch([absUrl, absToken, absHeaderName, absHeaderValue], () => {
  if (!showABSSetup.value) {
    return
  }
  resetABSConnectionResults()
})

watch([absLibrary], () => {
  if (!showABSSetup.value) {
    return
  }
  resetABSOperationResults()
})

watch([sourceFolder, absSQLitePath, absPathMappings], () => {
  if (!showABSSetup.value) {
    return
  }
  resetABSPathResults()
}, { deep: true })
</script>
