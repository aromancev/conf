import { reactive, readonly } from "vue"

export interface State {
  current?: string
}

export class ModalController<T extends string> {
  private _state: State
  private readState: State
  private submitFunc?: (id?: string) => void

  constructor(state?: T) {
    this._state = reactive({
      current: state,
    })
    this.readState = readonly(this._state)
  }

  get state(): { current: T } {
    return this.readState as { current: T }
  }

  async set(state?: T): Promise<string | undefined> {
    if (this.submitFunc) {
      this.submitFunc(undefined)
    }
    this._state.current = state
    return new Promise((resolve) => {
      this.submitFunc = (id?: string) => {
        resolve(id)
        this.submitFunc = undefined
      }
    })
  }

  submit(id: string): void {
    if (!this.submitFunc) {
      return
    }
    this.submitFunc(id)
    this._state.current = undefined
  }
}
