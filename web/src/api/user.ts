import { gql } from "@apollo/client/core"
import { Client, FetchPolicy } from "./api"
import { user } from "./schema"
import { User } from "./models/user"
export * from "./models/user"

interface OptionalFetchParams {
  policy?: FetchPolicy
}

interface FetchParams {
  policy: FetchPolicy
}

const defaultParams: FetchParams = {
  policy: "cache-first",
}

export class UserClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async fetchCurrent(params?: OptionalFetchParams): Promise<User> {
    params = {
      ...defaultParams,
      ...params,
    }
    const resp = await this.api.query<user>({
      query: gql`
        query user {
          user {
            id
            identifiers {
              platform
              value
            }
            hasPassword
          }
        }
      `,
      fetchPolicy: params.policy,
    })
    if (!resp.data) {
      throw new Error("No data in response.")
    }
    return resp.data.user
  }
}
