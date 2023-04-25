<template>
  <div class="form">
    <Input v-model="handle" :spellcheck="false" class="input" type="text" label="Handle" :errors="handleErrors" />
    <Input v-model="title" :spellcheck="false" class="input" type="text" label="Title" :errors="titleErrors" />
    <Textarea v-model="description" class="input" label="Description"></Textarea>
    <div class="controls">
      <div class="btn delete" @click="modal = 'delete'">Delete talk</div>
      <div class="btn save" :disabled="!hasUpdate || saving || !formValid ? true : null" @click="save">
        <div v-if="saving" class="save-loader">
          <PageLoader />
        </div>

        <span v-if="!saving">Save</span>
      </div>
    </div>
  </div>

  <ModalDialog :is-visible="modal === 'duplicate_entry'" :buttons="[{ text: 'OK' }]" @click="modal = 'none'">
    <p>Talk with this handle already exits.</p>
    <p>Try a different handle.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal === 'not_found'" :buttons="[{ text: 'OK' }]" @click="modal = 'none'">
    <p>Talk no longer exits.</p>
    <p>Maybe someone has changed the handle or archived it.</p>
  </ModalDialog>
  <ModalDialog
    :is-visible="modal === 'delete'"
    :buttons="[
      {
        text: 'Cancel',
        click: () => {
          modal = 'none'
        },
      },
      { text: 'Delete', click: deleteTalk },
    ]"
  >
    <p>Are you sure you want to delete talk "{{ talk.title || "Untitled" }}"?</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue"
import { api, errorCode, Code } from "@/api"
import { TalkClient } from "@/api/talk"
import { accessStore } from "@/api/models/access"
import { Talk, titleValidator, handleValidator } from "@/api/models/talk"
import { TalkUpdate } from "@/api/schema"
import { useRouter } from "vue-router"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"
import Input from "@/components/fields/InputField.vue"
import Textarea from "@/components/fields/TextareaField.vue"
import { route } from "@/router"
import { notificationStore } from "@/api/models/notifications"

type Modal = "none" | "duplicate_entry" | "not_found" | "delete"

const emit = defineEmits<{
  (e: "update", input: Talk): void
}>()

const props = defineProps<{
  confaHandle: string
  talk: Talk
}>()

const router = useRouter()

const modal = ref<Modal>("none")
const handle = ref(props.talk.handle)
const title = ref<string>(props.talk.title)
const description = ref(props.talk.description || "")
const update = ref<TalkUpdate>({})
const saving = ref(false)
const isSubmitted = ref(false)
const handleErrors = computed<string[]>(() => {
  return handleValidator.validate(handle.value)
})
const titleErrors = computed<string[]>(() => {
  if (!isSubmitted.value) {
    return []
  }
  return titleValidator.validate(title.value)
})
const hasUpdate = computed(() => {
  if (!update.value) {
    return 0
  }
  return Object.keys(update.value).length !== 0
})
const formValid = computed(() => {
  return !titleErrors.value.length && !handleErrors.value.length
})

watch([() => handle, title, description], () => {
  isSubmitted.value = false
})
watch(
  () => props.talk,
  (c) => {
    title.value = c.title
    handle.value = c.handle
    description.value = c.description || ""
  },
  {
    deep: true,
  },
)
watch(handle, (value) => {
  if (value === props.talk.handle) {
    delete update.value.handle
  } else {
    update.value.handle = value
  }
})
watch(title, (value) => {
  if (value === props.talk.title) {
    delete update.value.title
  } else {
    update.value.title = value
  }
})
watch(description, (value) => {
  if (value === props.talk.description) {
    delete update.value.description
  } else {
    update.value.description = value
  }
})
watch(
  () => props.talk,
  (value) => {
    if (value.ownerId !== accessStore.state.id) {
      router.replace({ name: "talk.overview", params: { talk: props.talk.handle } })
    }
  },
  { immediate: true },
)

async function save() {
  if (!hasUpdate.value) {
    return
  }

  isSubmitted.value = true
  if (saving.value || !formValid.value) {
    return
  }
  saving.value = true
  try {
    const currentUpdate = Object.assign({}, update.value)
    const updated = await new TalkClient(api).update({ id: props.talk.id }, currentUpdate)
    update.value = {}
    emit("update", updated)
  } catch (e) {
    switch (errorCode(e)) {
      case Code.DuplicateEntry:
        modal.value = "duplicate_entry"
        break
      case Code.NotFound:
        modal.value = "not_found"
        break
      default:
        notificationStore.error("failed to update talk")
        break
    }
  } finally {
    saving.value = false
  }
}

async function deleteTalk() {
  modal.value = "none"
  try {
    await new TalkClient(api).delete({ id: props.talk.id })
  } catch {
    notificationStore.error("failed to delete talk")
  }
  router.push(route.confa(props.confaHandle, "overview"))
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
  margin: 5px 0
  width: 100%
  max-width: theme.$form-width

.controls
  display: flex
  width: 100%
  max-width: theme.$form-width
  margin: 10px 0

.save-loader
  height: 20px
  width: 100%

.save
  width: 100px
  margin-left: auto

.delete
  color: var(--color-red)
</style>
