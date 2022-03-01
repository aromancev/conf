import { RoomEvent, Status } from "./schema"

export interface Peer {
  id: string
  joinedAt: number
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
          id: userId,
          joinedAt: event.createdAt || 0,
        })
        break
      case Status.Left:
        for (let i = 0; i < this.peers.length; i++) {
          if (this.peers[i].id === userId) {
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
      if (p.id == id) {
        return p
      }
    }
    return null
  }
}
