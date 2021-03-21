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
      isDark: false
    }
  },
  mounted() {
    this.isDark = localStorage.isDark === "true" ? true : false
  },
  watch: {
    isDark(newIsDark) {
      localStorage.isDark = newIsDark
    }
  },
  methods: {
    toggleTheme() {
      this.isDark = !this.isDark
    }
  }
})
</script>

<style lang="sass">
@use 'bootstrap-4-grid/scss/grid.scss'

.theme-light
  --color-background: #e4ebf5
  --color-concave: linear-gradient(145deg, #cdd4dd, #f4fbff)
  --color-shadow-start: #c2c8d0
  --color-shadow-end: #ffffff
  --color-font: #001f3f

.theme-dark
  --color-background: #444444
  --color-concave: linear-gradient(145deg, #3d3d3d, #494949)
  --color-shadow-start: #3a3a3a
  --color-shadow-end: #4e4e4e
  --color-font: #fefefe

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

button
  display: inline-block
  border: none
  margin: 0
  outline: none
  text-decoration: none
  font-family: inherit
  font-size: 1rem
  cursor: pointer
  text-align: center
  -webkit-appearance: none
  -moz-appearance: none

input
  -webkit-writing-mode: horizontal-tb !important
  text-rendering: auto
  letter-spacing: normal
  word-spacing: normal
  text-transform: none
  text-indent: 0
  text-shadow: none
  display: inline-block
  text-align: start
  appearance: auto
  -webkit-rtl-ordering: logical
  cursor: text
  margin: 0
  border: 0
  vertical-align: middle
  white-space: normal
  background: none
  line-height: 1
  font-family: inherit
  font-size: inherit
  color: inherit
  &:focus
    outline: 0

.theme-toggle
  user-select: none
  position: absolute
  right: 10px
  top: 10px
  color: inherit
  cursor: pointer

.btn
  font-weight: 500
  border-radius: 4px
  background: inherit
  color: inherit
  box-shadow: 5px 5px 10px var(--color-shadow-start), -5px -5px 10px var(--color-shadow-end)
  // &:hover
  //   background: var(--color-concave)
  &:hover
    box-shadow: 3px 3px 7px var(--color-shadow-start), -3px -3px 7px var(--color-shadow-end)
  &:active
    background: inherit
    box-shadow: inset 2px 2px 5px var(--color-shadow-start), inset -2px -2px 5px var(--color-shadow-end)

.input
  border-radius: 4px
  box-shadow: inset 2px 2px 3px var(--color-shadow-start), inset -2px -2px 3px var(--color-shadow-end)
</style>
