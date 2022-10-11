import { gql } from "@apollo/client/core"
import { Client, FetchPolicy } from "./api"
import { EventLookup, EventLimit, EventOrder, events, eventsVariables, EventFromInput } from "./schema"
import { RoomEvent } from "./room/schema"

interface OptionalFetchParams {
  policy?: FetchPolicy
  order?: EventOrder
  from?: EventFromInput
}

interface FetchParams {
  policy: FetchPolicy
  order: EventOrder
}

const defaultParams: FetchParams = {
  policy: "cache-first",
  order: EventOrder.ASC,
}

export class EventClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async fetchOne(lookup: EventLookup, params?: OptionalFetchParams): Promise<RoomEvent | null> {
    const iter = this.fetch(lookup, params)
    const events = await iter.next()
    if (events.length === 0) {
      return null
    }
    if (events.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return events[0]
  }

  fetch(lookup: EventLookup, params?: OptionalFetchParams): EventIterator {
    return new EventIterator(this.api, lookup, params)
  }
}

export class EventIterator {
  private api: Client
  private lookup: EventLookup
  private from?: EventFromInput
  private params: FetchParams
  private pages: number

  constructor(api: Client, lookup: EventLookup, params?: OptionalFetchParams) {
    this.api = api
    this.lookup = lookup
    this.params = {
      ...defaultParams,
      ...params,
    }
    this.from = params?.from
    this.pages = 0
  }

  pagesIterated(): number {
    return this.pages
  }

  async next(limit?: EventLimit): Promise<RoomEvent[]> {
    this.pages++

    const resp = await this.api.query<events, eventsVariables>({
      query: gql`
        query events($where: EventLookup!, $from: EventFromInput, $limit: EventLimit!, $order: EventOrder) {
          events(where: $where, limit: $limit, from: $from, order: $order) {
            items {
              id
              roomId
              createdAt
              payload
            }
            nextFrom {
              id
              createdAt
            }
          }
        }
      `,
      variables: {
        where: this.lookup,
        from: this.from,
        limit: limit || { count: 100, seconds: 0 },
        order: this.params.order,
      },
      fetchPolicy: this.params.policy,
    })

    this.from = resp.data.events.nextFrom || undefined
    const events: RoomEvent[] = []
    for (const item of resp.data.events.items) {
      events.push({
        id: item.id,
        roomId: item.roomId,
        createdAt: item.createdAt,
        payload: JSON.parse(item.payload),
      })
    }
    return events
  }
}
