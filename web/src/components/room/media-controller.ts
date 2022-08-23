import { reactive, readonly, watch, WatchStopHandle, WatchSource } from "vue"
import { MediaPlayer, MediaPlayerClass, PlaybackTimeUpdatedEvent } from "dashjs"
import { Media } from "./aggregators/media"

interface Watchers {
  media: WatchSource<Media | undefined>
  element: WatchSource<HTMLElement | undefined>
  isPlaying: WatchSource<boolean>
  unpausedAt: WatchSource<number>
  delta: WatchSource<number>
}

interface State {
  isActive: boolean
}

export class MediaController {
  private player: MediaPlayerClass
  private _state: State
  private startsAt?: number
  private stopWatch: WatchStopHandle
  private deltaStopWatch: WatchStopHandle

  constructor(watchers: Watchers) {
    this.player = MediaPlayer().create()
    this.player.on("playbackTimeUpdated", (event: PlaybackTimeUpdatedEvent) => {
      this._state.isActive = event.timeToEnd > 0 && (event.time || 0) > 0
    })
    this._state = reactive<State>({
      isActive: false,
    })
    this.deltaStopWatch = watch(watchers.delta, (delta) => {
      this.updateDelta(delta)
    })
    this.stopWatch = watch(
      [watchers.media, watchers.element, watchers.isPlaying, watchers.unpausedAt, watchers.delta],
      ([media, element, isPlaying, unpausedAt, delta]) => {
        this.update(media, element, isPlaying, unpausedAt, delta)
      },
    )
  }

  state(): State {
    return readonly(this._state)
  }

  private update(
    media: Media | undefined,
    element: HTMLElement | undefined,
    isPlaying: boolean,
    unpausedAt: number,
    delta: number,
  ): void {
    if (!media || !element) {
      return
    }
    this.startsAt = media.startsAt

    if (!this.player.isReady() || this.player.getSource() !== media.manifestUrl) {
      this.player.initialize(element, media.manifestUrl, isPlaying)
    }
    if (!media.startsAt) {
      this.player.seek(0)
      this.player.pause()
      return
    }

    const shouldStartByNow = Date.now() - unpausedAt + delta - media.startsAt
    if (isPlaying && shouldStartByNow) {
      this.player.play()
    } else {
      this.player.pause()
    }
  }

  private updateDelta(delta: number): void {
    if (!this.startsAt) {
      return
    }

    let seek = (delta - this.startsAt) / 1000 // Delta is in ms, but player seeks in seconds
    if (seek < 0) {
      seek = 0
    }
    if (seek > this.player.duration()) {
      seek = this.player.duration()
    }
    this.player.seek(seek)
  }

  close(): void {
    this.stopWatch()
    this.deltaStopWatch()
    if (this.player.isReady()) {
      this.player.destroy()
    }
  }
}
