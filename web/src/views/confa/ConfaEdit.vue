<template>
  <div class="form">
    <Input v-model="state.handle" :spellcheck="false" class="input" type="text" label="Handle" :errors="handleErrors" />
    <Input v-model="state.title" :spellcheck="false" class="input" type="text" label="Title" :errors="titleErrors" />
    <Textarea v-model="state.description" class="input description" label="Description"></Textarea>
    <div class="controls">
      <div class="save-indicator"></div>
      <div class="btn save" :disabled="!hasUpdate || state.isSaving || !formValid ? true : null" @click="save">
        <div v-if="state.isSaving" class="save-loader">
          <PageLoader />
        </div>

        <span v-if="!state.isSaving">{{ !hasUpdate ? "Saved" : "Save" }}</span>
      </div>
    </div>
  </div>

  <ModalDialog :is-visible="state.modal === 'duplicate_entry'" :buttons="{ ok: 'OK' }" @click="state.modal = 'none'">
    <p>Confa with this handle already exits.</p>
    <p>Try a different handle.</p>
  </ModalDialog>
  <ModalDialog :is-visible="state.modal === 'not_found'" :buttons="{ ok: 'OK' }" @click="state.modal = 'none'">
    <p>Confa no longer exits.</p>
    <p>Maybe someone has changed the handle or archived it.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { computed, watch, reactive } from "vue"
import { api, errorCode, Code } from "@/api"
import { ConfaClient } from "@/api/confa"
import { Confa } from "@/api/models/confa"
import { accessStore } from "@/api/models/access"
import { ConfaUpdate } from "@/api/schema"
import { useRouter } from "vue-router"
import { route } from "@/router"
import { titleValidator, handleValidator } from "@/api/models/confa"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"
import Input from "@/components/fields/InputField.vue"
import Textarea from "@/components/fields/TextareaField.vue"
import { notificationStore } from "@/api/models/notifications"

type Modal = "none" | "duplicate_entry" | "not_found"

const emit = defineEmits<{
  (e: "update", input: Confa): void
}>()

const props = defineProps<{
  confa: Confa
}>()

type State = {
  handle: string
  title: string
  description: string
  update: ConfaUpdate
  isSaving: boolean
  isSubmitted: boolean
  modal: Modal
}

const state = reactive<State>({
  handle: props.confa.handle,
  title: props.confa.title,
  description: props.confa.description,
  update: {},
  isSaving: false,
  isSubmitted: false,
  modal: "none",
})

const router = useRouter()

const handleErrors = computed<string[]>(() => {
  return handleValidator.validate(state.handle)
})
const titleErrors = computed<string[]>(() => {
  if (!state.isSubmitted) {
    return []
  }
  return titleValidator.validate(state.title)
})
const hasUpdate = computed(() => {
  if (!state.update) {
    return 0
  }
  return Object.keys(state.update).length !== 0
})
const formValid = computed(() => {
  return !titleErrors.value.length && !handleErrors.value.length
})

watch([() => state.handle, () => state.title, () => state.description], () => {
  state.isSubmitted = false
})
watch(
  () => props.confa,
  (c) => {
    state.title = c.title
    state.handle = c.handle
    state.description = c.description
  },
  {
    deep: true,
  },
)
watch(
  () => state.handle,
  (value) => {
    if (value === props.confa.handle) {
      delete state.update.handle
    } else {
      state.update.handle = value
    }
  },
)
watch(
  () => state.title,
  (value) => {
    if (value === props.confa.title) {
      delete state.update.title
    } else {
      state.update.title = value
    }
  },
)
watch(
  () => state.description,
  (value) => {
    if (value === props.confa.description) {
      delete state.update.description
    } else {
      state.update.description = value
    }
  },
)
watch(
  () => props.confa,
  (confa) => {
    if (confa.ownerId !== accessStore.state.id) {
      router.replace(route.confa(confa.handle, "overview"))
    }
  },
  { immediate: true },
)

async function save() {
  state.isSubmitted = true
  if (state.isSaving || !hasUpdate.value || !formValid.value) {
    return
  }
  state.isSaving = true
  try {
    const currentUpdate = Object.assign({}, state.update)
    const updated = await new ConfaClient(api).update({ id: props.confa.id }, currentUpdate)
    state.update = {}
    emit("update", updated)
  } catch (e) {
    switch (errorCode(e)) {
      case Code.DuplicateEntry:
        state.modal = "duplicate_entry"
        break
      case Code.NotFound:
        state.modal = "not_found"
        break
      default:
        notificationStore.error("failed to update conference")
        break
    }
  } finally {
    state.isSaving = false
  }
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
  text-align: right
  width: 100%
  max-width: theme.$form-width
  margin: 5px 0

.save-loader
  height: 20px
  width: 100%

.save
  width: 100px
  text-align: center
</style>
