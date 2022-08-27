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

export interface ThrottleParams {
  delayMs: number
}

export class Throttler<T> {
  func?: (() => Promise<T>) | (() => T)

  private params: ThrottleParams
  private promise: Promise<T>
  private resolve?: (t: T) => void
  private isReady: boolean
  private doAfterReady: boolean

  constructor(params: ThrottleParams) {
    this.params = params
    this.promise = new Promise((r) => {
      this.resolve = r
    })
    this.isReady = true
    this.doAfterReady = false
  }

  async do(): Promise<T> {
    if (!this.isReady) {
      this.doAfterReady = true
      return this.promise
    }

    if (!this.func) {
      throw new Error("Throttler function not defined.")
    }

    this.isReady = false
    const prevResolve = this.resolve
    this.promise = new Promise((r) => {
      this.resolve = r
    })
    const t = await this.func()
    if (prevResolve) {
      prevResolve(t)
    }

    setTimeout(() => {
      this.isReady = true
      if (this.doAfterReady) {
        this.doAfterReady = false
        this.do()
      }
    }, this.params.delayMs)

    return t
  }
}
