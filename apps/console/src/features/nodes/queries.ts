import { useQuery } from '@tanstack/vue-query'
import { api, unwrap } from '../../lib/api'
export function useNodes() {
  return useQuery({
    queryKey: ['nodes'],
    queryFn: async ({ signal }) =>
      unwrap(await api.GET('/api/v1/nodes', { signal })),
  })
}
