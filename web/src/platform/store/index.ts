import { reactive, readonly } from "vue"

export abstract class Store<T extends object> {
  protected readonly reactive: T
  private readonly readState: T

  constructor(state: T) {
    this.reactive = reactive<T>(state) as T
    this.readState = readonly(this.reactive) as T
  }

  get state(): T {
    return this.readState
  }
}
