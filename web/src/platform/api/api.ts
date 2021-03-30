import { Codes } from "@/platform/http/http"
import { AxiosInstance } from "axios"
import axios from "axios"

export class API {
  client: AxiosInstance

  constructor() {
    let protocol = "http"
    if (process.env.NODE_ENV == "development") {
      protocol = "http"
    }
    this.client = axios.create({
      baseURL: `${protocol}://${window.location.hostname}/api`
    })
  }

  async post(url: string, data?: object) {
    const resp = await this.client.post(url, data)
    if (resp.status !== Codes.OK && resp.status !== Codes.Created) {
      throw "unexpected code: " + resp.status
    }
  }
}

export const api: API = new API()
