<template>
  <div class="form">
    <div class="form-row">
      <div class="form-cell label">Handle</div>
      <div class="form-cell">
        <Input
          v-model="handle"
          spellcheck="false"
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
          spellcheck="false"
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
        <Textarea
          v-model="description"
          class="form-input description"
          placeholder="description"
        />
      </div>
    </div>
    <div class="form-row">
      <div class="form-cell"></div>
      <div class="form-cell controls">
        <div class="save-indicator"></div>
        <div
          class="btn save"
          :disabled="!hasUpdate || saving || !formValid ? true : null"
          @click="save"
        >
          <div v-if="saving" class="save-loader">
            <Loader />
          </div>

          <span v-if="!saving">
            {{ !hasUpdate ? "Saved" : "Save" }}
          </span>
        </div>
      </div>
    </div>
  </div>

  <Modal
    v-if="modal === Dialog.DuplicateEntry"
    @click="modal = Dialog.None"
    :buttons="{ ok: 'OK' }"
  >
    <p>Confa with this handle already exits.</p>
    <p>Try a different handle.</p>
  </Modal>
  <InternalError v-if="modal === Dialog.Error" @click="modal = Dialog.None" />
</template>

<script lang="ts">
import InternalError from "@/components/modals/InternalError.vue"
import Modal from "@/components/modals/Modal.vue"
import Loader from "@/components/Loader.vue"
import Input from "@/components/fields/Input.vue"
import Textarea from "@/components/fields/Textarea.vue"
import { defineComponent, PropType } from "vue"
import { confa, Confa, ConfaInput, errorCode, Code } from "@/api"

enum Dialog {
  None = "",
  Error = "error",
  DuplicateEntry = "duplicate_entry",
}

const defaultInputHeight = "1px"
const validHandle = new RegExp("^[a-z0-9-]{4,64}$")
const handleError = `• Must be from 4 to 64 characters long.
• Can only contain lower case letters, numbers, and '-'.
`
const validTitle = new RegExp("^[a-zA-Z0-9- ]{0,64}$")
const titleError = `• Must be from 0 to 64 characters long.
• Can only contain letters, numbers, spaces, and '-'.
`

export default defineComponent({
  name: "ConfaEdit",
  components: {
    InternalError,
    Modal,
    Loader,
    Input,
    Textarea,
  },
  emits: ["updated"],
  props: {
    confa: {
      type: Object as PropType<Confa>,
      required: true,
    },
  },
  data() {
    return {
      Dialog,
      modal: Dialog.None,
      handle: this.confa.handle,
      handleError: "",
      title: this.confa.title,
      titleError: "",
      description: this.confa.description,
      update: {} as ConfaInput,
      saving: false,
    }
  },
  computed: {
    hasUpdate(): boolean {
      return Object.keys(this.update).length !== 0
    },
    formValid(): boolean {
      return !this.titleError && !this.handleError
    },
  },
  watch: {
    handle(val: string) {
      if (!validHandle.test(val)) {
        this.handleError = handleError
        return
      }
      this.handleError = ""

      if (val === this.confa.handle) {
        delete this.update["handle"]
      } else {
        this.update.handle = val
      }
    },
    title(val: string) {
      if (!validTitle.test(val)) {
        this.titleError = titleError
        return
      }
      this.titleError = ""

      if (val === this.confa.title) {
        delete this.update["title"]
      } else {
        this.update.title = val
      }
    },
    description(val: string) {
      if (val === this.confa.description) {
        delete this.update["description"]
      } else {
        this.update.description = val
      }
    },
  },
  methods: {
    async save() {
      if (this.saving || !this.hasUpdate || !this.formValid) {
        return
      }
      const update = Object.assign({}, this.update)
      this.saving = true
      try {
        const updated = await confa.update({ id: this.confa.id }, update)
        this.update = {}
        if (updated === 0) {
          throw new Error("Nothing updated.")
        }
        this.$emit("updated", update)
        if (update.handle) {
          // Silently replace url without triggering re-render.
          const route = this.$router.resolve({
            name: "confaEdit",
            params: { confa: update.handle },
          })
          window.history.replaceState({}, "", route.path)
        }
      } catch (e) {
        if (errorCode(e) === Code.DuplicateEntry) {
          this.modal = Dialog.DuplicateEntry
        } else {
          this.modal = Dialog.Error
        }
      } finally {
        this.saving = false
      }
    },
  },
})
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

.description
  max-height: 500px
</style>
