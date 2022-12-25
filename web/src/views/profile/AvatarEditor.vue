<template>
  <ModalDialog
    :is-visible="props.isVisible"
    :buttons="{ upload: 'Upload', cancel: 'Cancel' }"
    :disabled="loading"
    @click="onClick"
  >
    <div class="avatar-box">
      <img class="avatar" :src="avatar" />
      <div v-if="loading" class="loader-box">
        <PageLoader></PageLoader>
      </div>
    </div>
  </ModalDialog>
</template>

<script setup lang="ts">
import ModalDialog from "@/components/modals/ModalDialog.vue"
import PageLoader from "@/components/PageLoader.vue"

const fullSize = 460
const thumbnailSize = 128

const emit = defineEmits<{
  (e: "close"): void
  (e: "update", full: string, thumbnail: string): void
}>()

const props = defineProps<{
  isVisible: boolean
  avatar: string
  loading?: boolean
}>()

async function onClick(btn: string) {
  switch (btn) {
    case "upload":
      emit(
        "update",
        await fill(props.avatar, fullSize, fullSize),
        await fill(props.avatar, thumbnailSize, thumbnailSize),
      )
      break
    case "cancel":
      emit("close")
      break
  }
}
</script>

<script lang="ts">
async function fill(imageURL: string, width: number, height: number): Promise<string> {
  const canvas = document.createElement("canvas")
  const ctx = canvas.getContext("2d")
  if (!ctx) {
    throw new Error("Failed to create rendering context.")
  }

  const img = new Image()
  await new Promise<void>((resolve) => {
    img.onload = () => {
      resolve()
    }
    img.src = imageURL
  })

  const sourceWidth = img.naturalWidth < img.naturalHeight ? img.naturalWidth : img.naturalHeight
  const sourceHeight = img.naturalHeight < img.naturalWidth ? img.naturalHeight : img.naturalWidth
  const sourceX = (img.naturalWidth - sourceWidth) / 2
  const sourceY = (img.naturalHeight - sourceHeight) / 2

  canvas.width = width
  canvas.height = height
  ctx.drawImage(img, sourceX, sourceY, sourceWidth, sourceHeight, 0, 0, width, height)
  return canvas.toDataURL("image/jpeg")
}
</script>

<style scoped lang="sass">
.avatar-box
  position: relative
  display: flex
  align-items: center
  justify-content: center
  margin: 10px

.avatar
  border-radius: 50%
  width: 250px
  height: 250px
  object-fit: cover

.loader-box
  border-radius: 15px
  position: absolute
  padding: 10px
  background: #000B
</style>
