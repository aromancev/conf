import { RoomEvent, Status } from "@/api/room/schema"

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

    const userId = event.ownerId || ""
    switch (state.status) {
      case Status.Joined:
        if (this.peers.has(userId)) {
          return
        }

        this.peers.set(userId, {
          userId: userId,
          joinedAt: event.createdAt || 0,
          profile: this.repo.profile(userId),
        })
        break
      case Status.Left:
        this.peers.delete(userId)
        break
    }
    return
  }
}
