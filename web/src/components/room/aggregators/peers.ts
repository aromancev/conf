import { reactive, readonly } from "vue"
import { RoomEvent, PeerStatus } from "@/api/room/schema"
import { FIFO } from "@/platform/cache"

interface Profile {
  handle: string
  name: string
  avatar: string
}

interface Repo {
  profile(id: string): Profile
}

export interface State {
  peers: Map<string, Peer>
}

export interface Peer {
  userId: string
  joinedAt: number
  profile: Profile
}

export class PeerAggregator {
  private _state: State
  private repo: Repo

  constructor(repo: Repo) {
    this.repo = repo
    this._state = reactive({
      peers: new FIFO(CAPACITY),
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
        if (this._state.peers.has(state.peerId)) {
          return
        }

        this._state.peers.set(state.peerId, {
          userId: state.peerId,
          joinedAt: event.createdAt || 0,
          profile: this.repo.profile(state.peerId),
        })
        break
      case PeerStatus.Left:
        this._state.peers.delete(state.peerId)
        break
    }
    return
  }

  reset(): void {
    this._state.peers.clear()
  }
}

const CAPACITY = 3000
