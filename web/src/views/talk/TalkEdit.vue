<template>
  <div class="form">
    <div class="form-row">
      <div class="form-cell label">Handle</div>
      <div class="form-cell">
        <Input
          v-model="handle"
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
          v-model="title"
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
        <Textarea v-model="description" class="form-input description" placeholder="description"></Textarea>
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

  <ModalDialog :is-visible="modal === 'duplicate_entry'" :buttons="{ ok: 'OK' }" @click="modal = 'none'">
    <p>Talk with this handle already exits.</p>
    <p>Try a different handle.</p>
  </ModalDialog>
  <ModalDialog :is-visible="modal === 'not_found'" :buttons="{ ok: 'OK' }" @click="modal = 'none'">
    <p>Talk no longer exits.</p>
    <p>Maybe someone has changed the handle or archived it.</p>
  </ModalDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue"
import { talkClient, errorCode, Code } from "@/api"
import { accessStore } from "@/api/models/access"
import { Talk, titleValidator, handleValidator } from "@/api/models/talk"
import { TalkUpdate } from "@/api/schema"
import { useRouter } from "vue-router"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"
import Input from "@/components/fields/InputField.vue"
import Textarea from "@/components/fields/TextareaField.vue"
import { notificationStore } from "@/api/models/notifications"

type Modal = "none" | "duplicate_entry" | "not_found"

const emit = defineEmits<{
  (e: "update", input: Talk): void
}>()

const props = defineProps<{
  talk: Talk
}>()

const router = useRouter()

const modal = ref<Modal>("none")
const handle = ref(props.talk.handle)
const title = ref<string>(props.talk.title)
const description = ref(props.talk.description || "")
const update = ref<TalkUpdate>({})
const saving = ref(false)
const handleErrors = computed<string[]>(() => handleValidator.validate(handle.value))
const titleErrors = computed<string[]>(() => titleValidator.validate(title.value))
const hasUpdate = computed(() => {
  if (!update.value) {
    return 0
  }
  return Object.keys(update.value).length !== 0
})
const formValid = computed(() => {
  return !titleErrors.value.length && !handleErrors.value.length
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
  if (saving.value || !hasUpdate.value || !formValid.value) {
    return
  }
  saving.value = true
  try {
    const currentUpdate = Object.assign({}, update.value)
    const updated = await talkClient.update({ id: props.talk.id }, currentUpdate)
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
</script>

<style scoped lang="sass">
@use '@/css/theme'

.form
  margin: 30px
  margin-left: 100px
  max-width: theme.$content-width
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
