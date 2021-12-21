<template>
  <LiveTalkRoom :roomId="talk?.roomId" />

  <InternalError
    v-if="modal === Dialog.Error"
    v-on:click="modal = Dialog.None"
  />
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import { defineComponent } from "vue"
import LiveTalkRoom from "@/views/room/LiveTalkRoom.vue"
import { confa, Talk, talk } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "Talk",
  components: {
    LiveTalkRoom,
    InternalError,
  },

  data() {
    return {
      Dialog,
      talk: null as Talk | null,
      modal: Dialog.None,
    }
  },

  async created() {
    this.talk = await this.fetchTalk()
    if (this.talk === null) {
      alert("talk not found")
      return
    }
  },

  methods: {
    async fetchTalk(): Promise<Talk | null> {
      const confaHanle = this.$route.params.confa as string
      const talkHandle = this.$route.params.talk as string

      if (talkHandle !== "new") {
        try {
          return await talk.fetchOne({
            handle: talkHandle,
          })
        } catch (e) {
          this.modal = Dialog.Error
        }
      }

      try {
        const conf = await confa.fetchOne({
          handle: confaHanle,
        })
        if (conf === null) {
          return null
        }
        const tlk = await talk.create(conf.id)
        this.$router.replace("/" + confaHanle + "/" + tlk.handle)
        return tlk
      } catch (e) {
        this.modal = Dialog.Error
      }

      return null
    },
  },
})
</script>
