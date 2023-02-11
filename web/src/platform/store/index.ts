import { reactive, readonly } from "vue"

export abstract class Store<T extends object> {
  protected _state: T
  protected readonly _readState: T

  constructor(data: T) {
    this._state = reactive(data) as T
    this._readState = readonly(this._state) as T
  }

  get state(): T {
    return this._readState
  }
}
