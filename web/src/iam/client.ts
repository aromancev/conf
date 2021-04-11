import { duration, Duration } from "@/platform/time"
import { api, Method, UnauthorisedError } from "@/platform/api"
import { userStore } from "./store"

const baseURL = "/iam/v1"
const minRefresh = 10 * Duration.second

interface Token {
  token: string
  expiresIn: number
}

class Client {
  refreshTimer = 0

  constructor() {
    this.refreshToken()
  }

  async login(email: string) {
    await api.do(
      Method.Post,
      baseURL + "/login",
      { email: email },
      {
        auth: false,
      },
    )
  }

  async session(token: string) {
    const resp = await api.do<Token>(
      Method.Post,
      baseURL + "/sessions",
      undefined,
      {
        auth: false,
        headers: {
          Authorization: "Bearer " + token,
        },
      },
    )
    this.setToken(resp.data)
  }

  private async refreshToken(): Promise<void> {
    clearTimeout(this.refreshTimer)

    try {
      const resp = await api.do<Token>(
        Method.Get,
        baseURL + "/token",
        undefined,
        {
          auth: false,
        },
      )
      this.setToken(resp.data)
    } catch (e) {
      if (!(e instanceof UnauthorisedError)) {
        throw e
      }
    }
  }

  private setToken(token: Token): void {
    api.setToken(token.token)
    userStore.login()
    this.scheduleRefresh(token.expiresIn)
  }

  private async scheduleRefresh(expiresIn: number): Promise<void> {
    const after = duration({ seconds: expiresIn }) - 2 * Duration.minute
    this.refreshTimer = setTimeout(
      this.refreshToken.bind(this),
      Math.max(after, minRefresh),
    )
  }
}

export const client: Client = new Client()
