<template>
  <div class="audience" @mousemove="select" @mouseleave="deselect">
    <div class="selected">{{ selected?.profile.name || "" }}</div>
    <div class="divider"></div>
    <div class="canvas">
      <canvas ref="audience" :style="{ display: isLoading ? 'none' : 'block' }"></canvas>
      <canvas ref="selection"></canvas>
      <router-link
        v-if="selected?.profile.handle"
        class="profile-link"
        :to="route.profile(selected.profile.handle, 'overview')"
        target="_blank"
      ></router-link>
      <PageLoader v-if="isLoading"></PageLoader>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from "vue"
import { Peer } from "./aggregators/peers"
import { route } from "@/router"
import PageLoader from "@/components/PageLoader.vue"

const props = defineProps<{
  isLoading?: boolean
  peers: Map<string, Peer>
}>()

const audience = ref<HTMLCanvasElement>()
const selection = ref<HTMLCanvasElement>()
const selected = ref(null as Peer | null)

let canvas = null as Canvas | null
let resizeInterval = 0

watch(
  () => props.peers,
  () => {
    canvas?.updatePeers()
  },
  { deep: true, immediate: true },
)

defineExpose({
  resize,
})

onMounted(() => {
  if (!audience.value || !selection.value) {
    console.error("not created")
    return
  }

  const audCtx = audience.value.getContext("2d")
  const selectionCtx = selection.value.getContext("2d")
  if (!audCtx || !selectionCtx) {
    throw new Error("Failed to get canvas context.")
  }
  canvas = new Canvas(
    props.peers,
    { audience: audCtx, selection: selectionCtx },
    audience.value.width,
    audience.value.height,
  )

  clearInterval(resizeInterval)
  resizeInterval = window.setInterval(resize, 1000)
  resize()
})

onUnmounted(() => {
  clearInterval(resizeInterval)
})

function resize() {
  if (document.fullscreenElement) {
    return
  }
  if (!audience.value || !selection.value) {
    return
  }

  const dpr = window.devicePixelRatio || 1
  const width = audience.value.offsetWidth * dpr
  const height = audience.value.offsetHeight * dpr
  if (audience.value.width === width && audience.value.height === height) {
    return
  }

  audience.value.width = width
  audience.value.height = height
  selection.value.width = width
  selection.value.height = height
  canvas?.resize(width, height)
}

function select(ev: MouseEvent) {
  if (!canvas) {
    return
  }
  const dpr = window.devicePixelRatio || 1
  const rect = (ev.target as HTMLElement).getBoundingClientRect()
  const userId = canvas.hover((ev.clientX - rect.left) * dpr, (ev.clientY - rect.top) * dpr)
  if (userId === (selected.value?.userId || null)) {
    return
  }
  if (userId) {
    selected.value = props.peers.get(userId) || null
    canvas.select(userId)
  } else {
    selected.value = null
    canvas.select("")
  }
}

function deselect() {
  if (!canvas) {
    return
  }
  canvas.select("")
}
</script>

<script lang="ts">
import { WaitGroup } from "@/platform/sync"

const compaction = 0.3
const padding = 0.25
const colorOutline = "#7f70f5"
const maxSize = 128
const basicBorder = 2
const selectedBorder = 4
const selectedScale = 1.1

interface CanvasPeer {
  userId: string
  row: number
  col: number
  x: number
  y: number
  image: HTMLImageElement
}

interface Context {
  audience: CanvasRenderingContext2D
  selection: CanvasRenderingContext2D
}

class Canvas {
  private context: Context

  private peers: Map<string, Peer>
  private cache: WeakMap<Peer, CanvasPeer>
  private ordered: CanvasPeer[]

  private width: number
  private height: number

  private rows: number
  private columns: number
  private padding: number
  private chess: boolean
  private renderSize: number
  private cellSize: number
  private shift: number
  private offsetY: number

  constructor(peers: Map<string, Peer>, context: Context, width: number, height: number) {
    this.context = context

    this.ordered = []
    this.cache = new WeakMap<Peer, CanvasPeer>()
    this.peers = peers

    this.height = height
    this.width = width

    this.rows = 0
    this.columns = 0
    this.padding = 0
    this.chess = false
    this.cellSize = 0
    this.renderSize = 0
    this.shift = 0
    this.offsetY = 0
  }

  resize(width: number, height: number): void {
    this.width = width
    this.height = height
    this.calcPositions()
    this.renderAudience()
    this.renderSelection()
  }

  async updatePeers(): Promise<void> {
    this.ordered = []

    const wg = new WaitGroup()
    this.peers.forEach((peer: Peer) => {
      // Create a new one if doesn't exist.
      let canvasPeer = this.cache.get(peer)
      if (!canvasPeer) {
        canvasPeer = {
          userId: peer.userId,
          row: 0,
          col: 0,
          x: 0,
          y: 0,
          image: new Image(),
        }
        this.cache.set(peer, canvasPeer)
      }

      // Update the avatar if changed.
      if (canvasPeer.image.src !== peer.profile.avatar) {
        wg.add(1)
        canvasPeer.image.onload = () => {
          wg.done()
        }
        canvasPeer.image.src = peer.profile.avatar
      }

      this.ordered.push(canvasPeer)
    })

    await wg.join()
    this.calcPositions()
    this.renderAudience()
    this.renderSelection()
  }

  hover(x: number, y: number): string | null {
    // Three possible rows because of compaction.
    const bottom = Math.floor((y - this.offsetY) / this.cellSize / compaction)
    const center = bottom - 1
    const top = center - 1

    // Two possible columns because of chess-like shift.
    const shift = bottom % 2 === 0 ? this.shift : -this.shift
    const left = Math.floor((x - this.padding - shift) / this.cellSize)
    const right = Math.floor((x - this.padding + shift) / this.cellSize)

    // Four combinations in total.
    const candidates = [this.at(top, left), this.at(center, right), this.at(bottom, left)]

    let minDist = Infinity
    let closestPeer = null as CanvasPeer | null
    for (const p of candidates) {
      if (!p) {
        continue
      }
      const dist = this.distance(p.x, p.y, x, y)
      if (dist < this.renderSize / 2 && dist < minDist) {
        closestPeer = p
        minDist = dist
      }
    }
    return closestPeer?.userId || null
  }

  select(id: string) {
    const peer = this.peers.get(id) || null
    if (peer) {
      this.renderSelection(this.cache.get(peer))
    } else {
      this.renderSelection()
    }
  }

  private calcPositions() {
    if (this.ordered.length <= 0) {
      return
    }
    const height = this.height / compaction
    const width = this.width

    // First size calculation round (approximating).
    let cellSize = Math.sqrt((height * width) / this.ordered.length)
    cellSize = Math.min(cellSize, maxSize) // Limiting the size of a cell.
    this.chess = Math.ceil(width / cellSize) < this.ordered.length
    this.columns = this.chess ? Math.ceil(width / cellSize) : this.ordered.length
    this.rows = Math.ceil(this.ordered.length / this.columns)

    // Second size calculation round (making sure all peers fit into the actual dimentions).
    cellSize = Math.min(cellSize, width, height) // Cell cannot be bigger that width or height.
    cellSize = Math.min(cellSize, (width - cellSize / 2) / this.columns) // Compensating for chess-like shift.
    const actualHeight = cellSize + (this.rows - 1) * cellSize * compaction // Calculating how much height was actually taken.
    cellSize = Math.min(cellSize, (cellSize * this.height) / actualHeight) // Compensating for the actual height.
    this.cellSize = cellSize
    this.offsetY = Math.min(
      (this.cellSize / 2) * (selectedScale - 1) + (this.cellSize * padding) / 2 + selectedBorder,
      (this.height - actualHeight) / 2,
    )
    this.padding = (this.width - cellSize * Math.min(this.columns, this.ordered.length)) / 2

    if (this.chess) {
      this.shift = this.cellSize * 0.25
      this.renderSize = this.cellSize * (1 - padding)
    } else {
      this.shift = 0
      this.renderSize = this.cellSize * 0.95
    }

    // Calculating coordinates for each peer.
    let index = 0
    for (let row = 0; row < this.rows; row++) {
      const shift = row % 2 === 0 ? this.shift : -this.shift
      for (let col = 0; col < this.columns; col++) {
        if (index >= this.ordered.length) {
          return
        }

        const peer = this.ordered[index]

        peer.row = row
        peer.col = col

        peer.x = col * this.cellSize // Base shift.
        peer.x += this.cellSize / 2 // Shift to the center of the cell.
        peer.x += this.padding // Compensate for outer padding.
        peer.x += shift // Compensate for chess-like shift.

        peer.y = row * this.cellSize // Base shift
        peer.y *= compaction // Compensate for compaction.
        peer.y += this.cellSize / 2 // Shift to the center of the cell.
        peer.y += this.offsetY
        index++
      }
    }
  }

  private renderAudience() {
    const ctx = this.context.audience

    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, this.width, this.height)

    for (const peer of this.ordered) {
      ctx.save()
      this.renderPeer(ctx, peer)
      ctx.restore()
    }
  }

  private renderSelection(selected?: CanvasPeer) {
    const ctx = this.context.selection

    ctx.save()
    ctx.clearRect(0, 0, this.width, this.height)

    if (selected) {
      this.renderPeer(ctx, selected, selectedBorder, selectedScale, 0.05)
    }

    ctx.restore()
  }

  private renderPeer(
    ctx: CanvasRenderingContext2D,
    peer: CanvasPeer,
    border = basicBorder,
    scale = 1,
    shift = 0,
  ): void {
    const renderSize = this.renderSize * scale
    const x = peer.x
    const y = peer.y - renderSize * shift
    // Clip overlapping peers.
    const overlapping = [
      this.bottomLeft(peer.row, peer.col),
      this.bottomMiddle(peer.row, peer.col),
      this.bottomRight(peer.row, peer.col),
    ]
    for (const p of overlapping) {
      if (!p) {
        continue
      }
      ctx.beginPath()
      ctx.rect(0, 0, this.width, this.height)
      ctx.arc(p.x, p.y, this.renderSize / 2, 0, Math.PI * 2, true)
      ctx.closePath()
      ctx.clip("evenodd")
    }

    // Clip outer circle boundary.
    ctx.save() // Saving to cancel clipping for the outline.
    ctx.beginPath()
    ctx.arc(x, y, renderSize / 2, 0, Math.PI * 2, true)
    ctx.closePath()
    ctx.clip("nonzero")

    // Icon.
    ctx.setTransform(1, 0, 0, 1, x - renderSize / 2, y - renderSize / 2)
    ctx.drawImage(peer.image, 0, 0, renderSize + 1, renderSize + 1)
    ctx.restore() // This will cancel clip and result in a smooth outline.

    // Outline.
    ctx.setTransform(1, 0, 0, 1, x, y)
    ctx.strokeStyle = colorOutline
    ctx.lineWidth = border
    ctx.beginPath()
    ctx.arc(0, 0, renderSize / 2, 0, Math.PI * 2, true)
    ctx.stroke()
  }

  private at(row: number, col: number): CanvasPeer | null {
    if (row < 0 || row >= this.rows || col < 0 || col >= this.columns) {
      return null
    }
    const i = row * this.columns + col
    if (i < 0 || i >= this.ordered.length) {
      return null
    }
    return this.ordered[i]
  }

  private topLeft(row: number, col: number): CanvasPeer | null {
    return this.at(row - 1, col - (row % 2))
  }

  private topRight(row: number, col: number): CanvasPeer | null {
    return this.at(row - 1, col + 1 - (row % 2))
  }

  private topMiddle(row: number, col: number): CanvasPeer | null {
    return this.at(row - 2, col)
  }

  private bottomMiddle(row: number, col: number): CanvasPeer | null {
    return this.at(row + 2, col)
  }

  private bottomLeft(row: number, col: number): CanvasPeer | null {
    return this.at(row + 1, col - (row % 2))
  }

  private bottomRight(row: number, col: number): CanvasPeer | null {
    return this.at(row + 1, col + 1 - (row % 2))
  }

  private distance(x1: number, y1: number, x2: number, y2: number): number {
    const dx = Math.pow(x2 - x1, 2)
    const dy = Math.pow(y2 - y1, 2)
    return Math.sqrt(dx + dy)
  }
}
</script>

<style scoped lang="sass">
.audience
  display: flex
  flex-direction: column
  background-color: transparent
  overflow: hidden

.selected
  margin: 10px
  height: 1em
  text-align: center

.divider
  height: 1px
  background: linear-gradient(to right, transparent 0, var(--color-highlight-background) 50%, transparent)

.canvas
  position: relative
  flex: 1
  display: flex
  justify-content: center

canvas
  position: absolute
  top: 0
  left: 0
  width: 100%
  height: 100%
  cursor: default

.loader
  height: 100%
  z-index: 100

.profile-link
  position: absolute
  top: 0
  left: 0
  width: 100%
  height: 100%
</style>
