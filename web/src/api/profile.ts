import { gql } from "@apollo/client/core"
import { Client, APIError, Code, errorCode } from "./api"
import {
  ProfileMask,
  ProfileLookup,
  profiles,
  profilesVariables,
  updateProfile,
  updateProfileVariables,
  requestAvatarUpload,
} from "./schema"
import { Profile, currentUser, profileStore } from "./models"
import { config } from "@/config"

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
              avatarThumbnail {
                format
                data
              }
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
    const profs: Profile[] = []
    for (const p of resp.data.profiles.items) {
      profs.push({
        id: p.id,
        ownerId: p.ownerId,
        handle: p.handle,
        displayName: p.displayName || "",
        avatarThumbnail: p.avatarThumbnail
          ? `data:image/${p.avatarThumbnail.format};base64,${p.avatarThumbnail.data}`
          : "",
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
    const p = resp.data.updateProfile
    return {
      id: p.id,
      ownerId: p.ownerId,
      handle: p.handle,
      displayName: p.displayName || "",
      avatarThumbnail: "",
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
      if (currentUser.id === "") {
        throw new APIError(Code.NotFound, "failed to fetch user")
      }
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
            avatarThumbnail: "",
          })
          break
      }
    }
  }

  fetch(lookup: ProfileLookup): ProfileIterator {
    return new ProfileIterator(this.api, lookup)
  }
}
