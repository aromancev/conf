import { reactive, readonly } from "vue"

export abstract class Store<T extends object> {
  protected readonly reactive: T
  private readonly readState: T
  private readonly init: T

  constructor(state: T) {
    this.init = Object.assign({}, state)
    this.reactive = reactive<T>(state) as T
    this.readState = readonly(this.reactive) as T
  }

  get state(): T {
    return this.readState
  }

  set(state: T): void {
    Object.assign(this.reactive, state)
  }

  reset(): void {
    Object.assign(this.reactive, this.init)
  }
}
