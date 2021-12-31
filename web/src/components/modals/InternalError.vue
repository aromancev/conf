<template>
  <ModalDialog :buttons="{ ok: 'OK' }" @click="click">
    <p>Oh snap! Something unexpected happen.</p>
    <p>Our engineers are already working on the problem. Please try again later.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import ModalDialog from "./ModalDialog.vue"

const emit = defineEmits<{
  (e: "click", id: string): void
}>()

function click(id: string) {
  emit("click", id)
}
</script>
