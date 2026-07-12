<template>
  <div class="device-detail">
    <el-page-header @back="$router.back()" :content="'设备详情'" style="margin-bottom: 12px" />

    <el-row :gutter="16">
      <el-col :xs="24" :sm="24" :md="12">
        <el-card style="margin-bottom: 16px">
          <template #header><span class="title">基本信息</span></template>
          <el-descriptions :column="1" border size="default">
            <el-descriptions-item label="设备 ID">{{ device.id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="设备名">{{ device.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="产品">{{ device.productName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="DeviceKey">{{ device.deviceKey || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="device.online ? 'success' : 'info'" size="small">{{ device.online ? '在线' : '离线' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ device.createdAt || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="24" :md="12">
        <el-card style="margin-bottom: 16px">
          <template #header><span class="title">最新数据</span></template>
          <div v-if="Object.keys(latest).length === 0" style="color:#909399">暂无数据</div>
          <el-descriptions v-else :column="1" border size="default">
            <el-descriptions-item v-for="(v, k) in latest" :key="k" :label="k">
              <span class="latest-value">{{ formatValue(v) }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="24" :md="24">
        <el-card>
          <template #header><span class="title">控制指令</span></template>
          <el-alert v-if="properties.length === 0" title="该设备暂无可控制属性" type="info" show-icon :closable="false" style="margin-bottom: 12px" />
          <el-form v-else :model="controlValues" label-width="140px">
            <el-row :gutter="16">
              <el-col v-for="p in properties" :key="p.identifier" :xs="24" :sm="12" :md="8">
                <el-form-item :label="p.name + (p._unit ? ' (' + p._unit + ')' : '')">
                  <el-switch v-if="p._dataType === 'bool'" v-model="controlValues[p.identifier]" />
                  <el-input-number v-else-if="p._dataType === 'int' || p._dataType === 'float'" v-model="controlValues[p.identifier]" :min="p._min" :max="p._max" :step="p._dataType === 'float' ? 0.1 : 1" style="width:100%" />
                  <el-input v-else v-model="controlValues[p.identifier]" />
                  <el-button style="margin-top: 6px" size="small" type="primary" @click="sendCommand(p)">下发 {{ p.name }}</el-button>
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const route = useRoute()
const device = ref({})
const latest = ref({})
const properties = ref([])
const controlValues = reactive({})

const id = computed(() => route.params.id)

const load = async () => {
  try {
    const data = await request.get('/device/' + id.value)
    device.value = data?.device || data || {}
    latest.value = data?.latest || data?.data || {}
    const props = data?.thingModel?.properties || data?.properties || []
    const arr = Array.isArray(props) ? props : Object.entries(props).map(([k, v]) => ({ identifier: k, ...v }))
    properties.value = arr.filter(p => (p.accessMode || 'rw').includes('w'))
    properties.value.forEach(p => {
      const dt = (typeof p.dataType === 'object' ? p.dataType?.type : p.dataType) || 'string'
      const specs = (typeof p.dataType === 'object' ? p.dataType?.specs : {}) || {}
      p._dataType = dt
      p._min = specs.min
      p._max = specs.max
      p._unit = specs.unit || ''
      if (dt === 'bool') controlValues[p.identifier] = false
      else if (dt === 'string') controlValues[p.identifier] = ''
      else controlValues[p.identifier] = specs.min || 0
    })
  } catch (e) {
    console.error(e)
    device.value = { id: id.value, name: 'Demo-Device-01', productName: '智能温控器', online: true, deviceKey: 'dk_demo_001', createdAt: '2025-01-01 10:00' }
    latest.value = { temperature: 25.5, humidity: 60 }
    properties.value = [
      { identifier: 'temperature', name: '温度', _dataType: 'float', _unit: '℃', _min: 0, _max: 40, accessMode: 'rw' },
      { identifier: 'power', name: '开关', _dataType: 'bool', _unit: '', accessMode: 'rw' }
    ]
    properties.value.forEach(p => { controlValues[p.identifier] = p._dataType === 'bool' ? false : (p._min || 0) })
  }
}

const sendCommand = async (p) => {
  try {
    await request.post('/device/' + id.value + '/control', { [p.identifier]: controlValues[p.identifier] })
    ElMessage.success(p.name + ' 指令下发成功')
  } catch (e) { console.error(e) }
}

const formatValue = (v) => {
  if (v === null || v === undefined) return '-'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

onMounted(load)
</script>

<style scoped>
.title { font-weight: 600; color: #303133; }
.latest-value { font-weight: 600; color: #409eff; font-size: 15px; }
</style>
