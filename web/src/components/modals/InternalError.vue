<template>
  <ModalDialog :is-visible="props.isVisible" :buttons="[{ text: 'OK' }]" :ctrl="props.ctrl" @click="click">
    <p>Oh snap! Something unexpected happen.</p>
    <p>Our engineers are already working on the problem. Please try again later.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import ModalDialog from "./ModalDialog.vue"
import { Controller } from "./ModalDialog.vue"

const emit = defineEmits<{
  (e: "click", id?: string): void
}>()

const props = defineProps<{
  isVisible: boolean
  disabled?: boolean
  ctrl?: Controller
}>()

function click(id?: string) {
  emit("click", id)
}
</script>
