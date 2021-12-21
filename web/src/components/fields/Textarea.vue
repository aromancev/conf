<template>
  <div class="field" :style="style">
    <textarea
      ref="textarea"
      :style="style"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :spellcheck="spellcheck"
      @input="$emit('update:modelValue', $event.target.value)"
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

<script lang="ts">
import { defineComponent } from "vue"

export default defineComponent({
  name: "Loader",
  props: {
    placeholder: {
      type: String,
    },
    spellcheck: {
      type: Boolean,
    },
    modelValue: {
      type: String,
      required: true,
    },
    error: {
      type: String,
    },
    disabled: {
      type: Boolean,
    },
  },
  watch: {
    modelValue() {
      this.alignHeight()
    },
  },
  data() {
    return {
      errExpanded: false,
      style: {
        height: "0",
      },
    }
  },
  mounted() {
    this.alignHeight()
  },
  methods: {
    alignHeight() {
      // Woodoo magic to auto-resize textarea based on content.
      if (this.modelValue.length === 0) {
        this.style["height"] = "0"
        return
      }
      const textarea = this.$refs.textarea as HTMLTextAreaElement
      textarea.style.height = "0"
      this.style["height"] = `${textarea.scrollHeight}px`
    },
  },
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

.field
  @include theme.shadow-inset-xs

  position: relative
  width: 150px
  border-radius: 4px
  min-height: 3em

textarea
  width: 100%
  min-height: 100%
  max-height: 100%
  resize: none
  padding: 1em 1em

textarea:disabled
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
 