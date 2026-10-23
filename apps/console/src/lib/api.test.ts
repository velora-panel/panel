import { describe, it, expect, vi, afterEach } from 'vitest'
import { api, unwrap } from './api'
import { requireSession } from '../router/guard'
afterEach(() => vi.restoreAllMocks())
describe('API boundary', () => {
  it('unwraps data and preserves useful safe errors', () => {
    expect(unwrap({ data: 1 })).toBe(1)
    expect(() => unwrap({ error: { error: { message: 'Denied' } } })).toThrow(
      'Denied',
    )
    expect(() => unwrap({})).toThrow('empty response')
  })
  it('checks the server session before entering protected routes', async () => {
    const get = vi.spyOn(api, 'GET')
    get.mockResolvedValue({
      data: { id: 'id', email: 'a@b.co', disabled: false, role: 'admin' },
      response: new Response(),
    })
    expect(await requireSession()).toBe(true)
    get.mockResolvedValue({
      error: {
        error: { code: 'UNAUTHORIZED', message: 'Unauthorized', details: {} },
        request_id: 'r',
      },
      response: new Response(null, { status: 401 }),
    })
    expect(await requireSession()).toBe('/login')
    get.mockResolvedValue({
      error: {
        error: { code: 'HTTP_ERROR', message: 'Unavailable', details: {} },
        request_id: 'r',
      },
      response: new Response(null, { status: 503 }),
    })
    expect(await requireSession()).toBe('/error')
    get.mockRejectedValue(new Error('offline'))
    expect(await requireSession()).toBe('/error')
  })
})
