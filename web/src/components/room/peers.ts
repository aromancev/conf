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
  private peers: { [k: string]: Peer }
  private repo: Repo

  constructor(repo: Repo, peers: { [k: string]: Peer }) {
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
        if (this.peers[userId]) {
          return
        }

        this.peers[userId] = {
          userId: userId,
          joinedAt: event.createdAt || 0,
          profile: this.repo.profile(userId),
        }
        break
      case Status.Left:
        delete this.peers[userId]
        break
    }
    return
  }
}
