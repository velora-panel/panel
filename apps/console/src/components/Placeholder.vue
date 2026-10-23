<script setup lang="ts">
import PageHeader from './PageHeader.vue'
import State from './State.vue'
defineProps<{ title: string; description: string }>()
</script>
<template>
  <main class="page">
    <PageHeader :title="title" :description="description" /><State
      kind="empty"
      title="Nothing here yet"
      :description="description"
      ><RouterLink to="/">Return to overview</RouterLink></State
    >
  </main>
</template>
