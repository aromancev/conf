import { Client } from "./api"
import { ConfaClient } from "./confa"
import { TalkClient } from "./talk"

export { User, Account, userStore, Confa, Talk } from "./models"

export const client: Client = new Client()
export const confa: ConfaClient = new ConfaClient(client)
export const talk: TalkClient = new TalkClient(client)
