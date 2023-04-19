<template>
  <InputField
    v-model="state.value"
    v-click-outside="submit"
    type="text"
    class="editable-field"
    :class="{ focused: state.isEditing }"
    :errors="errors"
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
  (e: "discard", value: string): void
}>()

const props = defineProps<{
  value: string
  validate?: (val: string) => string[]
}>()

type State = {
  value: string
  isEditing: boolean
}

const state = reactive<State>({
  value: props.value,
  isEditing: false,
})

const errors = computed<string[]>(() => {
  if (props.validate) {
    return props.validate(state.value)
  }
  return []
})

watch(
  () => props.value,
  (val) => {
    state.value = val.trim()
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
    submit()
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
  if (errors.value.length) {
    discard()
    return
  }
  state.value = state.value.trim()
  state.isEditing = false
  emit("update", state.value)
}

function discard() {
  const discardedValue = state.value.trim()
  state.isEditing = false
  state.value = props.value
  emit("discard", discardedValue)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.field.editable-field:not(.fake):not(.fake)
  border-radius: 4px
  padding: 0
  box-shadow: none
  cursor: text
  &:hover
    outline: 1px solid var(--color-outline)
  &.focused
    @include theme.shadow-inset-xs
    &:hover
      outline: none
</style>
