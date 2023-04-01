import { RoomEvent, RecordingEventStatus } from "@/api/room/schema"

export interface Recording {
  isRecording: boolean
}

export class RecordingAggregator {
  private readonly recording: Recording

  constructor(recording: Recording) {
    this.recording = recording
  }

  put(event: RoomEvent): void {
    const recording = event.payload.recording
    if (!recording) {
      return
    }
    switch (recording.status) {
      case RecordingEventStatus.Started:
        this.recording.isRecording = true
        break
      case RecordingEventStatus.Stopped:
        this.recording.isRecording = false
        break
    }
  }
}
