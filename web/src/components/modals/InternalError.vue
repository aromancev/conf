<template>
  <Modal v-on:click="click" :buttons="{ ok: 'OK' }">
    <p>Oh snap! Something unexpected happen.</p>
    <p>
      Our engineers are already working on the problem. Please try again later.
    </p>
  </Modal>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import Modal from "./Modal.vue"

export default defineComponent({
  name: "InternalError",
  emits: ["click"],
  components: {
    Modal,
  },
  methods: {
    click(id: string) {
      this.$emit("click", id)
    },
  },
})
</script>
