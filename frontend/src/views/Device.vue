<template>
  <div class="device">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">设备管理</span>
          <el-button type="primary" :icon="MagicStick" @click="openBatch()">批量生成</el-button>
        </div>
      </template>

      <el-table :data="list" stripe class="desktop-only" style="width:100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="设备名" min-width="160" />
        <el-table-column prop="productName" label="产品" min-width="140" show-overflow-tooltip />
        <el-table-column prop="deviceKey" label="DeviceKey" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.online ? 'success' : 'info'" size="small">{{ row.online ? '在线' : '离线' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="170" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="goDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="mobile-list mobile-only">
        <el-card v-for="row in list" :key="row.id" class="m-item" shadow="hover" @click="goDetail(row)">
          <div class="m-top">
            <span class="m-name">{{ row.name }}</span>
            <el-tag :type="row.online ? 'success' : 'info'" size="small">{{ row.online ? '在线' : '离线' }}</el-tag>
          </div>
          <div class="m-sub">产品: {{ row.productName || '-' }}</div>
          <div class="m-sub">DeviceKey: {{ row.deviceKey || '-' }}</div>
          <div class="m-sub">{{ row.createdAt }}</div>
          <div class="m-actions">
            <el-button size="small" type="primary" @click.stop="goDetail(row)">查看详情</el-button>
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

    <el-dialog v-model="batchVisible" title="批量生成设备" width="420px" destroy-on-close>
      <el-form :model="batchForm" label-width="100px">
        <el-form-item label="选择产品">
          <el-select v-model="batchForm.productId" placeholder="请选择产品" style="width:100%" filterable>
            <el-option v-for="p in products" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量">
          <el-input-number v-model="batchForm.count" :min="1" :max="500" style="width:100%" />
        </el-form-item>
        <el-form-item label="前缀">
          <el-input v-model="batchForm.prefix" placeholder="可选，如 DEV-" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitBatch">生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { MagicStick } from '@element-plus/icons-vue'
import request from '@/utils/request'

const router = useRouter()
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const products = ref([])
const batchVisible = ref(false)
const saving = ref(false)
const batchForm = ref({ productId: null, count: 10, prefix: 'DEV-' })

const load = async () => {
  try {
    const data = await request.get('/admin/device', { params: { page: page.value, size: pageSize.value } })
    list.value = Array.isArray(data?.list) ? data.list : Array.isArray(data?.items) ? data.items : Array.isArray(data) ? data : []
    total.value = data?.total ?? list.value.length
  } catch (e) {
    list.value = []
    total.value = 0
  }
}

const loadProducts = async () => {
  try {
    const data = await request.get('/admin/product', { params: { page: 1, size: 100 } })
    products.value = Array.isArray(data?.list) ? data.list : Array.isArray(data?.items) ? data.items : Array.isArray(data) ? data : []
  } catch (e) { products.value = [] }
}

const openBatch = () => {
  batchForm.value = { productId: products.value[0]?.id || null, count: 10, prefix: 'DEV-' }
  loadProducts()
  batchVisible.value = true
}

const submitBatch = async () => {
  if (!batchForm.value.productId) { ElMessage.warning('请选择产品'); return }
  saving.value = true
  try {
    await request.post('/admin/device/batch', batchForm.value)
    ElMessage.success(`成功生成 ${batchForm.value.count} 个设备`)
    batchVisible.value = false
    load()
  } finally { saving.value = false }
}

const goDetail = (row) => {
  router.push('/device/' + row.id)
}

onMounted(() => { load(); loadProducts() })
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.title { font-weight: 600; color: #303133; }
.m-item { margin-bottom: 10px; cursor: pointer; }
.m-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.m-name { font-weight: 600; color: #303133; }
.m-sub { color: #909399; font-size: 13px; margin-top: 4px; word-break: break-all; }
.m-actions { margin-top: 10px; }
.desktop-only { display: table; }
.mobile-only { display: none; }
@media (max-width: 768px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
}
</style>
