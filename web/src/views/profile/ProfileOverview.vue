<template>
  <div class="profile-overview">
    <div class="avatar">
      <img class="avatar-img" :src="avatar" />
      <div>{{ profile.givenName || genName(profile.ownerId) }} {{ profile.familyName }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { genName } from "@/platform/gen"
import { Profile } from "@/api/models/profile"

defineProps<{
  profile: Profile
  avatar: string
}>()
</script>

<style scoped lang="sass">
@use '@/css/theme'

.profile-overview
  padding: 30px 100px
  text-align: center

.avatar
  display: inline-block
  text-align: center

.avatar-img
  @include theme.shadow-l

  border-radius: 50%
  width: 200px
  margin-bottom: 20px
</style>
