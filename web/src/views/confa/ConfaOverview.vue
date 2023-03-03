<template>
  <div class="layout">
    <div class="talks">
      <div class="talks-header">
        <div>Talks</div>
        <router-link
          v-if="accessStore.state.id === confa.ownerId"
          class="btn create-talk"
          :to="route.talk(confa.handle, handleNew, 'watch')"
        >
          <span class="material-icons">add</span> New
        </router-link>
      </div>
      <div ref="talksList" class="talks-list" @scroll="onTalksScroll">
        <div v-if="state.isTalksLoading" class="talks-loader">
          <PageLoader />
        </div>
        <div v-if="!state.isTalksLoading" class="talks-items">
          <div v-for="talk in state.talks" :key="talk.id" class="talks-item">
            <router-link
              class="talks-link"
              :class="{ untitled: talk.title ? false : true }"
              :to="route.talk(confa.handle, talk.handle, 'watch')"
            >
              {{ talk.title || "Untitled" }}</router-link
            >
          </div>
        </div>
      </div>
    </div>
    <div class="description" :class="{ empty: !confa.description }">{{ confa.description || "No description" }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from "vue"
import { Confa } from "@/api/models/confa"
import { accessStore } from "@/api/models/access"
import { Talk } from "@/api/models/talk"
import { talkClient } from "@/api"
import { TalkIterator } from "@/api/talk"
import { route, handleNew } from "@/router"
import PageLoader from "@/components/PageLoader.vue"
import { notificationStore } from "@/api/models/notifications"

const props = defineProps<{
  confa: Confa
}>()

interface State {
  isTalksLoading: boolean
  isTalksFetchedAll: boolean
  talks: Talk[]
}

const state = reactive<State>({
  isTalksLoading: true,
  isTalksFetchedAll: false,
  talks: [],
})

const talksList = ref<HTMLElement>()

let talkIterator: TalkIterator | undefined

onMounted(() => {
  loadTalks()
})

async function onTalksScroll() {
  if (!talksList.value) {
    return
  }
  const scroll = talksList.value.scrollTop / (talksList.value.scrollHeight - talksList.value.clientHeight)
  if (scroll < 0.7) {
    return
  }
  loadTalks()
}

async function loadTalks() {
  if (!props.confa) {
    return
  }

  try {
    if (!talkIterator) {
      talkIterator = talkClient.fetch(
        { confaId: props.confa.id },
        {
          hydrated: false,
        },
      )
    }

    const fetched = await talkIterator.next()
    if (!fetched.length) {
      state.isTalksFetchedAll = true
    } else {
      state.talks = state.talks.concat(fetched)
    }
  } catch (e) {
    notificationStore.error("failed to load talks")
  } finally {
    state.isTalksLoading = false
  }
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.layout
  width: 100%
  height: 100%
  display: flex
  flex-direction: row
  justify-content: center
  padding: 30px

.talks
  width: 300px
  margin: 10px
  border-radius: 4px
  text-align: center
  display: flex
  flex-direction: column

.talks-header
  width: 100%
  display: flex
  flex-direction: row
  justify-content: flex-start
  align-items: flex-end
  margin: 20px 0

.create-talk
  display: flex
  flex-direction: row
  align-items: center
  margin-left: auto
  font-size: 14px
  padding: 3px 10px

.talks-list
  @include theme.shadow-inset-xs

  width: 100%
  height: 100%
  min-height: 300px
  position: relative
  overflow-y: scroll

.talks-loader
  position: absolute
  width: 100%
  height: 100%
  display: flex
  flex-direction: column
  justify-content: center
  align-items: center

.talks-items
  padding: 15px
  position: absolute

.talks-item
  @include theme.clickable

  width: 100%
  padding: 5px 0
  text-align: left
  overflow: hidden
  text-overflow: ellipsis
  white-space: nowrap

.talks-link
  color: var(--color-font)
  text-decoration: none
  &:hover
    text-decoration: underline
  &.untitled
    color: var(--color-font-disabled)

.description
  flex-grow: 1
  margin: 10px
  padding: 20px
  white-space: pre-wrap
  outline: 1px solid var(--color-outline)
  border-radius: 4px
  &.empty
    color: var(--color-font-disabled)
</style>
