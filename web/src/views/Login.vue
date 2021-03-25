<template>
  <div class="d-flex justify-content-center">
    <div class="content">
      <div class="title">Sign in to Confa</div>
      <div>
        <div class="error" :class="{active: invalid}">Email is invalid</div>
        <input
          v-model="email"
          spellcheck="false"
          class="email input px-3 py-2"
          type="text"
          placeholder="email address"
          @focus="resetInvalid"
          @change="resetInvalid"
        />
      </div>
      <button @click="login" class="submit btn px-3 py-1" :disabled="submitted">Sign in</button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { iam } from "@/platform/api/iam"
import { isValid } from "@/platform/email/email"

export default defineComponent({
  name: "Home",
  components: {},
    data() {
      return {
        email: "",
        submitted: false,
        invalid: false,
      }
  },
  methods: {
    resetInvalid() {
      this.invalid = false
    },

    async login() {
      if (!isValid(this.email)) {
        this.invalid = true
        return
      }
      this.submitted = true
      await iam.login(this.email)
    }
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
