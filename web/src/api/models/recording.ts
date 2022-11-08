export interface Recording {
  key: string
  roomId: string
  status: RecordingStatus
  createdAt: number
  startedAt: number
  stoppedAt?: number | null
}

export enum RecordingStatus {
  PROCESSING = "PROCESSING",
  READY = "READY",
  RECORDING = "RECORDING",
}
