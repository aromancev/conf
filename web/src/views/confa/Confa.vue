<template>
  <div v-if="loading" class="centered">
    <PageLoader />
  </div>

  <div v-if="!loading && confa" class="confa">
    <div class="title">{{ confa.title || confa.id }}</div>
    <div class="path">
      /
      <router-link
        class="path-link"
        :to="{
          name: 'confaOverview',
          params: { confa: confa.handle },
        }"
        >{{ confa.handle }}</router-link
      >
    </div>
    <div class="header">
      <router-link
        :to="{
          name: 'confaOverview',
          params: { confa: confa.handle },
        }"
        class="header-item"
        :class="{ active: tab === 'overview' }"
      >
        <span class="material-icons icon">remove_red_eye</span>
        Overview
      </router-link>
      <router-link
        :to="{
          name: 'confaEdit',
          params: { confa: confa.handle },
        }"
        class="header-item"
        :class="{ active: tab === 'edit' }"
      >
        <span class="material-icons icon">edit</span>
        Edit
      </router-link>
    </div>
    <div class="header-divider"></div>
    <div class="content">
      <div class="body">
        <ConfaPreview v-if="tab === 'overview'" :confa="confa" />
        <ConfaEdit v-if="tab === 'edit'" :confa="confa" @update="update" />
      </div>
    </div>
  </div>

  <NotFound v-if="!loading && !confa" />

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useRouter } from "vue-router"
import { confaClient, Confa, ConfaInput } from "@/api"
import InternalError from "@/components/modals/InternalError.vue"
import PageLoader from "@/components/PageLoader.vue"
import NotFound from "@/views/NotFound.vue"
import ConfaEdit from "./ConfaEdit.vue"
import ConfaPreview from "./ConfaOverview.vue"

enum Modal {
  None = "",
  Error = "error",
}

const props = defineProps<{
  tab: string
  handle: string
}>()

const router = useRouter()

const confa = ref<Confa | null>()
const loading = ref(false)
const modal = ref(Modal.None)

watch(
  () => props.handle,
  async (value) => {
    if (confa.value && props.handle === confa.value.handle) {
      return
    }
    loading.value = true
    try {
      if (value === "new") {
        confa.value = await confaClient.create()
        router.replace({ name: "confaOverview", params: { handle: confa.value.handle } })
      } else {
        confa.value = await confaClient.fetchOne({
          handle: value,
        })
        if (confa.value === null) {
          return
        }
      }
    } catch (e) {
      modal.value = Modal.Error
    } finally {
      loading.value = false
    }
  },
  { immediate: true },
)

function update(value: ConfaInput) {
  const current = Object.assign({}, confa.value)
  confa.value = Object.assign(current, value)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.centered
  height: 100%
  width: 100%
  display: flex
  flex-direction: column
  justify-content: center
  align-items: center

.title
  cursor: default
  font-size: 1.5em
  margin-top: 40px
  width: 100%
  max-width: theme.$content-width
  text-align: left
  padding: 0 20px

.path
  width: 100%
  text-align: left
  max-width: theme.$content-width
  padding: 0 20px
  margin-bottom: 10px

.path-link
  text-decoration: none
  color: var(--color-font-disabled)
  &:hover
    color: var(--color-font)
    text-decoration: underline

.confa
  width: 100%
  display: flex
  flex-direction: column
  justify-content: flex-start
  align-items: center

.content
  width: 100%
  max-width: theme.$content-width
  text-align: left

.body
  width: 100%
  display: flex
  flex-direction: row
  justify-content: flex-start

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
  height: 1px
  background: var(--color-outline)
</style>
