import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import AllUsers from './components/AllUsers.vue'
import UserDetail from './components/UserDetail.vue'
import PublicIPs from './components/PublicIPs.vue'
import SecurityGroups from './components/SecurityGroups.vue'
import SecurityGroupDetail from './components/SecurityGroupDetail.vue'
import Snapshots from './components/Snapshots.vue'
import EC2Instances from './components/EC2Instances.vue'
import EBSVolumes from './components/EBSVolumes.vue'
import S3Buckets from './components/S3Buckets.vue'
import AzureEnterpriseApps from './components/AzureEnterpriseApps.vue'

const routes = [
  { path: '/', name: 'AllUsers', component: AllUsers },
  { path: '/public-ips', name: 'PublicIPs', component: PublicIPs },
  { path: '/security-groups', name: 'SecurityGroups', component: SecurityGroups },
  { path: '/security-groups/:accountId/:region/:groupId', name: 'SecurityGroupDetail', component: SecurityGroupDetail, props: true },
  { path: '/snapshots', name: 'Snapshots', component: Snapshots },
  { path: '/ec2-instances', name: 'EC2Instances', component: EC2Instances },
  { path: '/ebs-volumes', name: 'EBSVolumes', component: EBSVolumes },
  { path: '/s3-buckets', name: 'S3Buckets', component: S3Buckets },
  { path: '/accounts/:accountId/users/:username', name: 'UserDetail', component: UserDetail, props: true },
  { path: '/azure/enterprise-apps', name: 'AzureEnterpriseApps', component: AzureEnterpriseApps }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.mount('#app')