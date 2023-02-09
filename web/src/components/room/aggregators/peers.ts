import { reactive, readonly } from "vue"
import { RoomEvent, PeerStatus } from "@/api/room/schema"
import { FIFOMap, FIFOSet } from "@/platform/cache"

const MAX_PEERS = 3000
const MAX_SESSIONS_PER_PEER = 10
const CLAP_TIMEOUT_MS = 7 * 1000

export type StatusKey = "clap"

export interface State {
  peers: Map<string, Peer>
  statuses: Map<string, Status>
}

export interface Peer {
  userId: string
  profile: Profile
  sessionIds: Set<string>
}

export interface Status {
  clap?: {
    startAt: number
  }
}

export class PeerAggregator {
  private _state: State
  private repo: Repo

  constructor(repo: Repo) {
    this.repo = repo
    this._state = reactive({
      peers: new FIFOMap(MAX_PEERS),
      statuses: new FIFOMap(MAX_PEERS),
    })
  }

  state(): State {
    return readonly(this._state) as State
  }

  setTime(time: number): void {
    for (const [id, status] of this._state.statuses) {
      if (status.clap && time > status.clap.startAt + CLAP_TIMEOUT_MS) {
        status.clap = undefined
      }
      if (isStatusEmpty(status)) {
        this._state.statuses.delete(id)
      }
    }
  }

  put(event: RoomEvent): void {
    const state = event.payload.peerState
    if (state?.status) {
      switch (state.status) {
        case PeerStatus.Joined:
          this.join(state.peerId, state.sessionId)
          break
        case PeerStatus.Left:
          this.leave(state.peerId, state.sessionId)
          break
      }
      return
    }

    const reaction = event.payload.reaction
    if (reaction) {
      if (!this._state.peers.has(reaction.fromId)) {
        return
      }
      let state = this._state.statuses.get(reaction.fromId)
      if (!state) {
        state = {}
        this._state.statuses.set(reaction.fromId, state)
      }
      if (reaction.reaction.clap) {
        if (reaction.reaction.clap.isStarting) {
          state.clap = {
            startAt: event.createdAt,
          }
        } else {
          state.clap = undefined
        }
      }
    }
  }

  private join(peerId: string, sessionId: string): void {
    const peer = this._state.peers.get(peerId)
    if (peer) {
      peer.sessionIds.add(sessionId)
      return
    }

    const sessions = new FIFOSet<string>(MAX_SESSIONS_PER_PEER)
    sessions.add(sessionId)
    this._state.peers.set(peerId, {
      userId: peerId,
      profile: this.repo.profile(peerId),
      sessionIds: sessions,
    })
  }

  private leave(peerId: string, sessionId: string): void {
    const peer = this._state.peers.get(peerId)
    if (!peer) {
      return
    }
    peer.sessionIds.delete(sessionId)
    if (peer.sessionIds.size === 0) {
      this._state.peers.delete(peerId)
      this._state.statuses.delete(peerId)
    }
  }

  reset(): void {
    this._state.peers.clear()
    this._state.statuses.clear()
  }
}

interface Profile {
  handle: string
  name: string
  avatar: string
}

interface Repo {
  profile(id: string): Profile
}

function isStatusEmpty(status: Status): boolean {
  if (status.clap) {
    return false
  }
  return true
}
