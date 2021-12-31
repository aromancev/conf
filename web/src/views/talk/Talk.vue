<template>
  <PageLoader v-if="loading" />

  <div v-if="!loading && talk" class="content">
    <div class="title">{{ talk.title || talk.id }}</div>
    <div class="path">
      /
      <router-link
        class="path-link"
        :to="{
          name: 'confaOverview',
          params: { confa: confaHandle },
        }"
      >
        {{ confaHandle }}
      </router-link>
      /
      <router-link
        class="path-link"
        :to="{
          name: 'talkOverview',
          params: { talk: talk.handle },
        }"
      >
        {{ talk.handle }}
      </router-link>
    </div>
    <div class="header">
      <router-link
        :to="{
          name: 'talkOverview',
          params: { talk: talk.handle },
        }"
        class="header-item"
        :class="{ active: tab === 'overview' }"
      >
        <span class="material-icons icon">remove_red_eye</span>
        Overview
      </router-link>
      <router-link
        :to="{
          name: 'talkOnline',
          params: { talk: talk.handle },
        }"
        class="header-item"
        :class="{ active: tab === 'online' }"
      >
        <span class="material-icons icon">podcasts</span>
        Online
      </router-link>
      <router-link
        v-if="talk.ownerId === user.id"
        :to="{
          name: 'talkEdit',
          params: { talk: talk.handle },
        }"
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
      <TalkOnline v-if="tab === 'online'" :talk="talk" :join-confirmed="joinConfirmed" @join="join" />
      <TalkEdit v-if="tab === 'edit'" :talk="talk" @update="update" />
    </div>
  </div>

  <NotFound v-if="!loading && !talk" />

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useRouter } from "vue-router"
import { talkClient, Talk, userStore, errorCode, Code } from "@/api"
import InternalError from "@/components/modals/InternalError.vue"
import PageLoader from "@/components/PageLoader.vue"
import NotFound from "@/views/NotFound.vue"
import TalkEdit from "./TalkEdit.vue"
import TalkOverview from "./TalkOverview.vue"
import TalkOnline from "./TalkOnline.vue"

enum Modal {
  None = "",
  Error = "error",
}

const props = defineProps<{
  tab: string
  confaHandle: string
  handle: string
}>()

const router = useRouter()
const user = userStore.getState()

const talk = ref<Talk | null>()
const loading = ref(false)
const modal = ref(Modal.None)
const joinConfirmed = ref(false)

watch(
  () => props.handle,
  async (value) => {
    if (talk.value && props.handle === talk.value.handle) {
      return
    }
    loading.value = true
    try {
      if (value === "new") {
        talk.value = await talkClient.create({ handle: props.confaHandle }, {})
        router.replace({ name: "talkOverview", params: { confa: props.confaHandle, talk: talk.value.handle } })
      } else {
        talk.value = await talkClient.fetchOne({
          handle: value,
        })
      }
    } catch (e) {
      switch (errorCode(e)) {
        case Code.NotFound:
          break
        default:
          modal.value = Modal.Error
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
}

function join(confirmed: boolean) {
  joinConfirmed.value = confirmed
  if (!confirmed) {
    router.push({ name: "talkOverview", params: { confa: props.confaHandle, talk: props.handle } })
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
