<template>
  <div class="container">
    <div class="row">
      <audience ref="audience" />
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
import { defineComponent } from "vue"
import { Event, client } from "@/api"
import { RecordProcessor } from "@/components/room"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "Room",
  components: {
    Audience,
    InternalError,
  },
  props: {
    roomId: String,
  },

  watch: {
    async roomId(val: string) {
      const rtc = await client.rtc(val)
      const aud = this.$refs.audience as RecordProcessor
      rtc.onevent = (event: Event) => {
        aud.processRecords([
          {
            event: event,
            live: true,
            forward: true,
          },
        ])
      }
    },
  },

  data() {
    return {
      Dialog,
      modal: Dialog.None,
    }
  },

  methods: {},
})
</script>

<style scoped lang="sass">
.audience
  margin: 100px
  width: 1000px
  height: 700px
</style>
