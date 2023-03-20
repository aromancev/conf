<template>
  <PageLoader v-if="state.isLoading" />

  <div v-if="!state.isLoading && state.confa" class="content">
    <div class="title">
      <EditableField
        v-if="accessStore.state.id === state.confa.ownerId"
        type="text"
        :value="state.confa.title || 'Untitled'"
        :validate="(v) => titleValidator.validate(v)"
        @update="updateTitle"
        @discard="discardTitle"
      ></EditableField>
      <div v-else>{{ state.confa.title || "Untitled" }}</div>
    </div>
    <div class="path">
      /
      <router-link class="path-link" :to="route.confa(handle, 'overview')">{{ state.confa.handle }}</router-link>
    </div>
    <div class="header">
      <router-link :to="route.confa(handle, 'overview')" class="header-item" :class="{ active: tab === 'overview' }">
        <span class="material-icons icon">remove_red_eye</span>
        Overview
      </router-link>
      <router-link
        v-if="state.confa.ownerId === accessStore.state.id"
        :to="route.confa(handle, 'edit')"
        class="header-item"
        :class="{ active: tab === 'edit' }"
      >
        <span class="material-icons icon">edit</span>
        Edit
      </router-link>
    </div>
    <div class="header-divider"></div>
    <div class="tab">
      <ConfaOverview v-if="tab === 'overview'" :confa="state.confa" />
      <ConfaEdit v-if="tab === 'edit'" :confa="state.confa" @update="update" />
    </div>
  </div>

  <NotFound v-if="!state.isLoading && !state.confa" />
</template>

<script setup lang="ts">
import { watch, reactive } from "vue"
import { useRouter } from "vue-router"
import { confaClient, errorCode, Code } from "@/api"
import { accessStore } from "@/api/models/access"
import { Confa } from "@/api/models/confa"
import { titleValidator } from "@/api/models/confa"
import { route, ConfaTab, handleNew } from "@/router"
import PageLoader from "@/components/PageLoader.vue"
import EditableField from "@/components/fields/EditableField.vue"
import NotFound from "@/views/NotFound.vue"
import ConfaEdit from "./ConfaEdit.vue"
import ConfaOverview from "./ConfaOverview.vue"
import { notificationStore } from "@/api/models/notifications"

const props = defineProps<{
  tab: ConfaTab
  handle: string
}>()

type State = {
  confa?: Confa
  isLoading: boolean
}

const state = reactive<State>({
  isLoading: false,
})

const router = useRouter()

watch(
  () => props.handle,
  async (handle) => {
    if (!accessStore.state.allowedWrite && (props.tab == "edit" || handle === handleNew)) {
      router.replace(route.login())
      return
    }

    if (state.confa && props.handle === state.confa.handle) {
      return
    }

    state.isLoading = true
    try {
      if (handle === handleNew) {
        state.confa = await confaClient.create()
        router.replace(route.confa(state.confa.handle, props.tab))
      } else {
        state.confa = await confaClient.fetchOne({
          handle: handle,
        })
      }
    } catch (e) {
      switch (errorCode(e)) {
        case Code.NotFound:
          break
        default:
          notificationStore.error("failed to load conference")
          break
      }
    } finally {
      state.isLoading = false
    }
  },
  { immediate: true },
)

async function updateTitle(title: string) {
  if (!state.confa || title === state.confa.title) {
    return
  }

  try {
    state.confa = await confaClient.update({ id: state.confa.id }, { title: title })
    notificationStore.info("title updated")
  } catch {
    notificationStore.error("failed to update title")
  }
}

function discardTitle() {
  notificationStore.info("title discarded")
}

function update(value: Confa) {
  const oldHandle = state.confa?.handle
  state.confa = value
  if (oldHandle !== value.handle) {
    router.replace(route.confa(value.handle, props.tab))
  }
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.content
  width: 100%
  min-height: 100%
  display: flex
  flex-direction: column
  justify-content: flex-start
  align-items: center

.title
  cursor: default
  font-size: 1.5em
  margin-top: 40px
  width: 100%
  max-width: theme.$content-width
  text-align: left
  padding: 0 30px

.path
  width: 100%
  text-align: left
  max-width: theme.$content-width
  padding: 0 30px
  margin-bottom: 10px
  font-size: 12px

.path-link
  text-decoration: none
  color: var(--color-font-disabled)
  &:hover
    color: var(--color-font)
    text-decoration: underline

.header
  width: 100%
  max-width: theme.$content-width
  display: flex
  flex-direction: row
  margin-bottom: -1px
  padding: 0 20px

.header-item
  @include theme.clickable

  display: flex
  flex-direction: row
  align-items: center
  justify-content: center
  text-decoration: none
  color: var(--color-font)
  padding: 10px
  width: 150px
  border-bottom: 1px solid transparent
  transition: border 0.3s
  &.active
    border-bottom: 1px solid var(--color-highlight-background)
  &:hover:not(.active)
    border-bottom: 1px solid var(--color-font)

  .icon
    margin-right: 5px
    font-size: 15px

.header-divider
  width: 100%
  height: 0
  border-bottom: 1px solid var(--color-outline)

.tab
  width: 100%
  max-width: theme.$content-width
  flex: 1
</style>
