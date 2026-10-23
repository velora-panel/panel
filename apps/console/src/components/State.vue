<script setup lang="ts">
defineProps<{
  kind: 'loading' | 'error' | 'empty'
  title: string
  description?: string
}>()
</script>
<template>
  <section
    class="state card"
    :role="kind === 'error' ? 'alert' : 'status'"
    :aria-busy="kind === 'loading'"
  >
    <span class="state-symbol" aria-hidden="true">{{
      kind === 'error' ? '!' : '◇'
    }}</span>
    <h2>{{ title }}</h2>
    <p v-if="description">{{ description }}</p>
    <slot />
  </section>
</template>
