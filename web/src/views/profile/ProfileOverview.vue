<template>
  <div class="profile-overview">
    <div class="avatar">
      <img class="avatar-img" :src="avatar" />
      <div>{{ name }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  avatar: string
  name: string
}>()
</script>

<style scoped lang="sass">
@use '@/css/theme'

.profile-overview
  width: 100%
  max-width: theme.$content-width
  padding: 30px 100px
  text-align: left

.avatar
  display: inline-block
  text-align: center

.avatar-img
  @include theme.shadow-l

  border-radius: 50%
  width: 200px
  margin-bottom: 20px
</style>
