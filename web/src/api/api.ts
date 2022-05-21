import {
  NormalizedCacheObject,
  OperationVariables,
  GraphQLRequest,
  MutationOptions,
  FetchResult,
  QueryOptions,
  ApolloQueryResult,
  from,
  ApolloClient,
  createHttpLink,
  InMemoryCache,
} from "@apollo/client/core"
import { onError, ErrorResponse } from "@apollo/client/link/error"
import { setContext } from "@apollo/client/link/context"
import { duration, Duration } from "@/platform/time"
import { ApolloError } from "@apollo/client/core"
import { userStore } from "./models"

const minRefresh = 10 * Duration.second
const httpOK = 200
const httpCreated = 202

export type FetchPolicy =
  // Apollo Client first executes the query against the cache. If all requested data is present in the cache, that data is returned. Otherwise, Apollo Client executes the query against your GraphQL server and returns that data after caching it.
  // Prioritizes minimizing the number of network requests sent by your application.
  // This is the default fetch policy.
  | "cache-first"
  // Apollo Client executes the full query against your GraphQL server, without first checking the cache. The query's result is stored in the cache.
  // Prioritizes consistency with server data, but can't provide a near-instantaneous response when cached data is available.
  | "network-only"
  // Apollo Client executes the query only against the cache. It never queries your server in this case.
  // A cache-only query throws an error if the cache does not contain data for all requested fields.
  | "cache-only"
  // Similar to network-only, except the query's result is not stored in the cache.
  | "no-cache"
  // Uses the same logic as cache-first, except this query does not automatically update when underlying field values change. You can still manually update this query with refetch and updateQueries.
  | "standby"

export class APIError extends Error {
  code: Code

  constructor(code: Code, message: string) {
    super(message)

    this.code = code
  }
}

export enum Code {
  DuplicateEntry = "DUPLICATE_ENTRY",
  NotFound = "NOT_FOUND",
  BadRequest = "BAD_REQUEST",
  Unknown = "UNKNOWN_CODE",
}

export function errorCode(e: unknown): Code {
  if (e instanceof APIError) {
    return e.code
  }

  const resp = e as ApolloError
  for (const err of resp.graphQLErrors || []) {
    return err.extensions?.code || Code.Unknown
  }

  return Code.Unknown
}

interface Token {
  token: string
  expiresIn: number
}

export class Client {
  private graph: ApolloClient<NormalizedCacheObject>
  private refreshTimer = 0
  private _token: Promise<string>
  private tokenResolve: ((token: string) => void) | null = null

  constructor() {
    const httpLink = createHttpLink({
      uri: `${window.location.protocol}/api/query`,
    })

    const errorLink = onError((resp: ErrorResponse) => {
      const traceId = resp.operation.getContext().response.headers.get("Trace-Id")
      const operation = resp.operation.operationName

      let msg = `Request "${operation}" failed:`
      msg += "\nTrace = " + traceId
      let log = console.warn
      for (const e of resp.response?.errors || resp.graphQLErrors || []) {
        const code = e.extensions?.code
        msg += `\n${code || Code.Unknown} (${e.message})`
        if (code === Code.Unknown) {
          log = console.error
        }
      }
      log(msg)
    })

    const authLink = setContext(async (operation: GraphQLRequest) => {
      // Those operations are used to fetch token.
      // We don't wait for token to avoid a deadlock.
      if (operation.operationName === "token" || operation.operationName === "createSession") {
        return
      }
      return {
        headers: {
          authorization: `Bearer ${await this._token}`,
        },
      }
    })

    this.graph = new ApolloClient({
      link: from([authLink, errorLink, httpLink]),
      cache: new InMemoryCache(),
    })

    this._token = new Promise<string>((resolve) => {
      resolve("")
    })
  }

  async token(): Promise<string> {
    return await this._token
  }

  async mutate<T = object, TVariables = OperationVariables>(
    options: MutationOptions<T, TVariables>,
  ): Promise<FetchResult<T>> {
    options.context = {
      token: await this._token,
    }
    return this.graph.mutate(options)
  }

  async query<T = object, TVariables = OperationVariables>(
    options: QueryOptions<TVariables, T>,
  ): Promise<ApolloQueryResult<T>> {
    options.context = {
      token: await this._token,
    }
    return this.graph.query(options)
  }

  async login(email: string): Promise<void> {
    const resp = await fetch("/api/iam/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        address: email,
      }),
    })
    if (resp.status !== httpOK) {
      throw new APIError(Code.Unknown, "Login failed.")
    }
  }

  async logout(): Promise<void> {
    const resp = await fetch("/api/iam/logout", {
      method: "POST",
    })
    if (resp.status !== httpOK) {
      throw new APIError(Code.Unknown, "Logout failed.")
    }
  }

  async createSession(emailToken: string): Promise<void> {
    this.setRefreshInProgress()

    const resp = await fetch("/api/iam/session", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        emailToken: emailToken,
      }),
    })
    if (resp.status === httpCreated) {
      const data = (await resp.json()) as Token
      this.setToken(data.token, data.expiresIn)
    } else {
      console.error("No token present in session response. Trying to refresh.")
      // Failed to acquire token from session. Try refreshing (hoping that the session cookie was set).
      this.refreshToken() // No point in waiting for it, so no `await`.
    }
  }

  async refreshToken() {
    this.setRefreshInProgress()

    const resp = await fetch("/api/iam/token", {
      method: "GET",
    })
    if (resp.status === httpOK) {
      const data = (await resp.json()) as Token
      this.setToken(data.token, data.expiresIn)
    } else {
      // Failed to refresh the token. Give up and set an empty token.
      this.setToken("", 0)
    }
  }

  async clearCache(): Promise<void> {
    await this.graph.clearStore()
  }

  // Be sure to ALWAYS call `setToken` after this.
  private setRefreshInProgress(): void {
    // Cancel refresh because this should not run in parallel.
    clearTimeout(this.refreshTimer)
    // Set a promise so all calls are waiting for refreshToken to finish to avoid race conditions.
    this._token = new Promise<string>((resolve) => {
      // Need to combine with oldResolve because some calls might be already waiting.
      // If we just set a new tokenResolve, they will hang forever.
      const oldResolve = this.tokenResolve
      this.tokenResolve = (token: string) => {
        if (oldResolve) {
          oldResolve(token)
        }
        resolve(token)
      }
    })
  }

  private setToken(token: string, expiresIn: number): void {
    // Set an instant promise for future calls.
    this._token = new Promise<string>((resolve) => {
      resolve(token)
    })
    // Release the previous promise with the token.
    if (this.tokenResolve) {
      this.tokenResolve(token)
    }

    if (!token) {
      return
    }
    // Set user in user store. This will trigger reactive state change.
    const claims = JSON.parse(window.atob(token.split(".")[1]))
    userStore.login(claims.uid, claims.acc)

    // If expiresIn is set, schedule a new refresh.
    if (expiresIn === 0) {
      return
    }
    const after = duration({ seconds: expiresIn }) - 2 * Duration.minute
    this.refreshTimer = window.setTimeout(this.refreshToken.bind(this), Math.max(after, minRefresh))
  }
}
