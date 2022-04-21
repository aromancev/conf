import { reactive } from "vue"
import { debounce } from "@/platform/debounce"
import { profileClient } from "@/api"
import { genName, genAvatar } from "@/platform/gen"
import { LRU } from "@/platform/containers"

const avatarSise = 64

export interface Profile {
  userId: string
  handle: string
  name: string
  avatar: string
}

interface Entry {
  profile: Profile
  fetched: boolean
}

export class ProfileRepository {
  private cache: LRU<Entry>
  private fetchDebounced: () => void

  constructor(cacheLimit: number, debounceFetchMS: number) {
    this.cache = new LRU<Entry>(cacheLimit)
    this.fetchDebounced = debounce(() => {
      this.fetch()
    }, debounceFetchMS)
  }

  profile(id: string): Profile {
    const fromCache = this.cache.get(id)
    if (fromCache) {
      return fromCache.profile
    }

    const entry = {
      profile: reactive<Profile>({
        userId: id,
        handle: "",
        name: genName(id),
        avatar: "",
      }),
      fetched: false,
    }
    genAvatar(id, avatarSise).then((src: string) => {
      entry.profile.avatar = src
    })
    this.cache.set(id, entry)
    this.fetchDebounced()

    return entry.profile
  }

  private async fetch(): Promise<void> {
    // Collect all the peer that do not have a profile fetched for them.
    const toFetch: Entry[] = []
    this.cache.forEach((entry: Entry): void => {
      if (entry.fetched) {
        return
      }
      toFetch.push(entry)
    })

    if (toFetch.length === 0) {
      return
    }

    // Fetch profiles. Only fetching one page.
    const iter = profileClient.fetch({ ownerIds: toFetch.map((e) => e.profile.userId) })
    const profiles = await iter.next()

    // Update info in all the entries.
    for (const prof of profiles) {
      const entry = this.cache.peek(prof.ownerId)
      if (!entry) {
        continue
      }
      entry.profile.handle = prof.handle || entry.profile.handle
      entry.profile.name = prof.displayName || entry.profile.name
    }

    // Mark all entries as fetched (even if they didn't have a profile).
    for (const entry of toFetch) {
      entry.fetched = true
    }
  }
}
