<script setup lang="ts">
import { defined } from "@/lib/defined"
import type { SelectItemTextProps } from 'reka-ui'
import { SelectItemText } from 'reka-ui'

const props = defineProps<SelectItemTextProps>()
</script>

<template>
  <SelectItemText
    data-slot="select-item-text"
    v-bind="defined(props)"
  >
    <slot />
  </SelectItemText>
</template>
