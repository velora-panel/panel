import { it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { Button } from '@/components/ui/button'
import FormField from './FormField.vue'
import PageHeader from './PageHeader.vue'
import State from './State.vue'
import Placeholder from './Placeholder.vue'
import { plugins } from '../test/helpers'
it('buttons preserve explicit type, disabled state and activation', async () => {
  const w = mount(Button, { slots: { default: 'Save' } })
  expect(w.element.tagName).toBe('BUTTON')
  await w.trigger('click')
  expect(w.emitted('click')).toHaveLength(1)
  await w.setProps({ variant: 'default' })
  const submit = mount(Button, { attrs: { type: 'submit', disabled: true } })
  expect(submit.attributes('type')).toBe('submit')
  expect(submit.attributes()).toHaveProperty('disabled')
})
it('associates input, label and validation error and emits edits', async () => {
  const w = mount(FormField, {
    props: { id: 'name', label: 'Name', modelValue: '' },
  })
  expect(w.find('label').attributes('for')).toBe('name')
  expect(w.find('input').attributes('aria-invalid')).toBe('false')
  await w.find('input').setValue('West')
  expect(w.emitted('update:modelValue')).toEqual([['West']])
  await w.setProps({ error: 'Required' })
  expect(w.find('input').attributes('aria-describedby')).toBe('name-error')
  expect(w.get('#name-error').text()).toBe('Required')
})
it('renders useful headers and all accessible status variants', async () => {
  const header = mount(PageHeader, {
    props: { title: 'Nodes', description: 'Registry' },
    slots: { default: 'Add' },
  })
  expect(header.text()).toContain('Registry')
  expect(header.text()).toContain('Add')
  const w = mount(State, { props: { kind: 'loading', title: 'Loading' } })
  expect(w.attributes('aria-busy')).toBe('true')
  await w.setProps({ kind: 'error', title: 'Failed', description: 'Try again' })
  expect(w.attributes('role')).toBe('alert')
  expect(w.text()).toContain('Try again')
  await w.setProps({ kind: 'empty' })
  expect(w.attributes('role')).toBe('status')
  const options = await plugins()
  const empty = mount(Placeholder, {
    ...options,
    props: { title: 'Blueprints', description: 'Coming later' },
  })
  expect(empty.text()).toContain('Coming later')
  expect(empty.get('a').attributes('href')).toBe('/')
})
