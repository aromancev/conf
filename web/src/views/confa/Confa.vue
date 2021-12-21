<template>
  <div v-if="loading" class="centered">
    <Loader />
  </div>

  <div v-if="!loading && confa" class="confa">
    <div class="title">{{ confa.title || confa.id }}</div>
    <div class="path">
      /<router-link
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
        <ConfaEdit v-if="tab === 'edit'" :confa="confa" @updated="updated" />
      </div>
    </div>
  </div>

  <NotFound v-if="!loading && !confa" />

  <InternalError
    v-if="modal === Dialog.Error"
    v-on:click="modal = Dialog.None"
  />
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import Loader from "@/components/Loader.vue"
import NotFound from "@/views/NotFound.vue"
import { defineComponent } from "vue"
import { confa, Confa, ConfaInput } from "@/api"
import ConfaEdit from "./ConfaEdit.vue"
import ConfaPreview from "./ConfaOverview.vue"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "Confa",
  components: {
    ConfaEdit,
    ConfaPreview,
    InternalError,
    NotFound,
    Loader,
  },
  props: {
    tab: {
      type: String,
      required: true,
    },
    handle: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      Dialog,
      confa: null as Confa | null,
      loading: false,
      modal: Dialog.None,
    }
  },
  watch: {
    handle: {
      immediate: true,
      async handler(handle: string) {
        if (this.confa && handle === this.confa.handle) {
          return
        }
        this.loading = true
        try {
          if (handle === "new") {
            this.confa = await confa.create()
            this.$router.replace("/" + this.confa.handle)
          } else {
            this.confa = await confa.fetchOne({
              handle: handle,
            })
            if (this.confa === null) {
              return
            }
          }
        } catch (e) {
          this.modal = Dialog.Error
        } finally {
          this.loading = false
        }
      },
    },
  },
  methods: {
    updated(confa: ConfaInput) {
      const current = Object.assign({}, this.confa)
      this.confa = Object.assign(current, confa)
    },
  },
})
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
