<template>
  <div
    class="audience"
    v-on:mousemove="select"
    v-on:onmouseleave="deselect"
    v-bind:style="{ cursor: cursor }"
  >
    <canvas ref="audience"></canvas>
    <canvas class="shade" ref="shade"></canvas>
    <div v-if="selected" class="badge">{{ selected.name }}</div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue"
import { EventType, PeerStatus, PayloadPeerStatus } from "@/api/models"
import { drawIcon } from "jdenticon"
import { Record } from "./record"
import { nameFromUUID, identiconConfig } from "./gen"

const compaction = 0.3
const padding = 0.25
const colorOutline = "#7f70f5"

interface Peer {
  id: string
  joinedAt: string
  row: number
  col: number
  x: number
  y: number
  dirty: boolean
  name: string
}

class Canvas {
  private audicence: CanvasRenderingContext2D
  private shade: CanvasRenderingContext2D

  private byId: { [key: string]: Peer }
  private ordered: Peer[]
  private selected: Peer | null

  private width: number
  private height: number

  private rows: number
  private columns: number
  private padding: number
  private chess: boolean
  private renderSize: number
  private cellSize: number
  private shift: number

  private allDirty: boolean

  constructor(
    audicence: CanvasRenderingContext2D,
    shade: CanvasRenderingContext2D,
    width: number,
    height: number,
  ) {
    this.audicence = audicence
    this.shade = shade

    this.byId = {}
    this.ordered = []
    this.selected = null

    this.height = height
    this.width = width

    this.rows = 0
    this.columns = 0
    this.padding = 0
    this.chess = false
    this.cellSize = 0
    this.renderSize = 0
    this.shift = 0

    this.allDirty = true
  }

  processRecords(records: Record[]): void {
    for (const record of records) {
      if (record.event.payload.type !== EventType.PeerStatus) {
        continue
      }
      const payload = record.event.payload.payload as PayloadPeerStatus
      const userId = record.event.ownerId || ""
      if (
        (record.forward && payload.status === PeerStatus.Joined) ||
        (!record.forward && payload.status === PeerStatus.Left)
      ) {
        if (this.byId[userId]) {
          continue
        }
        const p: Peer = {
          id: userId,
          joinedAt: record.event.createdAt || "",
          row: 0,
          col: 0,
          x: 0,
          y: 0,
          dirty: true,
          name: nameFromUUID(userId),
        }
        this.byId[userId] = p
        this.ordered.push(p)
      }
      if (
        (record.forward && payload.status === PeerStatus.Left) ||
        (!record.forward && payload.status === PeerStatus.Joined)
      ) {
        this.allDirty = true
        delete this.byId[userId]
        for (let i = 0; i < this.ordered.length; i++) {
          if (this.ordered[i].id === userId) {
            this.ordered.splice(i, 1)
            break
          }
        }
      }
    }

    this.calcPositions()
    this.renderAudience()
    this.renderShade()
  }

  hover(x: number, y: number): Peer | null {
    // Three possible rows because of compaction.
    const bottom = Math.floor(y / this.cellSize / compaction)
    const center = bottom - 1
    const top = center - 1

    // Two possible columns because of chess-like shift.
    const shift = bottom % 2 === 0 ? this.shift : -this.shift
    const left = Math.floor((x - this.padding - shift) / this.cellSize)
    const right = Math.floor((x - this.padding + shift) / this.cellSize)

    // Four combinations in total.
    const candidates = [
      this.at(top, left),
      this.at(center, right),
      this.at(bottom, left),
    ]

    let minDist = Infinity
    let closestPeer = null as Peer | null
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
    return closestPeer
  }

  select(id: string) {
    this.selected = this.byId[id]
    this.renderShade()
  }

  private calcPositions() {
    if (this.ordered.length <= 0) {
      return
    }
    this.chess = this.ordered.length > 3
    const height = this.height / compaction
    const width = this.width
    // First size calculation round (approximating).
    let cellSize = Math.sqrt((height * width) / this.ordered.length)
    this.columns = this.chess
      ? Math.ceil(width / cellSize)
      : this.ordered.length
    this.rows = Math.ceil(this.ordered.length / this.columns)

    // Second size calculation round (making sure all peers fit into the actual dimentions).
    cellSize = Math.min(cellSize, width, height) // Cell cannot be bigger that width or height.
    cellSize = Math.min(cellSize, (width - cellSize / 2) / this.columns) // Compensating for chess-like shift.
    const actualHeight = cellSize + (this.rows - 1) * cellSize * compaction // Calculating how much height was actually taken.
    cellSize = Math.min(cellSize, (cellSize * this.height) / actualHeight) // Compensating for the actual height.

    if (cellSize !== this.cellSize) {
      this.allDirty = true
    }
    this.cellSize = cellSize

    this.padding =
      (this.width - cellSize * Math.min(this.columns, this.ordered.length)) / 2

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
        index++
      }
    }
  }

  private renderAudience() {
    const ctx = this.audicence

    ctx.setTransform(1, 0, 0, 1, 0, 0)
    if (this.allDirty) {
      ctx.clearRect(0, 0, this.width, this.height)
    }

    for (const peer of this.ordered) {
      if (!this.allDirty && !peer.dirty) {
        continue
      }
      peer.dirty = false

      ctx.save()

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
      ctx.beginPath()
      ctx.arc(peer.x, peer.y, this.renderSize / 2, 0, Math.PI * 2, true)
      ctx.closePath()
      ctx.clip("nonzero")

      // Icon.
      ctx.setTransform(
        1,
        0,
        0,
        1,
        peer.x - this.renderSize / 2,
        peer.y - this.renderSize / 2,
      )
      drawIcon(ctx, peer.id, this.renderSize, identiconConfig)

      // Outline.
      ctx.setTransform(1, 0, 0, 1, peer.x, peer.y)
      ctx.strokeStyle = colorOutline
      ctx.lineWidth = 3
      ctx.beginPath()
      ctx.arc(0, 0, this.renderSize / 2, 0, Math.PI * 2, true)
      ctx.stroke()

      ctx.restore()
    }
  }

  private renderShade() {
    const ctx = this.shade

    ctx.save()
    ctx.clearRect(0, 0, this.width, this.height)

    if (this.selected) {
      // Fill everything.
      ctx.fillStyle = "black"
      ctx.fillRect(0, 0, this.width, this.height)

      // Cut a circle whole.
      ctx.globalCompositeOperation = "destination-out"
      ctx.beginPath()
      ctx.arc(
        this.selected.x,
        this.selected.y,
        this.renderSize / 2,
        0,
        Math.PI * 2,
        true,
      )
      ctx.closePath()
      ctx.fill()

      // Draw neighbors on top.
      ctx.globalCompositeOperation = "source-over"
      const overlapping = [
        this.bottomLeft(this.selected.row, this.selected.col),
        this.bottomMiddle(this.selected.row, this.selected.col),
        this.bottomRight(this.selected.row, this.selected.col),
      ]
      for (const p of overlapping) {
        if (!p) {
          continue
        }
        ctx.beginPath()
        ctx.arc(p.x, p.y, this.renderSize / 2, 0, Math.PI * 2, true)
        ctx.closePath()
        ctx.fill()
      }
    }

    ctx.restore()
  }

  private at(row: number, col: number): Peer | null {
    if (row < 0 || row >= this.rows || col < 0 || col >= this.columns) {
      return null
    }
    const i = row * this.columns + col
    if (i < 0 || i >= this.ordered.length) {
      return null
    }
    return this.ordered[i]
  }

  private topLeft(row: number, col: number): Peer | null {
    return this.at(row - 1, col - (row % 2))
  }

  private topRight(row: number, col: number): Peer | null {
    return this.at(row - 1, col + 1 - (row % 2))
  }

  private topMiddle(row: number, col: number): Peer | null {
    return this.at(row - 2, col)
  }

  private bottomMiddle(row: number, col: number): Peer | null {
    return this.at(row + 2, col)
  }

  private bottomLeft(row: number, col: number): Peer | null {
    return this.at(row + 1, col - (row % 2))
  }

  private bottomRight(row: number, col: number): Peer | null {
    return this.at(row + 1, col + 1 - (row % 2))
  }

  private distance(x1: number, y1: number, x2: number, y2: number): number {
    const dx = Math.pow(x2 - x1, 2)
    const dy = Math.pow(y2 - y1, 2)
    return Math.sqrt(dx + dy)
  }
}

export default defineComponent({
  name: "Audience",

  data() {
    return {
      canvas: null as Canvas | null,
      cursor: "default",
      selected: null as Peer | null,
    }
  },

  async mounted() {
    const aud = this.$refs.audience as HTMLCanvasElement
    aud.width = aud.offsetWidth
    aud.height = aud.offsetHeight
    const shade = this.$refs.shade as HTMLCanvasElement
    shade.width = shade.offsetWidth
    shade.height = shade.offsetHeight

    const audCtx = aud.getContext("2d")
    const shadeCtx = shade.getContext("2d")
    if (!audCtx || !shadeCtx) {
      throw new Error("Failed to get canvas context.")
    }

    this.canvas = new Canvas(audCtx, shadeCtx, aud.width, aud.height)
  },

  methods: {
    processRecords(records: Record[]) {
      this.canvas?.processRecords(records)
    },
    select(ev: MouseEvent) {
      if (!this.canvas) {
        return
      }
      const rect = (ev.target as HTMLElement).getBoundingClientRect()
      const peer = this.canvas.hover(
        ev.clientX - rect.left,
        ev.clientY - rect.top,
      )
      if (peer) {
        this.cursor = "pointer"
        this.selected = peer
        this.canvas.select(peer.id)
      } else {
        this.cursor = "default"
        this.selected = null
        this.canvas.select("")
      }
    },
    deselect() {
      this.cursor = "default"
      this.canvas?.select("")
    },
  },
})
</script>

<style scoped lang="sass">
@use '@/css/theme'

.audience
  @include theme.shadow-inset-s

  position: relative
  background-color: transparent
  border-radius: 2px
  overflow: hidden

.shade
  opacity: 0.7

.badge
  position: absolute
  width: 100%
  height: 2em
  bottom: 0
  font-weight: bold
  user-select: none
  color: #fefefe

canvas
  position: absolute
  top: 0
  left: 0
  width: 100%
  height: 100%
</style>
