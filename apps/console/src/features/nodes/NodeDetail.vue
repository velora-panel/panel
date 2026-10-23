<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { api, unwrap } from '../../lib/api'
import PageHeader from '../../components/PageHeader.vue'
import State from '../../components/State.vue'
const route = useRoute()
const id = computed(() => String(route.params.id))
const node = useQuery({
  queryKey: ['nodes', id],
  queryFn: async ({ signal }) =>
    unwrap(
      await api.GET('/api/v1/nodes/{id}', {
        params: { path: { id: id.value } },
        signal,
      }),
    ),
})
</script>
<template>
  <main class="page">
    <RouterLink to="/nodes">← Nodes</RouterLink
    ><State
      v-if="node.isPending.value"
      kind="loading"
      title="Loading Node"
    /><State
      v-else-if="node.isError.value"
      kind="error"
      title="Node unavailable"
      description="The record may not exist, or Core is unavailable."
    /><template v-else-if="node.data.value"
      ><PageHeader
        :title="node.data.value.name"
        :description="node.data.value.description"
      />
      <section class="card details">
        <h2>Registration</h2>
        <span class="badge">{{ node.data.value.status }}</span>
        <dl>
          <dt>Node ID</dt>
          <dd>{{ node.data.value.id }}</dd>
          <dt>Created</dt>
          <dd>{{ node.data.value.created_at }}</dd>
        </dl>
        <p>
          This is a metadata record. Beacon communication and workload
          management are not implemented.
        </p>
      </section></template
    >
  </main>
</template>
