<template>
  <div class="centered">
    <h1>Disclaimer</h1>
    This service is a hobby project and not intended to make profit. It does not provide any guarantees.
    <br />
    <br />
    All submitted data is public.
    <br />
    <br />
    Any data can be deleted at any moment.
    <br />
    <br />
    It is hosted on a very cheap server so it can be slow or crash under load.
    <br />
    <br />
    <br />
    <CopyText :value="'info@confa.io'"></CopyText>
  </div>
</template>

<script setup lang="ts">
import CopyText from "@/components/fields/CopyText.vue"
</script>

<style scoped lang="sass">
.centered
  height: 100%
  width: 100%
  display: flex
  flex-direction: column
  justify-content: center
  text-align: center
</style>
