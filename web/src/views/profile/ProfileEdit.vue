<template>
  <div class="form">
    <div class="form-row">
      <div class="form-cell"></div>
      <div class="form-cell">
        <div class="avatar">
          <img class="avatar-img" :src="avatar" />
          <div class="avatar-edit-icon material-icons" @click="editAvatar">edit</div>
        </div>
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell label">Handle</div>
      <div class="form-cell">
        <Input
          v-model="handle"
          :spellcheck="false"
          class="form-input"
          type="text"
          placeholder="handle"
          :error="handleError"
        />
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell label">Display Name</div>
      <div class="form-cell">
        <Input
          v-model="displayName"
          :spellcheck="false"
          class="form-input"
          type="text"
          placeholder="display name"
          :error="displayNameError"
        />
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell"></div>
      <div class="form-cell controls">
        <div class="save-indicator"></div>
        <div class="btn save" :disabled="!hasUpdate || saving || !formValid ? true : null" @click="save">
          <div v-if="saving" class="save-loader">
            <PageLoader />
          </div>

          <span v-if="!saving">{{ !hasUpdate ? "Saved" : "Save" }}</span>
        </div>
      </div>
    </div>
  </div>

  <AvatarEditor
    :is-visible="modal === 'avatar_edit'"
    :avatar="uploadedAvatar"
    :loading="saving"
    @close="modal = 'none'"
    @update="uploadAvatar"
  ></AvatarEditor>
  <ModalDialog :is-visible="modal === 'duplicate_entry'" :buttons="{ ok: 'OK' }" @click="modal = 'none'">
    <p>Profile with this handle already exits.</p>
    <p>Try a different handle.</p>
  </ModalDialog>
  <InternalError :is-visible="modal === 'error'" @click="modal = 'none'" />
</template>

<script lang="ts">
import { RegexValidator } from "@/platform/validator"

const handleValidator = new RegexValidator("^[a-z0-9-]{4,64}$", [
  "Must be from 4 to 64 characters long",
  "Can only contain lower case letters, numbers, and '-'",
])
const displayNameValidator = new RegexValidator("^[a-zA-Z ]{0,64}$", [
  "Must be from 0 to 64 characters long",
  "Can only contain letters and spaces",
])
</script>

<script setup lang="ts">
import { ref, computed, watch } from "vue"
import { profileClient, Profile, errorCode, Code, currentUser } from "@/api"
import { ProfileMask } from "@/api/schema"
import { useRouter } from "vue-router"
import { route } from "@/router"
import InternalError from "@/components/modals/InternalError.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"
import Input from "@/components/fields/InputField.vue"
import AvatarEditor from "./AvatarEditor.vue"

type Modal = "none" | "avatar_edit" | "error" | "duplicate_entry"

const emit = defineEmits<{
  (e: "update", input: Profile): void
  (e: "avatar", full: string, thumbnail: string): void
}>()

const props = defineProps<{
  profile: Profile
  avatar: string
}>()

const router = useRouter()

const modal = ref<Modal>("none")
const handle = ref(props.profile.handle)
const displayName = ref<string>(props.profile.displayName || "")
const update = ref<ProfileMask>({})
const saving = ref<boolean>(false)
const uploadedAvatar = ref<string>("")

const handleError = handleValidator.reactive(handle)
const displayNameError = displayNameValidator.reactive(displayName)
const hasUpdate = computed(() => {
  if (!update.value) {
    return 0
  }
  return Object.keys(update.value).length !== 0
})
const formValid = computed(() => {
  return !displayNameError.value && !handleError.value
})

watch(handle, (value) => {
  if (value === props.profile.handle) {
    delete update.value.handle
  } else {
    update.value.handle = value
  }
})
watch(displayName, (value) => {
  if (value === props.profile.displayName) {
    delete update.value.displayName
  } else {
    update.value.displayName = value
  }
})
watch(
  () => props.profile,
  (profile) => {
    if (profile.ownerId !== currentUser.id) {
      router.replace(route.profile(profile.handle, "overview"))
    }
  },
  { immediate: true },
)

async function save() {
  if (saving.value || !hasUpdate.value || !formValid.value) {
    return
  }
  saving.value = true
  try {
    const currentUpdate = Object.assign({}, update.value)
    const updated = await profileClient.update(currentUpdate)
    update.value = {}
    emit("update", updated)
  } catch (e) {
    switch (errorCode(e)) {
      case Code.DuplicateEntry:
        modal.value = "duplicate_entry"
        break
      default:
        modal.value = "error"
        break
    }
  } finally {
    saving.value = false
  }
}

async function editAvatar() {
  const file = await new Promise<File>((resolve) => {
    const input = document.createElement("input") as HTMLInputElement
    input.type = "file"
    input.onchange = () => {
      if (!input.files) {
        throw new Error("Failed to parse files.")
      }
      resolve(input.files[0])
    }
    input.click()
  })

  uploadedAvatar.value = await new Promise<string>((resolve) => {
    const reader = new FileReader()
    reader.onloadend = () => {
      resolve(reader.result as string)
    }
    reader.readAsDataURL(file)
  })
  modal.value = "avatar_edit"
}

async function uploadAvatar(full: string, thumbnail: string) {
  saving.value = true
  await profileClient.uploadAvatar(full)
  emit("avatar", full, thumbnail)
  modal.value = "none"
  saving.value = false
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.form
  margin: 30px
  margin-left: 100px
  display: table

.form-row
  display: table-row

.form-cell
  display: table-cell
  padding: 10px
  vertical-align: middle
  &.align-top
    vertical-align: top

.label
  text-align: right
  padding-right: 30px

.avatar
  position: relative
  display: flex
  margin-bottom: 20px
  width: 200px

.avatar-img
  @include theme.shadow-l

  border-radius: 50%
  width: 100%
  height: 100%

.avatar-edit-icon
  cursor: pointer
  border-radius: 10px
  padding: 10px
  background: #222
  color: white
  position: absolute
  right: 0
  bottom: 10%
  &:hover
    background: #2A2A2A

.form-input
  width: 400px

.controls
  display: flex
  flex-direction: row
  justify-content: flex-end

.save-loader
  height: 20px
  width: 100%

.save
  width: 100px
  text-align: center
</style>
