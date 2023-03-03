import { Client } from "./api"
import { ProfileClient } from "./profile"
import { ConfaClient } from "./confa"
import { TalkClient } from "./talk"
import { EventClient } from "./event"
import { RecordingClient } from "./recording"

export * from "./room"
export * from "./api"

export const api = new Client()
export const profileClient = new ProfileClient(api)
export const confaClient = new ConfaClient(api)
export const talkClient = new TalkClient(api)
export const eventClient = new EventClient(api)
export const recordingClient = new RecordingClient(api)
