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
import { gql } from "@apollo/client/core"
import {
  login,
  loginVariables,
  createSession,
  createSessionVariables,
  token,
} from "./schema"
import { userStore } from "./models"

enum Code {
  Unauthorised = "UNAUTHORIZED",
}

const minRefresh = 10 * Duration.second

export class Client {
  private graph: ApolloClient<NormalizedCacheObject>
  private refreshTimer = 0
  private token: Promise<string> | null = null
  private tokenResolve: ((token: string) => void) | null = null

  constructor() {
    const httpLink = createHttpLink({
      uri: `${window.location.protocol}/api/query`,
    })

    const errorLink = onError((resp: ErrorResponse) => {
      const traceId = resp.operation
        .getContext()
        .response.headers.get("Trace-Id")
      const operation = resp.operation.operationName

      let msg = `Request "${operation}" failed:`
      msg += "\nTrace = " + traceId
      for (const e of resp.response?.errors || resp.graphQLErrors || []) {
        const code = e.extensions?.code
        // Ignoring auth erros when fetching token.
        // There is no way to know if the user is authenticated or not without trying to fetch the token.
        if (operation === "token" && code === Code.Unauthorised) {
          return
        }
        msg += `\n${code || "UNKNOWN_CODE"} (${e.message})`
      }
      console.error(msg)
    })

    const authLink = setContext(async (operation: GraphQLRequest) => {
      // Those operations are used to fetch token.
      // We don't wait for token to avoid a deadlock.
      if (
        operation.operationName === "token" ||
        operation.operationName === "createSession"
      ) {
        return
      }
      return {
        headers: {
          authorization: `Bearer ${await this.token}`,
        },
      }
    })

    this.graph = new ApolloClient({
      link: from([authLink, errorLink, httpLink]),
      cache: new InMemoryCache(),
    })

    this.refreshToken()
  }

  async mutate<T = object, TVariables = OperationVariables>(
    options: MutationOptions<T, TVariables>,
  ): Promise<FetchResult<T>> {
    options.context = {
      token: await this.token,
    }
    return this.graph.mutate(options)
  }

  async query<T = object, TVariables = OperationVariables>(
    options: QueryOptions<TVariables, T>,
  ): Promise<ApolloQueryResult<T>> {
    options.context = {
      token: await this.token,
    }
    return this.graph.query(options)
  }

  async login(email: string): Promise<void> {
    await this.graph.mutate<login, loginVariables>({
      mutation: gql`
        mutation login($address: String!) {
          login(address: $address)
        }
      `,
      variables: {
        address: email,
      },
    })
  }

  async createSession(emailToken: string): Promise<void> {
    this.setRefreshInProgress()

    const resp = await this.graph.mutate<createSession, createSessionVariables>(
      {
        mutation: gql`
          mutation createSession($emailToken: String!) {
            createSession(emailToken: $emailToken) {
              token
              expiresIn
            }
          }
        `,
        variables: {
          emailToken: emailToken,
        },
      },
    )
    if (!resp.data) {
      console.error("No token present in session response. Trying to refresh.")
      // Failed to acquire token from session. Try refreshing (hoping that the session cookie was set).
      this.refreshToken() // No point in waiting for it, so no `await`.
      return
    }
    const token = resp.data.createSession
    this.setToken(token.token, token.expiresIn)
  }

  private async refreshToken() {
    this.setRefreshInProgress()

    try {
      const resp = await this.graph.query<token>({
        query: gql`
          query token {
            token {
              token
              expiresIn
            }
          }
        `,
      })
      const t = resp.data.token
      this.setToken(t.token, t.expiresIn)
    } catch {
      // Failed to refresh the token. Give up and set an empty token.
      this.setToken("", 0)
      return
    }
  }

  // Be sure to ALWAYS call `setToken` after this.
  private setRefreshInProgress(): void {
    // Cancel refresh because this should not run in parallel.
    clearTimeout(this.refreshTimer)
    // Set a promise so all calls are waiting for refreshToken to finish to avoid race conditions.
    this.token = new Promise<string>(resolve => {
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
    this.token = new Promise<string>(resolve => {
      resolve(token)
    })
    // Release the previous promise with the token.
    if (this.tokenResolve) {
      this.tokenResolve(token)
    }

    // Set user in user store. This will trigger reacive state change.
    userStore.id = "todo"

    // If expiresIn is set, schedule a new refresh.
    if (expiresIn === 0) {
      return
    }
    const after = duration({ seconds: expiresIn }) - 2 * Duration.minute
    this.refreshTimer = setTimeout(
      this.refreshToken.bind(this),
      Math.max(after, minRefresh),
    )
  }
}
