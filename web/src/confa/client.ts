import { api, Method } from "@/platform/api"
import { Confa } from "./store"

const baseURL = "/confa/v1"

interface Lookup {
  handle?: string
}

export interface Confas {
  items: Confa[]
}

class Client {
  async create(): Promise<Confa> {
    const resp = await api.do<Confa>(Method.Post, baseURL + "/confas", {})
    return new Promise(resolve => {
      resolve(resp.data)
    })
  }

  async confas(lookup: Lookup): Promise<Confas> {
    const params = {} as Record<string, string>
    if (lookup.handle) {
      params.handle = lookup.handle
    }
    const resp = await api.do<Confas>(
      Method.Get,
      baseURL + "/confas",
      undefined,
      {
        params: params,
      },
    )
    return new Promise(resolve => {
      resolve(resp.data)
    })
  }
}

export const client: Client = new Client()
