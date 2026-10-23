<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { useTitle } from '@vueuse/core'
import {
  LayoutDashboard,
  Server,
  Boxes,
  Layers,
  Settings,
  PanelLeft,
  LogOut,
} from '@lucide/vue'
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '@/components/ui/dropdown-menu'
import { TooltipProvider, Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip' 
import { usePreferences } from '../stores/preferences'
import { api, unwrap } from '../lib/api'
import { Button } from '@/components/ui/button'
import { SidebarProvider, Sidebar, SidebarContent, SidebarHeader, SidebarFooter, SidebarMenu, SidebarMenuItem, SidebarMenuButton, SidebarGroup, SidebarGroupLabel } from '@/components/ui/sidebar'
import mark from '../assets/velora-mark.svg'
const preferences = usePreferences()
const router = useRouter()
const cache = useQueryClient()
const failure = ref('')
const me = useQuery({ queryKey: ['me'], queryFn: async ({ signal }) => unwrap(await api.GET('/api/v1/auth/me', { signal })) })
useTitle('Velora Console')
const links = [
  { to: '/', label: 'Overview', icon: LayoutDashboard },
  { to: '/nodes', label: 'Nodes', icon: Server },
  { to: '/instances', label: 'Instances', icon: Boxes },
  { to: '/blueprints', label: 'Blueprints', icon: Layers },
  { to: '/settings', label: 'Settings', icon: Settings },
]
async function logout() {
  failure.value = ''
  try {
    const { response } = await api.POST('/api/v1/auth/logout')
    if (!response.ok) throw new Error('Logout failed')
    cache.clear()
    await router.push('/login')
  } catch {
    failure.value = 'Unable to sign out. Please try again.'
  }
}
</script>
<template>
  <SidebarProvider :open="!preferences.collapsed" :class="{ collapsed: preferences.collapsed }" @update:open="preferences.collapsed = !$event">
    <a class="skip-link" href="#content">Skip to content</a>
    <Sidebar collapsible="icon"><SidebarHeader>
      <RouterLink class="brand" to="/"
        ><img :src="mark" alt="Velora" width="36" height="36" /><span
          >velora<span class="brand-dot">.</span></span
        ></RouterLink
      >
</SidebarHeader><SidebarContent><SidebarGroup><SidebarGroupLabel>WORKSPACE</SidebarGroupLabel>
      <nav aria-label="Main navigation"><SidebarMenu>
        <SidebarMenuItem v-for="link in links" :key="link.to"><SidebarMenuButton as-child><RouterLink
          :to="link.to"
          :aria-label="link.label"
          ><component :is="link.icon" :size="18" /><span>{{
            link.label
          }}</span></RouterLink></SidebarMenuButton></SidebarMenuItem>
<SidebarMenuItem v-if="me.data.value?.role === 'admin'"><SidebarMenuButton as-child><RouterLink to="/users">Users</RouterLink></SidebarMenuButton></SidebarMenuItem></SidebarMenu></nav></SidebarGroup></SidebarContent><SidebarFooter>
      <div class="sidebar-footer">
        <span class="status-dot" /><span
          >Core foundation<br /><small>v0.1.0</small></span
        >
      </div>
</SidebarFooter></Sidebar>
    <div class="workspace min-w-0 flex-1">
      <header class="topbar">
        <TooltipProvider
          ><Tooltip
            ><TooltipTrigger as-child
              ><Button
                aria-label="Toggle sidebar"
                :aria-expanded="!preferences.collapsed"
                @click="preferences.toggle"
                ><PanelLeft :size="18" /></Button></TooltipTrigger
            ><TooltipContent class="tooltip"
                >Toggle sidebar</TooltipContent
              ></Tooltip
          ></TooltipProvider
        ><span class="breadcrumb">Workspace <span>/</span> Console</span
        ><DropdownMenu
          ><DropdownMenuTrigger as-child
            ><Button class="account-button"
              >Account</Button
            ></DropdownMenuTrigger
          ><DropdownMenuContent class="dropdown card"
              ><DropdownMenuItem class="menu-item" @select="logout"
                ><LogOut :size="16" /> Sign out</DropdownMenuItem
              ></DropdownMenuContent
            ></DropdownMenu
        >
      </header>
      <p v-if="failure" class="error" role="alert">{{ failure }}</p>
      <div id="content" tabindex="-1"><RouterView /></div>
      <footer class="page-footer">
        VELORA <span>Built for ownership.</span>
      </footer>
    </div>
  </SidebarProvider>
</template>
