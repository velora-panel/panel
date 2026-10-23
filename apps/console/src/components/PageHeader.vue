<script setup lang="ts">
defineProps<{ title: string; description: string }>()
</script>
<template>
  <header class="page-header">
    <div>
      <p class="eyebrow">VELORA CONSOLE</p>
      <h1>{{ title }}</h1>
      <p>{{ description }}</p>
    </div>
    <slot />
  </header>
</template>
