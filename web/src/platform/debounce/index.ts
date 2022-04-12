export function debounce(fn: () => void, ms: number): () => void {
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
