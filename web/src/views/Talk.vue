<template>
  <LiveTalkRoom v-if="talk" :room-id="talk.roomId" />

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { Talk, talkClient, confaClient } from "@/api"
import InternalError from "@/components/modals/InternalError.vue"
import LiveTalkRoom from "@/views/room/LiveTalkRoom.vue"

enum Modal {
  None = "",
  Error = "error",
}

const modal = ref(Modal.None)
const talk = ref<Talk | null>()

const props = defineProps<{
  handle: string
  confaHandle: string
  tab: string
}>()

watch(
  () => props.handle,
  async (value) => {
    if (value === "new") {
      const confa = await confaClient.fetchOne({ handle: props.confaHandle })
      if (!confa) {
        throw new Error("Failed to feth confa.")
      }
      talk.value = await talkClient.create(confa.id)
      return
    }

    try {
      talk.value = await talkClient.fetchOne({
        handle: value,
      })
    } catch (e) {
      modal.value = Modal.Error
    }
  },
  { immediate: true },
)
</script>
