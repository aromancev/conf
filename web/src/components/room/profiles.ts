import { profileClient } from "@/api"
import { Peer as APIPeer } from "@/api/room"
import { debounce } from "@/platform/debounce"

interface Peer extends APIPeer {
  profileFetched?: boolean
}

export class ProfileHydrator {
  private peers: Peer[]
  private debounced: () => void

  constructor(peers: Peer[], debounceMS: number) {
    this.peers = peers
    this.debounced = debounce(() => {
      this._hydrate()
    }, debounceMS)
  }

  hydrate(): void {
    this.debounced()
  }

  private async _hydrate(): Promise<void> {
    // Collect all the peer that do not have a profile fetched for them.
    const toFetch = new Map<string, Peer>()
    for (const peer of this.peers) {
      if (peer.profileFetched) {
        continue
      }
      toFetch.set(peer.userId, peer)
    }

    if (toFetch.size === 0) {
      return
    }

    // Fetch profiles. Only fetching one page.
    const iter = profileClient.fetch({ ownerIds: Array.from(toFetch.keys()) })
    const profiles = await iter.next()

    // Update info in all the peers.
    for (const prof of profiles) {
      const peer = toFetch.get(prof.ownerId)
      if (!peer) {
        continue
      }
      peer.handle = prof.handle
      peer.name = prof.displayName
    }

    // Mark all the peers as fetched (even if the profile wasn't found).
    toFetch.forEach((prof: Peer) => {
      prof.profileFetched = true
    })
  }
}
