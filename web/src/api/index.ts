import { Client } from "./api"
import { ConfaClient } from "./confa"

export const client: Client = new Client()
export const confa: ConfaClient = new ConfaClient(client)
