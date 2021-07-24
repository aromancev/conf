import { Client } from "./api"
import { ConfaClient } from "./confa"

export { User, userStore, Confa } from "./models"

export const client: Client = new Client()
export const confa: ConfaClient = new ConfaClient(client)
