import { RoomEvent, PeerStatus } from "@/api/room/schema"
import { FIFOSet } from "@/platform/cache"

const MAX_SESSIONS_PER_PEER = 10
const MAX_GHOST_SESSIONS = 100
const CLAP_TIMEOUT_MS = 7 * 1000

export type StatusKey = "clap"

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
  private readonly peers: Map<string, Peer>
  private readonly statuses: Map<string, Status>
  private readonly repo: Repo
  // ghostSessionIds stores sessions that seen `leave` event, but didn't see `join` even yet.
  // It's important to store those to ignore the `join` even that might come later.
  // It might happen because the backend doesn't give a guarantee on ordering of events.
  private ghostSessionIds: Set<string>

  constructor(peers: Map<string, Peer>, statuses: Map<string, Status>, repo: Repo) {
    this.peers = peers
    this.statuses = statuses
    this.repo = repo
    this.ghostSessionIds = new FIFOSet(MAX_GHOST_SESSIONS)
  }

  setTime(time: number): void {
    for (const [id, status] of this.statuses) {
      if (status.clap && time > status.clap.startAt + CLAP_TIMEOUT_MS) {
        status.clap = undefined
      }
      if (isStatusEmpty(status)) {
        this.statuses.delete(id)
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
      if (!this.peers.has(reaction.fromId)) {
        return
      }
      let state = this.statuses.get(reaction.fromId)
      if (!state) {
        state = {}
        this.statuses.set(reaction.fromId, state)
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

  private join(userId: string, sessionId: string): void {
    if (this.ghostSessionIds.has(sessionId)) {
      // This session have seen a `leave` event before. Skipping.
      return
    }

    const peer = this.peers.get(userId)
    if (peer) {
      peer.sessionIds.add(sessionId)
      return
    }

    const sessions = new FIFOSet<string>(MAX_SESSIONS_PER_PEER)
    sessions.add(sessionId)
    this.peers.set(userId, {
      userId: userId,
      profile: this.repo.profile(userId),
      sessionIds: sessions,
    })
  }

  private leave(userId: string, sessionId: string): void {
    const peer = this.peers.get(userId)
    if (!peer || !peer.sessionIds.has(sessionId)) {
      // No user exists for this `leave` event. This is a ghost session that we might see later.
      this.ghostSessionIds.add(sessionId)
      return
    }
    peer.sessionIds.delete(sessionId)
    if (peer.sessionIds.size === 0) {
      this.peers.delete(userId)
      this.statuses.delete(userId)
    }
  }

  reset(): void {
    this.peers.clear()
    this.statuses.clear()
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
