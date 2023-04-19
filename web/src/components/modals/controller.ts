import { reactive, readonly } from "vue"

export interface State {
  state?: string
}

export class ModalController<T extends string> {
  private readonly reactive: State
  private readonly readState: State
  private submitFunc?: (id?: string) => void

  constructor(state?: T) {
    this.reactive = reactive({
      state: state,
    })
    this.readState = readonly(this.reactive)
  }

  get state(): T {
    return this.readState.state as T
  }

  async set(state?: T): Promise<string | undefined> {
    if (this.submitFunc) {
      this.submitFunc(undefined)
    }
    this.reactive.state = state
    return new Promise((resolve) => {
      this.submitFunc = (id?: string) => {
        resolve(id)
        this.submitFunc = undefined
      }
    })
  }

  submit(id?: string): void {
    if (!this.submitFunc) {
      return
    }
    this.submitFunc(id)
    this.reactive.state = undefined
  }
}
