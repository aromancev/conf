export class LRUMap<K, V> extends Map<K, V> {
  private cap: number

  constructor(cap: number) {
    super()
    this.cap = cap
  }

  peek(key: K): V | undefined {
    return super.get(key)
  }

  get(key: K): V | undefined {
    const entry = super.get(key)
    if (!entry) {
      return undefined
    }
    // Peek the entry, re-insert for LRU strategy.
    super.delete(key)
    super.set(key, entry)

    return entry
  }

  set(key: K, value: V): this {
    if (super.size >= this.cap) {
      // least-recently used cache eviction strategy
      const first = super.keys().next().value
      super.delete(first)
    }

    return super.set(key, value)
  }
}
