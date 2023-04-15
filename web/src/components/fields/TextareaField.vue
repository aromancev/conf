<template>
  <div class="field">
    <div class="label">{{ label }}</div>
    <div class="area" :style="{ height: height }">
      <textarea
        ref="textarea"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :spellcheck="spellcheck"
        @input="input"
      ></textarea>
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
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue"

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void
}>()

const props = defineProps<{
  modelValue: string
  label?: string
  placeholder?: string
  disabled?: boolean
  spellcheck?: boolean
  error?: string
}>()

const errExpanded = ref(false)
const height = ref("0")
const textarea = ref<HTMLTextAreaElement | null>(null)

watch(
  () => props.modelValue,
  () => {
    alignHeight()
  },
)

onMounted(() => {
  alignHeight()
})

function input(event: Event) {
  emit("update:modelValue", (event.target as HTMLInputElement).value)
}

function alignHeight() {
  // Woodoo magic to auto-resize textarea based on content.
  if (!textarea.value) {
    return
  }

  if (props.modelValue.length === 0) {
    height.value = "0"
    return
  }

  textarea.value.style.height = "0"
  height.value = `${textarea.value.scrollHeight}px`
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.field
  @include theme.shadow-inset-xs

  border-radius: 4px
  padding: 0.5em 1em

.area
  position: relative
  min-height: 1.25em

textarea
  box-sizing: border-box
  width: 100%
  min-height: 100%
  max-height: 100%
  resize: none
  overflow: overlay
  &:disabled
    cursor: default

textarea:disabled
  color: var(--color-font-disabled)

.label
  text-align: left
  font-size: 0.7em
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
