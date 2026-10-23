import { it, expect, vi } from 'vitest'
import type { MiddlewareCallbackParams } from '@velora/api-client'
import { sessionMiddleware } from './session'
import { plugins } from '@/test/helpers'
it('clears cached account data on expiry, but preserves sessions after incorrect reauthentication or network errors', async () => {
  const { queryClient: cache, router } = await plugins()
  const fetch = vi.fn(async () => new Response(null, { status: 401 }))
  const middleware = sessionMiddleware(cache, router, fetch)
  const run = (path: string, status: number) =>
    middleware.onResponse!({
      request: new Request(`https://console.test${path}`),
      response: new Response(null, { status }),
    } as MiddlewareCallbackParams & { response: Response })
  cache.setQueryData(['secret'], 'private')
  await run('/api/v1/nodes', 200)
  expect(cache.getQueryData(['secret'])).toBe('private')
  await run('/api/v1/auth/login', 401)
  expect(fetch).not.toHaveBeenCalled()
  fetch.mockResolvedValue(new Response(null, { status: 403 }))
  await run('/api/v1/auth/password', 401)
  expect(cache.getQueryData(['secret'])).toBe('private')
  fetch.mockRejectedValue(new Error('offline'))
  await run('/api/v1/nodes', 401)
  expect(cache.getQueryData(['secret'])).toBe('private')
  await run('/api/v1/auth/me', 401)
  expect(cache.getQueryData(['secret'])).toBeUndefined()
  expect(router.currentRoute.value.path).toBe('/login')
  cache.setQueryData(['secret'], 'private')
  fetch.mockResolvedValue(new Response(null, { status: 401 }))
  await Promise.all([run('/api/v1/nodes', 401), run('/api/v1/users', 401)])
  expect(cache.getQueryData(['secret'])).toBeUndefined()
})
