import { gql } from "@apollo/client/core"
import { Client } from "./api"
import { requestSFUAccess, requestSFUAccessVariables } from "./schema"

export class RoomClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async requestSFUAccess(roomId: string): Promise<string> {
    const resp = await this.api.mutate<requestSFUAccess, requestSFUAccessVariables>({
      mutation: gql`
        mutation requestSFUAccess($roomId: String!) {
          requestSFUAccess(roomId: $roomId) {
            token
          }
        }
      `,
      variables: {
        roomId: roomId,
      },
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    return resp.data.requestSFUAccess.token
  }
}
