export class FIFOMap<K, V> extends Map<K, V> {
  private cap: number

  constructor(cap: number) {
    super()
    this.cap = cap
  }

  set(key: K, value: V): this {
    if (this.size >= this.cap) {
      const first = this.keys().next().value
      this.delete(first)
    }

    return super.set(key, value)
  }
}

export class FIFOSet<V> extends Set<V> {
  private cap: number

  constructor(cap: number) {
    super()
    this.cap = cap
  }

  add(value: V): this {
    if (this.size >= this.cap) {
      const first = super.keys().next().value
      super.delete(first)
    }

    return super.add(value)
  }
}
