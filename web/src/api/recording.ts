import { gql } from "@apollo/client/core"
import { Client, APIError, Code } from "./api"
import { RecordingLookup, RecordingFromInput, recordings, recordingsVariables } from "./schema"
import { Recording } from "./models"

class RecordingIterator {
  private api: Client
  private lookup: RecordingLookup
  private from: RecordingFromInput | null

  constructor(api: Client, lookup: RecordingLookup) {
    this.api = api
    this.lookup = lookup
    this.from = null
  }

  async next(): Promise<Recording[]> {
    const resp = await this.api.query<recordings, recordingsVariables>({
      query: gql`
        query recordings($where: RecordingLookup!, $limit: Int!, $from: RecordingFromInput) {
          recordings(where: $where, limit: $limit, from: $from) {
            items {
              key
              roomId
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

  async fetchOne(lookup: RecordingLookup): Promise<Recording> {
    const iter = this.fetch(lookup)
    const confas = await iter.next()
    if (confas.length === 0) {
      throw new APIError(Code.NotFound, "Confa not found.")
    }
    if (confas.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return confas[0]
  }

  fetch(lookup: RecordingLookup): RecordingIterator {
    return new RecordingIterator(this.api, lookup)
  }
}
