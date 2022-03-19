<template>
  <div class="messages">
    <PageLoader v-if="loading"></PageLoader>
    <div ref="browser" class="browser" @scroll="onScroll">
      <div v-if="!loading" class="message-list">
        <div
          v-for="msg in renderMessages"
          :key="msg.message.id"
          class="message"
          :class="{ me: msg.message.fromId === userId }"
        >
          <div class="message-body" :class="{ me: msg.message.fromId === userId, last: msg.isLastFrom }">
            <div v-if="msg.avatar" class="avatar">
              <!-- eslint-disable-next-line vue/no-v-html -->
              <div class="icon" v-html="msg.avatar"></div>
            </div>
            <div v-if="msg.fromName" class="from">
              {{ msg.fromName }}
            </div>
            {{ msg.message.text }}
            <div v-if="msg.message.fromId === userId" class="status material-icons">
              {{ msg.message.accepted ? "done" : "schedule" }}
            </div>
          </div>
        </div>
      </div>
    </div>
    <Textarea
      v-model="message"
      class="message-input"
      :spellcheck="false"
      placeholder="message"
      :disabled="loading"
      @keydown="keySend"
    ></Textarea>
    <div v-if="message" class="send material-icons" @click="send">send</div>
  </div>
</template>

<script setup lang="ts">
import { genName, genAvatar } from "@/platform/gen"
import { ref, watch, nextTick, computed } from "vue"
import { Message } from "@/api/room"
import Textarea from "@/components/fields/TextareaField.vue"
import PageLoader from "@/components/PageLoader.vue"

interface RenderMessage {
  message: Message
  fromName?: string
  avatar?: string
  isLastFrom: boolean
}

const emit = defineEmits<{
  (e: "message", value: string): void
}>()

const props = defineProps<{
  userId: string
  messages?: Message[]
  loading?: boolean
}>()

const browser = ref<HTMLElement>()
const message = ref("")

let autoScroll = true

const renderMessages = computed<RenderMessage[]>(() => {
  const messages: RenderMessage[] = []
  let lastMsg: RenderMessage | null = null
  for (const msg of props.messages || []) {
    const rMsg: RenderMessage = {
      message: msg,
      isLastFrom: false,
    }

    if (props.userId != msg.fromId && msg.fromId !== lastMsg?.message.fromId) {
      rMsg.fromName = genName(msg.fromId)
    }
    if (lastMsg && msg.fromId !== lastMsg?.message.fromId) {
      lastMsg.isLastFrom = true
    }
    if (lastMsg && lastMsg.message.fromId !== props.userId && msg.fromId !== lastMsg?.message.fromId) {
      lastMsg.avatar = genAvatar(msg.fromId, 32 + 1)
    }
    lastMsg = rMsg
    messages.push(lastMsg)
  }
  if (lastMsg && lastMsg.message.fromId !== props.userId) {
    lastMsg.avatar = genAvatar(lastMsg.message.fromId, 32 + 1)
  }
  if (lastMsg) {
    lastMsg.isLastFrom = true
  }

  return messages
})

watch(
  () => props.messages,
  () => {
    if (autoScroll) {
      alignScroll()
    }
  },
  { deep: true, immediate: true },
)

function keySend(ev: KeyboardEvent) {
  if (ev.shiftKey || ev.code !== "Enter" || message.value.length === 0) {
    return
  }
  ev.preventDefault()

  send()
}

async function send() {
  emit("message", message.value)
  message.value = ""
  autoScroll = true
  alignScroll()
}

function alignScroll() {
  nextTick(() => {
    if (!browser.value) {
      return
    }
    browser.value.scrollTop = browser.value.scrollHeight
  })
}

function onScroll() {
  if (!browser.value) {
    return
  }
  autoScroll = browser.value.clientHeight + browser.value.scrollTop >= browser.value.scrollHeight - 10
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.messages
  position: relative
  display: flex
  flex-direction: column
  align-items: center
  justify-content: center

.loader
  height: 100%

.browser
  position: relative
  width: 100%
  flex: 1
  overflow-y: auto

.message-list
  top: 0
  left: 0
  width: 100%
  position: absolute

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
  border-radius: 10px
  white-space: pre-line
  margin-left: 32px
  &.last:not(.me)
      border-bottom-left-radius: 0
  &.me
    text-align: right
    background-color: var(--color-highlight-background)
    color: var(--color-highlight-font)
    &.last
      border-bottom-right-radius: 0

.from
  color: var(--color-font-disabled)
  font-size: 0.7em

.field.message-input
  box-shadow: none
  max-height: 50%
  width: 100%
  padding-right: 2.5em
  color: var(--color-font)
  background: var(--color-fade-background)
  border-top-left-radius: 0
  border-top-right-radius: 0

.send
  position: absolute
  bottom: 0
  right: 0
  cursor: pointer
  font-size: 1.6em
  padding: 0.4em

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
