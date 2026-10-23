<script setup lang="ts">
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { ref } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useForm } from 'vee-validate'
import { z } from 'zod'
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogClose,
} from '@/components/ui/dialog'
import { Plus } from '@lucide/vue'
import { api, unwrap } from '../../lib/api'
import { useNodes } from './queries'
import PageHeader from '../../components/PageHeader.vue'
import State from '../../components/State.vue'
import { Button } from '@/components/ui/button'
import FormField from '../../components/FormField.vue'
const nodes = useNodes()
const cache = useQueryClient()
const open = ref(false)
const failure = ref('')
const nameSchema = z.string().trim().min(1).max(80)
const { defineField, errors, handleSubmit, isSubmitting, resetForm } = useForm({
  initialValues: { name: '', description: '' },
  validationSchema: {
    name: (v: string) =>
      nameSchema.safeParse(v).success ||
      'Use a name between 1 and 80 characters.',
    description: (v: string) =>
      v.length <= 500 || 'Use no more than 500 characters.',
  },
})
const [name] = defineField('name')
const [description] = defineField('description')
const submit = handleSubmit(async (values) => {
  failure.value = ''
  try {
    unwrap(await api.POST('/api/v1/nodes', { body: values }))
    await cache.invalidateQueries({ queryKey: ['nodes'] })
    open.value = false
    resetForm()
  } catch {
    failure.value =
      'Could not create Node. Check permissions, connection and duplicate names.'
  }
})
</script>
<template>
  <main class="page">
    <PageHeader
      title="Nodes"
      description="A registry for your compute infrastructure."
      ><Dialog v-model:open="open"
        ><DialogTrigger as-child
          ><Button variant="default"
            ><Plus :size="16" /> Add Node</Button
          ></DialogTrigger
        ><DialogContent
          ><DialogTitle as="h2">Add a Node</DialogTitle
          ><DialogDescription
            >Register metadata only. Beacon connectivity is not available yet.
            Administrator access is required.</DialogDescription
          >
          <form @submit="submit">
            <FormField
              id="node-name"
              v-model="name"
              label="Node name"
              :error="errors.name"
            /><FormField
              id="node-description"
              v-model="description"
              label="Description"
              :error="errors.description"
            />
            <p v-if="failure" role="alert" class="error">{{ failure }}</p>
            <div class="actions">
              <DialogClose as-child><Button>Cancel</Button></DialogClose
              ><Button type="submit" variant="default" :disabled="isSubmitting"
                >Create Node</Button
              >
            </div>
          </form></DialogContent
        ></Dialog
      ></PageHeader
    ><State
      v-if="nodes.isPending.value"
      kind="loading"
      title="Loading Nodes"
    /><State
      v-else-if="nodes.isError.value"
      kind="error"
      title="Unable to load Nodes"
      description="Check Core availability and your session."
      ><Button @click="nodes.refetch()">Retry</Button></State
    ><State
      v-else-if="nodes.data.value?.length === 0"
      kind="empty"
      title="Your fleet starts here"
      description="Add your first Node to organize your compute infrastructure. No daemon connection is required."
    />
    <Card v-else class="node-list">
      <RouterLink
        v-for="node in nodes.data.value"
        :key="node.id"
        :to="`/nodes/${node.id}`"
        class="node-row"
        ><span
          ><strong>{{ node.name }}</strong
          ><small>{{ node.description }}</small></span
        ><Badge>{{ node.status }}</Badge
        ><span aria-hidden="true">↗</span></RouterLink
      >
    </Card>
    <p class="small">Showing up to 100 most recently registered Nodes.</p>
  </main>
</template>
