import { reactive, readonly } from "vue"
import { RoomEvent, RecordingStatus } from "@/api/room/schema"

export interface State {
  isRecording: boolean
}

export class RecordingAggregator {
  private _state: State

  constructor() {
    this._state = reactive({
      isRecording: false,
    })
  }

  state(): State {
    return readonly(this._state) as State
  }

  put(event: RoomEvent): void {
    const recording = event.payload.recording
    if (!recording) {
      return
    }
    switch (recording.status) {
      case RecordingStatus.Started:
        this._state.isRecording = true
        break
      case RecordingStatus.Stopped:
        this._state.isRecording = false
        break
    }
  }
}
