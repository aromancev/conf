<template>
  <div v-if="user.loggedIn">
    <router-link class="btn px-3 py-2" to="/new">Create confa</router-link>
  </div>
  <div v-else>
    <router-link class="btn px-3 py-2" to="/login">Sign in</router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { userStore } from "@/iam"

export default defineComponent({
  name: "Home",
  components: {},
  data() {
    return {
      user: userStore.getState(),
    }
  },
})
</script>
