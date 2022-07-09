import { reactive, readonly } from "vue"
import { MediaPlayer, MediaPlayerClass, PlaybackTimeUpdatedEvent } from "dashjs"
import { Media } from "./aggregators/media"

interface Props {
  media?: Media
  element?: HTMLElement
  isPlaying: boolean
  delta: number
  unpausedAt: number
}

interface State {
  isActive: boolean
}

export class MediaController {
  private player: MediaPlayerClass
  private _state: State

  constructor() {
    this.player = MediaPlayer().create()
    this.player.on("playbackTimeUpdated", (event: PlaybackTimeUpdatedEvent) => {
      this._state.isActive = event.timeToEnd > 0 && (event.time || 0) > 0
    })
    this._state = reactive<State>({
      isActive: false,
    })
  }

  state(): State {
    return readonly(this._state)
  }

  update(props: Props) {
    if (!props.media || !props.element) {
      return
    }

    if (!this.player.isReady() || this.player.getSource() !== props.media.manifestUrl) {
      this.player.initialize(props.element, props.media.manifestUrl, props.isPlaying)
    }
    if (!props.media.startsAt) {
      this.player.seek(0)
      this.player.pause()
      return
    }
    let seek = (props.delta - props.media.startsAt) / 1000 // Delta is in ms, but player seeks in seconds
    if (seek < 0) {
      seek = 0
    }
    if (seek > this.player.duration()) {
      seek = this.player.duration()
    }
    this.player.seek(seek)
    const shouldStartByNow = Date.now() - props.unpausedAt + props.delta - props.media.startsAt
    const seekInsideVideo = seek > 0 && seek < this.player.duration()
    if (props.isPlaying && (shouldStartByNow || seekInsideVideo)) {
      this.player.play()
    } else {
      this.player.pause()
    }
  }

  close(): void {
    this.player.destroy()
  }
}
