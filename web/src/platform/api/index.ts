import { AxiosInstance } from "axios"
import axios from "axios"

export enum Method {
  Get = "GET",
  Post = "POST",
}

export enum Status {
  OK = 200,
  Created = 201,
  Unauthorized = 401,
}

export interface Params {
  headers?: Record<string, string>
  auth?: boolean
}

export interface Response<T> {
  data: T
  status: Status
}

export class UnauthorisedError extends Error {
  constructor() {
    super("Unauthorised request")
  }
}

const defaultParams: Params = {
  auth: true,
  headers: {},
}

class API {
  private client: AxiosInstance
  private token: Promise<string> | null = null
  private tokenResolve: ((token: string) => void) | null = null

  constructor() {
    let protocol = "https"
    if (process.env.NODE_ENV == "development") {
      protocol = "http"
    }
    this.client = axios.create({
      baseURL: `${protocol}://${window.location.hostname}/api`,
    })

    this.resetToken()
  }

  async do<T>(
    method: Method,
    url: string,
    data?: object,
    params?: Params,
  ): Promise<Response<T>> {
    params = Object.assign(defaultParams, params) as Params
    params.headers = params.headers || {}

    if (params.auth) {
      params.headers.Authorization = "Bearer" + (await this.token)
    }

    const resp = await this.client.request({
      method: method,
      url: url,
      data: data,
      headers: params.headers,
    })
    switch (resp.status) {
      case (Status.OK, Status.Created): {
        return new Promise<Response<T>>(resolve => {
          resolve({
            data: resp.data,
            status: resp.status,
          } as Response<T>)
        })
      }
      case Status.Unauthorized: {
        this.resetToken()
        throw new UnauthorisedError()
      }
      default: {
        throw new Error("Unexpected code: " + resp.status)
      }
    }
  }

  setToken(token: string): void {
    this.token = new Promise<string>(resolve => {
      resolve(token)
    })
    if (this.tokenResolve) {
      this.tokenResolve(token)
    }
  }

  resetToken(): void {
    if (this.tokenResolve) {
      this.tokenResolve("")
    }
    this.token = new Promise<string>(resolve => {
      this.tokenResolve = resolve
    })
  }
}

export const api: API = new API()
