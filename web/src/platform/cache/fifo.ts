export class FIFO<T> {
  private values: Map<string, T>
  private cap: number

  constructor(cap: number) {
    this.cap = cap
    this.values = new Map<string, T>()
  }

  get size(): number {
    return this.values.size
  }

  has(key: string): boolean {
    return this.values.has(key)
  }

  get(key: string): T | undefined {
    return this.values.get(key)
  }

  set(key: string, value: T) {
    if (this.values.size >= this.cap) {
      const first = this.values.keys().next().value
      this.values.delete(first)
    }

    this.values.set(key, value)
  }

  forEach(f: (entry: T) => void) {
    this.values.forEach(f)
  }
}
