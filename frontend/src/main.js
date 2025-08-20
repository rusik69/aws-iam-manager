import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import Accounts from './components/Accounts.vue'
import Users from './components/Users.vue'
import UserDetail from './components/UserDetail.vue'

const routes = [
  { path: '/', component: Accounts },
  { path: '/accounts/:accountId/users', component: Users, props: true },
  { path: '/accounts/:accountId/users/:username', component: UserDetail, props: true }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.mount('#app')