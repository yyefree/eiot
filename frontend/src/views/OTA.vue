<template>
  <div class="ota">
    <!-- Toolbar -->
    <div class="toolbar">
      <el-select v-model="filterProduct" placeholder="全部产品" clearable style="width: 200px" @change="load">
        <el-option v-for="p in products" :key="p.id" :label="p.name" :value="p.id" />
      </el-select>
      <el-button type="primary" :icon="Upload" @click="uploadVisible = true">上传固件</el-button>
    </div>

    <!-- Firmware List -->
    <el-table :data="list" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="version" label="版本号" width="140">
        <template #default="{ row }">
          <div class="firmware-name">
            <el-icon class="firmware-icon"><Document /></el-icon>
            {{ row.version }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="产品" width="140">
        <template #default="{ row }">{{ row.product_name || '-' }}</template>
      </el-table-column>
      <el-table-column prop="size" label="文件大小" width="100">
        <template #default="{ row }">{{ formatSize(row.size) }}</template>
      </el-table-column>
      <el-table-column prop="changelog" label="更新说明" min-width="180" show-overflow-tooltip />
      <el-table-column label="创建时间" width="160">
        <template #default="{ row }">{{ row.created_at || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'done' ? 'success' : row.status === 'pushing' ? 'warning' : 'info'" size="small">
            {{ row.status === 'done' ? '已完成' : row.status === 'pushing' ? '推送中' : '待推送' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openPush(row)">推送升级</el-button>
          <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
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

    <!-- Upload Firmware Dialog -->
    <el-dialog v-model="uploadVisible" title="上传固件" width="520px" destroy-on-close>
      <el-form :model="uploadForm" label-width="100px" size="default">
        <el-form-item label="版本号" required>
          <el-input v-model="uploadForm.version" placeholder="如 v1.0.0" />
        </el-form-item>
        <el-form-item label="目标产品" required>
          <el-select v-model="uploadForm.product_id" placeholder="选择产品" style="width: 100%">
            <el-option v-for="p in products" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="文件地址">
          <el-input v-model="uploadForm.file_url" placeholder="固件下载URL（可选）" />
        </el-form-item>
        <el-form-item label="文件大小">
          <el-input-number v-model="uploadForm.size" :min="0" :step="1024" />
        </el-form-item>
        <el-form-item label="更新说明">
          <el-input v-model="uploadForm.changelog" type="textarea" :rows="3" placeholder="本次更新内容说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">上传</el-button>
      </template>
    </el-dialog>

    <!-- Push to Devices Dialog -->
    <el-dialog v-model="pushVisible" title="推送OTA升级" width="560px" destroy-on-close>
      <el-form :model="pushForm" label-width="100px" size="default">
        <el-form-item label="固件版本">
          <el-input :model-value="pushForm.version" disabled />
        </el-form-item>
        <el-form-item label="目标设备">
          <el-radio-group v-model="pushForm.scope">
            <el-radio label="all">全部设备</el-radio>
            <el-radio label="custom">指定设备</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="pushForm.scope === 'custom'" label="设备列表">
          <el-input v-model="pushForm.deviceIds" type="textarea" :rows="3" placeholder="输入设备ID，每行一个" />
        </el-form-item>
        <el-form-item label="升级策略">
          <el-select v-model="pushForm.strategy" style="width: 100%">
            <el-option label="立即升级" value="immediate" />
            <el-option label="空闲时升级" value="idle" />
            <el-option label="定时升级" value="scheduled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pushVisible = false">取消</el-button>
        <el-button type="primary" :loading="pushing" @click="submitPush">确认推送</el-button>
      </template>
    </el-dialog>

    <!-- Detail Dialog -->
    <el-dialog v-model="detailVisible" title="固件详情" width="520px">
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="版本号">{{ detailData.version }}</el-descriptions-item>
        <el-descriptions-item label="目标产品">{{ detailData.product_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="文件大小">{{ formatSize(detailData.size) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detailData.status }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detailData.created_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="更新说明">{{ detailData.changelog || '无' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Document } from '@element-plus/icons-vue'
import request from '@/utils/request'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const products = ref([])
const filterProduct = ref(null)

const uploadVisible = ref(false)
const uploadForm = ref({ version: '', product_id: null, file_url: '', size: 0, changelog: '' })
const uploadFile = ref(null)
const uploading = ref(false)

const pushVisible = ref(false)
const pushForm = ref({ firmware_id: null, version: '', scope: 'all', deviceIds: '', strategy: 'immediate' })
const pushing = ref(false)

const detailVisible = ref(false)
const detailData = ref({})

const formatSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(1) + ' MB'
}

const loadProducts = async () => {
  try {
    const res = await request.get('/admin/product', { params: { page: 1, size: 100 } })
    products.value = res?.list || res?.items || []
  } catch (e) {
    console.error(e)
  }
}

const load = async () => {
  try {
    const params = { page: page.value, size: pageSize.value }
    if (filterProduct.value) params.product_id = filterProduct.value
    const data = await request.get('/admin/ota', { params })
    list.value = data?.list || data?.items || []
    total.value = data?.total ?? list.value.length
  } catch (e) {
    console.error(e)
    list.value = []
  }
}

const handleFileChange = (file) => {
  uploadFile.value = file.raw
}

const submitUpload = async () => {
  if (!uploadForm.value.version) { ElMessage.warning('请输入版本号'); return }
  if (!uploadForm.value.product_id) { ElMessage.warning('请选择目标产品'); return }
  uploading.value = true
  try {
    await request.post('/admin/ota', {
      product_id: uploadForm.value.product_id,
      version: uploadForm.value.version,
      changelog: uploadForm.value.changelog,
      file_url: uploadForm.value.file_url,
      size: uploadForm.value.size || 0
    })
    ElMessage.success('创建成功')
    uploadVisible.value = false
    load()
  } catch (e) {
    ElMessage.error('创建失败: ' + (e.message || '未知错误'))
  } finally {
    uploading.value = false
  }
}

const openPush = (row) => {
  pushForm.value = { firmware_id: row.id, version: row.version, scope: 'all', deviceIds: '', strategy: 'immediate' }
  pushVisible.value = true
}

const submitPush = async () => {
  pushing.value = true
  try {
    await request.post('/admin/ota/' + pushForm.value.firmware_id + '/push', {
      scope: pushForm.value.scope,
      device_ids: pushForm.value.scope === 'custom' ? pushForm.value.deviceIds.split('\n').filter(Boolean) : [],
      strategy: pushForm.value.strategy
    })
    ElMessage.success('推送成功')
    pushVisible.value = false
  } catch (e) {
    ElMessage.error('推送失败: ' + (e.message || '未知错误'))
  } finally {
    pushing.value = false
  }
}

const viewDetail = (row) => {
  detailData.value = row
  detailVisible.value = true
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除固件 v' + row.version + '？', '提示', { type: 'warning' })
    await request.delete('/admin/ota/' + row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') console.error(e)
  }
}

onMounted(() => { load(); loadProducts() })
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.firmware-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
}

.firmware-icon {
  color: #007DFF;
}
</style>
