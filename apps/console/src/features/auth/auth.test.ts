import { it, expect, vi, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import Login from './Login.vue'
import Settings from './Settings.vue'
import { api } from '../../lib/api'
import { plugins, user } from '../../test/helpers'
afterEach(() => vi.restoreAllMocks())
it('validates login, reports failure, and navigates after successful authentication', async () => {
  const options = await plugins('/login')
  const post = vi.spyOn(api, 'POST').mockRejectedValue(new Error('offline'))
  const w = mount(Login, options)
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Enter your email'))
  expect(post).not.toHaveBeenCalled()
  await w.get('#email').setValue('admin@example.com')
  await w.get('#password').setValue('long password here')
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Sign-in failed'))
  post.mockResolvedValue({ data: user, response: new Response() })
  await w.get('form').trigger('submit')
  await vi.waitFor(() =>
    expect(options.router.currentRoute.value.path).toBe('/'),
  )
  expect(post).toHaveBeenCalledWith('/api/v1/auth/login', {
    body: { email: 'admin@example.com', password: 'long password here' },
  })
})
it('shows current account and loading/error states', async () => {
  const get = vi
    .spyOn(api, 'GET')
    .mockImplementation(() => new Promise(() => {}))
  const w = mount(Settings, await plugins())
  expect(w.text()).toContain('Loading account')
  w.unmount()
  get.mockRejectedValue(new Error('offline'))
  const failed = mount(Settings, await plugins())
  await vi.waitFor(() => expect(failed.text()).toContain('Unable to load'))
  failed.unmount()
  get.mockResolvedValue({ data: user, response: new Response() })
  const loaded = mount(Settings, await plugins())
  await flushPromises()
  await vi.waitFor(() => expect(loaded.text()).toContain(user.email))
  expect(loaded.text()).toContain('admin')
})
