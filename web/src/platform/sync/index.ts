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

export function sleep(sig: AbortSignal, ms: number): Promise<void> {
  sig.throwIfAborted()
  return new Promise((res, rej) => {
    const t = setTimeout(res, ms)
    sig.addEventListener("abort", () => {
      rej(sig.reason)
      clearTimeout(t)
    })
  })
}

export function repeat(sig: AbortSignal, ms: number, f: () => void): void {
  sig.throwIfAborted()

  const i = setInterval(f, ms)
  sig.addEventListener("abort", () => {
    clearInterval(i)
  })
}

export class Semaphor {
  private count: number
  private waiting: Set<() => void>

  constructor(cap: number) {
    this.count = cap
    this.waiting = new Set()
  }

  async acquire(sig: AbortSignal): Promise<void> {
    sig.throwIfAborted()

    if (this.count > 0) {
      this.count--
      // Aborted after acquired. Resolve the first waiter.
      // Note even if the resolved waiter calls abort after,
      // nothing will happen because the promise will be permanently in resolved state at this point.
      sig.addEventListener("abort", () => this.releaseFirstInQueue())
      return
    }
    return new Promise((res, rej) => {
      this.waiting.add(res)
      // Aborted while waiting. Just remove it from the waiting list and reject the promise.
      sig.addEventListener("abort", () => {
        this.waiting.delete(res)
        rej(sig.reason)
      })
    })
  }

  private releaseFirstInQueue(): void {
    if (!this.waiting.size) {
      this.count++
      return
    }
    const res = this.waiting.keys().next().value
    this.waiting.delete(res)
    res()
  }
}

export class Backoff {
  private readonly min: number
  private readonly max: number
  private readonly factor: number
  private readonly jitter: number
  private counter: number

  constructor(factor: number, min: number, max: number, jitter?: number) {
    this.factor = factor
    this.min = min
    this.max = max
    this.jitter = jitter || 0
    this.counter = 0
  }

  get retries(): number {
    return this.counter
  }

  reset(): void {
    this.counter = 0
  }

  next(): number {
    let delay = this.min * Math.pow(this.factor, this.counter)
    if (delay > this.max) {
      delay = this.max
    }
    const absJitter = delay * this.jitter
    this.counter++
    // From -absJitter to absJitter.
    return delay + absJitter - absJitter * 2 * Math.random()
  }
}
