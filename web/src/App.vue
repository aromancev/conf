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
    <transition name="bubble">
      <div v-if="notificationStore.state.message" class="notification">
        <div class="notification-message" :class="{ error: notificationStore.state.level === 'error' }">
          {{ notificationStore.state.message }}
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { Theme } from "@/platform/theme"
import { notificationStore } from "./api/models/notifications"
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
  overflow: hidden

a
  text-decoration: none

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
  text-align: center

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

.notification
  width: 100%
  position: fixed
  bottom: 50px
  display: flex
  justify-content: center
  z-index: 200
  &.bubble-enter-active,
  &.bubble-leave-active
    transition: opacity .2s, bottom .2s

  &.bubble-enter-from,
  &.bubble-leave-to
    opacity: 0

  &.bubble-enter-from
    bottom: 0

.notification-message
  cursor: default
  border-radius: 4px
  background: #202124
  color: #fefefe
  padding: 0.5em 1em
  &.error
    background: var(--color-red)
</style>
