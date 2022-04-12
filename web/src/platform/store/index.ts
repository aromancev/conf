import { reactive, readonly } from "vue"

export abstract class Store<T extends object> {
  protected _state: T

  constructor() {
    const data = this.data()
    this._state = reactive(data) as T
  }

  protected abstract data(): T

  public state(): T {
    return readonly(this._state) as T
  }
}
