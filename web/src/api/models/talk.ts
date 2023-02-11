import { RegexValidator } from "@/platform/validator"

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
  state: TalkState
  title: string
  description?: string
}

export const handleValidator = new RegexValidator("^[a-z0-9-]{4,64}$", [
  "Must be from 4 to 64 characters long",
  "Can only contain lower case letters, numbers, and '-'",
])
export const titleValidator = new RegexValidator("^[a-zA-Z0-9- ]{0,64}$", [
  "Must be from 0 to 64 characters long",
  "Can only contain letters, numbers, spaces, and '-'",
])
