<template>
  <div class="background"></div>
  <div class="content">
    <div class="wrapper">
      <slot></slot>
    </div>
    <table v-if="buttons">
      <tr>
        <td v-for="(text, id) in buttons" :key="id" class="cell" :class="{ disabled: disabled }" @click="click(id)">
          {{ text }}
        </td>
      </tr>
    </table>
  </div>
</template>

<script setup lang="ts">
interface Controller {
  submit(id: string): void
}

const props = defineProps<{
  buttons: Record<string, string>
  disabled?: boolean
  ctrl?: Controller
}>()

const emit = defineEmits<{
  (e: "click", id: string): void
}>()

function click(id: string) {
  if (props.disabled) {
    return
  }
  if (props.ctrl) {
    props.ctrl.submit(id)
  }
  emit("click", id)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.background
  position: fixed
  left: 0
  top: 0
  height: 100vh
  width: 100vw
  backdrop-filter: blur(3px)
  background-color: var(--color-background)
  opacity: 0.6
  z-index: 200

.aligner
  position: fixed
  top: 0
  left: 0
  height: 100vh
  width: 100vw

.content
  @include theme.shadow-l
  position: fixed
  top: 50%
  left: 50%
  transform: translate(-50%, -50%)
  border-radius: 5px
  background-color: var(--color-background)
  text-align: center
  max-width: 500px
  z-index: 250

.wrapper
  padding: 1rem 3rem

table
  border-top: 1px solid var(--color-outline)
  width: 100%
  table-layout: fixed

.cell
  @include theme.clickable
  padding: 0.5rem 0
  font-weight: 500
  &.disabled
    cursor: default
    background-color: var(--color-fade-background)
  &:hover:not(.disabled)
    background-color: var(--color-highlight-background)

.cell + .cell
  border-left: 1px solid var(--color-outline)
</style>
