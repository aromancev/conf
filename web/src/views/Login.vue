<template>
  <div class="d-flex justify-content-center">
    <div class="content">
      <div class="title">Sign in to Confa</div>
      <div>
        <div class="error" :class="{ active: invalid }">Email is invalid</div>
        <input
          v-model="email"
          spellcheck="false"
          class="email input px-3 py-2"
          type="text"
          placeholder="email address"
          v-bind:disabled="submitted"
          @focus="resetInvalid"
          @change="resetInvalid"
        />
      </div>
      <button @click="login" class="submit btn px-3 py-1" :disabled="submitted">
        Sign in
      </button>
    </div>
  </div>

  <Modal
    v-if="modal == Dialog.EmailSent"
    v-on:click="onModalClick"
    :buttons="{ ok: 'OK' }"
  >
    <p>Email sent!</p>
    <p>Check your inbox to sign in.</p>
  </Modal>
  <Modal
    v-if="modal == Dialog.Error"
    v-on:click="onModalClick"
    :buttons="{ ok: 'OK' }"
  >
    <p>Oh snap! Something unexpected happen.</p>
    <p>
      Our engineers are already working on the problem. Please try again later.
    </p>
  </Modal>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { iam } from "@/platform/api/iam"
import { isValid } from "@/platform/email/email"
import Modal from "@/components/Modal.vue"

enum Dialog {
  None = "",
  EmailSent = "sent",
  Error = "error",
}

export default defineComponent({
  name: "Home",
  components: {
    Modal,
  },
  data() {
    return {
      Dialog,
      email: "",
      submitted: false,
      invalid: false,
      modal: Dialog.None,
    }
  },
  async beforeCreate() {
    const query = new URLSearchParams(window.location.search)
    const token = query.get("token")
    if (token) {
      try {
        await iam.session(token)
        this.$router.push("/")
      } catch (e) {
        this.modal = Dialog.Error
      }
    }
  },
  methods: {
    resetInvalid() {
      this.invalid = false
    },

    onModalClick() {
      if (this.modal === Dialog.EmailSent) {
        this.$router.push("/")
      }
      this.modal = Dialog.None
    },

    async login() {
      if (this.submitted) {
        return
      }
      if (!isValid(this.email)) {
        this.invalid = true
        return
      }
      this.submitted = true
      try {
        await iam.login(this.email)
        this.modal = Dialog.EmailSent
      } catch (e) {
        this.modal = Dialog.Error
        this.submitted = false
      }
    },
  },
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

.error
  color: rgba(0, 0, 0, 0)
  width: 100%
  text-align: left
  margin-bottom: 0.3em
  font-size: 0.9em
  &.active
    color: #ff5252
</style>
