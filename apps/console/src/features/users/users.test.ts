import { it, expect, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Users from './Users.vue'
import { api } from '@/lib/api'
import { plugins, user } from '@/test/helpers'
import { Dialog } from '@/components/ui/dialog'
import { Select } from '@/components/ui/select'
afterEach(() => vi.restoreAllMocks())
it('renders loading, access failures and account status', async () => {
  const get = vi
    .spyOn(api, 'GET')
    .mockImplementation(() => new Promise(() => {}))
  const loading = mount(Users, await plugins())
  expect(loading.text()).toContain('Loading users')
  loading.unmount()
  get.mockRejectedValue(new Error('Forbidden'))
  const denied = mount(Users, await plugins())
  await vi.waitFor(() =>
    expect(denied.text()).toContain('Administrator access'),
  )
  denied.unmount()
  get.mockResolvedValue({
    data: [user, { ...user, id: 'other', disabled: true }],
    response: new Response(),
  })
  const w = mount(Users, await plugins())
  await vi.waitFor(() => expect(w.text()).toContain('Disabled'))
  expect(w.text()).toContain('Active')
})
it('creates users, updates roles and disabled status, and revokes sessions after reauthentication', async () => {
  vi.spyOn(api, 'GET').mockResolvedValue({
    data: [user],
    response: new Response(),
  })
  const post = vi
    .spyOn(api, 'POST')
    .mockRejectedValue(new Error('Cannot create account'))
  const patch = vi
    .spyOn(api, 'PATCH')
    .mockRejectedValue(
      new Error('Last active administrator cannot be disabled'),
    )
  const options = await plugins()
  const w = mount(Users, {
    ...options,
    attachTo: document.body,
    global: {
      ...options.global,
      stubs: { DialogPortal: { template: '<div><slot /></div>' } },
    },
  })
  await vi.waitFor(() => expect(w.find('#user-email').exists()).toBe(true))
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('between 12 and 128'))
  await w.get('#user-email').setValue('new@example.com')
  await w.get('#user-password').setValue('new password here')
  await w.get('#admin-current').setValue('current password')
  w.getComponent(Select).vm.$emit('update:modelValue', 'admin')
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Cannot create account'))
  post.mockResolvedValue({ data: user, response: new Response() })
  await w.get('form').trigger('submit')
  await vi.waitFor(() =>
    expect(w.text()).toContain('Account updated successfully'),
  )
  expect(post).toHaveBeenCalledWith('/api/v1/users', {
    body: {
      email: 'new@example.com',
      password: 'new password here',
      role: 'admin',
      current_password: 'current password',
    },
  })
  await w.get(`[aria-label="Manage ${user.email}"]`).trigger('click')
  await vi.waitFor(() => expect(w.find('#edit-current').exists()).toBe(true))
  await w.get('#edit-current').setValue('current password')
  await w.get('#user-disabled').trigger('click')
  w.findAllComponents(Select).at(-1)!.vm.$emit('update:modelValue', 'user')
  await w.findAll('form').at(-1)!.trigger('submit')
  await vi.waitFor(() =>
    expect(w.text()).toContain('Last active administrator'),
  )
  patch.mockResolvedValue({ data: user, response: new Response() })
  await w.findAll('form').at(-1)!.trigger('submit')
  await vi.waitFor(() => expect(w.find('#edit-current').exists()).toBe(false))
  expect(patch).toHaveBeenCalledWith('/api/v1/users/{id}', {
    params: { path: { id: user.id } },
    body: {
      role: 'user',
      disabled: true,
      current_password: 'current password',
    },
  })
  await w.get(`[aria-label="Manage ${user.email}"]`).trigger('click')
  await vi.waitFor(() => expect(w.find('#edit-current').exists()).toBe(true))
  post.mockResolvedValue({
    error: {
      error: { code: 'FAILED', message: 'Revocation failed', details: {} },
      request_id: 'r',
    },
    response: new Response(null, { status: 500 }),
  })
  const revoke = () =>
    w.findAll('button').find((b) => b.text() === 'Revoke sessions')!
  await revoke().trigger('click')
  await vi.waitFor(() => expect(w.text()).toContain('Revocation failed'))
  post.mockResolvedValue({
    data: undefined,
    response: new Response(null, { status: 204 }),
  })
  await w.get('#edit-current').setValue('current password')
  await revoke().trigger('click')
  await vi.waitFor(() => expect(w.find('#edit-current').exists()).toBe(false))
  w.getComponent(Dialog).vm.$emit('update:open', false)
  expect(post).toHaveBeenLastCalledWith('/api/v1/users/{id}/revoke', {
    params: { path: { id: user.id } },
    body: { current_password: 'current password' },
  })
})
