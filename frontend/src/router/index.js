import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/login', component: () => import('../views/Login.vue') },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '概览', icon: 'Odometer' } },
      { path: 'project', name: 'Project', component: () => import('../views/Project.vue'), meta: { title: '项目管理', icon: 'FolderOpened' } },
      { path: 'product', name: 'Product', component: () => import('../views/Product.vue'), meta: { title: '产品管理', icon: 'Box' } },
      { path: 'device', name: 'Device', component: () => import('../views/Device.vue'), meta: { title: '设备管理', icon: 'Cpu' } },
      { path: 'device/:id', name: 'DeviceDetail', component: () => import('../views/DeviceDetail.vue'), meta: { title: '设备详情', icon: 'Monitor' }, hidden: true },
      { path: 'batch', name: 'Batch', component: () => import('../views/Device.vue'), meta: { title: '量产管理', icon: 'Tickets' } },
      { path: 'ota', name: 'OTA', component: () => import('../views/OTA.vue'), meta: { title: 'OTA升级', icon: 'Upload' } },
      { path: 'messages', name: 'Messages', component: () => import('../views/Messages.vue'), meta: { title: '消息中心', icon: 'Message' } },
      { path: 'users', name: 'Users', component: () => import('../views/Users.vue'), meta: { title: '用户管理', icon: 'User' } },
      { path: 'share', name: 'DeviceShare', component: () => import('../views/Share.vue'), meta: { title: '设备共享', icon: 'Share' } },
      { path: 'mobile-ui', name: 'MobileUI', component: () => import('../views/MobileUI.vue'), meta: { title: '移动端UI', icon: 'Mobile' } }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('eiot_token')
  if (!token && to.path !== '/login') {
    next('/login')
  } else if (to.path === '/users' || to.path === '/mobile-ui') {
    try {
      const user = JSON.parse(localStorage.getItem('eiot_user') || '{}')
      if (user.role !== 'admin') { next('/dashboard'); return }
    } catch (_) {}
    next()
  } else {
    next()
  }
})

export default router
