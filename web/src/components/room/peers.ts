import { RoomEvent, PeerStatus } from "@/api/room/schema"

interface Profile {
  handle: string
  name: string
  avatar: string
}

interface Repo {
  profile(id: string): Profile
}

export interface Peer {
  userId: string
  joinedAt: number
  profile: Profile
}

export class PeerAggregator {
  private peers: Map<string, Peer>
  private repo: Repo

  constructor(repo: Repo, peers: Map<string, Peer>) {
    this.repo = repo
    this.peers = peers
  }

  async put(event: RoomEvent): Promise<void> {
    const state = event.payload.peerState
    if (!state?.status) {
      return
    }

    switch (state.status) {
      case PeerStatus.Joined:
        if (this.peers.has(state.peerId)) {
          return
        }

        this.peers.set(state.peerId, {
          userId: state.peerId,
          joinedAt: event.createdAt || 0,
          profile: this.repo.profile(state.peerId),
        })
        break
      case PeerStatus.Left:
        this.peers.delete(state.peerId)
        break
    }
    return
  }
}
