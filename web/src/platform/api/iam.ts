import { api, API, Methods } from "./api"

export class IAM {
  api: API

  constructor(api: API) {
    this.api = api
  }

  async login(email: string) {
    await this.api.do(Methods.post, "/iam/v1/login", { email: email })
  }

  async session(token: string) {
    await this.api.do(Methods.post, "/iam/v1/session", undefined, {
      headers: {
        Authorization: "Bearer " + token,
      },
    })
  }
}

export const iam: IAM = new IAM(api)
