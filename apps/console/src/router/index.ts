import { createRouter, createWebHistory } from 'vue-router'
import { requireSession } from './guard'
import Login from '../features/auth/Login.vue'
import Shell from '../layouts/Shell.vue'
import Dashboard from '../features/dashboard/Dashboard.vue'
import Nodes from '../features/nodes/Nodes.vue'
import NodeDetail from '../features/nodes/NodeDetail.vue'
import Placeholder from '../components/Placeholder.vue'
import Users from '../features/users/Users.vue'
import Settings from '../features/auth/Settings.vue'
export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: Login },
    {
      path: '/',
      component: Shell,

      children: [
        { path: '', component: Dashboard },
        { path: 'nodes', component: Nodes },
        { path: 'nodes/:id', component: NodeDetail },
        {
          path: 'instances',
          component: Placeholder,
          props: {
            title: 'Instances',
            description:
              'Workloads will appear here when Beacon orchestration is available.',
          },
        },
        {
          path: 'instances/:id',
          component: Placeholder,
          props: {
            title: 'Instance unavailable',
            description: 'Instance management has not been implemented.',
          },
        },
        {
          path: 'blueprints',
          component: Placeholder,
          props: {
            title: 'Blueprints',
            description:
              'Reusable workload definitions are planned for a future release.',
          },
        },
        { path: 'settings', component: Settings },
        { path: 'users', component: Users },
      ],
    },
    {
      path: '/error',
      component: Placeholder,
      props: {
        title: 'Core unavailable',
        description: 'Check Core and PostgreSQL, then try again.',
      },
    },
    {
      path: '/:pathMatch(.*)*',
      component: Placeholder,
      props: {
        title: 'Page not found',
        description: 'This address does not exist in your Console.',
      },
    },
  ],
})

router.beforeEach((to) =>
  to.matched.some((record) => record.components?.default === Shell)
    ? requireSession()
    : true,
)
