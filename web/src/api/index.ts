import { Client } from "./api"
import { ConfaClient } from "./confa"
import { TalkClient } from "./talk"
import { EventsClient } from "./event"
export {
  User,
  Account,
  userStore,
  Confa,
  Talk,
  Event,
  EventType,
  EventProcessor,
} from "./models"
export { RTC } from "./rtc"

export const client: Client = new Client()
export const confa: ConfaClient = new ConfaClient(client)
export const talk: TalkClient = new TalkClient(client)
export const event: EventsClient = new EventsClient(client)
