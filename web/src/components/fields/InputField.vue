<template>
  <div class="field" tabindex="0" @focus="focusInput">
    <div class="label">{{ label }}</div>
    <input
      ref="input"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :spellcheck="spellcheck"
      :disabled="disabled"
      :autocomplete="autocomplete"
      tabindex="-1"
      @input="change"
      @keydown="keySend"
    />
    <div v-if="errors?.length && errExpanded" class="error">
      {{  errors.map((err: string) => "• " + err).join("\n") }}
    </div>
    <div
      v-if="errors?.length"
      class="error-badge material-icons"
      @mouseenter="errExpanded = true"
      @mouseleave="errExpanded = false"
    >
      priority_high
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void
}>()

defineProps<{
  type: string
  modelValue: string
  label?: string
  autocomplete?: string
  placeholder?: string
  spellcheck?: boolean
  disabled?: boolean
  errors?: string[]
}>()

const errExpanded = ref(false)
const input = ref<HTMLElement>()

function focusInput() {
  input.value?.focus()
}

function keySend(ev: KeyboardEvent) {
  if (ev.code !== "Enter" && ev.code !== "Escape") {
    return
  }
  ev.preventDefault()
  input.value?.blur()
}

function change(event: Event) {
  emit("update:modelValue", (event.target as HTMLInputElement).value)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.field
  @include theme.shadow-inset-xs

  position: relative
  border-radius: 4px
  padding: 0.5em 1em
  text-align: left

.label
  font-size: 0.7em
  color: var(--color-font-disabled)

input
  width: 100%

input:disabled
  color: var(--color-font-disabled)

.error-badge
  @include theme.clickable

  cursor: default
  right: 10px
  top: 50%
  transform: translateY(-50%)
  position: absolute
  color: var(--color-highlight-font)
  font-size: 12px
  padding: 4px
  background-color: var(--color-red)
  border-radius: 50%
  z-index: 100

.error
  position: absolute
  right: 15px
  top: 50%
  background-color: var(--color-red)
  color: var(--color-highlight-font)
  padding: 10px
  border-radius: 4px
  border-top-right-radius: 10px
  white-space: pre-wrap
  z-index: 100
  font-size: 12px
</style>
