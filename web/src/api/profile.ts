import { gql } from "@apollo/client/core"
import { Client, APIError, Code, errorCode } from "./api"
import {
  ProfileMask,
  ProfileLookup,
  profiles,
  profilesVariables,
  updateProfile,
  updateProfileVariables,
} from "./schema"
import { Profile, currentUser, profileStore } from "./models"

class ProfileIterator {
  private api: Client
  private lookup: ProfileLookup
  private from: string | null

  constructor(api: Client, lookup: ProfileLookup) {
    this.api = api
    this.lookup = lookup
    this.from = null
  }

  async next(): Promise<Profile[]> {
    const resp = await this.api.query<profiles, profilesVariables>({
      query: gql`
        query profiles($where: ProfileLookup!, $limit: Int!, $from: String) {
          profiles(where: $where, limit: $limit, from: $from) {
            items {
              id
              ownerId
              handle
              displayName
            }
            nextFrom
          }
        }
      `,
      variables: {
        where: this.lookup,
        limit: 100,
        from: this.from,
      },
    })

    this.from = resp.data.profiles.nextFrom
    return resp.data.profiles.items
  }
}

export class ProfileClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async update(request: ProfileMask = {}): Promise<Profile> {
    const resp = await this.api.mutate<updateProfile, updateProfileVariables>({
      mutation: gql`
        mutation updateProfile($request: ProfileMask!) {
          updateProfile(request: $request) {
            id
            ownerId
            handle
            displayName
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
    return resp.data.updateProfile
  }

  async fetchOne(input: ProfileLookup): Promise<Profile> {
    const iter = this.fetch(input)
    const confas = await iter.next()
    if (confas.length === 0) {
      throw new APIError(Code.NotFound, "Profile not found.")
    }
    if (confas.length > 1) {
      throw new Error("Unexpected response from API.")
    }
    return confas[0]
  }

  async refreshProfile(): Promise<void> {
    try {
      const profile = await this.fetchOne({ ownerIds: [currentUser.id] })
      profileStore.update(profile)
    } catch (e) {
      switch (errorCode(e)) {
        case Code.NotFound:
          profileStore.update({
            id: "",
            ownerId: currentUser.id,
            handle: "",
            displayName: "",
          })
          break
      }
    }
  }

  fetch(lookup: ProfileLookup): ProfileIterator {
    return new ProfileIterator(this.api, lookup)
  }
}
