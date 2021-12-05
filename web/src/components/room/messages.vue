<template>
  <div class="messages">
    <div class="browser" ref="browser" @scroll="onScroll">
      <div
        v-for="msg in ordered"
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
    <textarea
      class="message-input"
      spellcheck="false"
      placeholder="message"
      ref="input"
      v-on:keydown="keySend"
      v-model="message"
      :style="inputStyle"
      :disabled="!emitter"
    >
    </textarea>
    <div v-if="message" class="send material-icons" @click="send">send</div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from "vue"
import { EventType, PayloadMessage } from "@/api/models"
import { Record } from "./record"
import { genName, genAvatar } from "@/platform/gen"
import { Emitter } from "./rtc"

const defaultInputHeight = "1px"
const maxMessages = 100

interface Message {
  id: string
  from: string
  fromName: string
  avatar: string
  text: string
  isSent: boolean
  isFirstFrom: boolean
  isLatestFrom: boolean
}

export default defineComponent({
  name: "Messages",
  props: {
    userId: {
      type: String,
      required: true,
    },
    emitter: Object as PropType<Emitter>,
  },
  data() {
    return {
      message: "",
      byId: {} as { [key: string]: Message },
      ordered: [] as Message[],
      inputStyle: {
        height: defaultInputHeight,
      },
      autoScroll: true,
    }
  },

  watch: {
    message(val: string) {
      const el = this.$refs.input as HTMLTextAreaElement
      if (val.length !== 0) {
        el.style.height = "1px"
        this.inputStyle["height"] = `${el.scrollHeight}px`
      } else {
        el.style.height = defaultInputHeight
      }
    },
    ordered: {
      deep: true,
      handler() {
        const el = this.$refs.browser as HTMLElement
        this.$nextTick(() => {
          if (this.autoScroll) {
            el.scrollTop = el.scrollHeight
          }
        })
      },
    },
  },

  methods: {
    processRecords(records: Record[]) {
      for (const r of records) {
        if (r.event.payload.type !== EventType.Message) {
          continue
        }

        const payload = r.event.payload.payload as PayloadMessage

        let isFirstFrom = true
        if (this.ordered.length) {
          const latest = this.ordered[this.ordered.length - 1]
          if (latest.from === r.event.ownerId) {
            latest.isLatestFrom = false
            isFirstFrom = false
          }
        }
        const msg: Message = {
          id: r.event.id || "",
          from: r.event.ownerId || "",
          fromName: genName(r.event.ownerId || ""),
          avatar: genAvatar(r.event.ownerId || "", 32 + 1),
          text: payload.text,
          isSent: true,
          isFirstFrom: isFirstFrom,
          isLatestFrom: true,
        }

        const existing = this.byId[msg.id]
        if (existing) {
          existing.isSent = true
        } else {
          this.byId[msg.id] = msg
          this.ordered.push(msg)
        }

        if (this.ordered.length > maxMessages) {
          delete this.byId[this.ordered[0].id]
          this.ordered.shift()
        }
      }
    },
    keySend(ev: KeyboardEvent) {
      if (ev.shiftKey || ev.code !== "Enter" || this.message.length === 0) {
        return
      }
      ev.preventDefault()

      this.send()
    },
    async send() {
      if (!this.emitter) {
        return
      }

      const msg: Message = {
        id: "",
        from: this.userId,
        fromName: "",
        avatar: "",
        text: this.message,
        isSent: false,
        isFirstFrom: false,
        isLatestFrom: true,
      }
      const ev = {
        payload: {
          type: EventType.Message,
          payload: {
            text: this.message,
          },
        },
      }
      this.message = ""
      this.inputStyle["height"] = defaultInputHeight
      this.autoScroll = true

      this.ordered.push(msg)
      msg.id = await this.emitter.event(ev)
      this.byId[msg.id] = msg
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
  display: flex
  flex-direction: column

.browser
  width: 100%
  flex-grow: 1
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
  min-height: 2em
  width: 100%
  border-top: 1px solid var(--color-outline)
  padding: 0.5em
  padding-right: 2em
  color: var(--color-font)
  resize: none
  overflow: hidden
  &:disabled
    cursor: default
    background-color: var(--color-outline)

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
