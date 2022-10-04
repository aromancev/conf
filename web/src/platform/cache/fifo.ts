export class FIFO<K, V> extends Map<K, V> {
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
