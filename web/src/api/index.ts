import { Client } from "./api"
import { ConfaClient } from "./confa"
import { Signal } from "./rtc"

export const signal: Signal = new Signal()
export const client: Client = new Client()
export const confa: ConfaClient = new ConfaClient(client)
