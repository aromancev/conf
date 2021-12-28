<template>
  <div class="field">
    <input
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :spellcheck="spellcheck"
      :disabled="disabled"
      @input="input"
    />
    <div v-if="error && errExpanded" class="error">{{ error }}</div>
    <div
      v-if="error"
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
  placeholder?: string
  spellcheck?: boolean
  disabled?: boolean
  error?: string
}>()

const errExpanded = ref(false)

function input(event: Event) {
  emit("update:modelValue", (event.target as HTMLInputElement).value)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.field
  @include theme.shadow-inset-xs

  position: relative
  width: 150px
  border-radius: 4px
  padding: 0.5em 1em

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
</style>
