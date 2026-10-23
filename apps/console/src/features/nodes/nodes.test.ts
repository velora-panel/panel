import { it, expect, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import Nodes from './Nodes.vue'
import NodeDetail from './NodeDetail.vue'
import { api } from '../../lib/api'
import { plugins, node } from '../../test/helpers'
afterEach(() => {
  vi.restoreAllMocks()
})
it('renders pending, failure, retry, empty and list states', async () => {
  const get = vi
    .spyOn(api, 'GET')
    .mockImplementation(() => new Promise(() => {}))
  const pending = mount(Nodes, await plugins())
  expect(pending.text()).toContain('Loading Nodes')
  pending.unmount()
  get.mockRejectedValue(new Error('offline'))
  const failed = mount(Nodes, await plugins())
  await vi.waitFor(() => expect(failed.text()).toContain('Unable to load'))
  get.mockResolvedValue({ data: [], response: new Response() })
  await failed.get('.state button').trigger('click')
  await vi.waitFor(() =>
    expect(failed.text()).toContain('Your fleet starts here'),
  )
  failed.unmount()
  get.mockResolvedValue({ data: [node], response: new Response() })
  const loaded = mount(Nodes, await plugins())
  await vi.waitFor(() => expect(loaded.text()).toContain('West'))
  expect(loaded.get('.node-row').attributes('href')).toBe(`/nodes/${node.id}`)
})
it('validates and submits Node metadata, reports failures and closes on success', async () => {
  vi.spyOn(api, 'GET').mockResolvedValue({ data: [], response: new Response() })
  const post = vi.spyOn(api, 'POST').mockRejectedValue(new Error('denied'))
  const options = await plugins()
  const w = mount(Nodes, {
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
  await w.get('.page-header button').trigger('click')
  await vi.waitFor(() => expect(w.find('form').exists()).toBe(true))
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Use a name'))
  await w.get('#node-name').setValue('West')
  await w.get('#node-description').setValue('x'.repeat(501))
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('no more than 500'))
  await w.get('#node-description').setValue('Rack 1')
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.text()).toContain('Could not create Node'))
  post.mockResolvedValue({ data: node, response: new Response() })
  await w.get('form').trigger('submit')
  await vi.waitFor(() => expect(w.find('form').exists()).toBe(false))
  expect(post).toHaveBeenCalledWith('/api/v1/nodes', {
    body: { name: 'West', description: 'Rack 1' },
  })
})
it('renders Node detail and handles unavailable records', async () => {
  const get = vi
    .spyOn(api, 'GET')
    .mockImplementation(() => new Promise(() => {}))
  const pending = mount(NodeDetail, await plugins('/nodes/id'))
  expect(pending.text()).toContain('Loading Node')
  pending.unmount()
  get.mockRejectedValue(new Error('missing'))
  const missing = mount(NodeDetail, await plugins('/nodes/id'))
  await vi.waitFor(() => expect(missing.text()).toContain('Node unavailable'))
  missing.unmount()
  get.mockResolvedValue({ data: node, response: new Response() })
  const detail = mount(NodeDetail, await plugins('/nodes/id'))
  await vi.waitFor(() => expect(detail.text()).toContain('West'))
  expect(detail.text()).toContain('metadata record')
})
