import { gql } from "@apollo/client/core"
import { Client, FetchPolicy } from "./api"
import { EventLookup, events, eventsVariables, EventCursorInput } from "./schema"
import { RoomEvent } from "./room/schema"

interface OptionalFetchParams {
  policy?: FetchPolicy
  cursor?: EventCursorInput
}

interface FetchParams {
  policy: FetchPolicy
}

const defaultParams: FetchParams = {
  policy: "cache-first",
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
  private cursor?: EventCursorInput
  private params: FetchParams
  private pages: number

  constructor(api: Client, lookup: EventLookup, params?: OptionalFetchParams) {
    this.api = api
    this.lookup = lookup
    this.params = {
      ...defaultParams,
      ...params,
    }
    this.cursor = params?.cursor
    this.pages = 0
  }

  pagesIterated(): number {
    return this.pages
  }

  async next(limit?: number): Promise<RoomEvent[]> {
    this.pages++

    const resp = await this.api.query<events, eventsVariables>({
      query: gql`
        query events($where: EventLookup!, $limit: Int!, $cursor: EventCursorInput) {
          events(where: $where, limit: $limit, cursor: $cursor) {
            items {
              id
              roomId
              createdAt
              payload
            }
            next {
              id
              createdAt
            }
          }
        }
      `,
      variables: {
        where: this.lookup,
        cursor: this.cursor,
        limit: limit || 100,
      },
      fetchPolicy: this.params.policy,
    })

    this.cursor = resp.data.events.next || undefined
    const events: RoomEvent[] = []
    for (const item of resp.data.events.items) {
      events.push({
        id: item.id,
        roomId: item.roomId,
        createdAt: Number(item.createdAt),
        payload: JSON.parse(item.payload),
      })
    }
    return events
  }
}
