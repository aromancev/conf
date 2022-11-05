<template>
  <div class="layout">
    <div class="talks">
      <div class="talks-header">
        <div>Talks</div>
        <router-link
          v-if="currentUser.id === confa.ownerId"
          class="btn create-talk"
          :to="route.talk(confa.handle, handleNew, 'watch')"
        >
          <span class="material-icons">add</span> New
        </router-link>
      </div>
      <div class="talks-list">
        <div v-if="talksLoading" class="talks-loader">
          <PageLoader />
        </div>
        <div v-if="!talksLoading" class="talks-items">
          <div v-for="talk in talks" :key="talk.id" class="talks-item">
            /
            <router-link class="talks-link" :to="route.talk(confa.handle, talk.handle, 'watch')">{{
              talk.handle
            }}</router-link>
          </div>
        </div>
      </div>
    </div>
    <div class="description">{{ confa.description }}</div>
  </div>

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { currentUser, Confa, Talk } from "@/api/models"
import { talkClient } from "@/api"
import { route, handleNew } from "@/router"
import InternalError from "@/components/modals/InternalError.vue"
import PageLoader from "@/components/PageLoader.vue"

enum Modal {
  None = "",
  Error = "error",
}

const props = defineProps<{
  confa: Confa
}>()

const talks = ref<Talk[]>([])
const talksLoading = ref(false)
const modal = ref(Modal.None)

onMounted(() => {
  loadTalks()
})

async function loadTalks() {
  if (!props.confa) {
    return
  }
  talksLoading.value = true
  const iter = talkClient.fetch(
    { confaId: props.confa.id },
    {
      hydrated: false,
    },
  )
  talks.value = await iter.next()
  talksLoading.value = false
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.layout
  width: 100%
  display: flex
  flex-direction: row
  justify-content: center
  padding: 30px

.talks
  width: 300px
  min-width: 300px
  margin: 10px
  border-radius: 4px
  text-align: center

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
  font-size: 13px

.talks-list
  @include theme.shadow-inset-xs

  position: relative
  min-height: 300px
  width: 100%

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

.talks-item
  @include theme.clickable

  display: inline-block
  width: 100%
  padding: 5px 0
  text-align: left
  overflow: hidden
  text-overflow: ellipsis
  white-space: nowrap

.talks-link
  color: var(--color-font-disabled)
  text-decoration: none
  &:hover
    color: var(--color-font)
    text-decoration: underline

.description
  flex-grow: 1
  margin: 10px
  padding: 20px
  white-space: pre-wrap
</style>
