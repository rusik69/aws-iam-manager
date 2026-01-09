import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import Login from './components/Login.vue'
import AllUsers from './components/AllUsers.vue'
import UserDetail from './components/UserDetail.vue'
import PublicIPs from './components/PublicIPs.vue'
import SecurityGroups from './components/SecurityGroups.vue'
import SecurityGroupDetail from './components/SecurityGroupDetail.vue'
import Snapshots from './components/Snapshots.vue'
import EC2Instances from './components/EC2Instances.vue'
import EBSVolumes from './components/EBSVolumes.vue'
import S3Buckets from './components/S3Buckets.vue'
import Roles from './components/Roles.vue'
import LoadBalancers from './components/LoadBalancers.vue'
import VPCs from './components/VPCs.vue'
import NATGateways from './components/NATGateways.vue'
import AzureEnterpriseApps from './components/AzureEnterpriseApps.vue'
import AzureVMs from './components/AzureVMs.vue'
import AzureStorage from './components/AzureStorage.vue'

const routes = [
  { path: '/login', name: 'Login', component: Login },
  { path: '/', name: 'AllUsers', component: AllUsers, meta: { requiresAuth: true } },
  { path: '/public-ips', name: 'PublicIPs', component: PublicIPs, meta: { requiresAuth: true } },
  { path: '/security-groups', name: 'SecurityGroups', component: SecurityGroups, meta: { requiresAuth: true } },
  { path: '/security-groups/:accountId/:region/:groupId', name: 'SecurityGroupDetail', component: SecurityGroupDetail, props: true, meta: { requiresAuth: true } },
  { path: '/snapshots', name: 'Snapshots', component: Snapshots, meta: { requiresAuth: true } },
  { path: '/ec2-instances', name: 'EC2Instances', component: EC2Instances, meta: { requiresAuth: true } },
  { path: '/ebs-volumes', name: 'EBSVolumes', component: EBSVolumes, meta: { requiresAuth: true } },
  { path: '/s3-buckets', name: 'S3Buckets', component: S3Buckets, meta: { requiresAuth: true } },
  { path: '/roles', name: 'Roles', component: Roles, meta: { requiresAuth: true } },
  { path: '/load-balancers', name: 'LoadBalancers', component: LoadBalancers, meta: { requiresAuth: true } },
  { path: '/vpcs', name: 'VPCs', component: VPCs, meta: { requiresAuth: true } },
  { path: '/nat-gateways', name: 'NATGateways', component: NATGateways, meta: { requiresAuth: true } },
  { path: '/accounts/:accountId/users/:username', name: 'UserDetail', component: UserDetail, props: true, meta: { requiresAuth: true } },
  { path: '/azure/enterprise-apps', name: 'AzureEnterpriseApps', component: AzureEnterpriseApps, meta: { requiresAuth: true } },
  { path: '/azure/vms', name: 'AzureVMs', component: AzureVMs, meta: { requiresAuth: true } },
  { path: '/azure/storage', name: 'AzureStorage', component: AzureStorage, meta: { requiresAuth: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard to check authentication
router.beforeEach(async (to, from, next) => {
  // Skip auth check for login page
  if (to.path === '/login') {
    // Check if already authenticated
    try {
      const response = await fetch('/api/auth/check', {
        credentials: 'include'
      })
      const data = await response.json()
      if (data.authenticated) {
        next('/')
        return
      }
    } catch (err) {
      // Ignore errors, allow login page
    }
    next()
    return
  }

  // Check authentication for protected routes
  if (to.meta.requiresAuth) {
    try {
      const response = await fetch('/api/auth/check', {
        credentials: 'include'
      })
      const data = await response.json()
      if (data.authenticated) {
        next()
      } else {
        next('/login')
      }
    } catch (err) {
      next('/login')
    }
  } else {
    next()
  }
})

const app = createApp(App)
app.use(router)
app.mount('#app')