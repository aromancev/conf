<template>
  <div v-if="userId">
    <router-link class="btn px-3 py-2" to="/new">Create confa</router-link>
  </div>
  <div v-else>
    <div>{{ userId }}</div>
    <router-link class="btn px-3 py-2" to="/login">Sign in</router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { userStore } from "@/api"

export default defineComponent({
  name: "Home",
  components: {},
  computed: {
    userId() {
      return userStore.getState().id
    },
  },
})
</script>
