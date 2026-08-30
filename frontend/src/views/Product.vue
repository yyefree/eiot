<template>
  <div class="product">
    <!-- Filter Bar -->
    <div class="toolbar">
      <el-select v-model="filterProject" placeholder="全部项目" clearable size="default" style="width:200px" @change="load">
        <el-option label="未分类" :value="0" />
        <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
      </el-select>
      <el-button type="primary" :icon="Plus" @click="openCreate()">新建产品</el-button>
      <el-button :icon="Guide" @click="wizardVisible = true">产品开发向导</el-button>
    </div>

    <!-- Desktop Table -->
    <el-table :data="list" stripe class="desktop-only" style="width:100%">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="产品名称" min-width="160">
        <template #default="{ row }">
          <div class="product-name">
            <el-icon class="product-icon"><Box /></el-icon>
            {{ row.name }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="product_key" label="ProductKey" min-width="150" show-overflow-tooltip />
      <el-table-column label="联网方式" prop="network_type" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ netTypeLabel(row.network_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="属性数量" width="100">
        <template #default="{ row }">{{ propertyCount(row) }}</template>
      </el-table-column>
      <el-table-column prop="has_devices" label="绑定设备" width="100">
        <template #default="{ row }">
          <el-tag :type="row.has_devices ? 'success' : 'info'" size="small">{{ row.has_devices ? '已绑定' : '未绑定' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="openMobileUI(row)">编辑UI</el-button>
          <el-button link type="primary" @click="showQR(row)">绑定码</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Mobile List -->
    <div class="mobile-list mobile-only">
      <el-card v-for="row in list" :key="row.id" class="m-item" shadow="hover">
        <div class="m-name">{{ row.name }}</div>
        <div class="m-sub">PK: {{ row.product_key }} · {{ netTypeLabel(row.network_type) }}</div>
        <div class="m-sub">属性: {{ propertyCount(row) }} 项</div>
        <div class="m-actions">
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" @click="openMobileUI(row)">编辑UI</el-button>
          <el-button size="small" type="primary" @click="showQR(row)">绑定码</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </div>
      </el-card>
    </div>

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

    <!-- Product Development Wizard Dialog -->
    <el-dialog v-model="wizardVisible" title="产品开发向导" width="800px" destroy-on-close>
      <div class="wizard-steps">
        <el-steps :active="wizardStep" finish-status="success" align-center>
          <el-step title="功能定义" description="配置物模型" />
          <el-step title="人机交互" description="App面板配置" />
          <el-step title="设备调试" description="在线调试" />
          <el-step title="批量投产" description="量产部署" />
        </el-steps>
      </div>

      <div class="wizard-content">
        <!-- Step 1: Thing Model -->
        <div v-show="wizardStep === 0">
          <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
            定义产品的物模型属性，包括标识符、名称、数据类型和取值范围。
          </el-alert>
          <el-table :data="wizardProperties" size="small" border style="width:100%" empty-text="暂无属性，点击下方按钮添加">
            <el-table-column label="标识符" width="140">
              <template #default="{ row }"><el-input v-model="row.identifier" size="small" placeholder="如 temperature" /></template>
            </el-table-column>
            <el-table-column label="名称" width="130">
              <template #default="{ row }"><el-input v-model="row.name" size="small" placeholder="如 温度" /></template>
            </el-table-column>
            <el-table-column label="数据类型" width="120">
              <template #default="{ row }">
                <el-select v-model="row.dataType" size="small" style="width:100%">
                  <el-option v-for="t in dataTypes" :key="t" :label="t" :value="t" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="单位" width="80">
              <template #default="{ row }"><el-input v-model="row.unit" size="small" /></template>
            </el-table-column>
            <el-table-column label="下限" width="80">
              <template #default="{ row }"><el-input-number v-model="row.min" size="small" controls-position="right" style="width:100%" /></template>
            </el-table-column>
            <el-table-column label="上限" width="80">
              <template #default="{ row }"><el-input-number v-model="row.max" size="small" controls-position="right" style="width:100%" /></template>
            </el-table-column>
            <el-table-column label="操作" width="70" align="center">
              <template #default="{ $index }">
                <el-button link type="danger" size="small" @click="wizardProperties.splice($index, 1)"><el-icon><Delete /></el-icon></el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button size="small" type="primary" style="margin-top: 12px" :icon="Plus" @click="addWizardProperty">添加属性</el-button>
        </div>

        <!-- Step 2: App Panel -->
        <div v-show="wizardStep === 1">
          <el-empty description="人机交互配置">
            <template #image>
              <div style="width: 200px; height: 360px; border: 2px dashed #C9CDD4; border-radius: 20px; display: flex; align-items: center; justify-content: center; background: #F7F8FA;">
                <div style="text-align: center; color: #86909C;">
                  <el-icon :size="48"><Iphone /></el-icon>
                  <div style="margin-top: 12px; font-size: 13px;">App 面板预览</div>
                  <div style="font-size: 12px; margin-top: 4px;">开发中，敬请期待</div>
                </div>
              </div>
            </template>
            <el-button type="primary" disabled>配置面板（即将上线）</el-button>
          </el-empty>
        </div>

        <!-- Step 3: Device Debug -->
        <div v-show="wizardStep === 2">
          <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
            设备调试需要先完成产品开发并添加设备。
          </el-alert>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="调试状态">等待设备上线</el-descriptions-item>
            <el-descriptions-item label="最近心跳">-</el-descriptions-item>
            <el-descriptions-item label="调试设备数">0</el-descriptions-item>
            <el-descriptions-item label="在线设备数">0</el-descriptions-item>
          </el-descriptions>
          <el-empty description="暂无调试数据，请先添加设备">
            <el-button type="primary" @click="wizardVisible = false; $router.push('/device')">前往添加设备</el-button>
          </el-empty>
        </div>

        <!-- Step 4: Batch Production -->
        <div v-show="wizardStep === 3">
          <el-alert type="success" :closable="false" style="margin-bottom: 16px;">
            批量生成设备并导出 CSV 文件，用于量产烧录。
          </el-alert>
          <el-form :model="batchForm" label-width="120px" size="default">
            <el-form-item label="设备数量">
              <el-input-number v-model="batchForm.count" :min="1" :max="1000" />
            </el-form-item>
            <el-form-item label="设备前缀">
              <el-input v-model="batchForm.prefix" placeholder="如 dev_" style="width: 200px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="batchLoading" @click="batchGenerate">批量生成</el-button>
              <el-button :disabled="!batchResult.length" @click="exportCSV">导出 CSV</el-button>
            </el-form-item>
          </el-form>
          <el-table v-if="batchResult.length" :data="batchResult" size="small" style="margin-top: 12px;" max-height="240">
            <el-table-column prop="device_name" label="DeviceName" />
            <el-table-column prop="device_sn" label="DeviceSN" />
            <el-table-column prop="product_key" label="ProductKey" />
          </el-table>
        </div>
      </div>

      <template #footer>
        <div class="wizard-footer">
          <el-button @click="wizardVisible = false">取消</el-button>
          <el-button v-if="wizardStep > 0" @click="wizardStep--">上一步</el-button>
          <el-button v-if="wizardStep < 3" type="primary" @click="wizardStep++">下一步</el-button>
          <el-button v-if="wizardStep === 3" type="primary" @click="wizardVisible = false">完成</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Edit Product Dialog -->
    <el-dialog v-model="dialogVisible" :title="editing.id ? '编辑产品' : '新建产品'" width="760px" destroy-on-close>
      <el-form :model="editing" label-width="110px" size="default">
        <el-form-item label="产品名称" required><el-input v-model="editing.name" placeholder="请输入产品名称" /></el-form-item>
        <el-form-item label="所属项目">
          <el-select v-model="editing.project_id" placeholder="选择项目（可不选）" clearable style="width:100%">
            <el-option label="未分类" :value="0" />
            <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="联网方式">
          <el-select v-model="editing.network_type" style="width:100%">
            <el-option label="WiFi" value="wifi" />
            <el-option label="蓝牙 BLE" value="ble" />
            <el-option label="蜂窝移动网络" value="cellular" />
            <el-option label="WiFi + 蓝牙 BLE" value="wifi_ble" />
          </el-select>
        </el-form-item>
        <el-form-item label="ProductKey"><el-input v-model="editing.product_key" placeholder="留空自动生成" /></el-form-item>
        <el-form-item label="产品描述"><el-input v-model="editing.description" type="textarea" :rows="2" /></el-form-item>

        <el-divider>物模型 (properties)</el-divider>

        <el-form-item label="属性列表">
          <el-table :data="editing.properties" size="small" border style="width:100%" empty-text="暂无属性，点击下方按钮添加">
            <el-table-column label="标识符" width="140">
              <template #default="{ row }"><el-input v-model="row.identifier" size="small" placeholder="如 temperature" /></template>
            </el-table-column>
            <el-table-column label="名称" width="130">
              <template #default="{ row }"><el-input v-model="row.name" size="small" placeholder="如 温度" /></template>
            </el-table-column>
            <el-table-column label="数据类型" width="120">
              <template #default="{ row }">
                <el-select v-model="row.dataType" size="small" style="width:100%">
                  <el-option v-for="t in dataTypes" :key="t" :label="t" :value="t" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="单位" width="80">
              <template #default="{ row }"><el-input v-model="row.unit" size="small" /></template>
            </el-table-column>
            <el-table-column label="下限" width="80">
              <template #default="{ row }"><el-input-number v-model="row.min" size="small" controls-position="right" style="width:100%" /></template>
            </el-table-column>
            <el-table-column label="上限" width="80">
              <template #default="{ row }"><el-input-number v-model="row.max" size="small" controls-position="right" style="width:100%" /></template>
            </el-table-column>
            <el-table-column label="读写" width="90">
              <template #default="{ row }">
                <el-select v-model="row.accessMode" size="small" style="width:100%">
                  <el-option label="读" value="r" />
                  <el-option label="写" value="w" />
                  <el-option label="读写" value="rw" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="70" align="center">
              <template #default="{ $index }">
                <el-button link type="danger" size="small" @click="editing.properties.splice($index, 1)"><el-icon><Delete /></el-icon></el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button size="small" type="primary" style="margin-top: 10px" :icon="Plus" @click="addProperty()">添加属性</el-button>
        </el-form-item>
        <el-form-item v-if="editing.id && editing.has_devices">
          <el-alert type="warning" :closable="false" title="产品下已有设备，物模型结构锁定，不可修改" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <!-- QR Dialog -->
    <el-dialog v-model="qrVisible" title="设备绑定三元组" width="420px">
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="ProductKey">{{ qrData.product_key }}</el-descriptions-item>
        <el-descriptions-item label="DeviceName">{{ qrData.device_name }}</el-descriptions-item>
        <el-descriptions-item label="DeviceSN">{{ qrData.device_sn }}</el-descriptions-item>
      </el-descriptions>
      <el-alert type="info" :closable="false" style="margin-top:12px" title="绑定 URL（可扫码或复制到设备固件）" />
      <el-input v-model="qrData.bind_url" readonly size="small" style="margin-top:8px" />
      <template #footer>
        <el-button @click="qrVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Box, Guide, Iphone } from '@element-plus/icons-vue'
import { useRouter, useRoute } from 'vue-router'
import request from '@/utils/request'

const router = useRouter()
const route = useRoute()

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const dialogVisible = ref(false)
const saving = ref(false)
const projects = ref([])
const filterProject = ref(parseInt(route.query.project_id) || null)
const qrVisible = ref(false)
const qrData = ref({})

// Wizard state
const wizardVisible = ref(false)
const wizardStep = ref(0)
const wizardProperties = ref([])
const batchForm = ref({ count: 10, prefix: 'dev_' })
const batchLoading = ref(false)
const batchResult = ref([])

const editing = ref({ id: null, name: '', product_key: '', description: '', properties: [], has_devices: false, project_id: null, network_type: 'wifi' })

const dataTypes = ['int', 'float', 'bool', 'string']

const netTypeLabel = (t) => ({ wifi: 'WiFi', ble: '蓝牙', cellular: '蜂窝', wifi_ble: 'WiFi+BLE' }[t] || t || 'WiFi')

const propertyCount = (row) => {
  if (Array.isArray(row.properties)) return row.properties.length
  return 0
}

const loadProjects = async () => {
  try {
    const res = await request.get('/admin/project', { params: { page: 1, size: 100 } })
    projects.value = res?.list || []
  } catch (e) {
    console.error(e)
  }
}

const load = async () => {
  try {
    const params = { page: page.value, size: pageSize.value }
    if (filterProject.value) params.project_id = filterProject.value
    const data = await request.get('/admin/product', { params })
    list.value = Array.isArray(data?.list) ? data.list : Array.isArray(data?.items) ? data.items : Array.isArray(data) ? data : []
    total.value = data?.total ?? list.value.length
  } catch (e) {
    console.error(e)
    list.value = []
    total.value = 0
  }
}

const propertiesToJson = (props) => {
  return props.map(p => ({
    identifier: p.identifier,
    name: p.name,
    accessMode: p.accessMode || 'rw',
    dataType: {
      type: p.dataType || 'int',
      specs: {
        min: p.min,
        max: p.max,
        unit: p.unit || ''
      }
    }
  }))
}

const propertiesFromJson = (arr) => {
  if (!Array.isArray(arr)) return []
  return arr.map(p => ({
    identifier: p.identifier,
    name: p.name,
    dataType: p.dataType?.type || p.dataType || 'int',
    unit: p.dataType?.specs?.unit || '',
    min: p.dataType?.specs?.min,
    max: p.dataType?.specs?.max,
    accessMode: p.accessMode || 'rw'
  }))
}

const openCreate = () => {
  editing.value = { id: null, name: '', product_key: '', description: '', properties: [], has_devices: false, project_id: filterProject.value || null, network_type: 'wifi' }
  dialogVisible.value = true
}

const openEdit = async (row) => {
  try {
    const detail = await request.get('/admin/product/' + row.id)
    editing.value = {
      id: detail.id,
      name: detail.name,
      product_key: detail.product_key || '',
      description: detail.description || '',
      project_id: detail.project_id || null,
      network_type: detail.network_type || 'wifi',
      properties: propertiesFromJson(detail.properties || []),
      has_devices: !!detail.has_devices
    }
    dialogVisible.value = true
  } catch (e) {
    ElMessage.error('加载产品详情失败')
  }
}

const showQR = async (row) => {
  try {
    const res = await request.get('/admin/device', { params: { product_id: row.id, page: 1, size: 1 } })
    const dev = res?.list?.[0]
    if (!dev) { ElMessage.warning('该产品下暂无设备'); return }
    const qr = await request.get('/admin/device/' + dev.id + '/bind-qr')
    qrData.value = qr.data || qr
    qrVisible.value = true
  } catch (e) {
    ElMessage.error('获取绑定信息失败')
  }
}

const addProperty = () => {
  editing.value.properties.push({
    identifier: 'prop_' + (editing.value.properties.length + 1),
    name: '属性',
    dataType: 'int',
    unit: '',
    min: 0,
    max: 100,
    accessMode: 'rw'
  })
}

const addWizardProperty = () => {
  wizardProperties.value.push({
    identifier: 'prop_' + (wizardProperties.value.length + 1),
    name: '属性',
    dataType: 'int',
    unit: '',
    min: 0,
    max: 100
  })
}

const batchGenerate = async () => {
  batchLoading.value = true
  try {
    const res = await request.post('/admin/device/batch', {
      count: batchForm.value.count,
      prefix: batchForm.value.prefix,
      product_id: editing.value.id
    })
    batchResult.value = res?.devices || []
    ElMessage.success('生成 ' + batchResult.value.length + ' 台设备')
  } catch (e) {
    ElMessage.error('批量生成失败')
  } finally {
    batchLoading.value = false
  }
}

const exportCSV = () => {
  if (!batchResult.value.length) return
  const header = 'DeviceName,DeviceSN,ProductKey\n'
  const rows = batchResult.value.map(d => `${d.device_name},${d.device_sn},${d.product_key}`).join('\n')
  const blob = new Blob([header + rows], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = 'devices_' + Date.now() + '.csv'
  link.click()
}

const submit = async () => {
  if (!editing.value.name) { ElMessage.warning('请输入产品名称'); return }
  saving.value = true
  try {
    const propsJson = JSON.stringify(propertiesToJson(editing.value.properties))
    if (editing.value.id) {
      await request.put('/admin/product/' + editing.value.id, {
        name: editing.value.name,
        description: editing.value.description,
        icon: '',
        project_id: editing.value.project_id || 0,
        network_type: editing.value.network_type || 'wifi'
      })
      if (!editing.value.has_devices) {
        await request.put('/admin/product/' + editing.value.id + '/thing-model', {
          properties_json: propsJson,
          events_json: '[]',
          services_json: '[]'
        })
      }
      ElMessage.success('更新成功')
    } else {
      await request.post('/admin/product', {
        name: editing.value.name,
        description: editing.value.description,
        icon: '',
        product_key: editing.value.product_key,
        properties_json: propsJson,
        events_json: '[]',
        services_json: '[]',
        project_id: editing.value.project_id || 0,
        network_type: editing.value.network_type || 'wifi'
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    load()
  } catch (e) {
    ElMessage.error('保存失败: ' + (e?.response?.data?.message || e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

const openMobileUI = (row) => {
  router.push({ path: '/mobile-ui', query: { productId: row.id } })
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除 "' + row.name + '"？', '提示', { type: 'warning' })
    await request.delete('/admin/product/' + row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    if (e !== 'cancel') console.error(e)
  }
}

onMounted(() => { load(); loadProjects() })
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.product-name { display: flex; align-items: center; gap: 6px; font-weight: 600; }
.product-icon { color: #007DFF; }

.m-item { margin-bottom: 10px; }
.m-name { font-weight: 600; color: #1D2129; margin-bottom: 4px; }
.m-sub { color: #86909C; font-size: 13px; margin-top: 2px; }
.m-actions { margin-top: 8px; }

.desktop-only { display: table; }
.mobile-only { display: none; }

.wizard-steps { margin-bottom: 24px; padding: 0 20px; }

.wizard-content {
  min-height: 300px;
  padding: 20px 0;
}

.wizard-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
}

@media (max-width: 768px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
}
</style>
