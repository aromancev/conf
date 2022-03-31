import { RoomEvent, Status } from "./schema"

export interface Peer {
  userId: string
  joinedAt: number
  handle: string
  name: string
  avatar: string
}

export class PeerAggregator {
  private peers: Peer[]

  constructor(peers: Peer[]) {
    this.peers = peers
  }

  put(event: RoomEvent): void {
    const state = event.payload.peerState
    if (!state?.status) {
      return
    }

    const userId = event.ownerId || ""
    switch (state.status) {
      case Status.Joined:
        if (this.find(userId)) {
          return
        }

        this.peers.push({
          userId: userId,
          handle: "",
          name: "",
          avatar: "",
          joinedAt: event.createdAt || 0,
        })
        break
      case Status.Left:
        for (let i = 0; i < this.peers.length; i++) {
          if (this.peers[i].userId === userId) {
            this.peers.splice(i, 1)
            break
          }
        }
        break
    }
    return
  }

  // Since the number of messages is not too large, we don't use a separate map for simplicity.
  private find(id: string): Peer | null {
    for (const p of this.peers) {
      if (p.userId == id) {
        return p
      }
    }
    return null
  }
}
