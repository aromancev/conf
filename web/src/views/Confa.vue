<template>
  <div class="container">
    <div class="row">
      <div v-if="confa">
        <router-link
          class="btn px-3 py-2"
          :to="{ name: 'talk', params: { confa: confa.handle, talk: 'new' } }"
          >Create talk</router-link
        >
      </div>
      <h1 v-else>NOT FOUND</h1>

      <InternalError
        v-if="modal === Dialog.Error"
        v-on:click="modal = Dialog.None"
      />
    </div>
  </div>
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import { defineComponent } from "vue"
import { userStore, Confa } from "@/api/models"
import { confa } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
}

export default defineComponent({
  name: "Confa",
  components: {
    InternalError,
  },
  data() {
    return {
      Dialog,
      user: userStore,
      confa: null as Confa | null,
      modal: Dialog.None,
    }
  },
  async beforeCreate() {
    const handle = this.$route.params.confa as string
    try {
      if (handle === "new") {
        this.confa = await confa.create()
        this.$router.replace("/" + this.confa.handle)
      } else {
        this.confa = await confa.fetchOne({
          handle: handle,
        })
        if (this.confa === null) {
          alert("confa not found")
          return
        }
      }
    } catch (e) {
      this.modal = Dialog.Error
    }
  },
})
</script>
