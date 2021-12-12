<template>
  <div
    class="page"
    :class="{
      'theme-light': theme === Theme.Light,
      'theme-dark': theme === Theme.Dark,
    }"
  >
    <div class="page-header">
      <Header @theme="switchTheme" />
    </div>
    <div class="page-body">
      <router-view />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import Header from "@/components/header.vue"
import { Theme } from "@/platform/theme"

export default defineComponent({
  name: "App",
  components: {
    Header,
  },
  data() {
    return {
      Theme,
      theme: Theme.Light,
    }
  },
  methods: {
    switchTheme(theme: Theme) {
      this.theme = theme
    },
  },
})
</script>

<style lang="sass">
@use 'bootstrap-4-grid/scss/grid.scss'
@use '@/css/clear'
@use '@/css/theme'

html,
body,
#app
  height: 100vh

.page
  font-family: 'Roboto',-apple-system,BlinkMacSystemFont,'Segoe UI','Oxygen','Ubuntu','Cantarell','Fira Sans','Droid Sans','Helvetica Neue',sans-serif
  -webkit-font-smoothing: antialiased
  -moz-osx-font-smoothing: grayscale
  height: 100%
  margin: -8px
  text-align: center
  color: var(--color-font)
  background-color: var(--color-background)

  display: flex
  flex-flow: column

.page-header
  width: 100%
  height: 60px

.page-body
  flex-grow: 1

  width: 100%
  height: 100px

.theme-toggle
  @include theme.clickable

  position: absolute
  right: 10px
  top: 10px
  color: inherit
</style>
