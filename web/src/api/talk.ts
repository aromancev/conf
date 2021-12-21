import { gql, DocumentNode } from "@apollo/client/core"
import { Client, FetchPolicy, Policy } from "./api"
import {
  TalkInput,
  createTalk,
  createTalkVariables,
  startTalk,
  startTalkVariables,
  talks,
  talksVariables,
} from "./schema"
import { Talk } from "./models"

const queryHydrated = gql`
query talksHydrated($where: TalkInput!, $from: String) {
    talks(where: $where, from: $from) {
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

const query = gql`
query talks($where: TalkInput!, $from: String) {
  talks(where: $where, from: $from) {
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
  private input: TalkInput
  private from: string | null
  private query: DocumentNode
  private policy: FetchPolicy

  constructor(api: Client, input: TalkInput, hydrated: boolean, policy: FetchPolicy) {
    this.api = api
    this.input = input
    this.from = null
    this.query = hydrated ? queryHydrated : query
    this.policy = policy
  }

  async next(): Promise<Talk[]> {
    const resp = await this.api.query<talks, talksVariables>({
      query: this.query,
      variables: {
        where: this.input,
        from: this.from,
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

  async create(confaId: string): Promise<Talk> {
    const resp = await this.api.mutate<createTalk, createTalkVariables>({
      mutation: gql`
        mutation createTalk($confaId: String!) {
          createTalk(confaId: $confaId) {
            id
            ownerId
            confaId
            roomId
            handle
          }
        }
      `,
      variables: {
        confaId: confaId,
      },
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }

    await this.api.clearCache()

    return resp.data.createTalk
  }

  async start(talkId: string): Promise<void> {
    await this.api.mutate<startTalk, startTalkVariables>({
      mutation: gql`
        mutation startTalk($talkId: String!) {
          startTalk(talkId: $talkId)
        }
      `,
      variables: {
        talkId: talkId,
      },
    })
  }

  async fetchOne(input: TalkInput, hydrated = true): Promise<Talk | null> {
    const iter = this.fetch(input, hydrated)
    const talks = await iter.next()
    if (talks.length === 0) {
      return null
    }
    if (talks.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return talks[0]
  }

  fetch(input: TalkInput, hydrated = false, policy: FetchPolicy = Policy.CacheFirst): TalkIterator {
    return new TalkIterator(this.api, input, hydrated, policy)
  }
}
