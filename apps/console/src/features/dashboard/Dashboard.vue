<script setup lang="ts">
import { Server, ArrowUpRight } from '@lucide/vue'
import { useNodes } from '../nodes/queries'
import PageHeader from '../../components/PageHeader.vue'
import State from '../../components/State.vue'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
const nodes = useNodes()
</script>
<template>
  <main class="page">
    <PageHeader title="Overview" description="Your infrastructure, in focus." />
    <Card class="welcome">
      <div>
        <Badge>FOUNDATION RELEASE</Badge>
        <h2>Ready for your next Node.</h2>
        <p>
          Register your infrastructure today.<br />Workload orchestration comes
          next.
        </p>
        <RouterLink class="text-link" to="/nodes"
          >Manage Nodes <ArrowUpRight :size="16"
        /></RouterLink>
      </div>
      <Server :size="100" :stroke-width="1" class="welcome-icon" />
    </Card>
    <State
      v-if="nodes.isPending.value"
      kind="loading"
      title="Loading overview"
    /><State
      v-else-if="nodes.isError.value"
      kind="error"
      title="Overview unavailable"
    />
    <section v-else class="overview-grid">
      <article class="card metric">
        <p>Registered Nodes</p>
        <strong>{{ nodes.data.value?.length }}</strong
        ><small>Latest 100 metadata records</small>
      </article>
      <article class="card metric">
        <p>Beacon connectivity</p>
        <strong class="text-status">Not configured</strong
        ><small>External daemon integration is planned</small>
      </article>
    </section>
    <section class="card details">
      <h2>A deliberate starting point</h2>
      <p>
        Velora Core manages access and your Node registry. Instances and
        Blueprints will become available as the platform grows.
      </p>
      <RouterLink to="/blueprints" class="text-link"
        >Explore Blueprints <ArrowUpRight :size="16"
      /></RouterLink>
    </section>
  </main>
</template>
