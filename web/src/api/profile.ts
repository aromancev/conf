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
import { config } from "@/config"

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
              displayName
              hasAvatar
              avatarThumbnail {
                format
                data
              }
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
        hasAvatar: p.hasAvatar,
        displayName: p.displayName || "",
        avatarThumbnail: p.avatarThumbnail ? avatarThumbnail(p.avatarThumbnail) : "",
      })
    }
    return profs
  }
}

export class ProfileClient {
  private api: Client

  constructor(api: Client) {
    this.api = api
  }

  async update(request: ProfileUpdate = {}): Promise<Profile> {
    const resp = await this.api.mutate<updateProfile, updateProfileVariables>({
      mutation: gql`
        mutation updateProfile($request: ProfileUpdate!) {
          updateProfile(request: $request) {
            id
            ownerId
            handle
            displayName
            hasAvatar
            avatarThumbnail {
              format
              data
            }
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
      hasAvatar: p.hasAvatar,
      displayName: p.displayName || "",
      avatarThumbnail: p.avatarThumbnail ? avatarThumbnail(p.avatarThumbnail) : "",
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

  async fetchAvatar(ownerId: string, profileId: string): Promise<string> {
    const resp = await fetch(`${config.storage.baseURL}/user-public/${ownerId}/${profileId}`, {
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
      if (accessStore.state.id === "") {
        throw new APIError(Code.NotFound, "failed to fetch user")
      }
      const profile = await this.fetchOne({ ownerIds: [accessStore.state.id] })
      profileStore.update(profile)
    } catch (e) {
      switch (errorCode(e)) {
        case Code.NotFound:
          profileStore.update({
            id: "",
            ownerId: accessStore.state.id,
            handle: "",
            displayName: "",
            avatarThumbnail: "",
          })
          break
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
