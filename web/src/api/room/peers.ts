import { Event, EventType, PayloadPeerState, PeerStatus } from "@/api/models"

export interface Peer {
  id: string
  joinedAt: string
}

export class PeerAggregator {
  private peers: Peer[]

  constructor(peers: Peer[]) {
    this.peers = peers
  }

  put(event: Event, forwards: boolean): void {
    if (event.payload.type !== EventType.PeerState) {
      return
    }

    const payload = event.payload.payload as PayloadPeerState
    const userId = event.ownerId || ""
    if (!payload.status) {
      return
    }

    if ((forwards && payload.status === PeerStatus.Joined) || (!forwards && payload.status === PeerStatus.Left)) {
      if (this.find(userId)) {
        return
      }

      this.peers.push({
        id: userId,
        joinedAt: event.createdAt || "",
      })
      return
    }

    if ((forwards && payload.status === PeerStatus.Left) || (!forwards && payload.status === PeerStatus.Joined)) {
      for (let i = 0; i < this.peers.length; i++) {
        if (this.peers[i].id === userId) {
          this.peers.splice(i, 1)
          break
        }
      }
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
