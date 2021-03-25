import { api, API } from "./api"

export class IAM {
  api: API

  constructor(api: API) {
    this.api = api
  }

  async login(email: string) {
    await this.api.post("/iam/v1/login", { email: email })
  }
}

export const iam: IAM = new IAM(api)
