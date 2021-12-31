import { gql } from "@apollo/client/core"
import { Client } from "./api"
import { EventInput, EventLimit, EventOrder, events, eventsVariables } from "./schema"
import { Event, EventPayload, EventType } from "./models"

interface From {
  id: string
  createdAt: string
}

class EventIterator {
  private api: Client
  private input: EventInput
  private from: From | null
  private order: EventOrder | null

  constructor(api: Client, input: EventInput, order?: EventOrder) {
    this.api = api
    this.input = input
    this.from = null
    this.order = order || null
  }

  async next(limit?: EventLimit): Promise<Event[]> {
    const resp = await this.api.query<events, eventsVariables>({
      query: gql`
        query events($where: EventInput!, $from: EventFromInput, $limit: EventLimit!, $order: EventOrder) {
          events(where: $where, limit: $limit, from: $from, order: $order) {
            items {
              id
              ownerId
              roomId
              createdAt
              payload {
                type
                payload
              }
            }
            nextFrom {
              id
              createdAt
            }
          }
        }
      `,
      variables: {
        where: this.input,
        from: this.from,
        limit: limit || { count: 100, seconds: 0 },
        order: this.order,
      },
    })

    this.from = resp.data.events.nextFrom
    const events: Event[] = []
    for (const item of resp.data.events.items) {
      events.push({
        id: item.id,
        ownerId: item.ownerId,
        roomId: item.roomId,
        createdAt: item.createdAt,
        payload: {
          type: item.payload.type as EventType,
          payload: JSON.parse(item.payload.payload) as EventPayload,
        },
      })
    }
    events.sort((a: Event, b: Event): number => {
      const ac = a.createdAt || ""
      const bc = b.createdAt || ""

      if (ac < bc) {
        return -1
      }
      if (ac > bc) {
        return 1
      }
      return 0
    })
    return events
  }
}

export class EventClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async fetchOne(input: EventInput, order?: EventOrder): Promise<Event | null> {
    const iter = this.fetch(input, order)
    const events = await iter.next()
    if (events.length === 0) {
      return null
    }
    if (events.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return events[0]
  }

  fetch(input: EventInput, order?: EventOrder): EventIterator {
    return new EventIterator(this.api, input, order)
  }
}
