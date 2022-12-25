<template>
  <div class="content">
    <div class="title">Content Hub</div>
    <div ref="list" class="confas-list" @scroll="onScroll">
      <div v-if="state.isLoading" class="confas-loader">
        <PageLoader />
      </div>
      <div v-if="!state.isLoading" class="confas-items">
        <div v-for="confa in state.confas" :key="confa.id" class="confas-item">
          /
          <router-link class="confas-link" :to="route.confa(confa.handle, 'overview')">{{ confa.handle }}</router-link>
          <span v-if="confa.title" class="confas-title">{{ confa.title }}</span>
        </div>
      </div>
    </div>
  </div>
  <InternalError :is-visible="state.modal === 'error'" @click="state.modal = 'none'" />
</template>

<script setup lang="ts">
import { reactive, onMounted, ref } from "vue"
import { currentUser, Confa } from "@/api/models"
import { confaClient } from "@/api"
import { ConfaIterator } from "@/api/confa"
import { route } from "@/router"
import InternalError from "@/components/modals/InternalError.vue"
import PageLoader from "@/components/PageLoader.vue"

type Modal = "none" | "error"

interface State {
  isLoading: boolean
  isFetchedAll: boolean
  modal: Modal
  confas: Confa[]
}

const state = reactive<State>({
  isLoading: true,
  isFetchedAll: false,
  modal: "none",
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
      iterator = confaClient.fetch({ ownerId: currentUser.id })
    }
    const fetched = await iterator.next()
    if (!fetched.length) {
      state.isFetchedAll = true
    } else {
      state.confas = state.confas.concat(fetched)
    }
  } catch (e) {
    state.modal = "error"
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
  text-align: left
  max-height: 100%
  display: flex
  flex-direction: column
  padding: 50px

.title
  font-size: 40px
  margin-bottom: 30px

.confas-list
  @include theme.shadow-inset-xs

  min-height: 300px
  width: 400px
  max-width: 100%
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
  color: var(--color-font-disabled)
  text-decoration: none
  width: 100px
  display: inline-block
  &:hover
    color: var(--color-font)
    text-decoration: underline

.confas-title
  margin-left: 30px
</style>
