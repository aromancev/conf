import { Peer, Status } from "./aggregators/peers"
import { Sprite, Actor, Sound } from "./actor"
import { Semaphor } from "@/platform/sync"
import reactionsSpriteURL from "/static/room/reactions.webp"
import reactionsSoundURL from "/static/room/reactions.webm"

const COMPACTION = 0.3
const PADDING = 0.25
const COLOR_OUTLINE = "#7f70f5"
const MAX_SIZE = 128
const BASIC_BORDER = 2
const SELECTED_BORDER = 4
const SELECTED_SCALE = 1.1
const STATUS_SCALE = 0.8

type StatusKey = "clap"

interface RenderPeer {
  userId: string
  row: number
  col: number
  x: number
  y: number
  image?: ImageBitmap
  imageSrc: string
  status?: Map<StatusKey, Actor>
}

interface Context {
  audience: CanvasRenderingContext2D
  selection: CanvasRenderingContext2D
  statuses: CanvasRenderingContext2D
}

class ActorBuilder {
  private readonly sounds: Sound
  private readonly sprites: Sprite
  private readonly soundSemaphor: Semaphor

  constructor() {
    this.soundSemaphor = new Semaphor(40)
    this.sounds = new Sound(reactionsSoundURL, {
      clapV1: [
        [0, 200],
        [300, 200],
        [600, 200],
        [850, 200],
        [1150, 200],
      ],
      clapV2: [
        [1350, 200],
        [1600, 200],
        [1900, 200],
        [2150, 200],
        [2400, 200],
      ],
      clapV3: [
        [2650, 200],
        [2850, 200],
        [3100, 200],
        [3325, 200],
        [3550, 200],
      ],
      clapV4: [
        [3875, 150],
        [4050, 200],
        [4275, 200],
        [4475, 200],
        [4725, 200],
      ],
    })
    this.sprites = new Sprite(reactionsSpriteURL, 12, {
      clap: [0, 12],
    })
  }

  close() {
    this.sounds.close()
    this.sprites.close()
  }

  newClap(): Actor {
    const sounds = ["clapV1", "clapV2", "clapV3", "clapV4"]
    const randomSound = sounds[Math.floor(Math.random() * sounds.length)]
    return new Actor({
      periodMs: 250 + Math.random() * 100,
      render: {
        sprite: this.sprites,
        spriteKey: "clap",
      },
      play: {
        sound: this.sounds,
        soundKey: randomSound,
        delayMs: 250,
        semaphor: this.soundSemaphor,
      },
    })
  }
}

export interface RendererArgs {
  userId: string
  peers: Map<string, Peer>
  statuses: Map<string, Status>
  context: Context
  width: number
  height: number
  selfReactions: boolean
}

export class Renderer {
  private readonly context: Context

  private readonly userId: string
  private readonly peers: Map<string, Peer>
  private readonly statuses: Map<string, Status>
  private readonly actors: ActorBuilder
  private readonly selfReactions: boolean
  private canvasPeers: WeakMap<Peer, RenderPeer>
  private ordered: RenderPeer[]

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

  private isPlaying: boolean

  constructor(args: RendererArgs) {
    this.context = args.context

    this.userId = args.userId
    this.peers = args.peers
    this.statuses = args.statuses
    this.selfReactions = args.selfReactions
    this.ordered = []
    this.canvasPeers = new WeakMap<Peer, RenderPeer>()
    this.actors = new ActorBuilder()

    this.height = args.height
    this.width = args.width

    this.rows = 0
    this.columns = 0
    this.padding = 0
    this.chess = false
    this.cellSize = 0
    this.renderSize = 0
    this.shift = 0
    this.offsetY = 0

    this.isPlaying = false
  }

  resize(width: number, height: number): void {
    this.width = width
    this.height = height
    this.calcPositions()
    this.renderAudience()
    this.renderStatuses()
    this.renderSelection()
  }

  animate(): void {
    this.renderStatuses()
  }

  play(): void {
    this.isPlaying = true
    this.ordered.forEach((p) => {
      p.status?.forEach((a) => a.play())
    })
  }

  pause(): void {
    this.isPlaying = false
    this.ordered.forEach((p) => {
      p.status?.forEach((a) => a.pause())
    })
  }

  close(): void {
    this.isPlaying = false
    this.ordered.forEach((p) => {
      p.status?.forEach((a) => a.close())
    })
    this.actors.close()
  }

  async updatePeers(): Promise<void> {
    this.ordered = []
    const images: Promise<void>[] = []
    this.peers.forEach((peer: Peer) => {
      // Create a new one if doesn't exist.
      let canvasPeer = this.canvasPeers.get(peer)
      if (!canvasPeer) {
        canvasPeer = {
          userId: peer.userId,
          row: 0,
          col: 0,
          x: 0,
          y: 0,
          imageSrc: "",
        }
        this.canvasPeers.set(peer, canvasPeer)
      }
      const cp = canvasPeer

      // Update the avatar if changed.
      if (cp.imageSrc !== peer.profile.avatar) {
        images.push(
          new Promise((res) => {
            // TODO: Move this processing to a separate worker thread.
            // This is the bottleneck for rendering 1k+ peers. Hangs the UI.
            fetch(peer.profile.avatar).then((resp) => {
              resp.blob().then((blob) => {
                createImageBitmap(blob).then((bitmap) => {
                  cp.image = bitmap
                  res()
                })
              })
            })
          }),
        )
        cp.imageSrc = peer.profile.avatar
      }

      this.ordered.push(cp)
    })
    this.calcPositions()
    await Promise.all(images)

    this.updateStatuses() // Have to update because statuses are not rendered without peers.
    this.renderAudience()
    this.renderSelection()
  }

  updateStatuses(): void {
    for (const peer of this.ordered) {
      if (!this.selfReactions && peer.userId === this.userId) {
        continue
      }

      const status = this.statuses.get(peer.userId)
      if (!status) {
        peer.status?.forEach((a) => a.close())
        peer.status = undefined
        continue
      }
      if (!peer.status) {
        peer.status = new Map()
      }

      if (status.clap) {
        let clap = peer.status.get("clap")
        if (!clap) {
          clap = this.actors.newClap()
          peer.status.set("clap", clap)
        }
        if (this.isPlaying) {
          clap.play()
        }
      } else {
        peer.status.get("clap")?.close()
        peer.status.delete("clap")
      }
    }
    this.renderStatuses()
  }

  hover(x: number, y: number): string | null {
    // Three possible rows because of compaction.
    const bottom = Math.floor((y - this.offsetY) / this.cellSize / COMPACTION)
    const center = bottom - 1
    const top = center - 1

    // Two possible columns because of chess-like shift.
    const shift = bottom % 2 === 0 ? this.shift : -this.shift
    const left = Math.floor((x - this.padding - shift) / this.cellSize)
    const right = Math.floor((x - this.padding + shift) / this.cellSize)

    // Four combinations in total.
    const candidates = [this.at(top, left), this.at(center, right), this.at(bottom, left)]

    let minDist = Infinity
    let closestPeer = null as RenderPeer | null
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
    const peer = this.peers.get(id)
    if (peer) {
      this.renderSelection(this.canvasPeers.get(peer))
    } else {
      this.renderSelection()
    }
  }

  private calcPositions() {
    if (this.ordered.length <= 0) {
      return
    }
    const height = this.height / COMPACTION
    const width = this.width

    // First size calculation round (approximating).
    let cellSize = Math.sqrt((height * width) / this.ordered.length)
    cellSize = Math.min(cellSize, MAX_SIZE) // Limiting the size of a cell.
    this.chess = Math.ceil(width / cellSize) < this.ordered.length
    this.columns = this.chess ? Math.ceil(width / cellSize) : this.ordered.length
    this.rows = Math.ceil(this.ordered.length / this.columns)

    // Second size calculation round (making sure all peers fit into the actual dimentions).
    cellSize = Math.min(cellSize, width, height) // Cell cannot be bigger that width or height.
    cellSize = Math.min(cellSize, (width - cellSize / 2) / this.columns) // Compensating for chess-like shift.
    const actualHeight = cellSize + (this.rows - 1) * cellSize * COMPACTION // Calculating how much height was actually taken.
    cellSize = Math.min(cellSize, (cellSize * this.height) / actualHeight) // Compensating for the actual height.
    this.cellSize = cellSize
    this.offsetY = Math.min(
      (this.cellSize / 2) * (SELECTED_SCALE - 1) + (this.cellSize * PADDING) / 2 + SELECTED_BORDER,
      (this.height - actualHeight) / 2,
    )
    this.padding = (this.width - cellSize * Math.min(this.columns, this.ordered.length)) / 2

    if (this.chess) {
      this.shift = this.cellSize * 0.25
      this.renderSize = this.cellSize * (1 - PADDING)
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
        peer.y *= COMPACTION // Compensate for compaction.
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

  private async renderStatuses(): Promise<void> {
    const ctx = this.context.statuses
    const renderSize = this.renderSize * STATUS_SCALE

    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, this.width, this.height)

    for (const peer of this.ordered) {
      const status = peer.status
      if (!status) {
        continue
      }

      ctx.save()
      const clap = status.get("clap")
      if (clap) {
        // Need to adjust position a bit because it looks misaligned.
        ctx.setTransform(
          1,
          0,
          0,
          1,
          peer.x - renderSize / 2 - renderSize * 0.05,
          peer.y - renderSize / 2 - renderSize * 0.05,
        )
        await clap.render(ctx, renderSize, renderSize)
      }
      ctx.restore()
    }
  }

  private renderSelection(selected?: RenderPeer) {
    const ctx = this.context.selection

    ctx.save()
    ctx.clearRect(0, 0, this.width, this.height)

    if (selected) {
      this.renderPeer(ctx, selected, SELECTED_BORDER, SELECTED_SCALE, 0.05)
    }

    ctx.restore()
  }

  private renderPeer(
    ctx: CanvasRenderingContext2D,
    peer: RenderPeer,
    border = BASIC_BORDER,
    scale = 1,
    shiftY = 0,
  ): void {
    if (!peer.image) {
      return
    }
    const renderSize = this.renderSize * scale
    const x = peer.x
    const y = peer.y - renderSize * shiftY
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
    ctx.strokeStyle = COLOR_OUTLINE
    ctx.lineWidth = border
    ctx.beginPath()
    ctx.arc(0, 0, renderSize / 2, 0, Math.PI * 2, true)
    ctx.stroke()
  }

  private at(row: number, col: number): RenderPeer | null {
    if (row < 0 || row >= this.rows || col < 0 || col >= this.columns) {
      return null
    }
    const i = row * this.columns + col
    if (i < 0 || i >= this.ordered.length) {
      return null
    }
    return this.ordered[i]
  }

  private topLeft(row: number, col: number): RenderPeer | null {
    return this.at(row - 1, col - (row % 2))
  }

  private topRight(row: number, col: number): RenderPeer | null {
    return this.at(row - 1, col + 1 - (row % 2))
  }

  private topMiddle(row: number, col: number): RenderPeer | null {
    return this.at(row - 2, col)
  }

  private bottomMiddle(row: number, col: number): RenderPeer | null {
    return this.at(row + 2, col)
  }

  private bottomLeft(row: number, col: number): RenderPeer | null {
    return this.at(row + 1, col - (row % 2))
  }

  private bottomRight(row: number, col: number): RenderPeer | null {
    return this.at(row + 1, col + 1 - (row % 2))
  }

  private distance(x1: number, y1: number, x2: number, y2: number): number {
    const dx = Math.pow(x2 - x1, 2)
    const dy = Math.pow(y2 - y1, 2)
    return Math.sqrt(dx + dy)
  }
}
