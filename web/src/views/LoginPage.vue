<template>
  <div class="content">
    <div v-if="state === 'LOADING'" class="loader-wrapper">
      <PageLoader></PageLoader>
    </div>

    <div v-if="state === 'SIGN_IN_WITH_EMAIL'">
      <div class="title">Sign in</div>
      <GSIButton class="gsi" @token="loginWithGSI"></GSIButton>
      <div>or use you email to sign in:</div>
      <InputField
        v-model="email"
        :spellcheck="false"
        class="field"
        type="text"
        placeholder="Your email"
        autocomplete="email"
        :errors="emailErrors"
      />
      <button class="btn" :disabled="isSubmitted" @click="login">Email sign in link</button>
      <span class="link margin" @click="state = 'SIGN_IN_WITH_PASSWORD'">sign in with password</span>
    </div>

    <div v-if="state === 'SIGN_IN_WITH_PASSWORD'">
      <div class="title">Sign in</div>
      <InputField
        v-model="email"
        :spellcheck="false"
        class="field"
        type="text"
        placeholder="Your email"
        autocomplete="email"
        :errors="emailErrors"
      />
      <InputField v-model="password" class="field" type="password" placeholder="Password" autocomplete="password" />
      <button class="btn" :disabled="isSubmitted" @click="login">Sign in</button>
      <span class="link margin" @click="state = 'SIGN_IN_WITH_EMAIL'">sign in with email</span>
      <br />
      <br />
      <div class="link" @click="emailResetPassword">I forgot my password</div>
    </div>

    <div v-if="state === 'CREATE_PASSWORD'" class="centered">
      <ProfileAvatar
        v-if="accessStore.state.allowedWrite"
        class="avatar"
        :size="128"
        :user-id="accessStore.state.id"
        :src="profileStore.state.avatarThumbnail"
      ></ProfileAvatar>
      <div class="user">Logged in as {{ loggedInAs }}</div>
      <InputField
        v-model="password"
        :spellcheck="false"
        type="password"
        label="Password"
        autocomplete="password"
        :errors="passwordErrors"
      />
      <InputField
        v-model="confirmPassword"
        type="password"
        label="Confirm password"
        autocomplete="new-password"
        :errors="confirmPasswordErrors"
      />
      <button class="btn" :disabled="isSubmitted" @click="createPassword">Set password</button>
    </div>

    <div v-if="state === 'RESET_PASSWORD'">
      <InputField
        v-model="password"
        type="password"
        label="New password"
        autocomplete="new-password"
        :errors="passwordErrors"
      />
      <InputField
        v-model="confirmPassword"
        type="password"
        label="Confirm password"
        autocomplete="new-password"
        :errors="confirmPasswordErrors"
      />
      <button class="btn" :disabled="isSubmitted" @click="resetPassword">Set password</button>
    </div>

    <div v-if="state === 'SIGNED_IN'" class="centered">
      <ProfileAvatar
        v-if="accessStore.state.allowedWrite"
        class="avatar"
        :size="128"
        :user-id="accessStore.state.id"
        :src="profileStore.state.avatarThumbnail"
      ></ProfileAvatar>
      <div class="user">Logged in as {{ loggedInAs }}</div>
      <router-link class="btn create-talk" :to="route.talk(handleNew, handleNew, 'watch')"
        >Start broadcasting</router-link
      >
    </div>
  </div>

  <ModalDialog :is-visible="modal === 'EMAIL_LOGIN_SENT'" :buttons="[{ text: 'OK' }]" @click="modal = 'NONE'">
    <p>Email sent!</p>
    <p>Check your inbox to sign in.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal === 'EMAIL_RESET_SENT'" :buttons="[{ text: 'OK' }]" @click="modal = 'NONE'">
    <p>If such user exists, an email will be sent.</p>
    <p>Check your inbox to reset password.</p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal === 'PASSWORD_CREATED'"
    :buttons="[{ text: 'OK' }]"
    @click="
      () => {
        modal = 'NONE'
        state = 'SIGNED_IN'
        clearQuery()
      }
    "
  >
    <p>Passwword created.</p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal === 'PASSWORD_RESET'"
    :buttons="[{ text: 'OK' }]"
    @click="
      () => {
        modal = 'NONE'
        state = 'SIGN_IN_WITH_PASSWORD'
        clearQuery()
      }
    "
  >
    <p>New password set.</p>
    <p>You can use it to log in.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'BAD_REQUEST'" :buttons="[{ text: 'OK' }]" @click="modal = 'NONE'">
    <p>Incorrect email.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'NOT_FOUND'" :buttons="[{ text: 'OK' }]" @click="modal = 'NONE'">
    <p>Incorrect password or no such user.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'INVALID_TOKEN'" :buttons="[{ text: 'OK' }]" @click="modal = 'NONE'">
    <p>Email token has expired or has been used before.</p>
    <p>Please request a new email.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal == 'ALREADY_HAS_PASSWORD'" :buttons="[{ text: 'OK' }]" @click="modal = 'NONE'">
    <p>User already has a password.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue"
import { useRouter } from "vue-router"
import { api, errorCode, Code } from "@/api"
import { Account, accessStore } from "@/api/models/access"
import { emailValidator, passwordValidator } from "@/api/models/user"
import { route, handleNew, LoginAction } from "@/router"
import { notificationStore } from "@/api/models/notifications"
import { profileStore } from "@/api/models/profile"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import InputField from "@/components/fields/InputField.vue"
import PageLoader from "@/components/PageLoader.vue"
import ProfileAvatar from "@/components/profile/ProfileAvatar.vue"
import GSIButton from "@/components/GSIButton.vue"
import { genName } from "@/platform/gen"

type Modal =
  | "NONE"
  | "EMAIL_LOGIN_SENT"
  | "EMAIL_RESET_SENT"
  | "PASSWORD_CREATED"
  | "PASSWORD_RESET"
  | "BAD_REQUEST"
  | "NOT_FOUND"
  | "INVALID_TOKEN"
  | "ALREADY_HAS_PASSWORD"
type State =
  | "LOADING"
  | "SIGN_IN_WITH_EMAIL"
  | "SIGN_IN_WITH_PASSWORD"
  | "SIGNED_IN"
  | "CREATE_PASSWORD"
  | "RESET_PASSWORD"

const props = defineProps<{
  action?: LoginAction
  token?: string
}>()

const isSubmitted = ref<boolean>(false)
const password = ref<string>("")
const confirmPassword = ref<string>("")
const modal = ref<Modal>("NONE")
const state = ref<State>("LOADING")
const email = ref<string>("")
const emailErrors = computed<string[]>(() => (isSubmitted.value ? emailValidator.validate(email.value) : []))
const passwordErrors = computed<string[]>(() => (isSubmitted.value ? passwordValidator.validate(password.value) : []))
const confirmPasswordErrors = computed<string[]>(() => {
  if (!isSubmitted.value) {
    return []
  }
  if (password.value === confirmPassword.value) {
    return []
  }
  return ["Must be the same as the new password"]
})
const loggedInAs = computed<string>(() => {
  if (profileStore.state.givenName) {
    return `${profileStore.state.givenName} ${profileStore.state.familyName}`
  }
  return genName(accessStore.state.id)
})

const router = useRouter()
let resetPasswordToken = ""

watch(
  [() => accessStore.state.id, () => props.action, () => props.token],
  async () => {
    if (accessStore.state.id === "") {
      return
    }

    if (!props.action || !props.token) {
      state.value = accessStore.state.account === Account.Guest ? "SIGN_IN_WITH_EMAIL" : "SIGNED_IN"
      return
    }

    state.value = "LOADING"
    try {
      if (accessStore.state.account === Account.Guest && props.action === "login") {
        await api.createSessionWithEmail(props.token)
        state.value = "SIGNED_IN"
        router.replace(route.login())
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
      state.value = "SIGN_IN_WITH_EMAIL"
    }
  },
  { immediate: true },
)

watch([email, password, confirmPassword, state], () => (isSubmitted.value = false))

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
      await api.createSessionWithCredentials(email.value, password.value)
      clearForm()
      state.value = "SIGNED_IN"
    } else {
      await api.emailLogin(email.value)
      clearForm()
      modal.value = "EMAIL_LOGIN_SENT"
      state.value = "SIGN_IN_WITH_EMAIL"
    }
  } catch (e) {
    switch (errorCode(e)) {
      case Code.BadRequest:
        modal.value = "BAD_REQUEST"
        break
      case Code.NotFound:
        modal.value = "NOT_FOUND"
        state.value = "SIGN_IN_WITH_EMAIL"
        break
      default:
        notificationStore.error("failed to login")
        state.value = "SIGN_IN_WITH_EMAIL"
        break
    }
  }
}

async function loginWithGSI(token: string) {
  await api.createSessionWithGSI(token)
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
    state.value = "SIGN_IN_WITH_EMAIL"
  }
}

async function createPassword() {
  isSubmitted.value = true

  if (passwordErrors.value.length || confirmPasswordErrors.value.length) {
    return
  }

  state.value = "LOADING"
  try {
    await api.createPassword(resetPasswordToken, password.value)
    clearForm()
    accessStore.logout()
    api.refreshToken()
    modal.value = "PASSWORD_CREATED"
    state.value = "SIGNED_IN"
  } catch (e) {
    if (errorCode(e) === Code.NotFound) {
      modal.value = "ALREADY_HAS_PASSWORD"
      state.value = "SIGNED_IN"
      router.replace(route.login())
    } else {
      notificationStore.error("failed to create password")
      state.value = "SIGN_IN_WITH_EMAIL"
    }
  }
}

async function resetPassword() {
  isSubmitted.value = true

  if (passwordErrors.value.length || confirmPasswordErrors.value.length) {
    return
  }

  state.value = "LOADING"
  try {
    await api.resetPassword(resetPasswordToken, password.value)
    clearForm()
    accessStore.logout()
    api.refreshToken()
    modal.value = "PASSWORD_RESET"
    state.value = "SIGN_IN_WITH_PASSWORD"
  } catch {
    notificationStore.error("failed to create password")
    state.value = "SIGN_IN_WITH_PASSWORD"
  }
}

function clearForm() {
  email.value = ""
  password.value = ""
  confirmPassword.value = ""
  isSubmitted.value = false
}

function clearQuery() {
  window.history.replaceState(null, "", window.location.href.split("?")[0])
}
</script>

<style lang="sass" scoped>
@use '@/css/theme'

.content
  width: 400px
  height: 100%
  padding-top: 200px
  text-align: left

.loader-wrapper
  position: absolute
  top: 0
  left: 0
  display: flex
  justify-content: center
  align-items: center
  width: 100%
  height: 100%

.title
  font-size: 40px
  margin-bottom: 20px

.gsi
  max-width: 100%
  margin-bottom: 10px

.avatar
  width: 48px
  height: 48px
  border-radius: 50%
  overflow: hidden
  margin: 10px

.field
  margin: 10px 0

.create-talk
  margin: 15px 0

.user
  margin-bottom: 20px

.margin
  margin: 10px

.centered
  text-align: center
</style>
