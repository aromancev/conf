<template>
  <InternalError v-if="modal === 'error'" @click="modal = 'none'" />
</template>

<script setup lang="ts">
import { ref } from "vue"
import InternalError from "@/components/modals/InternalError.vue"

type Modal = "none" | "error"

const modal = ref<Modal>("none")
</script>

<style scoped lang="sass">
@use '@/css/theme'
</style>
