<template>
  <div class="layout">
    <div class="talks">
      <div class="talks-header">
        <div>Talks</div>
        <router-link
          v-if="userId === confa.ownerId"
          class="btn create-talk"
          :to="{
            name: 'talk',
            params: { confa: confa.handle, talk: 'new' },
          }"
          ><span class="material-icons">add</span> New</router-link
        >
      </div>
      <div class="talks-list">
        <div v-if="talksLoading" class="talks-loader"><Loader /></div>
        <div v-if="!talksLoading" class="talks-items">
          <div v-for="talk in talks" :key="talk.id" class="talks-item">
            /
            <router-link
              class="talks-link"
              :to="{
                name: 'talk',
                params: { confa: confa.handle, talk: talk.handle },
              }"
            >
              {{ talk.handle }}
            </router-link>
          </div>
        </div>
      </div>
    </div>
    <div class="description">{{ confa.description }}</div>
  </div>

  <InternalError
    v-if="modal === Dialog.Error"
    v-on:click="modal = Dialog.None"
  />
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import Loader from "@/components/Loader.vue"
import { defineComponent, PropType } from "vue"
import { userStore, Confa, Talk } from "@/api/models"
import { confa, talk, Policy } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "ConfaPreview",
  components: {
    InternalError,
    Loader,
  },
  props: {
    confa: {
      type: Object as PropType<Confa>,
      required: true,
    },
  },
  data() {
    return {
      Dialog,
      talksLoading: true,
      talks: [] as Talk[],
      modal: Dialog.None,
    }
  },
  computed: {
    handle(): string {
      return this.$route.params.confa as string
    },
    userId(): string {
      return userStore.getState().id
    },
  },
  mounted() {
    this.loadTalks()
  },
  methods: {
    async loadTalks() {
      if (!this.confa) {
        return
      }
      this.talksLoading = true
      const iter = talk.fetch({ confaId: this.confa.id }, false)
      this.talks = await iter.next()
      this.talksLoading = false
    },
  },
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

.layout
  width: 100%
  display: flex
  flex-direction: row
  justify-content: center
  padding: 20px

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
  overflow: hidden
  text-overflow: ellipsis
  white-space: nowrap

.talks-link
  color: var(--color-font-disabled)
  text-decoration: none
  &:hover
    color: var(--color-font)
    text-decoration: underline

.edit
  margin: 20px

.description
  flex-grow: 1
  margin: 10px
  padding: 20px
  white-space: pre-wrap
</style>
