<script setup lang="ts">
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
defineProps<{ id: string; label: string; error?: string | undefined }>()
defineOptions({ inheritAttrs: false })
const model = defineModel<string>({ required: true })
</script>
<template>
  <div class="field">
    <Label :for="id">{{ label }}</Label
    ><Input
      :id="id"
      v-model="model"
      v-bind="$attrs"
      :aria-invalid="Boolean(error)"
      :aria-describedby="error ? `${id}-error` : undefined"
    /><span v-if="error" :id="`${id}-error`" class="error">{{ error }}</span>
  </div>
</template>
