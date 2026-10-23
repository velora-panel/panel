<script setup lang="ts">
import { ref } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { api, unwrap } from '../../lib/api'
import PageHeader from '../../components/PageHeader.vue'
import State from '../../components/State.vue'
import FormField from '../../components/FormField.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
const cache = useQueryClient()
const email = ref('')
const current = ref('')
const next = ref('')
const confirmation = ref('')
const message = ref('')
const user = useQuery({ queryKey: ['me'], queryFn: async ({ signal }) => unwrap(await api.GET('/api/v1/auth/me', { signal })) })
const update = useMutation({
 mutationFn: async (kind: 'profile' | 'password') => {
  message.value = ''
  if (kind === 'profile') {
   const result = unwrap(await api.PATCH('/api/v1/auth/profile', { body: { email: email.value, current_password: current.value } }))
   cache.setQueryData(['me'], result)
  } else {
   if (next.value.length < 12 || next.value.length > 128) throw new Error('Use a password between 12 and 128 characters.')
   if (next.value !== confirmation.value) throw new Error('Passwords do not match.')
   const result = await api.PATCH('/api/v1/auth/password', { body: { current_password: current.value, new_password: next.value } })
   if (result.error) throw new Error(result.error.error.message)
  }
 },
 onSuccess: () => { current.value = ''; next.value = ''; confirmation.value = ''; message.value = 'Account updated successfully.' },
})
</script>
<template>
 <main class="page">
  <PageHeader title="Settings" description="Manage your profile and account security." />
  <State v-if="user.isPending.value" kind="loading" title="Loading account" />
  <State v-else-if="user.isError.value" kind="error" title="Unable to load account" />
  <div v-else class="settings-grid">
   <Card><CardHeader><CardTitle>Profile</CardTitle></CardHeader><CardContent>
    <p>{{ user.data.value?.email }} <Badge>{{ user.data.value?.role }}</Badge></p>
    <form @submit.prevent="update.mutate('profile')">
     <FormField id="profile-email" v-model="email" label="New email address" type="email" autocomplete="email" required />
     <FormField id="profile-current" v-model="current" label="Current password" type="password" autocomplete="current-password" required />
     <Button type="submit" :disabled="update.isPending.value">Save profile</Button>
    </form>
   </CardContent></Card>
   <Card><CardHeader><CardTitle>Password</CardTitle></CardHeader><CardContent>
    <p>Changing your password signs out your other sessions. This session remains active.</p>
    <form @submit.prevent="update.mutate('password')">
     <FormField id="password-current" v-model="current" label="Current password" type="password" autocomplete="current-password" required />
     <FormField id="password-new" v-model="next" label="New password" type="password" autocomplete="new-password" minlength="12" maxlength="128" required />
     <FormField id="password-confirm" v-model="confirmation" label="Confirm new password" type="password" autocomplete="new-password" required />
     <Button type="submit" :disabled="update.isPending.value">Change password</Button>
    </form>
   </CardContent></Card>
  </div>
  <p v-if="update.error.value" role="alert" class="error">{{ update.error.value.message }}</p>
  <p v-if="message" role="status">{{ message }}</p>
 </main>
</template>
