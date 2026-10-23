<script setup lang="ts">
import { defined } from '@/lib/defined'
import type { TabsRootEmits, TabsRootProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { reactiveOmit } from '@vueuse/core'
import { cn } from 'cn'
import { TabsRoot, useForwardPropsEmits } from 'reka-ui'

const props = defineProps<TabsRootProps & { class?: HTMLAttributes['class'] }>()
const emits = defineEmits<TabsRootEmits>()

const delegatedProps = reactiveOmit(props, 'class')
const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <TabsRoot
    v-slot="slotProps"
    data-slot="tabs"
    :data-orientation="forwarded.orientation || 'horizontal'"
    v-bind="defined(forwarded)"
    :class="cn('group/tabs flex gap-2 data-horizontal:flex-col', props.class)"
  >
    <slot v-bind="defined(slotProps)" />
  </TabsRoot>
</template>
