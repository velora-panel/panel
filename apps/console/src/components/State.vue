<script setup lang="ts">
import { Skeleton } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'
defineProps<{
  kind: 'loading' | 'error' | 'empty'
  title: string
  description?: string
}>()
</script>
<template>
  <Card
    class="state"
    :role="kind === 'error' ? 'alert' : 'status'"
    :aria-busy="kind === 'loading'"
  >
    <Skeleton
      v-if="kind === 'loading'"
      class="mx-auto mb-4 h-8 w-32"
      aria-hidden="true"
    />
    <span class="state-symbol" aria-hidden="true">{{
      kind === 'error' ? '!' : '◇'
    }}</span>
    <h2>{{ title }}</h2>
    <p v-if="description">{{ description }}</p>
    <slot />
  </Card>
</template>
