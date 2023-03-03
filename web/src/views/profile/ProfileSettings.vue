<template>
  <div class="profile-settings">
    <PageLoader v-if="isLoading"></PageLoader>
    <div v-else class="form">
      <div v-if="user.hasPassword">
        <InputField
          v-model="oldPassword"
          class="field"
          type="password"
          label="Old password"
          :errors="oldPasswordErrors"
          autocomplete="current-password"
        ></InputField>
        <InputField
          v-model="newPassword"
          class="field"
          type="password"
          label="New password"
          :errors="newPasswordErrors"
          autocomplete="new-password"
        ></InputField>
        <InputField
          v-model="confirmNewPassword"
          class="field"
          type="password"
          label="Confirm new password"
          :errors="confirmNewPasswordErrors"
          autocomplete="new-password"
        ></InputField>
        <button class="btn" :disabled="isSubmitted" @click="updatePassword">Update password</button>
        <span class="link" @click="emailResetPassword">I forgot my password</span>
      </div>
      <div v-else>
        <button class="btn" @click="emailCreatePassword">Create password</button>
      </div>
    </div>
  </div>
  <ModalDialog :is-visible="modal.state === 'PASSWORD_EMAIL_SENT'" :buttons="{ ok: 'OK' }" @click="modal.set()">
    <p>Email sent!</p>
    <p>Check your inbox to set password.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal.state === 'WRONG_PASSWORD'" :buttons="{ ok: 'OK' }" @click="modal.set()">
    <p>Wrong password. Please try again.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal.state === 'PASSWORD_UPDATED'" :ctrl="modal" :buttons="{ ok: 'OK' }">
    <p>Password updated.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue"
import InputField from "@/components/fields/InputField.vue"
import { accessStore } from "@/api/models/access"
import { passwordValidator } from "@/api/models/user"
import { User, UserClient, Platform } from "@/api/user"
import { api, Code, errorCode } from "@/api"
import { route } from "@/router"
import { notificationStore } from "@/api/models/notifications"
import PageLoader from "@/components/PageLoader.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import { ModalController } from "@/components/modals/controller"
import router from "@/router"

type Modal = "PASSWORD_EMAIL_SENT" | "PASSWORD_UPDATED" | "WRONG_PASSWORD"
const oldPassword = ref<string>("")
const newPassword = ref<string>("")
const isSubmitted = ref<boolean>(false)
const confirmNewPassword = ref<string>("")
const users = new UserClient(api)
const user = ref<User>({
  id: "",
  identifiers: [],
  hasPassword: false,
})
const modal = new ModalController<Modal>()
const isLoading = ref<boolean>(true)
const oldPasswordErrors = computed<string[]>(() => {
  if (!isSubmitted.value) {
    return []
  }
  if (oldPassword.value) {
    return []
  }
  return ["Must not be empty"]
})
const newPasswordErrors = computed<string[]>(() => {
  if (!isSubmitted.value) {
    return []
  }
  const errs = passwordValidator.validate(newPassword.value)
  if (oldPassword.value === newPassword.value) {
    errs.push("Must be different from the old password")
  }
  return errs
})
const confirmNewPasswordErrors = computed<string[]>(() => {
  if (!isSubmitted.value) {
    return []
  }
  if (newPassword.value === confirmNewPassword.value) {
    return []
  }
  return ["Must be the same as the new password"]
})

watch(
  accessStore.state,
  async () => {
    try {
      user.value = await users.fetchCurrent()
    } catch {
      notificationStore.error("failed to fetch settings")
    } finally {
      isLoading.value = false
    }
  },
  { immediate: true },
)

watch([newPassword, oldPassword, confirmNewPassword], () => (isSubmitted.value = false))

async function emailCreatePassword() {
  const email = user.value.identifiers.find((i) => i.platform === Platform.EMAIL)
  if (!email) {
    notificationStore.error("user doesn't have an email")
    return
  }
  isLoading.value = true
  try {
    await api.emailCreatePassword(email.value)
    modal.set("PASSWORD_EMAIL_SENT")
    clearForm()
  } catch {
    notificationStore.error("failed to send email")
  } finally {
    isLoading.value = false
  }
}

async function emailResetPassword() {
  const email = user.value.identifiers.find((i) => i.platform === Platform.EMAIL)
  if (!email) {
    notificationStore.error("user doesn't have an email")
    return
  }
  isLoading.value = true
  try {
    await api.emailResetPassword(email.value)
    clearForm()
    modal.set("PASSWORD_EMAIL_SENT")
  } catch {
    notificationStore.error("failed to send email")
  } finally {
    isLoading.value = false
  }
}

async function updatePassword() {
  if (isSubmitted.value) {
    return
  }
  isSubmitted.value = true

  if (oldPasswordErrors.value.length || newPasswordErrors.value.length || confirmNewPasswordErrors.value.length) {
    return
  }

  isLoading.value = true
  try {
    await api.updatePassword(oldPassword.value, newPassword.value)
    await modal.set("PASSWORD_UPDATED")
    clearForm()
    accessStore.logout()
    api.refreshToken()
    router.push(route.login())
  } catch (e) {
    if (errorCode(e) === Code.BadRequest) {
      modal.set("WRONG_PASSWORD")
    } else {
      notificationStore.error("failed to update password")
    }
  } finally {
    isLoading.value = false
  }
}

function clearForm() {
  oldPassword.value = ""
  newPassword.value = ""
  confirmNewPassword.value = ""
  isSubmitted.value = false
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.profile-settings
  height: 100%
  display: flex
  justify-content: center
  align-items: center

.field
  width: 100%
  margin: 10px 0

.form
  width: 400px
  height: 100%
  padding-top: 50px
  text-align: left

.link
  margin: 0 10px
</style>
