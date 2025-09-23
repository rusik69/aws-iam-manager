import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import AllUsers from './components/AllUsers.vue'
import UserDetail from './components/UserDetail.vue'
import PublicIPs from './components/PublicIPs.vue'
import SecurityGroups from './components/SecurityGroups.vue'
import SecurityGroupDetail from './components/SecurityGroupDetail.vue'

const routes = [
  { path: '/', name: 'AllUsers', component: AllUsers },
  { path: '/public-ips', name: 'PublicIPs', component: PublicIPs },
  { path: '/security-groups', name: 'SecurityGroups', component: SecurityGroups },
  { path: '/security-groups/:accountId/:region/:groupId', name: 'SecurityGroupDetail', component: SecurityGroupDetail, props: true },
  { path: '/accounts/:accountId/users/:username', name: 'UserDetail', component: UserDetail, props: true }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.mount('#app')