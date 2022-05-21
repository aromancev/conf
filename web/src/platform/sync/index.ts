export class WaitGroup {
  private promise: Promise<void>
  private resolve: (() => void) | null
  private counter: number
  private joined: boolean

  constructor() {
    this.joined = false
    this.counter = 0
    this.resolve = null
    this.promise = new Promise<void>((resolve) => {
      this.resolve = resolve
    })
  }

  add(val: number): void {
    if (this.joined) {
      throw new Error("Waitgroup trying to add after join")
    }
    this.counter += val
  }

  done(): void {
    this.counter--
    if (this.joined && this.counter <= 0) {
      if (this.resolve) this.resolve()
    }
  }

  async join(): Promise<void> {
    if (this.counter <= 0) {
      return
    }
    this.joined = true
    return this.promise
  }
}

export function debounce(fn: () => void, ms: number): () => void {
  let timeoutId: ReturnType<typeof setTimeout>
  return () => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => {
      fn()
    }, ms)
  }
}

export function throttle(fn: () => void, ms: number): () => void {
  let timeoutId: ReturnType<typeof setTimeout>
  let lastCalled: number
  return () => {
    clearTimeout(timeoutId)
    const sinceLastCall = Date.now() - lastCalled
    if (sinceLastCall >= ms) {
      fn()
      lastCalled = Date.now()
    } else {
      timeoutId = setTimeout(() => {
        fn()
        lastCalled = Date.now()
      }, ms - sinceLastCall)
    }
  }
}
