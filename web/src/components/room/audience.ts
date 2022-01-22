import { reactive, readonly } from "vue"
import { Record } from "./record"
import { EventType, PayloadPeerState, PeerStatus } from "@/api/models"
import { genName } from "@/platform/gen"

interface Peer {
  id: string
  joinedAt: string
  name: string
}

export class PeersProcessor {
  private byId: { [key: string]: Peer }
  private ordered: Peer[]

  constructor() {
    this.byId = {}
    this.ordered = reactive([])
  }

  peers(): Peer[] {
    return readonly(this.ordered) as Peer[]
  }

  processRecords(records: Record[]): void {
    for (const record of records) {
      if (record.event.payload.type !== EventType.PeerState) {
        continue
      }

      const payload = record.event.payload.payload as PayloadPeerState
      const userId = record.event.ownerId || ""
      if (!payload.status) {
        continue
      }
      if (
        (record.forward && payload.status === PeerStatus.Joined) ||
        (!record.forward && payload.status === PeerStatus.Left)
      ) {
        if (this.byId[userId]) {
          continue
        }
        const p: Peer = {
          id: userId,
          joinedAt: record.event.createdAt || "",
          name: genName(userId),
        }
        this.byId[userId] = p
        this.ordered.push(p)
      }
      if (
        (record.forward && payload.status === PeerStatus.Left) ||
        (!record.forward && payload.status === PeerStatus.Joined)
      ) {
        delete this.byId[userId]
        for (let i = 0; i < this.ordered.length; i++) {
          if (this.ordered[i].id === userId) {
            this.ordered.splice(i, 1)
            break
          }
        }
      }
    }
  }
}
