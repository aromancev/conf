<template>
  <div class="form">
    <div class="form-row">
      <div class="form-cell label">Handle</div>
      <div class="form-cell">
        <Input
          v-model="state.handle"
          :spellcheck="false"
          class="form-input"
          type="text"
          placeholder="handle"
          :errors="handleErrors"
        />
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell label">Title</div>
      <div class="form-cell">
        <Input
          v-model="state.title"
          :spellcheck="false"
          class="form-input"
          type="text"
          placeholder="title"
          :errors="titleErrors"
        />
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell label align-top">Description</div>
      <div class="form-cell">
        <Textarea v-model="state.description" class="form-input description" placeholder="description"></Textarea>
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell"></div>
      <div class="form-cell controls">
        <div class="save-indicator"></div>
        <div class="btn save" :disabled="!hasUpdate || state.isSaving || !formValid ? true : null" @click="save">
          <div v-if="state.isSaving" class="save-loader">
            <PageLoader />
          </div>

          <span v-if="!state.isSaving">{{ !hasUpdate ? "Saved" : "Save" }}</span>
        </div>
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
import { confaClient, errorCode, Code } from "@/api"
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
  modal: Modal
}

const state = reactive<State>({
  handle: props.confa.handle,
  title: props.confa.title,
  description: props.confa.description,
  update: {},
  isSaving: false,
  modal: "none",
})

const router = useRouter()

const handleErrors = computed<string[]>(() => handleValidator.validate(state.handle))
const titleErrors = computed<string[]>(() => titleValidator.validate(state.title))
const hasUpdate = computed(() => {
  if (!state.update) {
    return 0
  }
  return Object.keys(state.update).length !== 0
})
const formValid = computed(() => {
  return !titleErrors.value.length && !handleErrors.value.length
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
  if (state.isSaving || !hasUpdate.value || !formValid.value) {
    return
  }
  state.isSaving = true
  try {
    const currentUpdate = Object.assign({}, state.update)
    const updated = await confaClient.update({ id: props.confa.id }, currentUpdate)
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

.form-input
  width: 800px

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
