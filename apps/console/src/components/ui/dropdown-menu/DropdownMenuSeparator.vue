<script setup lang="ts">
import { defined } from "@/lib/defined"
import type { DropdownMenuSeparatorProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { reactiveOmit } from '@vueuse/core'
import { cn } from 'cn'
import {
  DropdownMenuSeparator,
} from 'reka-ui'

const props = defineProps<DropdownMenuSeparatorProps & {
  class?: HTMLAttributes['class']
}>()

const delegatedProps = reactiveOmit(props, 'class')
</script>

<template>
  <DropdownMenuSeparator
    data-slot="dropdown-menu-separator"
    v-bind="defined(delegatedProps)"
    :class="cn('bg-border -mx-1 my-1 h-px', props.class)"
  />
</template>
