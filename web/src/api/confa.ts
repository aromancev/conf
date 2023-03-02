import { gql } from "@apollo/client/core"
import { Client, APIError, Code } from "./api"
import {
  createConfa,
  createConfaVariables,
  confas,
  confasVariables,
  ConfaUpdate,
  ConfaLookup,
  ConfaCursorInput,
  updateConfa,
  updateConfaVariables,
} from "./schema"
import { Confa } from "./models/confa"

export class ConfaIterator {
  private api: Client
  private lookup: ConfaLookup
  private cursor?: ConfaCursorInput

  constructor(api: Client, lookup: ConfaLookup) {
    this.api = api
    this.lookup = lookup
  }

  async next(): Promise<Confa[]> {
    const resp = await this.api.query<confas, confasVariables>({
      query: gql`
        query confas($where: ConfaLookup!, $limit: Int!, $cursor: ConfaCursorInput) {
          confas(where: $where, limit: $limit, cursor: $cursor) {
            items {
              id
              ownerId
              handle
              title
              description
            }
            next {
              id
              createdAt
            }
          }
        }
      `,
      variables: {
        where: this.lookup,
        limit: 100,
        cursor: this.cursor,
      },
    })

    this.cursor = resp.data.confas.next || undefined
    return resp.data.confas.items
  }
}

export class ConfaClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async create(request: ConfaUpdate = {}): Promise<Confa> {
    const resp = await this.api.mutate<createConfa, createConfaVariables>({
      mutation: gql`
        mutation createConfa($request: ConfaUpdate!) {
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

    await this.api.clearCache()

    return resp.data.createConfa
  }

  async update(where: ConfaLookup, request: ConfaUpdate = {}): Promise<Confa> {
    const resp = await this.api.mutate<updateConfa, updateConfaVariables>({
      mutation: gql`
        mutation updateConfa($where: ConfaLookup!, $request: ConfaUpdate!) {
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
