import { it, expect, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Dashboard from './Dashboard.vue'
import { api } from '../../lib/api'
import { plugins, node } from '../../test/helpers'
afterEach(() => vi.restoreAllMocks())
it('reports loading and failure without inventing metrics', async () => {
  const get = vi
    .spyOn(api, 'GET')
    .mockImplementation(() => new Promise(() => {}))
  const pending = mount(Dashboard, await plugins())
  expect(pending.text()).toContain('Loading overview')
  pending.unmount()
  get.mockRejectedValue(new Error('offline'))
  const failed = mount(Dashboard, await plugins())
  await vi.waitFor(() =>
    expect(failed.text()).toContain('Overview unavailable'),
  )
})
it('shows the actual Node count and honest Beacon status', async () => {
  vi.spyOn(api, 'GET').mockResolvedValue({
    data: [node],
    response: new Response(),
  })
  const w = mount(Dashboard, await plugins())
  await vi.waitFor(() => expect(w.find('.metric strong').text()).toBe('1'))
  expect(w.text()).toContain('Not configured')
})
