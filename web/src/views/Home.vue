<template>
  <div></div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { userStore } from "@/api"

export default defineComponent({
  name: "Home",
  components: {},
  computed: {
    allowedWrite() {
      return userStore.getState().allowedWrite
    },
  },
})
</script>
