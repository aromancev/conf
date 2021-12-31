<template>
  <RoomOnlineTalk v-if="talk" :room-id="talk.roomId" :join-confirmed="joinConfirmed" @join="join" />
  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref } from "vue"
import { Talk } from "@/api"
import InternalError from "@/components/modals/InternalError.vue"
import RoomOnlineTalk from "@/views/room/RoomOnlineTalk.vue"

enum Modal {
  None = "",
  Error = "error",
}

const emit = defineEmits<{
  (e: "join", confirmed: boolean): void
}>()

const modal = ref(Modal.None)

defineProps<{
  talk: Talk
  joinConfirmed?: boolean
}>()

function join(confirmed: boolean) {
  emit("join", confirmed)
}
</script>
