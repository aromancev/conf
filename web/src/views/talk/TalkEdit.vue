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
          :error="handleError"
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
          :error="titleError"
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

  <ModalDialog v-if="modal === Modal.DuplicateEntry" :buttons="{ ok: 'OK' }" @click="modal = Modal.None">
    <p>Talk with this handle already exits.</p>
    <p>Try a different handle.</p>
  </ModalDialog>
  <ModalDialog v-if="modal === Modal.NotFound" :buttons="{ ok: 'OK' }" @click="modal = Modal.None">
    <p>Talk no longer exits.</p>
    <p>Maybe someone has changed the handle or archived it.</p>
  </ModalDialog>
  <InternalError v-if="modal === Modal.Error" @click="modal = Modal.None" />
</template>

<script lang="ts">
import { RegexValidator } from "@/platform/validator"

const handleValidator = new RegexValidator("^[a-z0-9-]{4,64}$", [
  "Must be from 4 to 64 characters long",
  "Can only contain lower case letters, numbers, and '-'",
])
const titleValidator = new RegexValidator("^[a-zA-Z0-9- ]{0,64}$", [
  "Must be from 0 to 64 characters long",
  "Can only contain letters, numbers, spaces, and '-'",
])
</script>

<script setup lang="ts">
import { ref, computed, watch } from "vue"
import { talkClient, Talk, TalkMask, errorCode, Code, currentUser } from "@/api"
import { useRouter } from "vue-router"
import InternalError from "@/components/modals/InternalError.vue"
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"
import Input from "@/components/fields/InputField.vue"
import Textarea from "@/components/fields/TextareaField.vue"

enum Modal {
  None = "",
  Error = "error",
  DuplicateEntry = "duplicate_entry",
  NotFound = "not_found",
}

const emit = defineEmits<{
  (e: "update", input: Talk): void
}>()

const props = defineProps<{
  talk: Talk
}>()

const router = useRouter()

const modal = ref(Modal.None)
const handle = ref(props.talk.handle)
const title = ref<string>(props.talk.title || "")
const description = ref(props.talk.description || "")
const update = ref<TalkMask>({})
const saving = ref(false)

const handleError = handleValidator.reactive(handle)
const titleError = titleValidator.reactive(title)
const hasUpdate = computed(() => {
  if (!update.value) {
    return 0
  }
  return Object.keys(update.value).length !== 0
})
const formValid = computed(() => {
  return !titleError.value && !handleError.value
})

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
    if (value.ownerId !== currentUser.id) {
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
        modal.value = Modal.DuplicateEntry
        break
      case Code.NotFound:
        modal.value = Modal.NotFound
        break
      default:
        modal.value = Modal.Error
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
