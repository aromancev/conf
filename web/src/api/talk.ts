import { gql } from "@apollo/client/core"
import { Client } from "./api"
import { TalkInput, createTalk, createTalkVariables, startTalk, startTalkVariables, talks, talksVariables } from "./schema"
import { Talk } from "./models"

class TalkIterator {
  private api: Client
  private input: TalkInput
  private from: string | null

  constructor(api: Client, input: TalkInput) {
    this.api = api
    this.input = input
    this.from = null
  }

  async next(): Promise<Talk[]> {
    const resp = await this.api.query<talks, talksVariables>({
      query: gql`
        query talks($where: TalkInput!, $from: String) {
          talks(where: $where, from: $from) {
            items {
              id
              ownerId
              confaId
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
            handle
          }
        }
      `,
      variables: {
        confaId: confaId,
      }
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }

    const talk = resp.data.createTalk
    return {
      id: talk.id,
      ownerId: talk.ownerId,
      confaId: talk.confaId,
      handle: talk.handle,
    }
  }

  async start(talkId: string): Promise<void> {
    const resp = await this.api.mutate<startTalk, startTalkVariables>({
      mutation: gql`
        mutation startTalk($talkId: String!) {
          startTalk(talkId: $talkId)
        }
      `,
      variables: {
        talkId: talkId,
      }
    })
  }

  async fetchOne(input: TalkInput): Promise<Talk | null> {
    const iter = this.fetch(input)
    const talks = await iter.next()
    if (talks.length === 0) {
      return null
    }
    if (talks.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return talks[0]
  }

  fetch(input: TalkInput): TalkIterator {
    return new TalkIterator(this.api, input)
  }
}
