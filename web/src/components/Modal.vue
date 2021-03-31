<template>
  <div @click="close" class="background"></div>
  <div class="content">
    <div class="px-5 py-2">
      <slot></slot>
    </div>
    <table v-if="buttons">
      <tr>
        <td
          class="py-2"
          v-for="(text, id) in buttons"
          v-bind:key="id"
          @click="click(id)"
        >
          {{ text }}
        </td>
      </tr>
    </table>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"

export default defineComponent({
  name: "Modal",
  props: {
    buttons: {} as Record<string, string>
  },
  emits: ["click"],
  methods: {
    click(id: string) {
      this.$emit("click", id)
    }
  }
})
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

table
  border-top: 1px solid var(--color-outline)
  min-width: 100%
  table-layout: fixed

td
  @include theme.clickable
  font-weight: 500
  &:hover
    background-color: var(--color-outline)

td + td
  border-left: 1px solid var(--color-outline)
</style>
