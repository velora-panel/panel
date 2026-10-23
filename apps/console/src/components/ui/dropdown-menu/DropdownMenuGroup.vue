<script setup lang="ts">
import { defined } from "@/lib/defined"
import type { DropdownMenuGroupProps } from 'reka-ui'
import { DropdownMenuGroup } from 'reka-ui'

const props = defineProps<DropdownMenuGroupProps>()
</script>

<template>
  <DropdownMenuGroup
    data-slot="dropdown-menu-group"
    v-bind="defined(props)"
  >
    <slot />
  </DropdownMenuGroup>
</template>
