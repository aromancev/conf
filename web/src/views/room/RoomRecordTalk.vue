<template>
  <div>Hello!</div>
  <input ref="rec" type="text" :style="{ border: '2px solid red' }" />
  <input ref="room" type="text" :style="{ border: '2px solid red' }" />
  <video ref="vid" controls="true" width="700" @click="start"></video>
</template>

<script setup lang="ts">
import { MediaPlayer } from "dashjs"
import { ref } from "vue"

const vid = ref<HTMLElement>()
const rec = ref<HTMLInputElement>()
const room = ref<HTMLInputElement>()

function start() {
  const player = MediaPlayer().create()
  const url = `http://localhost/api/storage/confa-tracks-public/${room.value?.value}/${rec.value?.value}/manifest`
  player.initialize(vid.value, url, false)
}
</script>

<style scoped lang="sass"></style>
