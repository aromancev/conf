import { gql } from "@apollo/client/core"
import { Client, FetchPolicy } from "./api"
import {
  EventLookup,
  EventLimit,
  EventOrder,
  events,
  eventsVariables,
  Status as GStatus,
  Hint as GHint,
  EventFromInput,
} from "./schema"
import { RoomEvent, EventPeerState, Status, Hint } from "./room/schema"

interface OptionalFetchParams {
  policy?: FetchPolicy
  order?: EventOrder
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

class EventIterator {
  private api: Client
  private lookup: EventLookup
  private from: EventFromInput | null
  private params: FetchParams

  constructor(api: Client, lookup: EventLookup, params?: OptionalFetchParams) {
    this.api = api
    this.lookup = lookup
    this.from = null
    this.params = {
      ...defaultParams,
      ...params,
    }
  }

  async next(limit?: EventLimit): Promise<RoomEvent[]> {
    const resp = await this.api.query<events, eventsVariables>({
      query: gql`
        query events($where: EventLookup!, $from: EventFromInput, $limit: EventLimit!, $order: EventOrder) {
          events(where: $where, limit: $limit, from: $from, order: $order) {
            items {
              id
              ownerId
              roomId
              createdAt
              payload {
                peerState {
                  status
                  tracks {
                    id
                    hint
                  }
                }
                message {
                  text
                }
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
        where: this.lookup,
        from: this.from,
        limit: limit || { count: 100, seconds: 0 },
        order: this.params.order,
      },
      fetchPolicy: this.params.policy,
    })

    this.from = resp.data.events.nextFrom
    const events: RoomEvent[] = []
    for (const item of resp.data.events.items) {
      const event: RoomEvent = {
        id: item.id,
        ownerId: item.ownerId,
        roomId: item.roomId,
        createdAt: item.createdAt,
        payload: {},
      }
      if (item.payload.message) {
        event.payload.message = item.payload.message
      }
      if (item.payload.peerState) {
        const state: EventPeerState = {
          status: item.payload.peerState.status ? fromGState(item.payload.peerState.status) : undefined,
        }
        if (item.payload.peerState.tracks) {
          state.tracks = []
          for (const t of item.payload.peerState.tracks) {
            state.tracks.push({
              id: t.id,
              hint: fromGHint(t.hint) || Hint.Camera,
            })
          }
        }
        event.payload.peerState = state
      }
      events.push(event)
    }
    return events
  }
}

function fromGState(s: GStatus): Status {
  switch (s) {
    case GStatus.joined:
      return Status.Joined
    case GStatus.left:
      return Status.Left
  }
}

function fromGHint(h: GHint): Hint {
  switch (h) {
    case GHint.camera:
      return Hint.Camera
    case GHint.device_audio:
      return Hint.DeviceAudio
    case GHint.screen:
      return Hint.Screen
    case GHint.user_audio:
      return Hint.UserAudio
  }
}
