<script setup lang="ts">
import { ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { api, unwrap } from '@/lib/api'
import PageHeader from '@/components/PageHeader.vue'
import State from '@/components/State.vue'
import FormField from '@/components/FormField.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableHeader,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
type Account = {
  id: string
  email: string
  role: 'admin' | 'user'
  disabled: boolean
}
const cache = useQueryClient()
const users = useQuery({
  queryKey: ['users'],
  queryFn: async ({ signal }) =>
    unwrap(await api.GET('/api/v1/users', { signal })),
})
const email = ref('')
const password = ref('')
const current = ref('')
const role = ref<'admin' | 'user'>('user')
const selected = ref<Account>()
const open = ref(false)
const message = ref('')
function edit(user: Account) {
  selected.value = { ...user }
  current.value = ''
  open.value = true
  action.reset()
}
const action = useMutation({
  mutationFn: async (kind: 'create' | 'save' | 'revoke') => {
    message.value = ''
    if (kind === 'create') {
      if (password.value.length < 12 || password.value.length > 128)
        throw new Error('Use a password between 12 and 128 characters.')
      unwrap(
        await api.POST('/api/v1/users', {
          body: {
            email: email.value,
            password: password.value,
            role: role.value,
            current_password: current.value,
          },
        }),
      )
    } else {
      const account = selected.value!
      const params = { path: { id: account.id } }
      if (kind === 'save')
        unwrap(
          await api.PATCH('/api/v1/users/{id}', {
            params,
            body: {
              role: account.role,
              disabled: account.disabled,
              current_password: current.value,
            },
          }),
        )
      else {
        const result = await api.POST('/api/v1/users/{id}/revoke', {
          params,
          body: { current_password: current.value },
        })
        if (result.error) throw new Error(result.error.error.message)
      }
    }
  },
  onSuccess: async () => {
    password.value = ''
    current.value = ''
    open.value = false
    message.value = 'Account updated successfully.'
    await cache.invalidateQueries({ queryKey: ['users'] })
  },
})
</script>
<template>
  <main class="page">
    <PageHeader
      title="Users"
      description="Control access to your Velora installation."
    />
    <State v-if="users.isPending.value" kind="loading" title="Loading users" />
    <State
      v-else-if="users.isError.value"
      kind="error"
      title="Unable to load users"
      description="Administrator access is required."
    />
    <template v-else>
      <Card
        ><CardHeader><CardTitle>Accounts</CardTitle></CardHeader
        ><CardContent>
          <Table
            ><TableHeader
              ><TableRow
                ><TableHead>Email</TableHead><TableHead>Role</TableHead
                ><TableHead>Status</TableHead
                ><TableHead>Actions</TableHead></TableRow
              ></TableHeader
            ><TableBody>
              <TableRow v-for="user in users.data.value" :key="user.id"
                ><TableCell>{{ user.email }}</TableCell
                ><TableCell>{{ user.role }}</TableCell
                ><TableCell
                  ><Badge>{{
                    user.disabled ? 'Disabled' : 'Active'
                  }}</Badge></TableCell
                ><TableCell
                  ><Button
                    variant="outline"
                    :aria-label="`Manage ${user.email}`"
                    @click="edit(user)"
                    >Manage</Button
                  ></TableCell
                ></TableRow
              >
            </TableBody></Table
          >
          <p class="small">Showing up to 200 accounts.</p>
        </CardContent></Card
      >
      <Card class="account-create"
        ><CardHeader><CardTitle>Create account</CardTitle></CardHeader
        ><CardContent>
          <form @submit.prevent="action.mutate('create')">
            <FormField
              id="user-email"
              v-model="email"
              label="Email"
              type="email"
              required
            />
            <FormField
              id="user-password"
              v-model="password"
              label="Initial password"
              type="password"
              autocomplete="new-password"
              minlength="12"
              maxlength="128"
              required
            />
            <Label for="user-role">Role</Label
            ><Select v-model="role"
              ><SelectTrigger id="user-role"><SelectValue /></SelectTrigger
              ><SelectContent
                ><SelectItem value="user">User</SelectItem
                ><SelectItem value="admin">Admin</SelectItem></SelectContent
              ></Select
            >
            <FormField
              id="admin-current"
              v-model="current"
              label="Your current password"
              type="password"
              autocomplete="current-password"
              required
            />
            <Button type="submit" :disabled="action.isPending.value"
              >Create user</Button
            >
          </form>
        </CardContent></Card
      >
    </template>
    <Dialog v-model:open="open"
      ><DialogContent
        ><DialogHeader
          ><DialogTitle>Manage account</DialogTitle
          ><DialogDescription>{{
            selected?.email
          }}</DialogDescription></DialogHeader
        >
        <form v-if="selected" @submit.prevent="action.mutate('save')">
          <p class="small">Account ID: {{ selected.id }}</p>
          <Label for="edit-role">Role</Label
          ><Select v-model="selected.role"
            ><SelectTrigger id="edit-role"><SelectValue /></SelectTrigger
            ><SelectContent
              ><SelectItem value="user">User</SelectItem
              ><SelectItem value="admin">Admin</SelectItem></SelectContent
            ></Select
          >
          <div class="actions">
            <Switch id="user-disabled" v-model="selected.disabled" /><Label
              for="user-disabled"
              >Account disabled</Label
            >
          </div>
          <FormField
            id="edit-current"
            v-model="current"
            label="Your current password"
            type="password"
            autocomplete="current-password"
            required
          />
          <p v-if="action.error.value" role="alert" class="error">
            {{ action.error.value.message }}
          </p>
          <div class="actions">
            <Button type="submit" :disabled="action.isPending.value"
              >Save changes</Button
            ><Button
              type="button"
              variant="outline"
              :disabled="action.isPending.value"
              @click="action.mutate('revoke')"
              >Revoke sessions</Button
            >
          </div>
        </form>
      </DialogContent></Dialog
    >
    <p v-if="action.error.value && !open" role="alert" class="error">
      {{ action.error.value.message }}
    </p>
    <p v-if="message" role="status">{{ message }}</p>
  </main>
</template>
