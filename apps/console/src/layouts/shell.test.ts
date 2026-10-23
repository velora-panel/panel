import { it, expect, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { DropdownMenuItem } from 'reka-ui'
import Shell from './Shell.vue'
import { api } from '../lib/api'
import { plugins } from '../test/helpers'
afterEach(() => {
  vi.restoreAllMocks()
})
it('toggles navigation and revokes the session before leaving', async () => {
  const post = vi
    .spyOn(api, 'POST')
    .mockRejectedValue(new Error('network error'))
  const options = await plugins()
  const w = mount(Shell, {
    ...options,
    attachTo: document.body,
    global: {
      ...options.global,
      stubs: {
        DialogPortal: { template: '<div><slot /></div>' },
        DropdownMenuPortal: { template: '<div><slot /></div>' },
        TooltipPortal: { template: '<div><slot /></div>' },
      },
    },
  })
  await w.get('[aria-label="Toggle sidebar"]').trigger('focus')
  await vi.waitFor(() => expect(w.find('.tooltip').exists()).toBe(true))
  await w.get('[aria-label="Toggle sidebar"]').trigger('click')
  expect(w.classes()).toContain('collapsed')
  await w.get('[aria-label="Toggle sidebar"]').trigger('click')
  expect(w.classes()).not.toContain('collapsed')
  await w.get('.account-button').trigger('keydown', { key: 'Enter' })
  await vi.waitFor(() =>
    expect(w.findComponent(DropdownMenuItem).exists()).toBe(true),
  )
  w.getComponent(DropdownMenuItem).vm.$emit('select')
  await vi.waitFor(() => expect(w.text()).toContain('Unable to sign out'))
  post.mockResolvedValue({
    error: {
      error: { code: 'FAIL', message: 'Failed', details: {} },
      request_id: 'r',
    },
    response: new Response(null, { status: 500 }),
  })
  w.getComponent(DropdownMenuItem).vm.$emit('select')
  await vi.waitFor(() => expect(w.text()).toContain('Unable to sign out'))
  post.mockResolvedValue({
    data: undefined,
    response: new Response(null, { status: 204 }),
  })
  w.getComponent(DropdownMenuItem).vm.$emit('select')
  await vi.waitFor(() =>
    expect(options.router.currentRoute.value.path).toBe('/login'),
  )
})
