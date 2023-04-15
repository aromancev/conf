import { gql } from "@apollo/client/core"
import { Client, FetchPolicy, APIError, Code } from "./api"
import {
  TalkLookup,
  TalkUpdate,
  TalkCursorInput,
  ConfaLookup,
  createTalk,
  createTalkVariables,
  talks,
  talksVariables,
  talksHydrated,
  talksHydratedVariables,
  updateTalk,
  updateTalkVariables,
  startTalkRecording,
  startTalkRecordingVariables,
  stopTalkRecording,
  stopTalkRecordingVariables,
  deleteTalk,
  deleteTalkVariables,
} from "./schema"
import { Talk } from "./models/talk"

interface FetchParams {
  policy: FetchPolicy
  hydrated: boolean
}

interface OptionalFetchParams {
  policy?: FetchPolicy
  hydrated?: boolean
}

const defaultParams: FetchParams = {
  policy: "cache-first",
  hydrated: false,
}

export class TalkIterator {
  private api: Client
  private lookup: TalkLookup
  private cursor?: TalkCursorInput
  private params: FetchParams

  constructor(api: Client, lookup: TalkLookup, params?: OptionalFetchParams) {
    this.api = api
    this.lookup = lookup
    this.params = {
      ...defaultParams,
      ...params,
    }
  }

  async next(): Promise<Talk[]> {
    if (this.params.hydrated) {
      const resp = await this.api.query<talksHydrated, talksHydratedVariables>({
        query: gql`
          query talksHydrated($where: TalkLookup!, $limit: Int!, $cursor: TalkCursorInput) {
            talks(where: $where, limit: $limit, cursor: $cursor) {
              items {
                id
                ownerId
                confaId
                roomId
                handle
                title
                description
                state
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
          cursor: this.cursor,
          limit: 100,
        },
        fetchPolicy: this.params.policy,
      })
      this.cursor = resp.data.talks.next || undefined
      return resp.data.talks.items
    }
    const resp = await this.api.query<talks, talksVariables>({
      query: gql`
        query talks($where: TalkLookup!, $limit: Int!, $cursor: TalkCursorInput) {
          talks(where: $where, limit: $limit, cursor: $cursor) {
            items {
              id
              ownerId
              confaId
              roomId
              handle
              title
              state
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
        cursor: this.cursor,
        limit: 100,
      },
      fetchPolicy: this.params.policy,
    })
    this.cursor = resp.data.talks.next || undefined
    return resp.data.talks.items
  }
}

export class TalkClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async create(where: ConfaLookup, request: TalkUpdate): Promise<Talk> {
    const resp = await this.api.mutate<createTalk, createTalkVariables>({
      mutation: gql`
        mutation createTalk($where: ConfaLookup!, $request: TalkUpdate!) {
          createTalk(where: $where, request: $request) {
            id
            ownerId
            confaId
            roomId
            handle
            title
            description
            state
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

  async update(where: TalkLookup, request: TalkUpdate = {}): Promise<Talk> {
    const resp = await this.api.mutate<updateTalk, updateTalkVariables>({
      mutation: gql`
        mutation updateTalk($where: TalkLookup!, $request: TalkUpdate!) {
          updateTalk(where: $where, request: $request) {
            id
            confaId
            roomId
            ownerId
            handle
            title
            description
            state
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

  async startRecording(where: TalkLookup): Promise<Talk> {
    const resp = await this.api.mutate<startTalkRecording, startTalkRecordingVariables>({
      mutation: gql`
        mutation startTalkRecording($where: TalkLookup!) {
          startTalkRecording(where: $where) {
            id
            ownerId
            confaId
            roomId
            handle
            title
            state
          }
        }
      `,
      variables: {
        where: where,
      },
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    return resp.data.startTalkRecording
  }

  async stopRecording(where: TalkLookup): Promise<Talk> {
    const resp = await this.api.mutate<stopTalkRecording, stopTalkRecordingVariables>({
      mutation: gql`
        mutation stopTalkRecording($where: TalkLookup!) {
          stopTalkRecording(where: $where) {
            id
            ownerId
            confaId
            roomId
            handle
            title
            state
          }
        }
      `,
      variables: {
        where: where,
      },
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    return resp.data.stopTalkRecording
  }

  async fetchOne(input: TalkLookup, params?: OptionalFetchParams): Promise<Talk> {
    const iter = this.fetch(input, params)
    const talks = await iter.next()
    if (talks.length === 0) {
      throw new APIError(Code.NotFound, "Talk not found.")
    }
    if (talks.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return talks[0]
  }

  fetch(lookup: TalkLookup, params?: OptionalFetchParams): TalkIterator {
    return new TalkIterator(this.api, lookup, params)
  }

  async delete(where: TalkLookup): Promise<void> {
    await this.api.mutate<deleteTalk, deleteTalkVariables>({
      mutation: gql`
        mutation deleteTalk($where: TalkLookup!) {
          deleteTalk(where: $where) {
            deletedCount
          }
        }
      `,
      variables: {
        where: where,
      },
    })

    await this.api.clearCache()
  }
}
