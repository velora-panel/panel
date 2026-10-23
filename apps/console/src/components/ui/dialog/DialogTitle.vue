<script setup lang="ts">
import { defined } from "@/lib/defined"
import type { DialogTitleProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { reactiveOmit } from '@vueuse/core'
import { cn } from 'cn'
import { DialogTitle, useForwardProps } from 'reka-ui'

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = reactiveOmit(props, 'class')

const forwardedProps = useForwardProps(delegatedProps)
</script>

<template>
  <DialogTitle
    data-slot="dialog-title"
    v-bind="defined(forwardedProps)"
    :class="cn('text-base leading-none font-medium cn-font-heading', props.class)"
  >
    <slot />
  </DialogTitle>
</template>
