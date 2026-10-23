<script setup lang="ts">
import { defined } from "@/lib/defined"
import type { DialogCloseProps } from 'reka-ui'
import { DialogClose } from 'reka-ui'

const props = defineProps<DialogCloseProps>()
</script>

<template>
  <DialogClose
    data-slot="dialog-close"
    v-bind="defined(props)"
  >
    <slot />
  </DialogClose>
</template>
