<template>
  <div class="share">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">设备共享记录</span>
          <el-button type="primary" :icon="Plus" @click="openCreate()">新增共享</el-button>
        </div>
      </template>

      <el-table :data="list" stripe class="desktop-only" style="width:100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="deviceName" label="设备" min-width="160" />
        <el-table-column prop="sharedTo" label="共享给" min-width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.revoked ? 'info' : 'success'" size="small">{{ row.revoked ? '已撤销' : '有效' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row.revoked" link type="danger" @click="revoke(row)">撤销</el-button>
            <span v-else style="color:#909399">-</span>
          </template>
        </el-table-column>
      </el-table>

      <div class="mobile-list mobile-only">
        <el-card v-for="row in list" :key="row.id" class="m-item" shadow="hover">
          <div class="m-top">
            <span class="m-name">{{ row.deviceName }}</span>
            <el-tag :type="row.revoked ? 'info' : 'success'" size="small">{{ row.revoked ? '已撤销' : '有效' }}</el-tag>
          </div>
          <div class="m-sub">共享给: {{ row.sharedTo }}</div>
          <div class="m-sub">{{ row.createdAt }}</div>
          <div class="m-actions">
            <el-button v-if="!row.revoked" size="small" type="danger" @click="revoke(row)">撤销</el-button>
          </div>
        </el-card>
      </div>

      <el-pagination
        style="margin-top: 16px; justify-content: flex-end; display: flex;"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="load"
        @current-change="load"
      />
    </el-card>

    <el-dialog v-model="dialogVisible" title="新增共享" width="420px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="选择设备">
          <el-select v-model="form.device_id" placeholder="请选择设备" style="width:100%" filterable>
            <el-option v-for="d in devices" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="共享给(用户ID)">
          <el-input-number v-model="form.share_user_id" :min="1" :controls="false" style="width:100%" placeholder="输入用户ID" />
        </el-form-item>
        <el-form-item label="权限">
          <el-select v-model="form.permission" style="width:100%">
            <el-option label="仅查看" value="read" />
            <el-option label="可控制" value="control" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效时长(小时)">
          <el-input-number v-model="form.hours" :min="0" :controls="false" style="width:100%" placeholder="0表示永久" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">共享</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const devices = ref([])
const dialogVisible = ref(false)
const saving = ref(false)
const form = ref({ device_id: null, share_user_id: null, permission: 'read', hours: 0 })

const load = async () => {
  try {
    const data = await request.get('/device/share', { params: { page: page.value, pageSize: pageSize.value } })
    list.value = data?.list || data?.items || data || []
    total.value = data?.total || list.value.length
  } catch (e) {
    list.value = []
    total.value = 0
  }
}

const loadDevices = async () => {
  try {
    const data = await request.get('/device', { params: { pageSize: 100 } })
    devices.value = data?.list || data?.items || data || []
  } catch (e) { devices.value = [] }
}

const openCreate = () => {
  form.value = { device_id: devices.value[0]?.id || null, share_user_id: null, permission: 'read', hours: 0 }
  loadDevices()
  dialogVisible.value = true
}

const submit = async () => {
  if (!form.value.device_id) { ElMessage.warning('请选择设备'); return }
  if (!form.value.share_user_id) { ElMessage.warning('请输入共享用户ID'); return }
  saving.value = true
  try {
    await request.post('/device/share', form.value)
    ElMessage.success('共享成功')
    dialogVisible.value = false
    load()
  } finally { saving.value = false }
}

const revoke = async (row) => {
  try {
    await ElMessageBox.confirm('确认撤销该共享？', '提示', { type: 'warning' })
    await request.delete('/device/share/' + row.id)
    ElMessage.success('已撤销')
    load()
  } catch (e) { if (e !== 'cancel') console.error(e) }
}

onMounted(() => { load(); loadDevices() })
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.title { font-weight: 600; color: #303133; }
.m-item { margin-bottom: 10px; }
.m-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.m-name { font-weight: 600; color: #303133; }
.m-sub { color: #909399; font-size: 13px; margin-top: 4px; word-break: break-all; }
.m-actions { margin-top: 8px; }
.desktop-only { display: table; }
.mobile-only { display: none; }
@media (max-width: 768px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
}
</style>
