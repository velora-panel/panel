import type { QueryClient } from '@tanstack/vue-query'
import type { Router } from 'vue-router'
import type { Middleware } from '@velora/api-client'
/** Check whether a 401 means session expiry or a failed sensitive-operation reauthentication. */
export function sessionMiddleware(
  cache: QueryClient,
  router: Router,
  fetchSession: () => Promise<Response>,
): Middleware {
  let checking: Promise<void> | undefined
  return {
    async onResponse({ request, response }) {
      if (
        response.status !== 401 ||
        new URL(request.url).pathname === '/api/v1/auth/login'
      )
        return response
      if (!checking) {
        checking = (async () => {
          try {
            const current =
              new URL(request.url).pathname === '/api/v1/auth/me'
                ? response
                : await fetchSession()
            if (current.status !== 401) return
            cache.clear()
            await router.replace('/login')
          } catch {
            /* A network failure does not establish that the session was revoked. */
          }
        })().finally(() => {
          checking = undefined
        })
      }
      await checking
      return response
    },
  }
}
