<template>
  <div class="content">
    <div class="title">Video conference platform</div>
    <div class="sub-title">One click broadcast and recording</div>
    <router-link class="cta btn" :to="{ name: 'talkOnline', params: { confa: 'new', talk: 'new' } }">
      Start a conference
    </router-link>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped lang="sass">
@use '@/css/theme'

.content
  max-width: theme.$content-width
  text-align: center

.title
  font-weight: bold
  font-size: 3rem
  margin: 20px
  margin-top: 100px
  text-align: center

.cta
  margin: 100px
</style>
