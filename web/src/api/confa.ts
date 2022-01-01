import { gql } from "@apollo/client/core"
import { Client, APIError, Code } from "./api"
import {
  createConfa,
  createConfaVariables,
  confas,
  confasVariables,
  ConfaMask,
  ConfaLookup,
  updateConfa,
  updateConfaVariables,
} from "./schema"
import { Confa } from "./models"

class ConfaIterator {
  private api: Client
  private lookup: ConfaLookup
  private from: string | null

  constructor(api: Client, lookup: ConfaLookup) {
    this.api = api
    this.lookup = lookup
    this.from = null
  }

  async next(): Promise<Confa[]> {
    const resp = await this.api.query<confas, confasVariables>({
      query: gql`
        query confas($where: ConfaLookup!, $from: ID) {
          confas(where: $where, from: $from) {
            items {
              id
              ownerId
              handle
              title
              description
            }
            nextFrom
          }
        }
      `,
      variables: {
        where: this.lookup,
        from: this.from,
      },
    })

    this.from = resp.data.confas.nextFrom
    return resp.data.confas.items
  }
}

export class ConfaClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async create(request: ConfaMask = {}): Promise<Confa> {
    const resp = await this.api.mutate<createConfa, createConfaVariables>({
      mutation: gql`
        mutation createConfa($request: ConfaMask!) {
          createConfa(request: $request) {
            id
            ownerId
            handle
            title
            description
          }
        }
      `,
      variables: {
        request: request,
      },
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    return resp.data.createConfa
  }

  async update(where: ConfaLookup, request: ConfaMask = {}): Promise<Confa> {
    const resp = await this.api.mutate<updateConfa, updateConfaVariables>({
      mutation: gql`
        mutation updateConfa($where: ConfaLookup!, $request: ConfaMask!) {
          updateConfa(where: $where, request: $request) {
            id
            ownerId
            handle
            title
            description
          }
        }
      `,
      variables: {
        where: where,
        request: request,
      },
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    return resp.data.updateConfa
  }

  async fetchOne(input: ConfaLookup): Promise<Confa> {
    const iter = this.fetch(input)
    const confas = await iter.next()
    if (confas.length === 0) {
      throw new APIError(Code.NotFound, "Confa not found.")
    }
    if (confas.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return confas[0]
  }

  fetch(lookup: ConfaLookup): ConfaIterator {
    return new ConfaIterator(this.api, lookup)
  }
}
