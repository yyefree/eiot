<template>
  <div class="dashboard">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon blue"><el-icon><Cpu /></el-icon></div>
          <div class="stat-info">
            <div class="stat-label">设备总数</div>
            <div class="stat-value">{{ stats.devices }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon green"><el-icon><CircleCheck /></el-icon></div>
          <div class="stat-info">
            <div class="stat-label">在线设备</div>
            <div class="stat-value">{{ stats.online }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon orange"><el-icon><Coin /></el-icon></div>
          <div class="stat-info">
            <div class="stat-label">物模型数量</div>
            <div class="stat-value">{{ stats.models }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon purple"><el-icon><Box /></el-icon></div>
          <div class="stat-info">
            <div class="stat-label">产品数量</div>
            <div class="stat-value">{{ stats.products }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="content-row">
      <el-col :xs="24" :md="16">
        <el-card>
          <template #header>
            <div class="card-title">最近 24 小时数据上报量</div>
          </template>
          <div ref="chartRef" style="height: 340px"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="devices-card">
          <template #header>
            <div class="card-title">设备列表</div>
          </template>
          <el-table :data="devices" size="small" style="width:100%" class="desktop-only">
            <el-table-column prop="name" label="设备名" min-width="120" />
            <el-table-column prop="productName" label="产品" min-width="120" show-overflow-tooltip />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.online ? 'success' : 'info'" size="small">{{ row.online ? '在线' : '离线' }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div class="mobile-list mobile-only">
            <div v-for="d in devices" :key="d.id" class="mobile-item">
              <div class="m-name">{{ d.name }}</div>
              <div class="m-sub">
                <span>{{ d.productName || '-' }}</span>
                <el-tag :type="d.online ? 'success' : 'info'" size="small">{{ d.online ? '在线' : '离线' }}</el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import request from '@/utils/request'

echarts.use([LineChart, GridComponent, TooltipComponent, TitleComponent, CanvasRenderer])

const router = useRouter()
const chartRef = ref(null)
let chart = null

const stats = ref({ devices: 0, online: 0, models: 0, products: 0 })
const devices = ref([])

const genMockData = () => {
  const hours = []
  const values = []
  for (let i = 23; i >= 0; i--) {
    hours.push((24 - i - 1) + ':00')
    values.push(Math.floor(Math.random() * 500 + 100))
  }
  return { hours, values }
}

const renderChart = () => {
  if (!chartRef.value) return
  chart = echarts.init(chartRef.value)
  const { hours, values } = genMockData()
  chart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 30, bottom: 40 },
    xAxis: { type: 'category', data: hours, axisLabel: { color: '#606266' } },
    yAxis: { type: 'value', axisLabel: { color: '#606266' } },
    series: [{
      data: values,
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      areaStyle: { color: 'rgba(64, 158, 255, 0.2)' },
      lineStyle: { color: '#409eff', width: 2 },
      itemStyle: { color: '#409eff' }
    }]
  })
}

const loadData = async () => {
  try {
    const data = await request.get('/dashboard')
    if (data) {
      stats.value.devices = data.devices || 0
      stats.value.online = data.online || 0
      stats.value.models = data.thingModels || 0
      stats.value.products = data.products || 0
      devices.value = data.devices_list || data.devices_list || []
    }
  } catch (e) {
    console.error(e)
    devices.value = [
      { id: 1, name: 'Demo-Device-01', productName: '智能温控器', online: true },
      { id: 2, name: 'Demo-Device-02', productName: '智能门锁', online: false },
      { id: 3, name: 'Demo-Device-03', productName: '环境监测', online: true },
      { id: 4, name: 'Demo-Device-04', productName: '智能插座', online: true }
    ]
    stats.value = { devices: 48, online: 32, models: 12, products: 8 }
  }
}

const handleResize = () => chart && chart.resize()

onMounted(async () => {
  await loadData()
  await nextTick()
  renderChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chart && chart.dispose()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { margin-bottom: 16px; display: flex; align-items: center; padding: 6px; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 14px; padding: 18px; }
.stat-icon {
  width: 48px; height: 48px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 22px;
}
.stat-icon.blue { background: linear-gradient(135deg, #409eff, #1890ff); }
.stat-icon.green { background: linear-gradient(135deg, #52c41a, #237804); }
.stat-icon.orange { background: linear-gradient(135deg, #faad14, #d48806); }
.stat-icon.purple { background: linear-gradient(135deg, #9254de, #531dab); }
.stat-label { color: #909399; font-size: 13px; }
.stat-value { font-size: 26px; font-weight: 600; color: #303133; }

.card-title { font-weight: 600; color: #303133; }
.content-row { margin-top: 0; }
.devices-card { height: 100%; }

.mobile-list { margin-top: 4px; }
.mobile-item {
  padding: 10px 0;
  border-bottom: 1px solid #f0f2f5;
}
.mobile-item:last-child { border-bottom: none; }
.m-name { font-size: 14px; font-weight: 500; color: #303133; }
.m-sub { display: flex; justify-content: space-between; align-items: center; color: #909399; font-size: 12px; margin-top: 4px; }

.desktop-only { display: table; }
.mobile-only { display: none; }

@media (max-width: 768px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
  .stat-value { font-size: 20px; }
}
</style>
