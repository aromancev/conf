<template>
  <PageLoader v-if="loading" />

  <div v-if="!loading && confa" class="content">
    <div class="title">{{ confa.title || confa.handle }}</div>
    <div class="path">
      /
      <router-link class="path-link" :to="route.confa(handle, ConfaTab.Overview)">{{ confa.handle }}</router-link>
    </div>
    <div class="header">
      <router-link
        :to="route.confa(handle, ConfaTab.Overview)"
        class="header-item"
        :class="{ active: tab === 'overview' }"
      >
        <span class="material-icons icon">remove_red_eye</span>
        Overview
      </router-link>
      <router-link
        v-if="confa.ownerId === user.id"
        :to="route.confa(handle, ConfaTab.Edit)"
        class="header-item"
        :class="{ active: tab === 'edit' }"
      >
        <span class="material-icons icon">edit</span>
        Edit
      </router-link>
    </div>
    <div class="header-divider"></div>
    <div class="tab">
      <ConfaOverview v-if="tab === 'overview'" :confa="confa" />
      <ConfaEdit v-if="tab === 'edit'" :confa="confa" @update="update" />
    </div>
  </div>

  <NotFound v-if="!loading && !confa" />

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useRouter } from "vue-router"
import { confaClient, Confa, userStore, errorCode, Code } from "@/api"
import { route, ConfaTab, handleNew } from "@/router"
import InternalError from "@/components/modals/InternalError.vue"
import PageLoader from "@/components/PageLoader.vue"
import NotFound from "@/views/NotFound.vue"
import ConfaEdit from "./ConfaEdit.vue"
import ConfaOverview from "./ConfaOverview.vue"

enum Modal {
  None = "",
  Error = "error",
}

const props = defineProps<{
  tab: ConfaTab
  handle: string
}>()

const router = useRouter()
const user = userStore.getState()

const confa = ref<Confa | null>()
const loading = ref(false)
const modal = ref(Modal.None)

watch(
  () => props.handle,
  async (value) => {
    if (props.tab == ConfaTab.Edit && !user.allowedWrite) {
      router.replace(route.login())
      return
    }

    if (confa.value && props.handle === confa.value.handle) {
      return
    }

    loading.value = true
    try {
      if (value === handleNew) {
        confa.value = await confaClient.create()
        router.replace(route.confa(props.handle, props.tab))
      } else {
        confa.value = await confaClient.fetchOne({
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

function update(value: Confa) {
  confa.value = value
  router.replace(route.confa(props.handle, props.tab))
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
