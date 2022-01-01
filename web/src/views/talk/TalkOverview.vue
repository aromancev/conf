<template>
  <div class="layout">
    <div class="description">{{ talk.description }}</div>
  </div>

  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref } from "vue"
import { Talk } from "@/api/models"
import InternalError from "@/components/modals/InternalError.vue"

enum Modal {
  None = "",
  Error = "error",
}

defineProps<{
  talk: Talk
}>()

const modal = ref(Modal.None)
</script>

<style scoped lang="sass">
@use '@/css/theme'

.layout
  width: 100%
  max-width: theme.$content-width
  display: flex
  flex-direction: row
  justify-content: center
  padding: 30px

.description
  flex-grow: 1
  margin: 10px
  padding: 20px
  white-space: pre-wrap
</style>
