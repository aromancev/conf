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
  onBuffer?: (bufferMs: number, durationMs: number) => void

  private player: MediaPlayerClass
  private _state: State
  private startsAt?: number
  private stopWatch: WatchStopHandle
  private stopWatchProgress: WatchStopHandle

  constructor(watchers: Watchers) {
    this.player = MediaPlayer().create()
    this.player.on("playbackTimeUpdated", (event: PlaybackTimeUpdatedEvent) => {
      this._state.isActive = event.timeToEnd > 0 && (event.time || 0) > 0
    })
    this.player.on("fragmentLoadingCompleted", (e) => {
      if (!this.onBuffer) {
        return
      }
      const sec = e.request.mediaStartTime + e.request.duration
      if (!sec) {
        return
      }
      this.onBuffer(sec * 1000, this.player.duration() * 1000)
    })
    this.player.on("bufferStalled", () => {
      if (this.onBuffer) {
        this.onBuffer(0, this.player.duration() * 1000)
      }
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
        immediate: false,
      },
    )
    this.stopWatchProgress = watch(
      watchers.progress,
      (progress) => {
        this.updateProgress(progress)
      },
      {
        deep: true,
        immediate: false,
      },
    )
  }

  state(): State {
    return readonly(this._state)
  }

  close(): void {
    this.stopWatch()
    this.stopWatchProgress()
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
        this.onBuffer(0, 0)
      }
      this.player.initialize(element, media.manifestUrl, isPlaying)
    }
    const progressNow = this.progressForNow(progress)
    if (progressNow < media.startsAt || progressNow > media.startsAt + this.player.duration() * 1000) {
      this.player.pause()
      return
    }

    if (isPlaying && !isBuffering) {
      this.player.play()
    } else {
      this.player.pause()
    }
  }

  private updateProgress(progress: Progress): void {
    const progressNow = this.progressForNow(progress)
    if (!this.startsAt) {
      return
    }

    let seek = (progressNow - this.startsAt) / 1000 // Progress is in ms, but player seeks in seconds
    if (seek < 0) {
      seek = 0
    }
    if (seek > this.player.duration()) {
      return
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
