<template>
  <PageLoader v-if="loading" />

  <div v-if="!loading && talk" class="content">
    <div class="title">{{ talk.title || talk.handle }}</div>
    <div class="path">
      /
      <router-link class="path-link" :to="route.confa(confaHandle, 'overview')">
        {{ confaHandle }}
      </router-link>
      /
      <router-link class="path-link" :to="route.talk(confaHandle, handle, 'watch')">
        {{ talk.handle }}
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
        v-if="talk.ownerId === user.id"
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
      <TalkOverview v-if="tab === 'overview'" :talk="talk" />
      <TalkLive
        v-if="tab === 'watch' && talk.state !== TalkState.ENDED"
        :talk="talk"
        :confa-handle="confaHandle"
        :join-confirmed="joinConfirmed"
        :invite-link="inviteLink"
        @join="join"
        @update="update"
      />
      <TalkReplay v-if="tab === 'watch' && talk.state === TalkState.ENDED" :talk="talk" />
      <TalkEdit v-if="tab === 'edit'" :talk="talk" @update="update" />
    </div>
  </div>

  <NotFound v-if="!loading && !talk" />

  <InternalError v-if="modal === 'error'" @click="modal = 'none'" />
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue"
import { useRouter } from "vue-router"
import { talkClient, Talk, userStore, confaClient, errorCode, Code, TalkState } from "@/api"
import { route, TalkTab, handleNew } from "@/router"
import InternalError from "@/components/modals/InternalError.vue"
import PageLoader from "@/components/PageLoader.vue"
import NotFound from "@/views/NotFound.vue"
import TalkEdit from "./TalkEdit.vue"
import TalkOverview from "./TalkOverview.vue"
import TalkLive from "./TalkLive.vue"
import TalkReplay from "./TalkReplay.vue"

type Modal = "none" | "error"

const props = defineProps<{
  tab: TalkTab
  confaHandle: string
  handle: string
}>()

const router = useRouter()
const user = userStore.state()

const talk = ref<Talk | null>()
const loading = ref(false)
const modal = ref<Modal>("none")
const joinConfirmed = ref(false)

const inviteLink = computed(() => {
  return window.location.host + router.resolve(route.talk(props.confaHandle, props.handle, "watch")).fullPath
})

watch(
  () => props.handle,
  async (value) => {
    if (!user.allowedWrite && (props.tab == "edit" || props.confaHandle === handleNew)) {
      router.replace(route.login())
      return
    }

    if (talk.value && props.handle === talk.value.handle) {
      return
    }

    loading.value = true
    let confaHandle = props.confaHandle
    let talkHandle = value
    try {
      if (props.confaHandle === handleNew) {
        const confa = await confaClient.create()
        confaHandle = confa.handle
      }
      if (talkHandle === handleNew) {
        if (!user.allowedWrite) {
          router.replace(route.login())
          return
        }
        talk.value = await talkClient.create({ handle: confaHandle }, {})
        talkHandle = talk.value.handle
      } else {
        talk.value = await talkClient.fetchOne(
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
          modal.value = "error"
          break
      }
    } finally {
      loading.value = false
    }
  },
  { immediate: true },
)

function update(value: Talk) {
  talk.value = value
  router.replace(route.talk(props.confaHandle, value.handle, props.tab))
}

function join(confirmed: boolean) {
  joinConfirmed.value = confirmed
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
  display: flex
  flex-direction: column
  align-items: center
  flex: 1
</style>
