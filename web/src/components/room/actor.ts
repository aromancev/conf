import { sleep, repeat, Semaphor } from "@/platform/sync"

const SOUND_GAIN = 0.5

export class Sprite {
  private readonly frames: number
  private readonly fragments: Record<string, number[]>
  private readonly sprite: Promise<ImageBitmap | undefined>

  constructor(source: string, frames: number, fragments: Record<string, [number, number]>) {
    this.frames = frames
    this.fragments = fragments
    this.sprite = new Promise((res) => {
      fetch(source)
        .then((resp) => {
          resp.blob().then((blob) => {
            createImageBitmap(blob).then((bitmap) => {
              res(bitmap)
            })
          })
        })
        .catch((e) => {
          console.error("Failed to fetch sprite from source", e)
          res(undefined)
        })
    })
  }

  async close(): Promise<void> {
    const sprite = await this.sprite
    if (!sprite) {
      return
    }
    sprite.close()
  }

  async render(
    ctx: CanvasRenderingContext2D,
    spriteKey: string,
    progress: number,
    width: number,
    height: number,
  ): Promise<void> {
    const sprite = await this.sprite
    if (!sprite) {
      return
    }
    const [startFrame, durFrames] = this.fragments[spriteKey]
    if (startFrame === undefined || durFrames === undefined) {
      throw new Error(`Sprite key doesn't exist: ${spriteKey}`)
    }
    const frameWidth = sprite.width / this.frames
    const frame = startFrame + Math.floor(progress * durFrames)
    ctx.drawImage(sprite, frameWidth * frame, 0, frameWidth, sprite.height, 0, 0, width, height)
  }
}

export class Sound {
  private readonly fragments: Record<string, number[][]>
  private context?: AudioContext
  private gain?: GainNode
  private buffer?: Promise<AudioBuffer>
  private sourcePayload: Promise<ArrayBuffer | undefined>

  constructor(source: string, fragments: Record<string, [number, number][]>) {
    this.fragments = fragments
    this.sourcePayload = new Promise((res) => {
      fetch(source)
        .then((resp) => {
          resp.arrayBuffer().then((buff) => {
            res(buff)
          })
        })
        .catch((e) => {
          console.error("Failed to fetch sound from source", e)
          res(undefined)
        })
    })
  }

  async close(): Promise<void> {
    await this.context?.close()
  }

  async play(sig: AbortSignal, fragmentKey: string, delayMs: number): Promise<void> {
    sig.throwIfAborted()

    const src = await this.audioSource()
    if (!src) {
      return
    }

    const [startMs, durationMs] = this.randomFragment(fragmentKey)
    const startS = startMs / 1000
    const delayS = delayMs > 0 ? delayMs / 1000 : 0
    const offsetS = delayMs < 0 ? -delayMs / 1000 : 0
    const durationS = durationMs / 1000 - offsetS
    if (durationS <= 0) {
      return
    }
    // Waiting manually instead of using the first argument of `start`
    // because it doesn't work as expected.
    if (delayS > 0) {
      try {
        await sleep(sig, delayS * 1000)
      } catch (e) {
        return
      }
    }
    src.start(0, startS + offsetS, durationS)

    sig.addEventListener("abort", () => {
      src.stop()
    })
  }

  private async audioSource(): Promise<AudioBufferSourceNode | undefined> {
    // Wait for source to be fetched.
    const payload = await this.sourcePayload
    if (!payload) {
      return
    }
    if (!this.buffer) {
      this.buffer = new Promise((res) => {
        this.context = new AudioContext()
        this.context.decodeAudioData(payload).then((buff) => res(buff))
        this.gain = this.context.createGain()
        this.gain.connect(this.context.destination)
        this.gain.gain.value = SOUND_GAIN
      })
    }
    if (!this.context || !this.gain) {
      return undefined
    }
    const buff = cloneAudioBuffer(await this.buffer)
    const src = this.context.createBufferSource()
    src.buffer = buff
    src.connect(this.gain)
    return src
  }

  private randomFragment(key: string): number[] {
    const fragments = this.fragments[key]
    if (!fragments) {
      throw new Error(`Sound key doesn't exist: ${key}`)
    }
    const i = Math.floor(Math.random() * fragments.length)
    return fragments[i]
  }
}

function cloneAudioBuffer(b: AudioBuffer) {
  const buff = new AudioBuffer({
    length: b.length,
    numberOfChannels: b.numberOfChannels,
    sampleRate: b.sampleRate,
  })
  for (let i = 0; i < buff.numberOfChannels; ++i) {
    const samples = b.getChannelData(i)
    buff.copyToChannel(samples, i)
  }
  return buff
}

export interface RenderArgs {
  sprite: Sprite
  spriteKey: string
}

export interface PlayArgs {
  sound: Sound
  soundKey: string
  delayMs?: number
  semaphor?: Semaphor
}

export interface ActorArgs {
  periodMs: number
  render?: RenderArgs
  play?: PlayArgs
}

export class Actor {
  private readonly periodMs: number
  private readonly renderArgs?: RenderArgs
  private readonly playArgs?: PlayArgs
  private startAt: number
  private currentAt: number
  private isPaused: boolean
  private isSoundAllowed: boolean
  private playCtrl: AbortController
  private actorCtrl: AbortController

  constructor(p: ActorArgs) {
    this.playArgs = p.play
    this.renderArgs = p.render
    this.periodMs = p.periodMs
    this.startAt = 0
    this.currentAt = 0
    this.isPaused = true
    this.playCtrl = new AbortController()
    this.actorCtrl = new AbortController()
    this.isSoundAllowed = false

    if (!p.play) {
      return
    }

    if (!p.play.semaphor) {
      this.isSoundAllowed = true
      return
    }

    p.play.semaphor
      .acquire(this.actorCtrl.signal)
      .then(() => {
        this.isSoundAllowed = true
        this.startSound(this.playCtrl.signal)
      })
      .catch(() => {}) // eslint-disable-line @typescript-eslint/no-empty-function
  }

  play(): void {
    // Was started but is not paused (still playing).
    if (this.startAt && !this.isPaused) {
      return
    }
    this.playCtrl.abort()
    this.playCtrl = new AbortController()

    this.isPaused = false
    if (!this.startAt) {
      // Starting for the first time.
      this.startAt = Date.now() % this.periodMs
      this.currentAt = 0
    }

    this.startSound(this.playCtrl.signal)
  }

  pause(): void {
    this.playCtrl.abort()
    this.isPaused = true
    this.currentAt = this.currentAtNow()
  }

  close(): void {
    this.actorCtrl.abort()
    this.playCtrl.abort()
  }

  async render(ctx: CanvasRenderingContext2D, width: number, height: number): Promise<void> {
    if (!this.renderArgs) {
      return
    }

    if (!this.isPaused) {
      this.currentAt = this.currentAtNow()
    }

    const sprite = this.renderArgs.sprite
    const key = this.renderArgs.spriteKey
    await sprite.render(ctx, key, this.currentAt / this.periodMs, width, height)
  }

  private async startSound(signal: AbortSignal): Promise<void> {
    if (!this.playArgs || this.isPaused || !this.isSoundAllowed) {
      return
    }

    try {
      signal.throwIfAborted()

      const delayMs = this.playArgs.delayMs || 0
      const sound = this.playArgs.sound
      const key = this.playArgs.soundKey
      this.currentAt = this.currentAtNow()
      sound.play(signal, key, delayMs - this.currentAt)
      await sleep(signal, this.periodMs - this.currentAt)
      sound.play(signal, key, delayMs)
      repeat(signal, this.periodMs, () => sound.play(signal, key, delayMs))
    } catch (e) {} // eslint-disable-line no-empty
  }

  private currentAtNow(): number {
    return (Date.now() - this.startAt) % this.periodMs
  }
}
