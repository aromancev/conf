import { gql, DocumentNode } from "@apollo/client/core"
import { Client, FetchPolicy, Policy, APIError, Code } from "./api"
import {
  TalkLookup,
  TalkMask,
  ConfaLookup,
  createTalk,
  createTalkVariables,
  talks,
  talksVariables,
  updateTalk,
  updateTalkVariables,
} from "./schema"
import { Talk } from "./models"

const queryHydrated = gql`
  query talksHydrated($where: TalkLookup!, $limit: Int!, $from: String) {
    talks(where: $where, limit: $limit, from: $from) {
      items {
        id
        ownerId
        confaId
        roomId
        handle
        title
        description
      }
      nextFrom
    }
  }
`

const query = gql`
  query talks($where: TalkLookup!, $limit: Int!, $from: String) {
    talks(where: $where, limit: $limit, from: $from) {
      items {
        id
        ownerId
        confaId
        roomId
        handle
      }
      nextFrom
    }
  }
`

class TalkIterator {
  private api: Client
  private lookup: TalkLookup
  private from: string | null
  private query: DocumentNode
  private policy: FetchPolicy

  constructor(api: Client, lookup: TalkLookup, hydrated: boolean, policy: FetchPolicy) {
    this.api = api
    this.lookup = lookup
    this.from = null
    this.query = hydrated ? queryHydrated : query
    this.policy = policy
  }

  async next(): Promise<Talk[]> {
    const resp = await this.api.query<talks, talksVariables>({
      query: this.query,
      variables: {
        where: this.lookup,
        from: this.from,
        limit: 100,
      },
      fetchPolicy: this.policy,
    })

    this.from = resp.data.talks.nextFrom
    return resp.data.talks.items
  }
}

export class TalkClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async create(where: ConfaLookup, request: TalkMask): Promise<Talk> {
    const resp = await this.api.mutate<createTalk, createTalkVariables>({
      mutation: gql`
        mutation createTalk($where: ConfaLookup!, $request: TalkMask!) {
          createTalk(where: $where, request: $request) {
            id
            ownerId
            confaId
            roomId
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

    await this.api.clearCache()

    return resp.data.createTalk
  }

  async update(where: TalkLookup, request: TalkMask = {}): Promise<Talk> {
    const resp = await this.api.mutate<updateTalk, updateTalkVariables>({
      mutation: gql`
        mutation updateTalk($where: TalkLookup!, $request: TalkMask!) {
          updateTalk(where: $where, request: $request) {
            id
            confaId
            roomId
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
    return resp.data.updateTalk
  }

  async fetchOne(input: TalkLookup, hydrated = true): Promise<Talk | null> {
    const iter = this.fetch(input, hydrated)
    const talks = await iter.next()
    if (talks.length === 0) {
      throw new APIError(Code.NotFound, "Talk not found.")
    }
    if (talks.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return talks[0]
  }

  fetch(lookup: TalkLookup, hydrated = false, policy: FetchPolicy = Policy.CacheFirst): TalkIterator {
    return new TalkIterator(this.api, lookup, hydrated, policy)
  }
}
