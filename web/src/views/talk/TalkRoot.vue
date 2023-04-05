<template>
  <PageLoader v-if="state.isLoading" />

  <div v-if="!state.isLoading && state.talk" class="content">
    <div class="title">
      <EditableField
        v-if="accessStore.state.id === state.talk.ownerId"
        type="text"
        :value="state.talk.title || 'Untitled'"
        :validate="(v) => titleValidator.validate(v)"
        @update="updateTitle"
        @discard="discardTitle"
      ></EditableField>
      <div v-else>{{ state.talk.title || "Untitled" }}</div>
    </div>
    <div class="path">
      /
      <router-link class="path-link" :to="route.confa(confaHandle, 'overview')">
        {{ confaHandle }}
      </router-link>
      /
      <router-link class="path-link" :to="route.talk(confaHandle, handle, 'watch')">
        {{ state.talk.handle }}
      </router-link>
    </div>
    <div class="header">
      <router-link
        :to="route.talk(confaHandle, handle, 'watch')"
        class="header-item"
        :class="{ active: tab === 'watch' }"
      >
        <span class="material-icons icon">remove_red_eye</span>
        Watch
      </router-link>
      <router-link
        :to="route.talk(confaHandle, handle, 'overview')"
        class="header-item"
        :class="{ active: tab === 'overview' }"
      >
        <span class="material-icons icon">feed</span>
        Overview
      </router-link>
      <router-link
        v-if="state.talk.ownerId === accessStore.state.id"
        :to="route.talk(confaHandle, handle, 'edit')"
        class="header-item"
        :class="{ active: tab === 'edit' }"
      >
        <span class="material-icons icon">edit</span>
        Edit
      </router-link>
    </div>
    <div class="header-divider"></div>
    <div class="tab">
      <TalkOverview v-if="tab === 'overview'" :talk="state.talk" />
      <TalkLive
        v-if="tab === 'watch' && state.talk.state !== TalkState.ENDED"
        :talk="state.talk"
        :confa-handle="confaHandle"
        :join-confirmed="state.isJoinConfirmed"
        :invite-link="inviteLink"
        @join="join"
        @update="update"
      />
      <TalkReplay v-if="tab === 'watch' && state.talk.state === TalkState.ENDED" :talk="state.talk" />
      <TalkEdit v-if="tab === 'edit'" :talk="state.talk" @update="update" />
    </div>
  </div>

  <NotFound v-if="!state.isLoading && !state.talk" />
</template>

<script setup lang="ts">
import { watch, computed, reactive } from "vue"
import { useRouter } from "vue-router"
import { api, errorCode, Code } from "@/api"
import { ConfaClient } from "@/api/confa"
import { TalkClient } from "@/api/talk"
import { accessStore } from "@/api/models/access"
import { Talk, TalkState, titleValidator } from "@/api/models/talk"
import { route, TalkTab, handleNew } from "@/router"
import PageLoader from "@/components/PageLoader.vue"
import NotFound from "@/views/NotFound.vue"
import EditableField from "@/components/fields/EditableField.vue"
import TalkEdit from "./TalkEdit.vue"
import TalkOverview from "./TalkOverview.vue"
import TalkLive from "./TalkLive.vue"
import TalkReplay from "./TalkReplay.vue"
import { notificationStore } from "@/api/models/notifications"

const props = defineProps<{
  tab: TalkTab
  confaHandle: string
  handle: string
}>()

type State = {
  talk?: Talk
  isLoading: boolean
  isJoinConfirmed: boolean
}

const state = reactive<State>({
  isLoading: false,
  isJoinConfirmed: false,
})

const router = useRouter()

const inviteLink = computed(() => {
  return window.location.host + router.resolve(route.talk(props.confaHandle, props.handle, "watch")).fullPath
})

watch(
  () => props.handle,
  async (value) => {
    if (!accessStore.state.allowedWrite && (props.tab == "edit" || props.confaHandle === handleNew)) {
      router.replace(route.login())
      return
    }

    if (state.talk && props.handle === state.talk.handle) {
      return
    }

    state.isLoading = true
    let confaHandle = props.confaHandle
    let talkHandle = value
    try {
      if (props.confaHandle === handleNew) {
        const confa = await new ConfaClient(api).create()
        confaHandle = confa.handle
      }
      if (talkHandle === handleNew) {
        if (!accessStore.state.allowedWrite) {
          router.replace(route.login())
          return
        }
        state.talk = await new TalkClient(api).create({ handle: confaHandle }, {})
        talkHandle = state.talk.handle
      } else {
        state.talk = await new TalkClient(api).fetchOne(
          {
            handle: talkHandle,
          },
          {
            hydrated: true,
          },
        )
      }
      if (props.confaHandle != confaHandle || props.handle != talkHandle) {
        router.replace(route.talk(confaHandle, talkHandle, props.tab))
      }
    } catch (e) {
      switch (errorCode(e)) {
        case Code.NotFound:
          break
        default:
          notificationStore.error("failed to load talk")
          break
      }
    } finally {
      state.isLoading = false
    }
  },
  { immediate: true },
)

async function updateTitle(title: string) {
  if (!state.talk || title === state.talk.title) {
    return
  }

  try {
    state.talk = await new TalkClient(api).update({ id: state.talk.id }, { title: title })
    notificationStore.info("title updated")
  } catch {
    notificationStore.error("failed to update title")
  }
}

function discardTitle() {
  notificationStore.info("title discarded")
}

function update(value: Talk) {
  const oldHandle = state.talk?.handle
  state.talk = value
  if (oldHandle !== value.handle) {
    router.replace(route.talk(props.confaHandle, value.handle, props.tab))
  }
  state.talk = value
}

function join(confirmed: boolean) {
  state.isJoinConfirmed = confirmed
  if (!confirmed) {
    router.push(route.talk(props.confaHandle, props.handle, "overview"))
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
  font-size: 25px
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
  display: flex
  flex-direction: column
  align-items: center
  flex: 1
</style>
