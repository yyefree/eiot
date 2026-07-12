<template>
  <el-container class="layout-container" v-if="!isMobile">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
      <div class="aside-logo">
        <el-icon :size="24" color="#fff"><Link /></el-icon>
        <span v-if="!isCollapse" class="logo-text">EIOT</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        router
        background-color="#001529"
        text-color="#cfd8dc"
        active-text-color="#409eff"
        class="layout-menu"
      >
        <el-menu-item v-for="item in visibleMenus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <div class="header-left">
          <el-icon class="collapse-btn" :size="18" @click="isCollapse = !isCollapse"><Fold v-if="!isCollapse" /><Expand v-else /></el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ route.meta?.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" :icon="UserFilled" />
              <span class="username">{{ userName }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout"><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in"><component :is="Component" /></transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>

  <div class="mobile-container" v-else>
    <div class="mobile-header">
      <el-icon :size="22" @click="drawerVisible = true"><Menu /></el-icon>
      <span class="mobile-title">{{ route.meta?.title || 'EIOT' }}</span>
      <el-dropdown @command="handleCommand">
        <el-avatar :size="28" :icon="UserFilled" style="cursor:pointer" />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div class="mobile-content">
      <router-view />
    </div>

    <el-drawer v-model="drawerVisible" title="菜单" size="260px" direction="ltr">
      <div class="mobile-aside-logo">
        <el-icon :size="22"><Link /></el-icon>
        <span class="logo-text">EIOT 平台</span>
      </div>
      <el-menu :default-active="activeMenu" router @select="drawerVisible = false" class="mobile-menu">
        <el-menu-item v-for="item in visibleMenus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DataAnalysis, Coin, Box, Cpu, Monitor, User, Share, Iphone,
  Link, Fold, Expand, ArrowDown, SwitchButton, Menu, UserFilled
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const isCollapse = ref(false)
const drawerVisible = ref(false)
const isMobile = ref(false)
const userInfo = ref(null)

const menus = [
  { path: '/dashboard', title: '概览看板', icon: DataAnalysis, adminOnly: false },
  { path: '/project', title: '项目管理', icon: Coin, adminOnly: false },
  { path: '/product', title: '产品管理', icon: Box, adminOnly: false },
  { path: '/device', title: '设备管理', icon: Cpu, adminOnly: false },
  { path: '/users', title: '用户列表', icon: User, adminOnly: true },
  { path: '/share', title: '共享记录', icon: Share, adminOnly: false },
  { path: '/mobile-ui', title: '移动端UI', icon: Iphone, adminOnly: false }
]

const visibleMenus = computed(() => {
  const isAdmin = userInfo.value?.role === 'admin' || userInfo.value?.is_admin
  return menus.filter(m => !m.adminOnly || isAdmin)
})

const userName = computed(() => {
  return userInfo.value?.nickname || userInfo.value?.phone || userInfo.value?.username || '用户'
})

const activeMenu = computed(() => {
  const p = route.path
  if (p.startsWith('/device/')) return '/device'
  return p
})

const checkSize = () => {
  const w = window.innerWidth
  isMobile.value = w <= 768
  isCollapse.value = w <= 1024
}

const loadUser = async () => {
  try {
    const data = await request.get('/user/info')
    userInfo.value = data || {}
    localStorage.setItem('eiot_user', JSON.stringify(userInfo.value))
  } catch (e) {
    const cached = localStorage.getItem('eiot_user')
    if (cached) userInfo.value = JSON.parse(cached)
  }
}

const handleCommand = async (cmd) => {
  if (cmd === 'logout') {
    try {
      await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
      localStorage.removeItem('eiot_token')
      localStorage.removeItem('eiot_user')
      router.push('/login')
      ElMessage.success('已退出')
    } catch (e) {}
  }
}

onMounted(() => {
  checkSize()
  loadUser()
  window.addEventListener('resize', checkSize)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkSize)
})
</script>

<style scoped>
.layout-container { height: 100vh; }
.layout-aside {
  background: #001529;
  transition: width 0.25s;
  overflow-x: hidden;
}
.aside-logo {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: bold;
  font-size: 18px;
  gap: 8px;
  background: #000c17;
}
.layout-menu {
  border-right: none;
  height: calc(100vh - 56px);
}
.layout-header {
  background: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0,21,41,.08);
  padding: 0 20px;
}
.header-left { display: flex; align-items: center; gap: 16px; }
.header-right { display: flex; align-items: center; }
.collapse-btn { cursor: pointer; color: #333; }
.user-info { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.username { font-size: 14px; color: #333; }
.layout-main { background: #f5f7fa; padding: 16px; }

.mobile-container { display: flex; flex-direction: column; height: 100vh; background: #f5f7fa; }
.mobile-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 16px; background: #fff; box-shadow: 0 1px 4px rgba(0,21,41,.08);
}
.mobile-title { font-size: 16px; font-weight: 600; color: #333; }
.mobile-content { flex: 1; overflow-y: auto; padding: 12px; }
.mobile-aside-logo {
  height: 56px; display: flex; align-items: center; justify-content: center;
  gap: 8px; background: #001529; color: #fff; margin: -20px -20px 10px;
}
.mobile-menu { border-right: none; }

.fade-enter-active, .fade-leave-active { transition: opacity .2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
