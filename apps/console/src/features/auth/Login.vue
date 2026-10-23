<script setup lang="ts">
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useForm } from 'vee-validate'
import { api, unwrap } from '../../lib/api'
import FormField from '../../components/FormField.vue'
import { Button } from '@/components/ui/button'
import mark from '../../assets/velora-mark.svg'
const router = useRouter()
const failure = ref('')
const { defineField, handleSubmit, isSubmitting, errors } = useForm({
  initialValues: { email: '', password: '' },
  validationSchema: {
    email: (v: string) => /.+@.+\..+/.test(v) || 'Enter your email address.',
    password: (v: string) => v.length >= 12 || 'Use at least 12 characters.',
  },
})
const [email] = defineField('email')
const [password] = defineField('password')
const submit = handleSubmit(async (values) => {
  failure.value = ''
  try {
    unwrap(await api.POST('/api/v1/auth/login', { body: values }))
    await router.push('/')
  } catch {
    failure.value =
      'Sign-in failed. Check your credentials and Core availability.'
  }
})
</script>
<template>
  <main class="login-page">
    <div class="login-intro">
      <img :src="mark" alt="Velora" width="72" height="72" />
      <p class="eyebrow">YOUR INFRASTRUCTURE. YOUR CONTROL.</p>
      <h1>A clear view.<br />A solid foundation.</h1>
      <p>One console for the fleet you’re building.</p>
      <Badge>CORE / FOUNDATION</Badge>
    </div>
    <Card class="login-form"
      ><form @submit="submit">
        <h2>Welcome to Velora</h2>
        <p>Sign in to your Console.</p>
        <FormField
          id="email"
          v-model="email"
          label="Email address"
          type="email"
          autocomplete="username"
          :error="errors.email"
        /><FormField
          id="password"
          v-model="password"
          label="Password"
          type="password"
          autocomplete="current-password"
          :error="errors.password"
        />
        <p v-if="failure" role="alert" class="error">{{ failure }}</p>
        <Button type="submit" variant="default" :disabled="isSubmitting"
          >Sign in</Button
        >
        <p class="small">
          First time here? Bootstrap an administrator with the Core CLI.
        </p>
      </form></Card
    >
  </main>
</template>
