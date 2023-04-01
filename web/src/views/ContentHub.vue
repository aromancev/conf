<template>
  <div class="content">
    <div class="confas">
      <div class="confas-header">
        <div class="title">My conferences</div>
        <router-link class="btn create-confa" :to="route.confa(handleNew, 'overview')">
          <span class="material-icons">add</span> New
        </router-link>
      </div>
      <div ref="list" class="confas-list" @scroll="onScroll">
        <div v-if="state.isLoading" class="confas-loader">
          <PageLoader />
        </div>
        <div v-if="!state.isLoading" class="confas-items">
          <div v-for="confa in state.confas" :key="confa.id" class="confas-item">
            <router-link
              class="confas-link"
              :class="{ untitled: confa.title ? false : true }"
              :to="route.confa(confa.handle, 'overview')"
              >{{ confa.title || "Untitled" }}</router-link
            >
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, onMounted, ref } from "vue"
import { Confa } from "@/api/models/confa"
import { accessStore } from "@/api/models/access"
import { api } from "@/api"
import { ConfaClient } from "@/api/confa"
import { ConfaIterator } from "@/api/confa"
import { handleNew, route } from "@/router"
import PageLoader from "@/components/PageLoader.vue"
import { notificationStore } from "@/api/models/notifications"

interface State {
  isLoading: boolean
  isFetchedAll: boolean
  confas: Confa[]
}

const state = reactive<State>({
  isLoading: true,
  isFetchedAll: false,
  confas: [],
})
const list = ref<HTMLElement>()

let iterator: ConfaIterator | undefined

onMounted(() => {
  loadConfas()
})

async function onScroll() {
  if (!list.value) {
    return
  }
  const scroll = list.value.scrollTop / (list.value.scrollHeight - list.value.clientHeight)
  if (scroll < 0.7) {
    return
  }
  loadConfas()
}

async function loadConfas() {
  if (state.isFetchedAll) {
    return
  }

  try {
    if (!iterator) {
      iterator = new ConfaClient(api).fetch({ ownerId: accessStore.state.id })
    }
    const fetched = await iterator.next()
    if (!fetched.length) {
      state.isFetchedAll = true
    } else {
      state.confas = state.confas.concat(fetched)
    }
  } catch (e) {
    notificationStore.error("failed to load conferences")
  } finally {
    state.isLoading = false
  }
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.content
  width: theme.$content-width
  max-width: 100%
  max-height: 100%
  padding: 50px
  height: 100%

.confas
  width: 400px
  height: 100%
  display: flex
  flex-direction: column

.confas-header
  width: 100%
  display: flex
  flex-direction: row
  justify-content: flex-start
  align-items: flex-end
  margin: 20px 0

.title
  font-size: 20px

.create-confa
  display: flex
  flex-direction: row
  align-items: center
  margin-left: auto
  font-size: 14px
  padding: 3px 10px

.confas-list
  @include theme.shadow-inset-xs

  min-height: 300px
  width: 100%
  overflow-y: scroll

.confas-loader
  position: absolute
  width: 100%
  height: 100%
  display: flex
  flex-direction: column
  justify-content: center
  align-items: center

.confas-items
  padding: 15px

.confas-item
  @include theme.clickable

  display: inline-block
  width: 100%
  padding: 5px 0
  text-align: left
  overflow: hidden
  text-overflow: ellipsis
  white-space: nowrap

.confas-link
  color: var(--color-font)
  text-decoration: none
  display: inline-block
  &:hover
    color: var(--color-font)
    text-decoration: underline
  &.untitled
    color: var(--color-font-disabled)

.confas-title
  margin-left: 30px
</style>
