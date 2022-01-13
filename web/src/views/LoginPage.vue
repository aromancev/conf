<template>
  <div class="d-flex justify-content-center">
    <div class="content">
      <div class="title">Sign in to Confa</div>
      <div>
        <InputField
          v-model="email"
          :spellcheck="false"
          class="email"
          type="text"
          placeholder="email address"
          :disabled="submitted"
          :error="valid ? '' : emailError"
        />
      </div>
      <button class="submit btn" :disabled="submitted || !valid" @click="login">Sign in</button>
    </div>
  </div>

  <ModalDialog v-if="modal == Modal.EmailSent" :buttons="{ ok: 'OK' }" @click="router.push({ name: 'home' })">
    <p>Email sent!</p>
    <p>Check your inbox to sign in.</p>
  </ModalDialog>
  <ModalDialog v-if="modal == Modal.BadRequest" :buttons="{ ok: 'OK' }" @click="modal = Modal.None">
    <p>Incorrect email.</p>
  </ModalDialog>
  <InternalError v-if="modal == Modal.Error" @click="modal = Modal.None" />
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useRouter } from "vue-router"
import { client, errorCode, Code } from "@/api"
import { userStore } from "@/api/models"
import { isValid } from "@/platform/email"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import InputField from "@/components/fields/InputField.vue"
import InternalError from "@/components/modals/InternalError.vue"

enum Modal {
  None = "",
  EmailSent = "sent",
  BadRequest = "bad_request",
  Error = "error",
}

const emailError = "• Must be a valid email"

const email = ref("")
const submitted = ref(false)
const valid = ref(false)
const modal = ref(Modal.None)

const router = useRouter()
const user = userStore.getState()

const props = defineProps<{
  token?: string
}>()

watch(
  user,
  () => {
    if (user.allowedWrite) {
      router.replace({ name: "home" })
    }
  },
  { immediate: true },
)

watch(
  () => props.token,
  async (value) => {
    if (!value) {
      return
    }
    try {
      await client.createSession(value)
    } catch (e) {
      modal.value = Modal.Error
    }
  },
  { immediate: true },
)

watch(email, (value) => {
  valid.value = isValid(value)
})

async function login() {
  if (submitted.value || !valid.value) {
    return
  }
  submitted.value = true
  try {
    await client.login(email.value)
    modal.value = Modal.EmailSent
  } catch (e) {
    if (errorCode(e) === Code.BadRequest) {
      modal.value = Modal.BadRequest
    } else {
      modal.value = Modal.Error
    }
    submitted.value = false
  }
}
</script>

<style lang="sass" scoped>

.content
  margin-top: 7rem
  width: 300px

.title
  font-size: 2em
  margin-bottom: 2rem
  text-align: center

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
