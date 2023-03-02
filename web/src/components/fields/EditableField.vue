<template>
  <InputField
    v-model="state.value"
    v-click-outside="submit"
    type="text"
    class="input"
    :class="{ focused: state.isEditing }"
    :error="error"
    @click="startEdit"
    @focus="startEdit"
    @keydown="keySend"
  ></InputField>
</template>
<script setup lang="ts">
import { reactive, watch, computed } from "vue"
import InputField from "@/components/fields/InputField.vue"

const emit = defineEmits<{
  (e: "update", value: string): void
}>()

const props = defineProps<{
  value: string
  disabled?: boolean
  validate?: (val: string) => string
}>()

type State = {
  value: string
  isEditing: boolean
}

const state = reactive<State>({
  value: props.value,
  isEditing: false,
})

const error = computed<string>(() => {
  if (props.validate) {
    return props.validate(state.value)
  }
  return ""
})

watch(
  () => props.value,
  (val) => {
    state.value = val
  },
)

function startEdit() {
  if (state.isEditing) {
    return
  }
  state.isEditing = true
  state.value = props.value
}

function keySend(ev: KeyboardEvent) {
  if (ev.code === "Enter") {
    ev.preventDefault()
    if (error.value) {
      discard()
    } else {
      submit()
    }
    return
  }
  if (ev.code === "Escape") {
    ev.preventDefault()
    discard()
    return
  }
}

function submit() {
  if (!state.isEditing) {
    return
  }
  if (error.value) {
    discard()
  }
  state.isEditing = false
  emit("update", state.value)
}

function discard() {
  state.isEditing = false
  state.value = props.value
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.input
  border-radius: 4px
  padding: 0.5em 0
  box-shadow: none
  cursor: text
  &:hover
    outline: 1px solid var(--color-outline)
  &.focused
    @include theme.shadow-inset-xs
    &:hover
      outline: none
</style>
