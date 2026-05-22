<template>
  <div class="template-builder">
    <label>{{ label }}</label>
    <div class="template-builder-input-row">
      <input
        ref="templateInput"
        :value="modelValue"
        :aria-label="label"
        :placeholder="placeholder"
        @input="updateTemplate"
      />
      <button
        class="icon-button"
        type="button"
        :disabled="undoStack.length === 0"
        aria-label="Undo template edit"
        title="Undo"
        @click="undoTemplateEdit"
      >
        <RotateCcw :size="16" />
      </button>
      <button
        class="icon-button"
        type="button"
        :disabled="!modelValue"
        aria-label="Clear template"
        title="Clear"
        @click="clearTemplate"
      >
        <X :size="16" />
      </button>
    </div>
    <div class="template-token-toolbar" aria-label="Insert template field">
      <button
        v-for="field in fields"
        :key="field.value"
        class="template-token-button"
        :class="field.kind"
        type="button"
        @click="insertTemplateToken(field.value)"
      >
        <Plus :size="14" /> {{ field.label }}
      </button>
    </div>
    <div class="template-preview" aria-label="Template preview">
      <template v-for="(part, index) in templateParts" :key="index">
        <span v-if="part.kind === 'text'">{{ part.value }}</span>
        <code v-else class="template-token" :class="part.kind">{{ part.value }}</code>
      </template>
    </div>
    <p v-if="hint" class="hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, RotateCcw, X } from 'lucide-vue-next'

export type TemplateFieldKind = 'author' | 'series' | 'title' | 'other'

export type TemplateField = {
  value: string
  label: string
  kind: TemplateFieldKind
}

type TemplatePart = {
  kind: TemplateFieldKind | 'text'
  value: string
}

const props = defineProps<{
  modelValue: string
  label: string
  placeholder: string
  fields: TemplateField[]
  emptyText?: string
  hint?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const templateInput = ref<HTMLInputElement | null>(null)
const undoStack = ref<string[]>([])

const fieldKindByValue = computed(() => {
  return new Map(props.fields.map((field) => [field.value.toLowerCase(), field.kind]))
})

const templateParts = computed<TemplatePart[]>(() => tokenizeTemplate(props.modelValue))

function updateTemplate(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}

function insertTemplateToken(field: string) {
  const token = `{${field}}`
  const input = templateInput.value
  pushUndoState()

  if (!input) {
    emit('update:modelValue', `${props.modelValue}${token}`)
    return
  }

  const start = input.selectionStart ?? props.modelValue.length
  const end = input.selectionEnd ?? start
  emit('update:modelValue', `${props.modelValue.slice(0, start)}${token}${props.modelValue.slice(end)}`)
  window.setTimeout(() => {
    input.focus()
    input.setSelectionRange(start + token.length, start + token.length)
  }, 0)
}

function clearTemplate() {
  if (!props.modelValue) {
    return
  }
  pushUndoState()
  emit('update:modelValue', '')
  window.setTimeout(() => templateInput.value?.focus(), 0)
}

function undoTemplateEdit() {
  const previous = undoStack.value.at(-1)
  if (previous === undefined) {
    return
  }
  undoStack.value = undoStack.value.slice(0, -1)
  emit('update:modelValue', previous)
  window.setTimeout(() => templateInput.value?.focus(), 0)
}

function pushUndoState() {
  const previous = props.modelValue
  if (undoStack.value.at(-1) === previous) {
    return
  }
  undoStack.value = [...undoStack.value.slice(-9), previous]
}

function tokenizeTemplate(template: string): TemplatePart[] {
  const parts: TemplatePart[] = []
  const tokenPattern = /\{([^{}]+)\}/g
  let cursor = 0

  for (const match of template.matchAll(tokenPattern)) {
    const index = match.index ?? 0
    if (index > cursor) {
      parts.push({ kind: 'text', value: template.slice(cursor, index) })
    }
    const token = match[0]
    parts.push({ kind: tokenKind(match[1]), value: token })
    cursor = index + token.length
  }

  if (cursor < template.length) {
    parts.push({ kind: 'text', value: template.slice(cursor) })
  }

  return parts.length > 0 ? parts : [{ kind: 'text', value: props.emptyText ?? 'Select fields to build a template.' }]
}

function tokenKind(field: string): TemplatePart['kind'] {
  const normalized = field.split('|')[0]?.trim().toLowerCase() ?? ''
  return fieldKindByValue.value.get(normalized) ?? 'other'
}
</script>
