import { gql } from "@apollo/client/core"
import { Client, FetchPolicy } from "./api"
import {
  EventLookup,
  EventLimit,
  EventOrder,
  events,
  eventsVariables,
  PeerStatus as GPeerStatus,
  RecordingStatus as GRecordingStatus,
  Hint as GHint,
  EventFromInput,
} from "./schema"
import { RoomEvent, EventPeerState, PeerStatus, RecordingStatus, Hint } from "./room/schema"

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
              payload {
                peerState {
                  peerId
                  status
                  tracks {
                    id
                    hint
                  }
                }
                message {
                  fromId
                  text
                }
                recording {
                  status
                }
                trackRecording {
                  id
                  trackId
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

    this.from = resp.data.events.nextFrom || undefined
    const events: RoomEvent[] = []
    for (const item of resp.data.events.items) {
      const event: RoomEvent = {
        id: item.id,
        roomId: item.roomId,
        createdAt: item.createdAt,
        payload: {},
      }
      if (item.payload.message) {
        event.payload.message = item.payload.message
      }
      if (item.payload.peerState) {
        const state: EventPeerState = {
          peerId: item.payload.peerState.peerId,
          status: item.payload.peerState.status ? fromGPeerState(item.payload.peerState.status) : undefined,
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
      if (item.payload.recording) {
        event.payload.recording = {
          status: fromGRecordingState(item.payload.recording.status),
        }
      }
      if (item.payload.trackRecording) {
        event.payload.trackRecording = item.payload.trackRecording
      }
      events.push(event)
    }
    return events
  }
}

function fromGPeerState(s: GPeerStatus): PeerStatus {
  switch (s) {
    case GPeerStatus.joined:
      return PeerStatus.Joined
    case GPeerStatus.left:
      return PeerStatus.Left
  }
}

function fromGRecordingState(s: GRecordingStatus): RecordingStatus {
  switch (s) {
    case GRecordingStatus.started:
      return RecordingStatus.Started
    case GRecordingStatus.stopped:
      return RecordingStatus.Stopped
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
