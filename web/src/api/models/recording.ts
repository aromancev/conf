export interface Recording {
  key: string
  roomId: string
  createdAt: number
  startedAt: number
  stoppedAt?: number | null
}
