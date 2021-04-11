<template>
  <div class="container">
    <div class="row">
      <h1>{{ confa.handle }}</h1>
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
import { userStore } from "@/iam"
import { client, Confa } from "@/confa"

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
      confa: {} as Confa,
      modal: Dialog.None,
    }
  },
  async beforeCreate() {
    const handle = this.$route.params.confa as string
    try {
      if (handle === "new") {
        this.confa = await client.create()
        this.$router.replace("/" + this.confa.handle)
      } else {
        const confas = await client.confas({
          handle: handle,
        })
        this.confa = confas.items[0]
      }
    } catch (e) {
      this.modal = Dialog.Error
    }
  },
})
</script>
