import { reactive, readonly, watch, WatchStopHandle, WatchSource } from "vue"
import { MediaPlayer, MediaPlayerClass, PlaybackTimeUpdatedEvent } from "dashjs"
import { Media } from "./aggregators/media"

interface Watchers {
  media: WatchSource<Media | undefined>
  element: WatchSource<HTMLElement | undefined>
  isPlaying: WatchSource<boolean>
  isBuffering: WatchSource<boolean>
  progress: WatchSource<Progress>
}

interface Progress {
  value: number
  increasingSince: number
}

interface State {
  isActive: boolean
}

export class MediaController {
  onBuffer?: (ms: number) => void

  private player: MediaPlayerClass
  private _state: State
  private startsAt?: number
  private stopWatch: WatchStopHandle

  constructor(watchers: Watchers) {
    this.player = MediaPlayer().create()
    this.player.on("playbackTimeUpdated", (event: PlaybackTimeUpdatedEvent) => {
      this._state.isActive = event.timeToEnd > 0 && (event.time || 0) > 0
    })
    this.player.on("fragmentLoadingCompleted", (e) => {
      if (!this.onBuffer) {
        return
      }
      const ms = e.request.mediaStartTime + e.request.duration
      if (!ms) {
        return
      }
      if (ms >= this.player.duration()) {
        this.onBuffer(Infinity)
        return
      }
      this.onBuffer(ms * 1000)
    })

    this._state = reactive<State>({
      isActive: false,
    })
    this.stopWatch = watch(
      [watchers.media, watchers.element, watchers.isPlaying, watchers.isBuffering, watchers.progress],
      ([media, element, isPlaying, isBuffering, progress]) => {
        this.update(media, element, isPlaying, isBuffering, progress)
      },
      {
        deep: true,
        immediate: true,
      },
    )
  }

  state(): State {
    return readonly(this._state)
  }

  close(): void {
    this.stopWatch()
    if (this.player.isReady()) {
      this.player.destroy()
    }
  }

  private update(
    media: Media | undefined,
    element: HTMLElement | undefined,
    isPlaying: boolean,
    isBuffering: boolean,
    progress: Progress,
  ): void {
    if (!media || !element) {
      return
    }
    this.startsAt = media.startsAt

    if (!this.player.isReady() || this.player.getSource() !== media.manifestUrl) {
      if (this.onBuffer) {
        this.onBuffer(0)
      }
      this.player.initialize(element, media.manifestUrl, isPlaying)
    }
    const progressNow = this.progressForNow(progress)
    const shouldStartByNow = progressNow >= media.startsAt
    if (!shouldStartByNow) {
      this.player.seek(0)
      this.player.pause()
      return
    }

    if (isPlaying && !isBuffering) {
      this.player.play()
    } else {
      this.player.pause()
    }
    this.seek(progressNow)
  }

  private seek(progress: number): void {
    if (!this.startsAt) {
      return
    }

    let seek = (progress - this.startsAt) / 1000 // Progress is in ms, but player seeks in seconds
    if (seek < 0) {
      seek = 0
    }
    if (seek > this.player.duration()) {
      seek = this.player.duration()
    }
    this.player.seek(seek)
  }

  private progressForNow(progress: Progress): number {
    if (!progress.increasingSince) {
      return progress.value
    }
    return Date.now() - progress.increasingSince + progress.value
  }
}
