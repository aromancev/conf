<template>
  <img class="avatar-img" :src="avatar" />
  <InternalError v-if="modal === 'error'" @click="modal = 'none'" />
</template>

<script setup lang="ts">
import { ref } from "vue"
import InternalError from "@/components/modals/InternalError.vue"

type Modal = "none" | "error"

defineProps<{
  avatar: string
}>()

const modal = ref<Modal>("none")
</script>

<style scoped lang="sass">
@use '@/css/theme'

.avatar-img
  @include theme.shadow-l

  border-radius: 50%
  width: 200px
  margin: 30px
  margin-left: 100px
</style>
