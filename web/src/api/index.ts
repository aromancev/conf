import { Client } from "./api"
import { ConfaClient } from "./confa"
import { TalkClient } from "./talk"
import { EventClient } from "./event"
import { ConfaMask as SchemaConfaMask, TalkMask as SchemaTalkMask } from "./schema"

export * from "./models"
export * from "./room"
export * from "./api"
export type ConfaMask = SchemaConfaMask
export type TalkMask = SchemaTalkMask

export const client = new Client()
export const confaClient = new ConfaClient(client)
export const talkClient = new TalkClient(client)
export const eventClient = new EventClient(client)
