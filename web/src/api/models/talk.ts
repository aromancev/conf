export enum TalkState {
  CREATED = "CREATED",
  ENDED = "ENDED",
  LIVE = "LIVE",
  RECORDING = "RECORDING",
}

export interface Talk {
  id: string
  ownerId: string
  confaId: string
  roomId: string
  handle: string
  title?: string
  description?: string
  state?: TalkState
}
