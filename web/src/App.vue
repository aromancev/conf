<template>
  <div
    class="page"
    v-bind:class="{ 'theme-light': !isDark, 'theme-dark': isDark }"
  >
    <div @click="toggleTheme" class="theme-toggle material-icons">
      {{ this.isDark ? "light_mode" : "dark_mode" }}
    </div>
    <router-view />
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"

export default defineComponent({
  name: "App",
  data() {
    return {
      isDark: false,
    }
  },
  async mounted() {
    this.isDark = localStorage.isDark === "true" ? true : false
  },
  watch: {
    isDark(newIsDark) {
      localStorage.isDark = newIsDark
    },
  },
  methods: {
    toggleTheme() {
      this.isDark = !this.isDark
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
  height: 100%

.page
  font-family: 'Roboto',-apple-system,BlinkMacSystemFont,'Segoe UI','Oxygen','Ubuntu','Cantarell','Fira Sans','Droid Sans','Helvetica Neue',sans-serif
  -webkit-font-smoothing: antialiased
  -moz-osx-font-smoothing: grayscale
  height: 100%
  margin: -8px
  text-align: center
  color: var(--color-font)
  background-color: var(--color-background)

.theme-toggle
  @include theme.clickable

  position: absolute
  right: 10px
  top: 10px
  color: inherit
</style>
