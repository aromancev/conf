<template>
  <div>Hello!</div>
  <input ref="input" type="text" :style="{ border: '2px solid red' }" />
  <video ref="vid" controls="true" width="700" @click="start"></video>
</template>

<script setup lang="ts">
import { MediaPlayer } from "dashjs"
import { ref } from "vue"

const vid = ref<HTMLElement>()
const input = ref<HTMLInputElement>()

function start() {
  const player = MediaPlayer().create()
  const url = `http://localhost/api/storage/confa-tracks-public/a617c780-2986-4178-bcef-1e28a51fd031/${input.value?.value}/manifest`
  player.initialize(vid.value, url, false)
}
</script>

<style scoped lang="sass"></style>
