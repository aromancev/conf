import { gql } from "@apollo/client/core"
import { Client } from "./api"
import { createConfa, confas, confasVariables, ConfaInput } from "./schema"
import { Confa } from "./models"

class ConfaIterator {
  private api: Client
  private input: ConfaInput
  private from: string | null

  constructor(api: Client, input: ConfaInput) {
    this.api = api
    this.input = input
    this.from = null
  }

  async next(): Promise<Confa[]> {
    const resp = await this.api.query<confas, confasVariables>({
      query: gql`
        query confas($where: ConfaInput!, $from: String) {
          confas(where: $where, from: $from) {
            items {
              id
              ownerId
              handle
            }
            nextFrom
          }
        }
      `,
      variables: {
        where: this.input,
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

  async create(): Promise<Confa> {
    const resp = await this.api.mutate<createConfa>({
      mutation: gql`
        mutation createConfa {
          createConfa {
            id
            ownerId
            handle
          }
        }
      `,
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    const confa = resp.data.createConfa
    return {
      id: confa.id,
      ownerId: confa.ownerId,
      handle: confa.handle,
    }
  }

  async fetchOne(input: ConfaInput): Promise<Confa | null> {
    const iter = this.fetch(input)
    const confas = await iter.next()
    if (confas.length === 0) {
      return null
    }
    if (confas.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return confas[0]
  }

  fetch(input: ConfaInput): ConfaIterator {
    return new ConfaIterator(this.api, input)
  }
}
