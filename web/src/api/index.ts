import { Client } from "./api"
import { ProfileClient } from "./profile"
import { ConfaClient } from "./confa"
import { TalkClient } from "./talk"
import { EventClient } from "./event"

export * from "./models"
export * from "./room"
export * from "./api"

export const client = new Client()
export const profileClient = new ProfileClient(client)
export const confaClient = new ConfaClient(client)
export const talkClient = new TalkClient(client)
export const eventClient = new EventClient(client)
