import { reactive, readonly } from "vue"
import { RoomEvent, PeerStatus } from "@/api/room/schema"
import { FIFOMap, FIFOSet } from "@/platform/cache"

export interface State {
  peers: Map<string, Peer>
}

export interface Peer {
  userId: string
  profile: Profile
  sessionIds: Set<string>
}

export class PeerAggregator {
  private _state: State
  private repo: Repo

  constructor(repo: Repo) {
    this.repo = repo
    this._state = reactive({
      peers: new FIFOMap(MAX_PEERS),
    })
  }

  state(): State {
    return readonly(this._state) as State
  }

  put(event: RoomEvent): void {
    const state = event.payload.peerState
    if (!state?.status) {
      return
    }
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
    }
  }

  reset(): void {
    this._state.peers.clear()
  }
}

const MAX_PEERS = 3000
const MAX_SESSIONS_PER_PEER = 10

interface Profile {
  handle: string
  name: string
  avatar: string
}

interface Repo {
  profile(id: string): Profile
}
