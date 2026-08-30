<template>
  <div class="dashboard">
    <!-- Stat Cards -->
    <div class="card-grid">
      <div class="stat-card">
        <div class="stat-icon blue"><el-icon :size="24"><Cpu /></el-icon></div>
        <div class="stat-body">
          <div class="stat-label">设备总数</div>
          <div class="stat-value">{{ stats.devices }}</div>
          <div class="stat-trend up">
            <el-icon :size="12"><Top /></el-icon>较昨日 +2.1%
          </div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon green"><el-icon :size="24"><CircleCheck /></el-icon></div>
        <div class="stat-body">
          <div class="stat-label">在线设备</div>
          <div class="stat-value">{{ stats.online }}</div>
          <div class="stat-trend up">
            <el-icon :size="12"><Top /></el-icon>在线率 {{ onlineRate }}%
          </div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon orange"><el-icon :size="24"><Box /></el-icon></div>
        <div class="stat-body">
          <div class="stat-label">产品数量</div>
          <div class="stat-value">{{ stats.products }}</div>
          <div class="stat-trend up">
            <el-icon :size="12"><Top /></el-icon>本月新增 +1
          </div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon red"><el-icon :size="24"><Warning /></el-icon></div>
        <div class="stat-body">
          <div class="stat-label">今日告警</div>
          <div class="stat-value">{{ stats.alerts || 0 }}</div>
          <div class="stat-trend down">
            <el-icon :size="12"><Bottom /></el-icon>较昨日 -5.3%
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :xs="24" :md="16">
        <div class="alibaba-card">
          <div class="card-header">
            <span class="card-title">24小时设备在线趋势</span>
            <el-radio-group v-model="chartRange" size="small">
              <el-radio-button label="24h">24小时</el-radio-button>
              <el-radio-button label="7d">7天</el-radio-button>
            </el-radio-group>
          </div>
          <div ref="trendChartRef" style="height: 320px;"></div>
        </div>
      </el-col>
      <el-col :xs="24" :md="8">
        <div class="alibaba-card">
          <div class="card-header">
            <span class="card-title">设备类型分布</span>
          </div>
          <div ref="pieChartRef" style="height: 320px;"></div>
        </div>
      </el-col>
    </el-row>

    <!-- Bottom Row -->
    <el-row :gutter="16">
      <el-col :xs="24" :md="14">
        <div class="alibaba-card">
          <div class="card-header">
            <span class="card-title">最近操作日志</span>
            <el-button link type="primary" size="small">查看全部</el-button>
          </div>
          <el-table :data="logs" size="small" style="width: 100%;">
            <el-table-column prop="time" label="时间" width="160" />
            <el-table-column prop="user" label="操作人" width="100" />
            <el-table-column prop="action" label="操作" min-width="120" />
            <el-table-column prop="target" label="目标" min-width="140" show-overflow-tooltip />
            <el-table-column label="结果" width="80">
              <template #default="{ row }">
                <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                  {{ row.success ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-col>
      <el-col :xs="24" :md="10">
        <div class="alibaba-card">
          <div class="card-header">
            <span class="card-title">快捷入口</span>
          </div>
          <div class="quick-entry">
            <div class="quick-entry-item" @click="$router.push('/product')">
              <div class="entry-icon"><el-icon><Box /></el-icon></div>
              <span class="entry-label">新建产品</span>
            </div>
            <div class="quick-entry-item" @click="$router.push('/device')">
              <div class="entry-icon"><el-icon><Cpu /></el-icon></div>
              <span class="entry-label">添加设备</span>
            </div>
            <div class="quick-entry-item" @click="$router.push('/ota')">
              <div class="entry-icon"><el-icon><Upload /></el-icon></div>
              <span class="entry-label">OTA升级</span>
            </div>
            <div class="quick-entry-item" @click="$router.push('/project')">
              <div class="entry-icon"><el-icon><FolderOpened /></el-icon></div>
              <span class="entry-label">项目管理</span>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import {
  Cpu, CircleCheck, Box, Warning, Top, Bottom, Upload, FolderOpened
} from '@element-plus/icons-vue'
import request from '@/utils/request'

echarts.use([LineChart, PieChart, GridComponent, TooltipComponent, TitleComponent, LegendComponent, CanvasRenderer])

const trendChartRef = ref(null)
const pieChartRef = ref(null)
let trendChart = null
let pieChart = null
const chartRange = ref('24h')

const stats = ref({ devices: 0, online: 0, products: 0, alerts: 0 })
const logs = ref([])

const onlineRate = computed(() => {
  if (!stats.value.devices) return 0
  return ((stats.value.online / stats.value.devices) * 100).toFixed(1)
})

const loadData = async () => {
  try {
    const data = await request.get('/admin/stats')
    if (data) {
      stats.value.devices = data.devices || 0
      stats.value.online = data.online || 0
      stats.value.products = data.products || 0
      stats.value.alerts = data.alerts || 0
    }
  } catch (e) {
    console.error(e)
  }

  // Mock log data
  logs.value = [
    { time: '2026-08-30 14:32:15', user: 'admin', action: '创建产品', target: '智能温控器', success: true },
    { time: '2026-08-30 13:18:42', user: 'admin', action: 'OTA推送', target: '固件 v2.1.0 → 128台设备', success: true },
    { time: '2026-08-30 11:05:33', user: 'user01', action: '添加设备', target: 'SN:20260830001', success: true },
    { time: '2026-08-30 09:22:10', user: 'admin', action: '删除项目', target: '旧测试项目', success: true },
    { time: '2026-08-30 08:45:00', user: 'system', action: '设备离线', target: 'dev_001 连接超时', success: false }
  ]
}

const renderTrendChart = () => {
  if (!trendChartRef.value) return
  trendChart = echarts.init(trendChartRef.value)

  const hours = []
  const onlineData = []
  const offlineData = []

  for (let i = 23; i >= 0; i--) {
    const h = new Date()
    h.setHours(h.getHours() - i)
    hours.push(h.getHours().toString().padStart(2, '0') + ':00')
    onlineData.push(Math.floor(Math.random() * 30) + 70)
    offlineData.push(Math.floor(Math.random() * 10) + 5)
  }

  trendChart.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(255,255,255,0.95)',
      borderColor: '#E5E6EB',
      textStyle: { color: '#1D2129', fontSize: 13 }
    },
    legend: {
      data: ['在线设备', '离线设备'],
      right: 0,
      top: 0,
      textStyle: { color: '#86909C', fontSize: 12 }
    },
    grid: { left: 50, right: 20, top: 40, bottom: 40 },
    xAxis: {
      type: 'category',
      data: hours,
      axisLabel: { color: '#86909C', fontSize: 12 },
      axisLine: { lineStyle: { color: '#E5E6EB' } },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#86909C', fontSize: 12 },
      axisLine: { show: false },
      splitLine: { lineStyle: { color: '#F2F3F5', type: 'dashed' } }
    },
    series: [
      {
        name: '在线设备',
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: onlineData,
        lineStyle: { color: '#007DFF', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(0,125,255,0.3)' },
            { offset: 1, color: 'rgba(0,125,255,0.02)' }
          ])
        }
      },
      {
        name: '离线设备',
        type: 'line',
        smooth: true,
        symbol: 'none',
        data: offlineData,
        lineStyle: { color: '#C9CDD4', width: 2, type: 'dashed' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(201,205,212,0.15)' },
            { offset: 1, color: 'rgba(201,205,212,0.02)' }
          ])
        }
      }
    ]
  })
}

const renderPieChart = () => {
  if (!pieChartRef.value) return
  pieChart = echarts.init(pieChartRef.value)

  pieChart.setOption({
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(255,255,255,0.95)',
      borderColor: '#E5E6EB',
      textStyle: { color: '#1D2129', fontSize: 13 },
      formatter: '{b}: {c}台 ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center',
      textStyle: { color: '#86909C', fontSize: 12 }
    },
    color: ['#007DFF', '#00B42A', '#FF7D00', '#F53F3F', '#722ED1', '#14C9C9'],
    series: [
      {
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: false,
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold' }
        },
        data: [
          { value: 45, name: 'WiFi设备' },
          { value: 28, name: '蓝牙设备' },
          { value: 15, name: '蜂窝设备' },
          { value: 8, name: '网关设备' },
          { value: 4, name: '其他' }
        ]
      }
    ]
  })
}

const handleResize = () => {
  trendChart && trendChart.resize()
  pieChart && pieChart.resize()
}

onMounted(async () => {
  await loadData()
  await nextTick()
  renderTrendChart()
  renderPieChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  trendChart && trendChart.dispose()
  pieChart && pieChart.dispose()
})
</script>

<style scoped>
.card-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

@media (max-width: 1200px) {
  .card-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .card-grid { grid-template-columns: 1fr; }
}

.stat-card {
  background: #fff;
  border-radius: var(--radius, 8px);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  border: 1px solid #F2F3F5;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
  transform: translateY(-2px);
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.stat-icon.blue { background: linear-gradient(135deg, #007DFF, #40A9FF); }
.stat-icon.green { background: linear-gradient(135deg, #00B42A, #52C41A); }
.stat-icon.orange { background: linear-gradient(135deg, #FF7D00, #FAAD14); }
.stat-icon.red { background: linear-gradient(135deg, #F53F3F, #FF7875); }

.stat-body { flex: 1; }

.stat-label {
  font-size: 13px;
  color: #86909C;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1D2129;
  line-height: 1.2;
}

.stat-trend {
  font-size: 12px;
  margin-top: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.stat-trend.up { color: #00B42A; }
.stat-trend.down { color: #F53F3F; }

.alibaba-card {
  background: #fff;
  border-radius: var(--radius, 8px);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  border: 1px solid #F2F3F5;
  padding: 20px;
}

.alibaba-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.alibaba-card .card-title {
  font-size: 16px;
  font-weight: 600;
  color: #1D2129;
}

.quick-entry {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.quick-entry-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 12px;
  border-radius: 8px;
  border: 1px solid #F2F3F5;
  cursor: pointer;
  transition: all 0.3s ease;
  text-decoration: none;
  color: #1D2129;
}

.quick-entry-item:hover {
  border-color: #007DFF;
  background: #E6F7FF;
}

.entry-icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #E6F7FF;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 10px;
  color: #007DFF;
  font-size: 20px;
}

.entry-label {
  font-size: 13px;
  color: #1D2129;
  font-weight: 500;
}
</style>
