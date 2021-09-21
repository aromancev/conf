<template>
  <div class="container">
    <div class="row">
      <audience ref="audience" />
      <messages ref="messages" :userId="userId" :emitter="rtc" />
      <InternalError
        v-if="modal === Dialog.Error"
        v-on:click="modal = Dialog.None"
      />
    </div>
  </div>
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import Audience from "@/components/room/audience.vue"
import Messages from "@/components/room/messages.vue"
import { defineComponent } from "vue"
import { userStore, RTC, Event, client, event } from "@/api"
import { RecordProcessor, BufferedProcessor } from "@/components/room"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "Room",
  components: {
    Audience,
    Messages,
    InternalError,
  },
  props: {
    roomId: String,
  },

  data() {
    return {
      Dialog,
      modal: Dialog.None,
      rtc: null as RTC | null,
    }
  },

  computed: {
    userId() {
      return userStore.getState().id
    },
  },

  watch: {
    async roomId(val: string) {
      const processors = [
        this.$refs.audience,
        this.$refs.messages,
      ] as RecordProcessor[]
      const buffered = new BufferedProcessor(processors, 500)

      const rtc = await client.rtc(val)
      rtc.onevent = (event: Event) => {
        buffered.put([event], true)
      }

      const iter = event.fetch({ roomId: val })
      const events = await iter.next({ count: 500, seconds: 2 * 60 * 60 })
      buffered.flush()
      buffered.put(events, false)
      buffered.autoflush = true

      this.rtc = rtc
    },
  },

  methods: {},
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

.audience
  position: absolute
  left: 0
  margin: 100px
  width: 300px
  height: 500px

.messages
  @include theme.shadow-inset-s
  position: absolute
  right: 0
  margin: 100px
  width: 500px
  height: 500px
</style>
