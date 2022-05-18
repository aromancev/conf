export class LRU<T> {
  private values: Map<string, T>
  private cap: number

  constructor(cap: number) {
    this.cap = cap
    this.values = new Map<string, T>()
  }

  peek(key: string): T | undefined {
    return this.values.get(key)
  }

  get(key: string): T | undefined {
    const entry = this.values.get(key)
    if (!entry) {
      return undefined
    }
    // Peek the entry, re-insert for LRU strategy.
    this.values.delete(key)
    this.values.set(key, entry)

    return entry
  }

  set(key: string, value: T) {
    if (this.values.size >= this.cap) {
      // least-recently used cache eviction strategy
      const first = this.values.keys().next().value
      this.values.delete(first)
    }

    this.values.set(key, value)
  }

  forEach(f: (entry: T) => void) {
    this.values.forEach(f)
  }
}
