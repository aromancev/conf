import { Client } from "./api"
import { ConfaClient } from "./confa"
import { TalkClient } from "./talk"
import { EventClient } from "./event"
import { ConfaInput as SchemaConfaInput } from "./schema"

export * from "./models"
export * from "./rtc"
export * from "./api"
export type ConfaInput = SchemaConfaInput

export const client = new Client()
export const confaClient = new ConfaClient(client)
export const talkClient = new TalkClient(client)
export const eventClient = new EventClient(client)
