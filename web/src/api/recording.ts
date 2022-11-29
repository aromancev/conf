import { gql } from "@apollo/client/core"
import { Client, FetchPolicy, APIError, Code } from "./api"
import { RecordingLookup, RecordingFromInput, recordings, recordingsVariables } from "./schema"
import { Recording } from "./models"

interface OptionalFetchParams {
  policy?: FetchPolicy
}

interface FetchParams {
  policy: FetchPolicy
}

const defaultParams: FetchParams = {
  policy: "cache-first",
}

class RecordingIterator {
  private api: Client
  private lookup: RecordingLookup
  private params: FetchParams
  private from: RecordingFromInput | null

  constructor(api: Client, lookup: RecordingLookup, params?: OptionalFetchParams) {
    this.api = api
    this.lookup = lookup
    this.from = null
    this.params = {
      ...defaultParams,
      ...params,
    }
  }

  async next(): Promise<Recording[]> {
    const resp = await this.api.query<recordings, recordingsVariables>({
      query: gql`
        query recordings($where: RecordingLookup!, $limit: Int!, $from: RecordingFromInput) {
          recordings(where: $where, limit: $limit, from: $from) {
            items {
              key
              roomId
              status
              createdAt
              startedAt
              stoppedAt
            }
            nextFrom {
              key
            }
          }
        }
      `,
      variables: {
        where: this.lookup,
        limit: 100,
        from: this.from,
      },
      fetchPolicy: this.params.policy,
    })

    this.from = resp.data.recordings.nextFrom
    return resp.data.recordings.items
  }
}

export class RecordingClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async fetchOne(lookup: RecordingLookup, params?: OptionalFetchParams): Promise<Recording> {
    const iter = this.fetch(lookup, params)
    const confas = await iter.next()
    if (confas.length === 0) {
      throw new APIError(Code.NotFound, "Confa not found.")
    }
    if (confas.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return confas[0]
  }

  fetch(lookup: RecordingLookup, params?: OptionalFetchParams): RecordingIterator {
    return new RecordingIterator(this.api, lookup, params)
  }
}
