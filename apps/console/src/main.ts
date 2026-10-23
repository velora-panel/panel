import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { api } from './lib/api'
import { sessionMiddleware } from './lib/session'
import App from './app/App.vue'
import { router } from './router'
import '../src/styles/main.css'
const queryClient = new QueryClient()
api.use(sessionMiddleware(queryClient, router, () => fetch('/api/v1/auth/me', { credentials: 'same-origin' })))
createApp(App).use(createPinia()).use(VueQueryPlugin, { queryClient }).use(router).mount('#app')
