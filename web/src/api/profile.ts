import { gql } from "@apollo/client/core"
import { Client, APIError, Code, errorCode, FetchPolicy } from "./api"
import {
  ProfileUpdate,
  ProfileLookup,
  ProfileCursorInput,
  profiles,
  profilesVariables,
  updateProfile,
  updateProfileVariables,
  requestAvatarUpload,
} from "./schema"
import { Profile, profileStore } from "./models/profile"
import { accessStore } from "./models/access"

type OptionalFetchParams = {
  policy?: FetchPolicy
}

type FetchParams = {
  policy: FetchPolicy
}

const defaultParams: FetchParams = {
  policy: "cache-first",
}

type Image = {
  format: string
  data: string
}

class ProfileIterator {
  private api: Client
  private lookup: ProfileLookup
  private cursor?: ProfileCursorInput
  private params: FetchParams

  constructor(api: Client, lookup: ProfileLookup, params?: OptionalFetchParams) {
    this.api = api
    this.lookup = lookup
    this.params = {
      ...defaultParams,
      ...params,
    }
  }

  async next(): Promise<Profile[]> {
    const resp = await this.api.query<profiles, profilesVariables>({
      query: gql`
        query profiles($where: ProfileLookup!, $limit: Int!, $cursor: ProfileCursorInput) {
          profiles(where: $where, limit: $limit, cursor: $cursor) {
            items {
              id
              ownerId
              handle
              givenName
              familyName
              avatarThumbnail {
                format
                data
              }
              avatarUrl
            }
            next {
              id
            }
          }
        }
      `,
      variables: {
        where: this.lookup,
        limit: 100,
        cursor: this.cursor,
      },
      fetchPolicy: this.params.policy,
    })

    this.cursor = resp.data.profiles.next || undefined
    const profs: Profile[] = []
    for (const p of resp.data.profiles.items) {
      profs.push({
        id: p.id,
        ownerId: p.ownerId,
        handle: p.handle,
        givenName: p.givenName || "",
        familyName: p.familyName || "",
        avatarThumbnail: p.avatarThumbnail ? avatarThumbnail(p.avatarThumbnail) : "",
        avatarUrl: p.avatarUrl || "",
      })
    }
    return profs
  }
}

export class ProfileClient {
  private api: Client
  private refreshCtrl: AbortController

  constructor(api: Client) {
    this.api = api
    this.refreshCtrl = new AbortController()
  }

  async update(request: ProfileUpdate = {}): Promise<Profile> {
    const resp = await this.api.mutate<updateProfile, updateProfileVariables>({
      mutation: gql`
        mutation updateProfile($request: ProfileUpdate!) {
          updateProfile(request: $request) {
            id
            ownerId
            handle
            givenName
            familyName
            avatarThumbnail {
              format
              data
            }
            avatarUrl
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
    const p = resp.data.updateProfile
    return {
      id: p.id,
      ownerId: p.ownerId,
      handle: p.handle,
      givenName: p.givenName || "",
      familyName: p.familyName || "",
      avatarThumbnail: p.avatarThumbnail ? avatarThumbnail(p.avatarThumbnail) : "",
      avatarUrl: p.avatarUrl || "",
    }
  }

  async uploadAvatar(avatarURL: string): Promise<void> {
    const resp = await this.api.mutate<requestAvatarUpload>({
      mutation: gql`
        mutation requestAvatarUpload {
          requestAvatarUpload {
            url
            formData
          }
        }
      `,
    })

    const data = resp.data?.requestAvatarUpload
    if (!data) {
      throw new Error("No data in response.")
    }

    const form = new FormData()
    for (const [k, v] of Object.entries(JSON.parse(data.formData))) {
      form.append(k, v as string)
    }
    const res = await fetch(avatarURL)
    form.append("file", await res.blob())
    const minioResp = await fetch(data.url, {
      method: "POST",
      body: form,
    })
    if (minioResp.status >= 400) {
      throw new Error("Failed to upload file.")
    }
  }

  async fetchAvatar(url: string): Promise<string> {
    const resp = await fetch(url, {
      method: "GET",
      cache: "default",
    })
    if (!resp.ok) {
      return ""
    }
    const blob = await resp.blob()
    return new Promise((resolve) => {
      const reader = new FileReader()
      reader.onloadend = () => resolve(reader.result as string)
      reader.readAsDataURL(blob)
    })
  }

  async fetchOne(input: ProfileLookup, params?: OptionalFetchParams): Promise<Profile> {
    const iter = this.fetch(input, params)
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
    const userId = accessStore.state.id
    try {
      const profile = await this.fetchOne({ ownerIds: [userId] })
      profileStore.set(profile)
    } catch (e) {
      profileStore.set({
        id: "",
        ownerId: userId,
        handle: "",
        avatarThumbnail: "",
        avatarUrl: "",
      })
      if (errorCode(e) !== Code.NotFound) {
        throw e
      }
    }
  }

  fetch(lookup: ProfileLookup, params?: OptionalFetchParams): ProfileIterator {
    return new ProfileIterator(this.api, lookup, params)
  }
}

function avatarThumbnail(img: Image): string {
  return `data:image/${img.format};base64,${img.data}`
}
