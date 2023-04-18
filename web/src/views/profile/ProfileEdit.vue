<template>
  <div class="form">
    <div class="avatar">
      <img class="avatar-img" :src="avatar" />
      <div class="avatar-edit-icon material-icons" @click="editAvatar">edit</div>
    </div>
    <InputField v-model="handle" :spellcheck="false" class="input" type="text" label="Handle" :errors="handleErrors" />
    <InputField
      v-model="givenName"
      :spellcheck="false"
      class="input"
      type="text"
      label="Given name"
      :errors="givenNameErrors"
    />
    <InputField
      v-model="familyName"
      :spellcheck="false"
      class="input"
      type="text"
      label="Family name"
      :errors="familyNameErrors"
    />
    <div class="controls">
      <div class="save-indicator"></div>
      <div class="btn save" :disabled="!hasUpdate || saving || !isFormValid ? true : null" @click="save">
        <div v-if="saving" class="save-loader">
          <PageLoader />
        </div>
        <span v-if="!saving">{{ !hasUpdate ? "Saved" : "Save" }}</span>
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
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue"
import { api, errorCode, Code } from "@/api"
import { ProfileClient } from "@/api/profile"
import { Profile, handleValidator, nameValidator } from "@/api/models/profile"
import { accessStore } from "@/api/models/access"
import { ProfileUpdate } from "@/api/schema"
import { useRouter } from "vue-router"
import { route } from "@/router"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"
import InputField from "@/components/fields/InputField.vue"
import AvatarEditor from "./AvatarEditor.vue"
import { notificationStore } from "@/api/models/notifications"

type Modal = "none" | "avatar_edit" | "duplicate_entry"

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
const givenName = ref<string>(props.profile.givenName || "")
const familyName = ref<string>(props.profile.familyName || "")
const update = ref<ProfileUpdate>({})
const saving = ref<boolean>(false)
const uploadedAvatar = ref<string>("")

const handleErrors = computed<string[]>(() => handleValidator.validate(handle.value))
const givenNameErrors = computed<string[]>(() => nameValidator.validate(givenName.value))
const familyNameErrors = computed<string[]>(() => nameValidator.validate(familyName.value))
const hasUpdate = computed(() => {
  if (!update.value) {
    return 0
  }
  return Object.keys(update.value).length !== 0
})
const isFormValid = computed(() => {
  return !givenNameErrors.value.length && !familyNameErrors.value.length && !handleErrors.value.length
})

watch(handle, (value) => {
  if (value === props.profile.handle) {
    delete update.value.handle
  } else {
    update.value.handle = value
  }
})
watch(givenName, (value) => {
  if (value === props.profile.givenName) {
    delete update.value.givenName
  } else {
    update.value.givenName = value.trim()
  }
})
watch(familyName, (value) => {
  if (value === props.profile.familyName) {
    delete update.value.familyName
  } else {
    update.value.familyName = value.trim()
  }
})
watch(
  () => props.profile,
  (profile) => {
    if (profile.ownerId !== accessStore.state.id) {
      router.replace(route.profile(profile.handle, "overview"))
    }
  },
  { immediate: true },
)

async function save() {
  if (saving.value || !hasUpdate.value || !isFormValid.value) {
    return
  }
  saving.value = true
  try {
    const currentUpdate = Object.assign({}, update.value)
    const updated = await new ProfileClient(api).update(currentUpdate)
    update.value = {}
    emit("update", updated)
  } catch (e) {
    switch (errorCode(e)) {
      case Code.DuplicateEntry:
        modal.value = "duplicate_entry"
        break
      default:
        notificationStore.error("failed to update profile")
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
  await new ProfileClient(api).uploadAvatar(full)
  emit("avatar", full, thumbnail)
  modal.value = "none"
  saving.value = false
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.form
  padding: 30px
  display: flex
  flex-direction: column
  align-items: center
  width: 100%

.input
  width: 100%
  margin: 5px 0
  max-width: theme.$form-width

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

.controls
  text-align: right
  margin: 5px 0
  width: 100%
  max-width: theme.$form-width

.save-loader
  height: 20px
  width: 100%

.save
  width: 100px
  text-align: center
</style>
