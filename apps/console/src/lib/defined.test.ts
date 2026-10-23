import { it, expect } from 'vitest'
import { defined } from './defined'
import { cn } from './utils'
it('omits undefined forwarded props while preserving explicit false and null', () => {
  expect(
    defined({ value: undefined, disabled: false, label: null, name: 'a' }),
  ).toEqual({ disabled: false, label: null, name: 'a' })
  expect(cn('px-2', 'px-4', false)).toBe('px-4')
})
