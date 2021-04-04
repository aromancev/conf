import { duration, Duration } from "@/platform/time"
import { api, Method, API, UnauthorisedError } from "@/platform/api"

const baseURL = "/iam/v1"
const minRefresh = 10 * Duration.second

interface Token {
  token: string
  expiresIn: number
}

export class IAM {
  api: API
  refreshTimer = 0


  constructor(api: API) {
    this.api = api
    this.refreshToken()
  }

  async login(email: string) {
    await this.api.do(Method.Post, baseURL + "/login", { email: email }, {
      auth: false,
    })
  }

  async session(token: string) {
    const resp = await this.api.do<Token>(Method.Post, baseURL + "/sessions", undefined, {
      auth: false,
      headers: {
        Authorization: "Bearer " + token,
      },
    })

    this.api.setToken(resp.data.token)
    this.scheduleRefresh(resp.data.expiresIn)
  }

  private async refreshToken(): Promise<void> {
    clearTimeout(this.refreshTimer)
    
    try {
      const resp = await this.api.do<Token>(Method.Get, baseURL + "/token", undefined, {
        auth: false,
      })
      this.api.setToken(resp.data.token)
      this.scheduleRefresh(resp.data.expiresIn)
    } catch (e) {
     if (!(e instanceof UnauthorisedError)) {
       throw e
     }
    }
  }

  private async scheduleRefresh(expiresIn: number): Promise<void> {
    const after = duration({seconds: expiresIn}) - 5 * Duration.minute
    this.refreshTimer = setTimeout(this.refreshToken, Math.max(after, minRefresh))
  }
}

export const iam: IAM = new IAM(api)
