<template>
  <div
    class="page"
    :class="{
      'theme-light': theme === Theme.Light,
      'theme-dark': theme === Theme.Dark,
    }"
  >
    <div class="page-header">
      <PageHeader @theme="switchTheme"></PageHeader>
    </div>
    <div class="page-body">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { Theme } from "@/platform/theme"
import PageHeader from "@/components/PageHeader.vue"

const theme = ref(Theme.Light)

function switchTheme(value: Theme) {
  theme.value = value
}
</script>

<style lang="sass">
@use '@/css/clear'
@use '@/css/theme'

$header-height: 60px

html, body, #app
  margin: 0
  height: 100vh
  width: 100vw

div
  box-sizing: border-box

.page
  font-family: 'Roboto',-apple-system,BlinkMacSystemFont,'Segoe UI','Oxygen','Ubuntu','Cantarell','Fira Sans','Droid Sans','Helvetica Neue',sans-serif
  -webkit-font-smoothing: antialiased
  -moz-osx-font-smoothing: grayscale
  min-height: 100vh
  width: 100vw
  color: var(--color-font)
  background-color: var(--color-background)

.page-body
  width: 100vw
  height: calc(100vh - $header-height)
  display: flex
  flex-direction: column
  justify-content: center
  align-items: center
  overflow-y: overlay
  overflow-x: hidden

.page-header
  width: 100%
  height: $header-height
</style>
