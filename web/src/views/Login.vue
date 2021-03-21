<template>
  <div class="d-flex justify-content-center">
    <div class="content">
      <div class="title">Sign in to Confa</div>
      <div>
        <input
          spellcheck="false"
          class="email input px-3 py-2"
          type="text"
          placeholder="email address"
        />
      </div>
      <button class="submit btn px-3 py-1">Sign in</button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"

export default defineComponent({
  name: "Home",
  components: {}
})
</script>

<style lang="sass" scoped>
.content
  margin-top: 7rem
  witdh: 10rem

.title
  font-size: 2em
  margin-bottom: 2rem

.email
  width: 100%


.submit
  width: 100%
  margin: 1rem 0
</style>
