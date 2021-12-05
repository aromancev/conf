<template>
  <div
    class="shade"
    :class="{ active: controlOpened }"
    @click="switchControl"
  ></div>
  <div class="header">
    <div class="end">
      <div v-if="allowedWrite" class="name" @click="switchControl">
        {{ profile.name }}
      </div>
      <div
        v-if="allowedWrite"
        class="avatar"
        v-html="profile.avatar"
        @click="switchControl"
      ></div>
      <router-link
        v-if="!allowedWrite"
        class="px-3 py-2 btn-convex login"
        to="/login"
        >Sign in</router-link
      >
      <div
        v-if="!allowedWrite"
        @click="toggleTheme"
        class="theme-switch material-icons"
      >
        {{ theme === Theme.Dark ? "light_mode" : "dark_mode" }}
      </div>
    </div>

    <div v-if="allowedWrite" class="control" :class="{ opened: controlOpened }">
      <div class="control-item">
        <span class="icon material-icons">person</span>
        Profile
      </div>
      <div class="control-item" @click="toggleTheme">
        <span class="icon material-icons">{{
          theme === Theme.Dark ? "light_mode" : "dark_mode"
        }}</span>
        Theme
      </div>
      <div class="control-divider"></div>
      <div class="control-item">
        <span class="icon material-icons">logout</span>
        Sign out
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { userStore } from "@/api"
import { genAvatar, genName } from "@/platform/gen"
import { Theme } from "@/platform/theme"

interface Profile {
  avatar?: string
  name?: string
}

export default defineComponent({
  name: "Header",
  data() {
    return {
      Theme,
      controlOpened: false,
      theme: Theme.Light,
    }
  },
  emits: ["theme"],
  computed: {
    allowedWrite(): boolean {
      return userStore.getState().allowedWrite
    },
    profile(): Profile {
      return {
        avatar: genAvatar(userStore.getState().id, 35),
        name: genName(userStore.getState().id),
      }
    },
  },
  watch: {
    theme(theme: string) {
      this.$emit("theme", theme)
      localStorage.theme = theme
    },
  },
  mounted() {
    this.theme = localStorage.theme || Theme.Light
  },
  methods: {
    switchControl() {
      this.controlOpened = !this.controlOpened
    },
    toggleTheme() {
      this.theme = this.theme == Theme.Light ? Theme.Dark : Theme.Light
    },
  },
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

$height: 100%

.header
  @include theme.shadow-m

  position: relative
  top: 0

  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-end

  background: var(--color-background)
  width: 100%
  height: $height
  z-index: 100

.avatar
  width: 34px
  height: 34px
  border-radius: 50%
  overflow: hidden
  margin: 10px
  cursor: pointer

.end
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-end

  padding: 20px

  width: 300px
  height: 100%
  z-index: 60
  background: var(--color-background)

.control
  @include theme.shadow-m

  position: absolute
  top: $height
  text-align: left
  right: 20px
  height: 0
  width: 200px
  overflow: hidden
  background: var(--color-background)

  border-radius: 0 0 5px 5px
  z-index: 50
  &.opened
    height: auto

.control-item
  @include theme.clickable

  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start

  width: 100%
  height: 40px
  line-height: 40px
  color: var(--color-font)
  background: var(--color-background)
  &:hover
    color: var(--color-highlight-font)
    background: var(--color-highlight-background)

  .icon
    margin: 10px
    font-size: 20px

.control-divider
  height: 0
  width: 100%
  margin: 5px 0
  border-top: 1px solid var(--color-outline)

.name
  @include theme.clickable

.login
  margin: 10px

.theme-switch
  @include theme.clickable

.shade
  position: fixed
  left: 0
  top: 0
  height: 100vh
  width: 100vw
  background-color: var(--color-background)
  transition: opacity 1s ease-in-out
  opacity: 0.8
  display: none
  z-index: 10
  &.active
    display: block
</style>
