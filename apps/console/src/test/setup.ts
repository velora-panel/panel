import { afterEach } from 'vitest'
import { enableAutoUnmount } from '@vue/test-utils'
enableAutoUnmount(afterEach)
// jsdom has no layout engine; browser tests exercise actual observer behavior.
globalThis.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
