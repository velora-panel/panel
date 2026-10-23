import { it, expect, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Settings from './Settings.vue'
import { api } from '@/lib/api'
import { plugins, user } from '@/test/helpers'
afterEach(() => vi.restoreAllMocks())
it('updates profile and password with reauthentication, validation, and safe failures', async () => {
  vi.spyOn(api, 'GET').mockResolvedValue({
    data: user,
    response: new Response(),
  })
  const patch = vi
    .spyOn(api, 'PATCH')
    .mockRejectedValue(new Error('Service unavailable'))
  const options = await plugins()
  const w = mount(Settings, options)
  await vi.waitFor(() => expect(w.find('#profile-email').exists()).toBe(true))
  await w.get('#profile-email').setValue('new@example.com')
  await w.get('#profile-current').setValue('current password')
  await w.findAll('form')[0]!.trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Service unavailable'))
  patch.mockResolvedValue({
    data: { ...user, email: 'new@example.com' },
    response: new Response(),
  })
  await w.findAll('form')[0]!.trigger('submit')
  await vi.waitFor(() =>
    expect(w.text()).toContain('Account updated successfully'),
  )
  expect(options.queryClient.getQueryData(['me'])).toEqual({
    ...user,
    email: 'new@example.com',
  })
  await w.findAll('form')[1]!.trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('between 12 and 128'))
  await w.get('#password-new').setValue('new password here')
  await w.findAll('form')[1]!.trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Passwords do not match'))
  await w.get('#password-confirm').setValue('new password here')
  await w.get('#password-current').setValue('current password')
  patch.mockResolvedValue({
    error: {
      error: {
        code: 'UNAUTHORIZED',
        message: 'Current password is incorrect',
        details: {},
      },
      request_id: 'r',
    },
    response: new Response(null, { status: 401 }),
  })
  await w.findAll('form')[1]!.trigger('submit')
  await vi.waitFor(() =>
    expect(w.text()).toContain('Current password is incorrect'),
  )
  patch.mockResolvedValue({
    data: undefined,
    response: new Response(null, { status: 204 }),
  })
  await w.findAll('form')[1]!.trigger('submit')
  await vi.waitFor(() =>
    expect((w.get('#password-new').element as HTMLInputElement).value).toBe(''),
  )
  expect(patch).toHaveBeenLastCalledWith('/api/v1/auth/password', {
    body: {
      current_password: 'current password',
      new_password: 'new password here',
    },
  })
})
