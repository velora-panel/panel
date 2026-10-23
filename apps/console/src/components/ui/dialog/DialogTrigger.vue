<script setup lang="ts">
import { defined } from '@/lib/defined'
import type { DialogTriggerProps } from 'reka-ui'
import { DialogTrigger } from 'reka-ui'

const props = defineProps<DialogTriggerProps>()
</script>

<template>
  <DialogTrigger data-slot="dialog-trigger" v-bind="defined(props)">
    <slot />
  </DialogTrigger>
</template>
