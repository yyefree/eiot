<template>
  <div class="messages">
    <!-- Toolbar -->
    <div class="toolbar">
      <el-radio-group v-model="filterType" @change="filterMessages">
        <el-radio-button label="all">全部</el-radio-button>
        <el-radio-button label="unread">未读</el-radio-button>
        <el-radio-button label="read">已读</el-radio-button>
      </el-radio-group>
      <el-select v-model="filterCategory" placeholder="消息类型" clearable style="width: 140px" @change="filterMessages">
        <el-option label="系统通知" value="system" />
        <el-option label="设备告警" value="alert" />
        <el-option label="OTA升级" value="ota" />
        <el-option label="操作日志" value="operation" />
      </el-select>
      <el-button type="primary" @click="markAllRead" :disabled="!unreadCount">全部标记已读</el-button>
    </div>

    <!-- Message List -->
    <div class="message-list">
      <div v-if="filteredMessages.length === 0" class="empty-state">
        <el-empty description="暂无消息" />
      </div>
      <div v-for="msg in filteredMessages" :key="msg.id" class="message-item" :class="{ unread: !msg.read }" @click="readMessage(msg)">
        <div class="message-dot" :class="{ read: msg.read }"></div>
        <div class="message-content">
          <div class="message-header">
            <span class="message-title">{{ msg.title }}</span>
            <el-tag :type="getCategoryType(msg.category)" size="small">{{ getCategoryLabel(msg.category) }}</el-tag>
          </div>
          <div class="message-desc">{{ msg.content }}</div>
          <div class="message-time">{{ msg.time }}</div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <el-pagination
      v-if="total > pageSize"
      style="margin-top: 16px; justify-content: flex-end; display: flex;"
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
    />

    <!-- Message Detail Dialog -->
    <el-dialog v-model="detailVisible" title="消息详情" width="520px">
      <div class="detail-content">
        <h3 style="margin: 0 0 12px; color: #1D2129;">{{ detailMsg.title }}</h3>
        <div style="color: #86909C; font-size: 12px; margin-bottom: 16px;">
          {{ detailMsg.time }} · {{ getCategoryLabel(detailMsg.category) }}
        </div>
        <div style="color: #4E5969; line-height: 1.8;">{{ detailMsg.content }}</div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import request from '@/utils/request'

const messages = ref([])
const filterType = ref('all')
const filterCategory = ref('')
const page = ref(1)
const pageSize = ref(20)
const detailVisible = ref(false)
const detailMsg = ref({})

const total = computed(() => messages.value.length)
const unreadCount = computed(() => messages.value.filter(m => !m.read).length)

const filteredMessages = computed(() => {
  let result = messages.value
  if (filterType.value === 'unread') result = result.filter(m => !m.read)
  if (filterType.value === 'read') result = result.filter(m => m.read)
  if (filterCategory.value) result = result.filter(m => m.category === filterCategory.value)
  return result.slice((page.value - 1) * pageSize.value, page.value * pageSize.value)
})

const getCategoryType = (category) => {
  const map = { system: 'info', alert: 'danger', ota: 'warning', operation: '' }
  return map[category] || ''
}

const getCategoryLabel = (category) => {
  const map = { system: '系统通知', alert: '设备告警', ota: 'OTA升级', operation: '操作日志' }
  return map[category] || '其他'
}

const loadMessages = async () => {
  try {
    const data = await request.get('/message', { params: { page: 1, size: 100 } })
    const list = data?.list || data || []
    messages.value = list.map(m => ({
      ...m,
      category: m.type || 'system',
      time: m.created_at || '-'
    }))
  } catch (e) {
    console.error(e)
    messages.value = []
  }
}

const filterMessages = () => {
  page.value = 1
}

const readMessage = async (msg) => {
  msg.read = true
  detailMsg.value = msg
  detailVisible.value = true
  try { await request.put('/message/' + msg.id + '/read', {}) } catch (e) { /* ignore */ }
}

const markAllRead = async () => {
  try {
    await request.put('/message/read-all', {})
    messages.value.forEach(m => m.read = true)
  } catch (e) { console.error(e) }
}

onMounted(() => { loadMessages() })
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.message-list {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  border: 1px solid #F2F3F5;
  overflow: hidden;
}

.empty-state {
  padding: 60px 0;
}

.message-item {
  display: flex;
  align-items: flex-start;
  padding: 16px 20px;
  border-bottom: 1px solid #F2F3F5;
  cursor: pointer;
  transition: background 0.2s ease;
}

.message-item:hover {
  background: #F7F8FA;
}

.message-item:last-child {
  border-bottom: none;
}

.message-item.unread {
  background: #E6F7FF;
}

.message-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #007DFF;
  margin-top: 6px;
  margin-right: 12px;
  flex-shrink: 0;
}

.message-dot.read {
  background: #C9CDD4;
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.message-title {
  font-size: 14px;
  font-weight: 500;
  color: #1D2129;
}

.message-desc {
  font-size: 13px;
  color: #86909C;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.message-time {
  font-size: 12px;
  color: #C9CDD4;
  margin-top: 6px;
}

.detail-content {
  line-height: 1.8;
}
</style>
