<template>
  <div class="shade" :class="{ active: modal !== 'none' }" @click="switchModal('none')"></div>
  <div class="header">
    <div class="start">
      <div class="menu material-icons" @click="switchModal('sidebar')">menu</div>
      <router-link :to="route.home()">
        <ConfaLogo></ConfaLogo>
      </router-link>
    </div>
    <router-link class="disclaimer" :to="route.disclaimer()">
      <span class="material-icons disclaimer-icon">warning</span>
      NOT A COMMERCIAL PRODUCT
    </router-link>
    <div class="end">
      <ProfileAvatar
        v-if="accessStore.state.allowedWrite"
        class="avatar"
        :size="128"
        :user-id="accessStore.state.id"
        :src="profileStore.state.avatarThumbnail"
        @click="switchModal('profile')"
      ></ProfileAvatar>
      <router-link v-if="!accessStore.state.allowedWrite" class="btn-convex login" :to="route.login()"
        >Sign in</router-link
      >
    </div>

    <div v-if="modal === 'sidebar'" class="sidebar">
      <router-link
        v-if="accessStore.state.allowedWrite"
        class="control-item"
        :to="route.contentHub()"
        @click="switchModal('none')"
      >
        <span class="icon material-icons">hub</span>
        Content hub
      </router-link>
      <router-link v-if="accessStore.state.allowedWrite" class="control-item" to="/new" @click="switchModal('none')">
        <span class="icon material-icons">add</span>
        Create conference
      </router-link>
      <div class="control-divider"></div>
      <div class="control-item" @click="toggleTheme">
        <span class="icon material-icons">{{ styleStore.state.theme === "dark" ? "light_mode" : "dark_mode" }}</span>
        {{ styleStore.state.theme === "dark" ? "Light mode" : "Dark mode" }}
      </div>
      <CopyText class="sidebar-end info" :value="'info@confa.io'"></CopyText>
    </div>

    <div
      v-if="accessStore.state.allowedWrite"
      class="control"
      :class="{ opened: modal === 'profile' }"
      @click="modal = 'none'"
    >
      <router-link class="control-item" :to="route.profile(profileStore.state.handle || handleNew, 'overview')">
        <span class="icon material-icons">person</span>
        My profile
      </router-link>
      <div class="control-divider"></div>
      <div class="control-item" @click="logout">
        <span class="icon material-icons">logout</span>
        Sign out
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { api } from "@/api"
import { accessStore } from "@/api/models/access"
import { profileStore } from "@/api/models/profile"
import { styleStore } from "@/api/models/style"
import router, { route, handleNew } from "@/router"
import ConfaLogo from "@/components/ConfaLogo.vue"
import CopyText from "@/components/fields/CopyText.vue"
import ProfileAvatar from "@/components/profile/ProfileAvatar.vue"

type Modal = "none" | "sidebar" | "profile"

const modal = ref<Modal>("none")

async function logout() {
  await api.logout()
  router.push(route.login())
}
function switchModal(val: Modal) {
  if (modal.value === val) {
    modal.value = "none"
    return
  }
  modal.value = val
}
function toggleTheme() {
  styleStore.setTheme(styleStore.state.theme === "light" ? "dark" : "light")
}
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
  justify-content: flex-start
  background: var(--color-background)
  width: 100%
  height: $height
  z-index: 100

.menu
  @include theme.clickable

.avatar
  @include theme.clickable

  width: 34px
  height: 34px
  border-radius: 50%
  overflow: hidden
  margin: 10px

.start
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start
  z-index: 60
  padding: 0 30px
  height: 100%
  background: var(--color-background)

.logo
  font-size: 1.5rem
  margin: 20px
  cursor: pointer

.disclaimer
  font-size: 14px
  color: var(--color-font)
  display: flex
  align-items: center
  padding: 5px
  border: 2px solid #fb8c00
  border-radius: 4px

.disclaimer-icon
  margin-right: 5px

.end
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-end
  margin-left: auto
  padding: 0 30px
  width: 300px
  height: 100%
  z-index: 60
  background: var(--color-background)

.sidebar
  @include theme.shadow-m

  position: absolute
  top: $height
  left: 0
  text-align: left
  height: calc(100vh - $height)
  width: 200px
  overflow: hidden
  background: var(--color-background)

  border-radius: 0 0 4px 4px
  z-index: 50

.sidebar-end
  position: absolute
  bottom: 0
  width: 100%

.info
  text-align: center
  padding: 10px

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

  border-radius: 0 0 4px 4px
  z-index: 50
  &.opened
    height: auto

.control-item
  @include theme.clickable

  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start

  text-decoration: none
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
  opacity: 0.5
  display: none
  z-index: 10
  &.active
    display: block
</style>
