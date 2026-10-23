<script setup lang="ts">
import { defined } from "@/lib/defined"
import type { SelectValueProps } from 'reka-ui'
import { SelectValue } from 'reka-ui'

const props = defineProps<SelectValueProps>()
</script>

<template>
  <SelectValue
    data-slot="select-value"
    v-bind="defined(props)"
  >
    <slot />
  </SelectValue>
</template>
