<template>
  <div class="layout-wrapper" v-if="!isMobile">
    <!-- Top Header Bar -->
    <header class="top-header">
      <div class="header-left">
        <div class="logo-area" @click="router.push('/dashboard')">
          <div class="logo-icon">飞</div>
          <span class="logo-text">飞燕IoT</span>
        </div>
        <el-divider direction="vertical" />
        <el-dropdown trigger="click">
          <div class="project-selector">
            <span class="project-name">{{ currentProject || '我的项目' }}</span>
            <el-icon><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item>我的项目</el-dropdown-item>
              <el-dropdown-item divided>切换项目</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="header-right">
        <el-badge :value="notifyCount" :hidden="!notifyCount" class="notify-badge">
          <el-icon class="header-icon" :size="20" @click="router.push('/messages')"><Bell /></el-icon>
        </el-badge>
        <el-dropdown @command="handleCommand">
          <div class="user-dropdown">
            <el-avatar :size="28" :icon="UserFilled" class="user-avatar" />
            <span class="user-name">{{ userName }}</span>
            <el-icon><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">
                <el-icon><User /></el-icon>个人信息
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <!-- Main Layout -->
    <div class="main-layout">
      <!-- Left Sidebar -->
      <aside class="side-nav" :class="{ collapsed: isCollapse }">
        <div class="nav-scroll">
          <div v-for="group in menuGroups" :key="group.label" class="nav-group">
            <div class="nav-group-title" v-if="!isCollapse">{{ group.label }}</div>
            <div v-else class="nav-group-divider"></div>
            <template v-for="item in group.children" :key="item.path">
              <div
                class="nav-item"
                :class="{ active: isActive(item.path), hidden: item.adminOnly && !isAdmin }"
                @click="router.push(item.path)"
              >
                <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
                <span class="nav-label" v-if="!isCollapse">{{ item.title }}</span>
              </div>
            </template>
          </div>
        </div>
        <div class="collapse-trigger" @click="isCollapse = !isCollapse">
          <el-icon v-if="isCollapse"><Expand /></el-icon>
          <el-icon v-else><Fold /></el-icon>
        </div>
      </aside>

      <!-- Content Area -->
      <div class="content-area">
        <!-- Breadcrumb -->
        <div class="breadcrumb-bar">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta?.title">{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <!-- Page Content -->
        <div class="page-content">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </div>
    </div>
  </div>

  <!-- Mobile Layout -->
  <div class="mobile-container" v-else>
    <div class="mobile-header">
      <el-icon :size="22" @click="drawerVisible = true"><Menu /></el-icon>
      <div class="mobile-logo">
        <div class="mobile-logo-icon">飞</div>
        <span>飞燕IoT</span>
      </div>
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

    <el-drawer v-model="drawerVisible" title="导航菜单" size="280px" direction="ltr">
      <div class="mobile-nav-header">
        <div class="mobile-nav-logo">
          <div class="mobile-logo-icon">飞</div>
          <span>飞燕IoT</span>
        </div>
      </div>
      <div class="mobile-nav-scroll">
        <div v-for="group in menuGroups" :key="group.label" class="nav-group">
          <div class="nav-group-title">{{ group.label }}</div>
          <div
            v-for="item in group.children"
            :key="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path), hidden: item.adminOnly && !isAdmin }"
            @click="router.push(item.path); drawerVisible = false"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span class="nav-label">{{ item.title }}</span>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Odometer, FolderOpened, Box, Cpu, User, Share, Iphone,
  ArrowDown, SwitchButton, Menu, UserFilled, Bell,
  Fold, Expand, Setting, Tickets, Upload, Message, ChatDotRound, Connection
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const isCollapse = ref(false)
const drawerVisible = ref(false)
const isMobile = ref(false)
const userInfo = ref(null)
const currentProject = ref('我的项目')
const notifyCount = ref(0)

const loadUnreadCount = async () => {
  try {
    const data = await request.get('/message/unread')
    notifyCount.value = typeof data === 'number' ? data : (data?.count || 0)
  } catch (e) { /* ignore */ }
}

const menuGroups = [
  {
    label: '工作台',
    children: [
      { path: '/dashboard', title: '概览', icon: Odometer }
    ]
  },
  {
    label: '设备管理',
    children: [
      { path: '/project', title: '项目管理', icon: FolderOpened },
      { path: '/product', title: '产品管理', icon: Box },
      { path: '/device', title: '设备管理', icon: Cpu },
      { path: '/batch', title: '量产管理', icon: Tickets }
    ]
  },
  {
    label: '运营管理',
    children: [
      { path: '/ota', title: 'OTA升级', icon: Upload },
      { path: '/messages', title: '消息中心', icon: Message }
    ]
  },
  {
    label: '系统管理',
    children: [
      { path: '/users', title: '用户管理', icon: User, adminOnly: true },
      { path: '/share', title: '设备共享', icon: Share }
    ]
  }
]

const isAdmin = computed(() => {
  return userInfo.value?.role === 'admin' || userInfo.value?.is_admin
})

const userName = computed(() => {
  return userInfo.value?.nickname || userInfo.value?.phone || userInfo.value?.username || '用户'
})

const isActive = (path) => {
  const p = route.path
  if (path === '/dashboard') return p === '/dashboard' || p === '/'
  if (path === '/device') return p.startsWith('/device')
  return p === path
}

const checkSize = () => {
  const w = window.innerWidth
  isMobile.value = w <= 768
  if (w <= 1024) isCollapse.value = true
}

const loadUser = async () => {
  try {
    const data = await request.get('/user/info')
    userInfo.value = data || {}
    localStorage.setItem('eiot_user', JSON.stringify(userInfo.value))
  } catch (e) {
    try {
      const cached = localStorage.getItem('eiot_user')
      if (cached) userInfo.value = JSON.parse(cached)
    } catch (_) {}
  }
}

const handleCommand = async (cmd) => {
  if (cmd === 'logout') {
    try {
      await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
      localStorage.removeItem('eiot_token')
      localStorage.removeItem('eiot_user')
      router.push('/login')
      ElMessage.success('已退出登录')
    } catch (e) {}
  }
}

onMounted(() => {
  checkSize()
  loadUser()
  loadUnreadCount()
  window.addEventListener('resize', checkSize)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkSize)
})
</script>

<style scoped>
/* Top Header */
.top-header {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid var(--border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  box-shadow: var(--shadow-sm);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #007DFF, #00B0FF);
  border-radius: 8px;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 16px;
}

.logo-text {
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
  background: linear-gradient(135deg, #007DFF, #00B0FF);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.project-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 6px;
  transition: background 0.2s;
}

.project-selector:hover {
  background: var(--bg);
}

.project-name {
  font-size: 14px;
  color: var(--text);
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.header-icon {
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s;
}

.header-icon:hover {
  color: var(--primary);
}

.notify-badge {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-dropdown:hover {
  background: var(--bg);
}

.user-avatar {
  background: var(--primary);
}

.user-name {
  font-size: 14px;
  color: var(--text);
}

/* Main Layout */
.layout-wrapper {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-layout {
  flex: 1;
  display: flex;
  margin-top: 56px;
  overflow: hidden;
}

/* Side Navigation */
.side-nav {
  width: 200px;
  background: var(--sidebar-bg);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  flex-shrink: 0;
}

.side-nav.collapsed {
  width: 60px;
}

.nav-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.nav-scroll::-webkit-scrollbar {
  width: 4px;
}

.nav-scroll::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.2);
  border-radius: 2px;
}

.nav-group {
  margin-bottom: 8px;
}

.nav-group-title {
  padding: 12px 20px 8px;
  font-size: 12px;
  color: rgba(255,255,255,0.4);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
  overflow: hidden;
}

.nav-group-divider {
  height: 1px;
  background: rgba(255,255,255,0.1);
  margin: 8px 16px;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 10px 20px;
  color: rgba(255,255,255,0.65);
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  position: relative;
}

.nav-item:hover {
  color: #fff;
  background: rgba(255,255,255,0.08);
}

.nav-item.active {
  color: #fff;
  background: var(--primary);
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #fff;
  border-radius: 0 2px 2px 0;
}

.nav-item.hidden {
  display: none;
}

.nav-icon {
  font-size: 18px;
  margin-right: 10px;
  flex-shrink: 0;
}

.collapsed .nav-icon {
  margin-right: 0;
}

.nav-label {
  font-size: 14px;
}

.collapse-trigger {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255,255,255,0.45);
  cursor: pointer;
  border-top: 1px solid rgba(255,255,255,0.1);
  transition: color 0.2s;
}

.collapse-trigger:hover {
  color: #fff;
  background: rgba(255,255,255,0.08);
}

/* Content Area */
.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg);
}

.breadcrumb-bar {
  padding: 12px 20px;
  background: #fff;
  border-bottom: 1px solid var(--border-light);
}

.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

/* Mobile Layout */
.mobile-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg);
}

.mobile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: #fff;
  box-shadow: var(--shadow-sm);
}

.mobile-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--text);
}

.mobile-logo-icon {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, #007DFF, #00B0FF);
  border-radius: 6px;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 13px;
}

.mobile-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.mobile-nav-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-light);
}

.mobile-nav-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 700;
  color: var(--text);
}

.mobile-nav-scroll {
  padding: 8px 0;
}

/* Override el-divider for header */
.top-header .el-divider--vertical {
  height: 24px;
  margin: 0 4px;
}
</style>
