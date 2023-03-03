<template>
  <div class="content">
    <PageLoader v-if="state === 'LOADING'"></PageLoader>

    <div v-if="state === 'GUEST'">
      <InputField
        v-model="email"
        :spellcheck="false"
        class="field"
        type="text"
        label="Email address"
        autocomplete="email"
        :errors="emailErrors"
      />
      <InputField v-model="password" class="field" type="password" label="Password" autocomplete="password" />
      <button class="submit btn" :disabled="isSubmitted" @click="login">
        {{ password.length ? "Log in with password" : "Send login link" }}
      </button>
      <span class="link" @click="emailResetPassword">I forgot my password</span>
    </div>

    <div v-if="state === 'CREATE_PASSWORD'">
      <div class="title">Logged in as {{ profileStore.state.displayName || genName(accessStore.state.id) }}</div>
      <InputField
        v-model="password"
        :spellcheck="false"
        type="password"
        label="Password"
        autocomplete="password"
        :error="passwordErrors"
      />
      <button class="submit btn" :disabled="isSubmitted || passwordErrors.length !== 0" @click="resetPassword">
        Set password
      </button>
      <div class="or">or</div>
      <router-link class="btn create-talk" :to="route.talk(handleNew, handleNew, 'watch')"
        >Start broadcasting</router-link
      >
    </div>

    <div v-if="state === 'RESET_PASSWORD'">
      <InputField
        v-model="password"
        :spellcheck="false"
        type="password"
        label="Password"
        autocomplete="password"
        :error="passwordErrors"
      />
      <button class="submit btn" :disabled="isSubmitted || passwordErrors.length !== 0" @click="resetPassword">
        Set password
      </button>
    </div>

    <div v-if="state === 'HAS_PASSWORD'">
      <div class="title">Logged in as {{ profileStore.state.displayName || genName(accessStore.state.id) }}</div>
      <router-link class="btn create-talk" :to="route.talk(handleNew, handleNew, 'watch')"
        >Start broadcasting</router-link
      >
    </div>
  </div>

  <ModalDialog :is-visible="modal === 'EMAIL_LOGIN_SENT'" :buttons="{ ok: 'OK' }" @click="modal = 'NONE'">
    <p>Email sent!</p>
    <p>Check your inbox to sign in.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal === 'EMAIL_RESET_SENT'" :buttons="{ ok: 'OK' }" @click="modal = 'NONE'">
    <p>Email sent!</p>
    <p>Check your inbox to reset password.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'BAD_REQUEST'" :buttons="{ ok: 'OK' }" @click="modal = 'NONE'">
    <p>Incorrect email.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'NOT_FOUND'" :buttons="{ ok: 'OK' }" @click="modal = 'NONE'">
    <p>Incorrect password or no such user.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'INVALID_TOKEN'" :buttons="{ ok: 'OK' }" @click="modal = 'NONE'">
    <p>Email token has expired or has been used before.</p>
    <p>Please request a new email.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue"
import { useRouter } from "vue-router"
import { api, errorCode, Code } from "@/api"
import { Account, accessStore } from "@/api/models/access"
import { emailValidator, passwordValidator } from "@/api/models/user"
import { route, handleNew, LoginAction } from "@/router"
import { genName } from "@/platform/gen"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import InputField from "@/components/fields/InputField.vue"
import PageLoader from "@/components/PageLoader.vue"
import { notificationStore } from "@/api/models/notifications"
import { profileStore } from "@/api/models/profile"

type Modal = "NONE" | "EMAIL_LOGIN_SENT" | "EMAIL_RESET_SENT" | "BAD_REQUEST" | "NOT_FOUND" | "INVALID_TOKEN"
type State = "LOADING" | "GUEST" | "HAS_PASSWORD" | "CREATE_PASSWORD" | "RESET_PASSWORD"

const props = defineProps<{
  action?: LoginAction
  token?: string
}>()

const password = ref<string>("")
const modal = ref<Modal>("NONE")
const state = ref<State>("LOADING")
const email = ref<string>("")
const emailErrors = computed<string[]>(() => (isSubmitted.value ? emailValidator.validate(email.value) : []))
const passwordErrors = computed<string[]>(() => (isSubmitted.value ? passwordValidator.validate(password.value) : []))
const isSubmitted = ref<boolean>(false)

const router = useRouter()
let resetPasswordToken = ""

watch(
  accessStore.state,
  async () => {
    if (!props.action || !props.token) {
      state.value = accessStore.state.account === Account.Guest ? "GUEST" : "HAS_PASSWORD"
      return
    }

    state.value = "LOADING"
    try {
      if (accessStore.state.account !== Account.Guest || props.action === "login") {
        const session = await api.createSessionByEmail(props.token)
        if (session?.createPasswordToken) {
          resetPasswordToken = session.createPasswordToken
          state.value = "CREATE_PASSWORD"
        } else {
          state.value = "HAS_PASSWORD"
        }
        return
      }
      switch (props.action) {
        case "create-password":
          resetPasswordToken = props.token
          state.value = "CREATE_PASSWORD"
          break
        case "reset-password":
          resetPasswordToken = props.token
          state.value = "RESET_PASSWORD"
          break
      }
    } catch (e) {
      if (errorCode(e) === Code.Unauthorized) {
        modal.value = "INVALID_TOKEN"
        router.replace(route.login())
      } else {
        notificationStore.error("failed to verify email")
      }
      state.value = "GUEST"
    }
  },
  { immediate: true },
)

watch([email, password], () => (isSubmitted.value = false))

async function login() {
  if (isSubmitted.value) {
    return
  }

  isSubmitted.value = true
  if (emailErrors.value.length) {
    return
  }

  try {
    state.value = "LOADING"
    if (password.value.length) {
      await api.createSessionByLogin(email.value, password.value)
      clearForm()
      state.value = "HAS_PASSWORD"
    } else {
      await api.emailLogin(email.value)
      clearForm()
      modal.value = "EMAIL_LOGIN_SENT"
    }
  } catch (e) {
    switch (errorCode(e)) {
      case Code.BadRequest:
        modal.value = "BAD_REQUEST"
        break
      case Code.NotFound:
        modal.value = "NOT_FOUND"
        state.value = "GUEST"
        break
      default:
        notificationStore.error("failed to login")
        state.value = "GUEST"
        break
    }
  }
}

async function emailResetPassword() {
  isSubmitted.value = true
  if (emailErrors.value.length) {
    return
  }
  state.value = "LOADING"
  try {
    await api.emailResetPassword(email.value)
    clearForm()
    modal.value = "EMAIL_RESET_SENT"
  } catch {
    notificationStore.error("failed to send email")
  } finally {
    state.value = "GUEST"
  }
}

async function resetPassword() {
  state.value = "LOADING"
  try {
    await api.resetPassword(resetPasswordToken, password.value)
    clearForm()
    accessStore.logout()
    api.refreshToken()
    state.value = "GUEST"
    router.push(route.login())
  } catch {
    notificationStore.error("failed to create password")
    state.value = "GUEST"
  }
}

function clearForm() {
  email.value = ""
  password.value = ""
  isSubmitted.value = false
}
</script>

<style lang="sass" scoped>

.content
  width: 300px

.title
  font-size: 20px
  margin-bottom: 20px

.field
  margin: 10px 0

.submit
  width: 100%
  margin: 15px 0

.create-talk
  margin: 15px 0

.or
  color: var(--color-font-disabled)
</style>
