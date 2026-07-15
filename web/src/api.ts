export type HealthResponse = {
  status: string
}

export type Option = {
  value: string
  label: string
  description?: string
}

export type FieldMapping = {
  title_field?: string
  series_field?: string
  author_fields?: string[]
  track_field?: string
  disc_field?: string
}

export type OrganizerConfig = {
  base_dir: string
  output_dir: string
  replace_space: string
  dry_run: boolean
  remove_empty: boolean
  use_embedded_metadata: boolean
  flat: boolean
  skip_errors: boolean
  layout: string
  layout_template: string
  author_format: string
  field_mapping: FieldMapping
  allowed_source_paths?: string[]
  metadata_source?: string
  abs?: ABSConfig
}

export type MoveSummary = {
  from: string
  to: string
}

export type OrganizerSummary = {
  MetadataFound: string[]
  MetadataMissing: string[]
  Moves: MoveSummary[]
  EmptyDirsRemoved: string[]
}

export type OrganizePreviewResponse = {
  summary: OrganizerSummary
  log_path?: string
}

export type OrganizeRunResponse = {
  summary: OrganizerSummary
  log_path?: string
}

export type RenameConfig = {
  base_dir: string
  template: string
  dry_run: boolean
  author_format: string
  recursive: boolean
  field_mapping: FieldMapping
  replace_space: string
  strict_mode: boolean
  preserve_path: boolean
  use_embedded_metadata: boolean
  allowed_current_paths?: string[]
}

export type RenameMetadata = Record<string, unknown>

export type RenameCandidate = {
  CurrentPath: string
  ProposedPath: string
  Metadata: RenameMetadata
  IsNoOp: boolean
  IsConflict: boolean
  Error: string
}

export type RenameSummary = {
  FilesScanned: number
  FilesRenamed: number
  FilesSkipped: number
  ConflictsFound: number
  Errors: string[]
}

export type RenamePreviewResponse = {
  candidates: RenameCandidate[]
  summary: RenameSummary
}

export type RenameRunResponse = {
  candidates: RenameCandidate[]
  summary: RenameSummary
  log_path?: string
}

export type PathMapping = {
  abs_prefix: string
  local_prefix: string
}

export type HeaderConfig = {
  name: string
  value?: string
}

export type ABSConfig = {
  url: string
  token?: string
  library_id: string
  sqlite_path?: string
  path_mappings?: PathMapping[]
  all_libraries: boolean
  header_file?: string
  headers?: HeaderConfig[]
}

export type ABSLibraryFolder = {
  id: string
  path: string
  fullPath: string
  libraryId?: string
}

export type ABSLibrary = {
  id: string
  name: string
  mediaType: string
  folders?: ABSLibraryFolder[]
}

export type ABSLibrariesResponse = {
  libraries: ABSLibrary[]
}

export type ABSPathMappingResponse = {
  mappings: PathMapping[]
}

export type ABSMetadataItem = {
  title: string
  authors: string[]
  series: string[]
  source_type: string
  source_path: string
}

export type ABSItemsResponse = {
  items: ABSMetadataItem[]
}

export type ABSLibraryItem = {
  id: string
  path: string
  rel_path: string
  is_missing: boolean
  is_invalid: boolean
  media_type: string
  title?: string
}

export type ABSLibraryStateResponse = {
  library_id: string
  items: ABSLibraryItem[]
}

export type ABSScanTriggerResponse = {
  triggered: boolean
  library_id: string
}

export type ABSCleanMissingResponse = {
  cleaned: boolean
  library_id: string
}

export type PathValidationItem = {
  id: string
  path: string
  kind: 'existing-directory' | 'output-directory'
}

export type PathValidationResponse = {
  results: Array<{
    id: string
    path: string
    valid: boolean
    error?: string
  }>
}

export type WebConfig = {
  host: string
  port: number
  open: boolean
  initial: {
    input_dir: string
    output_dir: string
  }
  organizer: OrganizerConfig
  rename: RenameConfig
  abs: ABSConfig
}

export type OptionsResponse = {
  layouts: Option[]
  scan_modes: Option[]
  author_formats: Option[]
  field_mappings: Record<string, FieldMapping>
}

const token = new URLSearchParams(window.location.search).get('token') ?? ''

export const hasWebSessionToken = token !== ''

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    headers: token ? { 'X-Audiobook-Organizer-Token': token } : undefined,
  })
  return decode<T>(response)
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'X-Audiobook-Organizer-Token': token } : {}),
    },
    body: JSON.stringify(body),
  })
  return decode<T>(response)
}

async function decode<T>(response: Response): Promise<T> {
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    const message = typeof payload.error === 'string' ? payload.error : response.statusText
    throw new Error(message)
  }
  return payload as T
}
