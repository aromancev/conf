export enum EventType {
  PeerState = "peer_state",
  Message = "message",
}

export type EventPayload = PayloadPeerState | PayloadMessage

export interface Event {
  id?: string
  ownerId?: string
  roomId?: string
  createdAt?: string
  payload: {
    type: EventType
    payload: EventPayload
  }
}

export interface EventProcessor {
  forward(event: Event): void
  backward(event: Event): void
}

export enum PeerStatus {
  Joined = "joined",
  Left = "left",
}

export enum Hint {
  Camera = "camera",
  Screen = "screen",
  UserAudio = "user_audio",
  DeviceAudio = "device_audio",
}

export interface Track {
  hint: Hint
}

export interface PayloadPeerState {
  status?: PeerStatus
  tracks?: { [key: string]: Track }
}

export interface PayloadMessage {
  text: string
}
