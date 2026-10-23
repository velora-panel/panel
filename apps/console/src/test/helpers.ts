import type { Plugin } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { createPinia } from 'pinia'
export async function plugins(path = '/') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }],
  })
  await router.push(path)
  await router.isReady()
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  return {
    router,
    queryClient,
    global: {
      plugins: [
        router,
        createPinia(),
        [VueQueryPlugin, { queryClient }] as [
          Plugin,
          { queryClient: QueryClient },
        ],
      ],
    },
  }
}
export const node = {
  id: '00000000-0000-0000-0000-000000000001',
  name: 'West',
  description: 'Rack 1',
  status: 'unconfigured' as const,
  created_at: '2026-01-01T00:00:00Z',
}
export const user = {
  id: node.id,
  email: 'admin@example.com',
  disabled: false,
  role: 'admin' as const,
}
