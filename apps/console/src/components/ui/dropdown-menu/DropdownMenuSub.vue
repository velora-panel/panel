<script setup lang="ts">
import { defined } from '@/lib/defined'
import type { DropdownMenuSubEmits, DropdownMenuSubProps } from 'reka-ui'
import { DropdownMenuSub, useForwardPropsEmits } from 'reka-ui'

const props = defineProps<DropdownMenuSubProps>()
const emits = defineEmits<DropdownMenuSubEmits>()

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <DropdownMenuSub
    v-slot="slotProps"
    data-slot="dropdown-menu-sub"
    v-bind="defined(forwarded)"
  >
    <slot v-bind="defined(slotProps)" />
  </DropdownMenuSub>
</template>
