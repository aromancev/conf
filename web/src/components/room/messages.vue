<template>
  <div class="messages">
    <div class="browser" ref="browser" @scroll="onScroll">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="message"
        v-bind:class="{ me: msg.from === userId }"
      >
        <div class="message-body" v-bind:class="{ me: msg.from === userId }">
          <div v-if="msg.from !== userId && msg.isLatestFrom" class="avatar">
            <div class="icon" v-html="msg.avatar"></div>
          </div>
          <div v-if="msg.from !== userId && msg.isFirstFrom" class="from">
            {{ msg.fromName }}
          </div>
          {{ msg.text }}
          <div v-if="msg.from === userId" class="status material-icons">
            {{ msg.isSent ? "done" : "schedule" }}
          </div>
        </div>
      </div>
    </div>
    <Textarea
      class="message-input"
      :spellcheck="false"
      placeholder="message"
      v-on:keydown="keySend"
      v-model="message"
    >
    </Textarea>
    <div v-if="message" class="send material-icons" @click="send">send</div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from "vue"
import { Message } from "./messages"
import Textarea from "@/components/fields/Textarea.vue"

export default defineComponent({
  name: "Messages",
  components: {
    Textarea,
  },
  emits: ["message"],
  props: {
    userId: {
      type: String,
      required: true,
    },
    messages: {
      type: Array as PropType<Array<Message>>,
      required: true,
    },
  },
  data() {
    return {
      message: "",
      byId: {} as { [key: string]: Message },
      ordered: [] as Message[],
      autoScroll: true,
    }
  },
  watch: {
    messages: {
      deep: true,
      handler() {
        if (this.autoScroll) {
          this.alignScroll()
        }
      },
    },
    message() {
      this.alignScroll()
    },
  },
  mounted() {
    this.alignScroll()
  },
  methods: {
    keySend(ev: KeyboardEvent) {
      if (ev.shiftKey || ev.code !== "Enter" || this.message.length === 0) {
        return
      }
      ev.preventDefault()

      this.send()
    },
    async send() {
      this.$emit("message", this.message)
      this.message = ""
      this.autoScroll = true
    },
    alignScroll() {
      // Because of vue3 bug $nextTick doesn't compile. Need to upgrade the version.
      eval(
        "this.$nextTick(()=>{ this.$refs.browser.scrollTop = this.$refs.browser.scrollHeight })",
      )
    },
    onScroll() {
      const el = this.$refs.browser as HTMLElement
      this.autoScroll = el.clientHeight + el.scrollTop >= el.scrollHeight - 10
    },
  },
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

.messages
  position: relative
  display: flex
  flex-direction: column

.browser
  width: 100%
  flex: 1
  overflow-y: auto

.message
  width: 100%
  text-align: left
  padding: 0.3em 1em
  color: var(--color-font)
  &.me
    text-align: right

.message-body
  @include theme.shadow-xs
  position: relative
  display: inline-block
  max-width: 70%
  word-wrap: break-word
  padding: 0.3em 1.1em 0.3em 0.3em
  border-radius: 5px
  white-space: pre-line
  margin-left: 32px
  &.me
    text-align: right
    background-color: var(--color-highlight-background)
    color: var(--color-highlight-font)

.from
  color: var(--color-font-disabled)
  font-size: 0.7em

.message-input
  box-shadow: none
  max-height: 50%
  width: 100%
  padding-right: 2.5em
  color: var(--color-font)
  background: var(--color-outline)
  border-top-left-radius: 0
  border-top-right-radius: 0

.send
  position: absolute
  bottom: 0
  right: 0
  cursor: pointer
  font-size: 1.6em
  padding: 0.2em

.status
  position: absolute
  right: 0
  bottom: 0
  font-size: 0.8em
  padding: 0.5em 0.3em
  cursor: default

.avatar
  position: absolute
  bottom: 0
  left: -32px
  margin-left: -0.5em
  height: 32px
  width: 32px
  border-radius: 32px
  overflow: hidden
  margin-right: 0.5em
</style>
