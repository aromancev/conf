import { Codes } from "@/platform/http/http"
import { AxiosInstance } from "axios"
import axios from "axios"

export enum Methods {
  get = "GET",
  post = "POST",
}

export interface Params {
  headers?: Record<string, string>
}

export class API {
  client: AxiosInstance

  constructor() {
    let protocol = "https"
    if (process.env.NODE_ENV == "development") {
      protocol = "http"
    }
    this.client = axios.create({
      baseURL: `${protocol}://${window.location.hostname}/api`,
    })
  }

  async do(method: Methods, url: string, data?: object, params?: Params) {
    const resp = await this.client.request({
      method: method,
      url: url,
      data: data,
      headers: params?.headers,
    })
    if (resp.status !== Codes.OK && resp.status !== Codes.Created) {
      throw "unexpected code: " + resp.status
    }
  }
}

export const api: API = new API()
