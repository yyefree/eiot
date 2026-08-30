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

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :xs="24" :sm="24" :md="12">
        <el-card style="margin-bottom: 16px">
          <template #header><span class="title">设备影子</span></template>
          <div v-if="!shadow" style="color:#909399">暂无影子数据</div>
          <template v-else>
            <div style="margin-bottom:8px;font-size:12px;color:#909399">版本: {{ shadow.version }}</div>
            <el-row :gutter="12">
              <el-col :span="12">
                <div class="shadow-label">期望值 (Desired)</div>
                <pre class="shadow-json">{{ prettyJson(shadow.desired_json) }}</pre>
              </el-col>
              <el-col :span="12">
                <div class="shadow-label">上报值 (Reported)</div>
                <pre class="shadow-json">{{ prettyJson(shadow.reported_json) }}</pre>
              </el-col>
            </el-row>
            <el-button size="small" type="primary" style="margin-top:10px" @click="showShadowDialog = true">设置期望值</el-button>
          </template>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="24" :md="12">
        <el-card style="margin-bottom: 16px">
          <template #header><span class="title">历史数据</span></template>
          <el-select v-model="chartProperty" placeholder="选择属性" size="small" style="width:200px;margin-bottom:12px" @change="loadChartData">
            <el-option v-for="p in allProperties" :key="p.identifier" :label="p.name" :value="p.identifier" />
          </el-select>
          <div ref="chartRef" style="width:100%;height:280px"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top:0">
      <template #header><span class="title">事件与服务</span></template>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="事件" name="events">
          <div style="margin-bottom:12px">
            <el-button type="primary" size="small" @click="showEventDialog = true">上报事件</el-button>
          </div>
          <el-table :data="eventList" border size="small" v-loading="eventLoading">
            <el-table-column prop="event_id" label="事件ID" width="140" />
            <el-table-column prop="event_name" label="事件名称" width="140" />
            <el-table-column label="输出参数">
              <template #default="{ row }">
                <span class="mono">{{ formatValue(row.output_json) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" width="180" />
          </el-table>
          <el-pagination
            style="margin-top:12px;justify-content:flex-end"
            small
            layout="total, prev, pager, next"
            :total="eventTotal"
            :page-size="eventPageSize"
            v-model:current-page="eventPage"
            @current-change="loadEvents"
          />
        </el-tab-pane>

        <el-tab-pane label="服务" name="services">
          <div style="margin-bottom:12px">
            <el-button type="primary" size="small" @click="showServiceDialog = true">调用服务</el-button>
          </div>
          <el-table :data="serviceList" border size="small" v-loading="serviceLoading">
            <el-table-column prop="service_id" label="服务ID" width="140" />
            <el-table-column prop="service_name" label="服务名称" width="140" />
            <el-table-column label="输入参数">
              <template #default="{ row }">
                <span class="mono">{{ formatValue(row.input_json) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="输出参数">
              <template #default="{ row }">
                <span class="mono">{{ formatValue(row.output_json) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">{{ row.status === 'success' ? '成功' : '失败' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" width="180" />
          </el-table>
          <el-pagination
            style="margin-top:12px;justify-content:flex-end"
            small
            layout="total, prev, pager, next"
            :total="serviceTotal"
            :page-size="servicePageSize"
            v-model:current-page="servicePage"
            @current-change="loadServices"
          />
        </el-tab-pane>

        <el-tab-pane label="调试" name="debug">
          <!-- 调试标签页内容 -->
          <el-tabs v-model="debugActiveTab" style="margin-top: 12px">
            <!-- 1. 模拟上报属性 -->
            <el-tab-pane label="模拟上报属性" name="debug-property">
              <el-alert v-if="writableProperties.length === 0" title="该设备暂无可写属性" type="info" show-icon :closable="false" style="margin-bottom: 12px" />
              <el-form v-else :model="propertyReportForm" label-width="140px" size="default">
                <el-row :gutter="16">
                  <el-col v-for="p in writableProperties" :key="p.identifier" :xs="24" :sm="12" :md="8">
                    <el-form-item :label="p.name + (p._unit ? ' (' + p._unit + ')' : '')">
                      <el-switch v-if="p._dataType === 'bool'" v-model="propertyReportForm[p.identifier]" />
                      <el-input-number v-else-if="p._dataType === 'int' || p._dataType === 'float'" v-model="propertyReportForm[p.identifier]" :min="p._min" :max="p._max" :step="p._dataType === 'float' ? 0.1 : 1" :precision="p._dataType === 'float' ? 2 : 0" style="width:100%" />
                      <el-select v-else-if="p._dataType === 'enum' && p._specs?.enumList" v-model="propertyReportForm[p.identifier]" placeholder="请选择" style="width:100%" filterable>
                        <el-option v-for="opt in p._specs.enumList" :key="opt.value" :label="opt.label || opt.value" :value="opt.value" />
                      </el-select>
                      <el-input v-else v-model="propertyReportForm[p.identifier]" placeholder="请输入" />
                    </el-form-item>
                  </el-col>
                </el-row>
                <div style="margin-top: 16px; display: flex; gap: 12px; align-items: center;">
                  <el-button type="primary" @click="reportProperties" :loading="propertyReportLoading" :disabled="Object.keys(propertyReportForm).length === 0">
                    一键上报
                  </el-button>
                  <el-button @click="resetPropertyForm">重置</el-button>
                  <el-divider direction="vertical" />
                  <span v-if="lastPropertyReport" class="report-info">
                    最近上报: {{ formatTime(lastPropertyReport.time) }} · 
                    <el-tag :type="lastPropertyReport.success ? 'success' : 'danger'" size="small">
                      {{ lastPropertyReport.success ? '成功' : '失败' }}
                    </el-tag>
                    <el-tooltip :content="lastPropertyReport.response" placement="top">
                      <el-icon style="margin-left: 4px; cursor: pointer;"><Search /></el-icon>
                    </el-tooltip>
                  </span>
                </div>
              </el-form>
            </el-tab-pane>

            <!-- 2. 模拟上报事件 -->
            <el-tab-pane label="模拟上报事件" name="debug-event">
              <el-alert v-if="allEvents.length === 0" title="该设备暂无事件定义" type="info" show-icon :closable="false" style="margin-bottom: 12px" />
              <el-form v-else :model="debugEventForm" label-width="100px" size="default">
                <el-form-item label="事件">
                  <el-select v-model="debugEventForm.event_id" placeholder="选择事件" style="width: 100%" @change="onDebugEventSelect">
                    <el-option v-for="e in allEvents" :key="e.identifier" :label="e.name" :value="e.identifier" />
                  </el-select>
                </el-form-item>
                <template v-if="debugEventForm.params.length > 0">
                  <el-form-item v-for="(param, idx) in debugEventForm.params" :key="param.identifier" :label="param.name">
                    <el-switch v-if="param._dataType === 'bool'" v-model="debugEventForm.values[param.identifier]" />
                    <el-input-number v-else-if="param._dataType === 'int' || param._dataType === 'float'" v-model="debugEventForm.values[param.identifier]" style="width:100%" :precision="param._dataType === 'float' ? 2 : 0" />
                    <el-select v-else-if="param._dataType === 'enum' && param._specs?.enumList" v-model="debugEventForm.values[param.identifier]" placeholder="请选择" style="width:100%" filterable>
                      <el-option v-for="opt in param._specs.enumList" :key="opt.value" :label="opt.label || opt.value" :value="opt.value" />
                    </el-select>
                    <el-input v-else v-model="debugEventForm.values[param.identifier]" placeholder="请输入" />
                  </el-form-item>
                </template>
                <template v-else>
                  <el-form-item label="输出参数">
                    <el-alert title="该事件无输出参数" type="info" show-icon :closable="false" />
                  </el-form-item>
                </template>
                <div style="margin-top: 16px; display: flex; gap: 12px; align-items: center;">
                  <el-button type="primary" @click="reportDebugEvent" :loading="debugEventLoading" :disabled="!debugEventForm.event_id">
                    上报事件
                  </el-button>
                  <el-button @click="resetDebugEventForm">重置</el-button>
                  <el-divider direction="vertical" />
                  <span v-if="lastDebugEvent" class="report-info">
                    最近上报: {{ formatTime(lastDebugEvent.time) }} · 
                    <el-tag :type="lastDebugEvent.success ? 'success' : 'danger'" size="small">
                      {{ lastDebugEvent.success ? '成功' : '失败' }}
                    </el-tag>
                  </span>
                </div>
              </el-form>
            </el-tab-pane>

            <!-- 3. 调用服务 -->
            <el-tab-pane label="调用服务" name="debug-service">
              <el-alert v-if="allServices.length === 0" title="该设备暂无服务定义" type="info" show-icon :closable="false" style="margin-bottom: 12px" />
              <div v-else>
                <el-form :model="debugServiceForm" label-width="100px" size="default" style="margin-bottom: 16px;">
                  <el-form-item label="服务">
                    <el-select v-model="debugServiceForm.service_id" placeholder="选择服务" style="width: 100%" @change="onDebugServiceSelect">
                      <el-option v-for="s in allServices" :key="s.identifier" :label="s.name" :value="s.identifier" />
                    </el-select>
                  </el-form-item>
                  <template v-if="debugServiceForm.params.length > 0">
                    <el-row :gutter="16">
                      <el-col v-for="(param, idx) in debugServiceForm.params" :key="param.identifier" :xs="24" :sm="12">
                        <el-form-item :label="param.name">
                          <el-switch v-if="param._dataType === 'bool'" v-model="debugServiceForm.values[param.identifier]" />
                          <el-input-number v-else-if="param._dataType === 'int' || param._dataType === 'float'" v-model="debugServiceForm.values[param.identifier]" style="width:100%" :precision="param._dataType === 'float' ? 2 : 0" />
                          <el-select v-else-if="param._dataType === 'enum' && param._specs?.enumList" v-model="debugServiceForm.values[param.identifier]" placeholder="请选择" style="width:100%" filterable>
                            <el-option v-for="opt in param._specs.enumList" :key="opt.value" :label="opt.label || opt.value" :value="opt.value" />
                          </el-select>
                          <el-input v-else v-model="debugServiceForm.values[param.identifier]" placeholder="请输入" />
                        </el-form-item>
                      </el-col>
                    </el-row>
                  </template>
                  <template v-else>
                    <el-form-item label="输入参数">
                      <el-alert title="该服务无输入参数" type="info" show-icon :closable="false" />
                    </el-form-item>
                  </template>
                  <div style="display: flex; gap: 12px; align-items: center;">
                    <el-button type="primary" @click="invokeDebugService" :loading="debugServiceLoading" :disabled="!debugServiceForm.service_id">
                      调用服务
                    </el-button>
                    <el-button @click="resetDebugServiceForm">重置</el-button>
                  </div>
                </el-form>

                <!-- 服务调用历史 -->
                <el-divider style="margin: 16px 0">服务调用历史</el-divider>
                <el-table :data="debugServiceHistory" border size="small" v-loading="debugServiceHistoryLoading">
                  <el-table-column prop="service_id" label="服务ID" width="140" />
                  <el-table-column prop="service_name" label="服务名称" width="140" />
                  <el-table-column label="输入参数">
                    <template #default="{ row }">
                      <span class="mono">{{ formatValue(row.input_json) }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="输出参数">
                    <template #default="{ row }">
                      <span class="mono">{{ formatValue(row.output_json) }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="状态" width="90">
                    <template #default="{ row }">
                      <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                        {{ row.status === 'success' ? '成功' : '失败' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="created_at" label="时间" width="180" />
                </el-table>
                <el-pagination
                  style="margin-top:12px;justify-content:flex-end"
                  small
                  layout="total, prev, pager, next"
                  :total="debugServiceHistoryTotal"
                  :page-size="debugServiceHistoryPageSize"
                  v-model:current-page="debugServiceHistoryPage"
                  @current-change="loadDebugServiceHistory"
                />
              </div>
            </el-tab-pane>

            <!-- 4. MQTT 原始报文 -->
            <el-tab-pane label="MQTT 原始报文" name="debug-mqtt">
              <div style="margin-bottom: 12px; display: flex; gap: 12px; align-items: center; flex-wrap: wrap;">
                <el-switch v-model="mqttAutoRefresh" active-text="自动刷新" inactive-text="手动刷新" @change="toggleMqttAutoRefresh" />
                <el-button v-if="!mqttAutoRefresh" @click="loadMqttMessages" :loading="mqttLoading">刷新</el-button>
                <el-select v-model="mqttDirectionFilter" placeholder="方向筛选" size="small" style="width: 140px" @change="loadMqttMessages">
                  <el-option label="全部" value="" />
                  <el-option label="上行 (设备→云)" value="up" />
                  <el-option label="下行 (云→设备)" value="down" />
                </el-select>
                <el-select v-model="mqttTopicFilter" placeholder="主题筛选" size="small" style="width: 200px" @change="loadMqttMessages">
                  <el-option label="全部" value="" />
                  <el-option v-for="t in mqttTopics" :key="t" :label="t" :value="t" />
                </el-select>
              </div>
              <el-table :data="mqttMessages" border size="small" v-loading="mqttLoading" style="width: 100%">
                <el-table-column prop="topic" label="主题" min-width="200" show-overflow-tooltip />
                <el-table-column label="方向" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.direction === 'up' ? 'success' : 'info'" size="small">
                      {{ row.direction === 'up' ? '上行' : '下行' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="Payload" min-width="300" show-overflow-tooltip>
                  <template #default="{ row }">
                    <span class="mono">{{ formatValue(row.payload) }}</span>
                  </template>
                </el-table-column>
                <el-table-column prop="timestamp" label="时间" width="180" />
              </el-table>
              <div v-if="mqttMessages.length === 0 && !mqttLoading" style="text-align: center; padding: 24px; color: #909399;">
                暂无 MQTT 报文记录
              </div>
            </el-tab-pane>

            <!-- 5. 在线调试日志 -->
            <el-tab-pane label="在线调试日志" name="debug-log">
              <div style="margin-bottom: 12px; display: flex; gap: 12px; align-items: center; flex-wrap: wrap;">
                <el-select v-model="logTypeFilter" placeholder="类型筛选" size="small" style="width: 160px" @change="loadDebugLogs">
                  <el-option label="全部" value="" />
                  <el-option label="属性" value="property" />
                  <el-option label="事件" value="event" />
                  <el-option label="服务" value="service" />
                  <el-option label="影子" value="shadow" />
                </el-select>
                <el-button @click="loadDebugLogs" :loading="debugLogLoading">刷新</el-button>
              </div>
              <el-table :data="debugLogs" border size="small" v-loading="debugLogLoading" style="width: 100%">
                <el-table-column prop="type" label="类型" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getLogTypeTagType(row.type)" size="small">{{ row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="action" label="操作" width="120" />
                <el-table-column label="详情" min-width="300" show-overflow-tooltip>
                  <template #default="{ row }">
                    <span class="mono">{{ formatValue(row.details) }}</span>
                  </template>
                </el-table-column>
                <el-table-column prop="operator" label="操作人" width="120" />
                <el-table-column prop="created_at" label="时间" width="180" />
              </el-table>
              <el-pagination
                style="margin-top:12px;justify-content:flex-end"
                small
                layout="total, prev, pager, next"
                :total="debugLogTotal"
                :page-size="debugLogPageSize"
                v-model:current-page="debugLogPage"
                @current-change="loadDebugLogs"
              />
              <div v-if="debugLogs.length === 0 && !debugLogLoading" style="text-align: center; padding: 24px; color: #909399;">
                暂无调试日志
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Event Report Dialog -->
    <el-dialog v-model="showEventDialog" title="上报事件" width="500px">
      <el-form label-width="100px">
        <el-form-item label="事件">
          <el-select v-model="eventForm.event_id" placeholder="选择事件" style="width:100%" @change="onEventSelect">
            <el-option v-for="e in allEvents" :key="e.identifier" :label="e.name" :value="e.identifier" />
          </el-select>
        </el-form-item>
        <el-form-item label="输出参数">
          <el-input v-model="eventForm.outputStr" type="textarea" :rows="5" placeholder='JSON 格式，如 {"key": "value"}' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEventDialog = false">取消</el-button>
        <el-button type="primary" @click="reportEvent" :loading="eventSubmitting">上报</el-button>
      </template>
    </el-dialog>

    <!-- Service Invocation Dialog -->
    <el-dialog v-model="showServiceDialog" title="调用服务" width="550px">
      <el-form label-width="100px">
        <el-form-item label="服务">
          <el-select v-model="serviceForm.service_id" placeholder="选择服务" style="width:100%" @change="onServiceSelect">
            <el-option v-for="s in allServices" :key="s.identifier" :label="s.name" :value="s.identifier" />
          </el-select>
        </el-form-item>
        <template v-if="serviceForm.params.length > 0">
          <el-form-item v-for="(param, idx) in serviceForm.params" :key="param.identifier" :label="param.name">
            <el-switch v-if="param._dataType === 'bool'" v-model="serviceForm.values[param.identifier]" />
            <el-input-number v-else-if="param._dataType === 'int' || param._dataType === 'float'" v-model="serviceForm.values[param.identifier]" style="width:100%" />
            <el-input v-else v-model="serviceForm.values[param.identifier]" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="showServiceDialog = false">取消</el-button>
        <el-button type="primary" @click="invokeService" :loading="serviceSubmitting">调用</el-button>
      </template>
    </el-dialog>

    <!-- Shadow Desired Dialog -->
    <el-dialog v-model="showShadowDialog" title="设置期望值" width="500px">
      <el-input v-model="shadowDesiredStr" type="textarea" :rows="8" placeholder='JSON 格式，如 {"temp_01": 25}' />
      <template #footer>
        <el-button @click="showShadowDialog = false">取消</el-button>
        <el-button type="primary" @click="updateShadow" :loading="shadowSubmitting">设置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, computed, nextTick, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import request from '@/utils/request'
import { Search } from '@element-plus/icons-vue'

const route = useRoute()
const device = ref({})
const latest = ref({})
const properties = ref([])
const controlValues = reactive({})

const id = computed(() => route.params.id)

const allProperties = ref([])
const allEvents = ref([])
const allServices = ref([])

const activeTab = ref('events')

const shadow = ref(null)
const showShadowDialog = ref(false)
const shadowDesiredStr = ref('{}')
const shadowSubmitting = ref(false)

const eventList = ref([])
const eventTotal = ref(0)
const eventPage = ref(1)
const eventPageSize = 10
const eventLoading = ref(false)
const showEventDialog = ref(false)
const eventForm = reactive({ event_id: '', event_name: '', outputStr: '{}' })
const eventSubmitting = ref(false)

const serviceList = ref([])
const serviceTotal = ref(0)
const servicePage = ref(1)
const servicePageSize = 10
const serviceLoading = ref(false)
const showServiceDialog = ref(false)
const serviceForm = reactive({ service_id: '', service_name: '', params: [], values: {} })
const serviceSubmitting = ref(false)

const chartProperty = ref('')
const chartRef = ref(null)
let chartInstance = null

// ===== Debug Tab State =====
const debugActiveTab = ref('debug-property')

// 1. 模拟上报属性
const writableProperties = computed(() => allProperties.value.filter(p => (p.accessMode || 'rw').includes('w')))
const propertyReportForm = reactive({})
const propertyReportLoading = ref(false)
const lastPropertyReport = ref(null)

// 2. 模拟上报事件
const debugEventForm = reactive({ event_id: '', event_name: '', params: [], values: {} })
const debugEventLoading = ref(false)
const lastDebugEvent = ref(null)

// 3. 调用服务
const debugServiceForm = reactive({ service_id: '', service_name: '', params: [], values: {} })
const debugServiceLoading = ref(false)
const debugServiceHistory = ref([])
const debugServiceHistoryTotal = ref(0)
const debugServiceHistoryPage = ref(1)
const debugServiceHistoryPageSize = 10
const debugServiceHistoryLoading = ref(false)

// 4. MQTT 原始报文
const mqttMessages = ref([])
const mqttLoading = ref(false)
const mqttAutoRefresh = ref(false)
const mqttDirectionFilter = ref('')
const mqttTopicFilter = ref('')
const mqttTopics = computed(() => [...new Set(mqttMessages.value.map(m => m.topic))])
let mqttAutoRefreshTimer = null

// 5. 在线调试日志
const debugLogs = ref([])
const debugLogLoading = ref(false)
const debugLogTotal = ref(0)
const debugLogPage = ref(1)
const debugLogPageSize = 20
const logTypeFilter = ref('')

const load = async () => {
  try {
    const data = await request.get('/device/' + id.value)
    device.value = data?.device || data || {}
    latest.value = data?.latest || data?.data || {}
    const thingModel = data?.thingModel || {}

    const props = thingModel?.properties || data?.properties || []
    const arr = Array.isArray(props) ? props : Object.entries(props).map(([k, v]) => ({ identifier: k, ...v }))
    allProperties.value = arr
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

    const events = thingModel?.events || []
    allEvents.value = Array.isArray(events) ? events : Object.entries(events).map(([k, v]) => ({ identifier: k, ...v }))

    const svcs = thingModel?.services || []
    allServices.value = Array.isArray(svcs) ? svcs : Object.entries(svcs).map(([k, v]) => ({ identifier: k, ...v }))
  } catch (e) {
    console.error(e)
    ElMessage.error('加载设备详情失败')
  }

  loadEvents()
  loadServices()
  loadShadow()
  if (allProperties.value.length > 0 && !chartProperty.value) {
    chartProperty.value = allProperties.value[0].identifier
    nextTick(() => loadChartData())
  }
}

const sendCommand = async (p) => {
  try {
    await request.post('/device/' + id.value + '/control', { [p.identifier]: controlValues[p.identifier] })
    ElMessage.success(p.name + ' 指令下发成功')
  } catch (e) {
    ElMessage.error(p.name + ' 指令下发失败')
  }
}

const formatValue = (v) => {
  if (v === null || v === undefined) return '-'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

const prettyJson = (obj) => {
  if (!obj) return '{}'
  try { return JSON.stringify(typeof obj === 'string' ? JSON.parse(obj) : obj, null, 2) }
  catch { return String(obj) }
}

// Events
const loadEvents = async () => {
  eventLoading.value = true
  try {
    const data = await request.get('/device/' + id.value + '/event', { params: { page: eventPage.value, size: eventPageSize } })
    eventList.value = data?.list || []
    eventTotal.value = data?.total || 0
  } catch { /* ignore */ }
  eventLoading.value = false
}

const onEventSelect = (id) => {
  const ev = allEvents.value.find(e => e.identifier === id)
  eventForm.event_name = ev?.name || ''
  const outParams = ev?.outputParams || ev?.output || []
  if (Array.isArray(outParams) && outParams.length > 0) {
    const obj = {}
    outParams.forEach(p => { obj[p.identifier || p.name] = p.defaultValue || '' })
    eventForm.outputStr = JSON.stringify(obj)
  } else {
    eventForm.outputStr = '{}'
  }
}

const reportEvent = async () => {
  let output
  try { output = JSON.parse(eventForm.outputStr) } catch { ElMessage.warning('输出参数格式错误'); return }
  eventSubmitting.value = true
  try {
    await request.post('/device/' + id.value + '/event', { event_id: eventForm.event_id, event_name: eventForm.event_name, output })
    ElMessage.success('事件上报成功')
    showEventDialog.value = false
    loadEvents()
  } catch { /* handled by interceptor */ }
  eventSubmitting.value = false
}

// Services
const loadServices = async () => {
  serviceLoading.value = true
  try {
    const data = await request.get('/device/' + id.value + '/service', { params: { page: servicePage.value, size: servicePageSize } })
    serviceList.value = data?.list || []
    serviceTotal.value = data?.total || 0
  } catch { /* ignore */ }
  serviceLoading.value = false
}

const onServiceSelect = (sid) => {
  const svc = allServices.value.find(s => s.identifier === sid)
  serviceForm.service_name = svc?.name || ''
  const inParams = svc?.inputParams || svc?.input || []
  const params = Array.isArray(inParams) ? inParams : Object.entries(inParams).map(([k, v]) => ({ identifier: k, ...v }))
  serviceForm.params = params.map(p => {
    const dt = (typeof p.dataType === 'object' ? p.dataType?.type : p.dataType) || 'string'
    return { ...p, _dataType: dt }
  })
  serviceForm.values = {}
  serviceForm.params.forEach(p => {
    if (p._dataType === 'bool') serviceForm.values[p.identifier] = false
    else if (p._dataType === 'int' || p._dataType === 'float') serviceForm.values[p.identifier] = p.defaultValue || 0
    else serviceForm.values[p.identifier] = p.defaultValue || ''
  })
}

const invokeService = async () => {
  serviceSubmitting.value = true
  try {
    await request.post('/device/' + id.value + '/service', {
      service_id: serviceForm.service_id,
      service_name: serviceForm.service_name,
      input: { ...serviceForm.values }
    })
    ElMessage.success('服务调用成功')
    showServiceDialog.value = false
    loadServices()
  } catch { /* handled by interceptor */ }
  serviceSubmitting.value = false
}

// Shadow
const loadShadow = async () => {
  try {
    shadow.value = await request.get('/device/' + id.value + '/shadow')
  } catch { shadow.value = null }
}

const updateShadow = async () => {
  let desired
  try { desired = JSON.parse(shadowDesiredStr.value) } catch { ElMessage.warning('JSON 格式错误'); return }
  shadowSubmitting.value = true
  try {
    await request.put('/device/' + id.value + '/shadow', { desired })
    ElMessage.success('期望值设置成功')
    showShadowDialog.value = false
    loadShadow()
  } catch { /* handled by interceptor */ }
  shadowSubmitting.value = false
}

// Chart
const loadChartData = async () => {
  if (!chartProperty.value) return
  const sn = device.value.sn || device.value.device_sn || device.value.deviceKey
  if (!sn) return
  try {
    const data = await request.get('/device/data/' + sn, { params: { property: chartProperty.value, limit: 200 } })
    const list = data?.list || []
    await nextTick()
    if (!chartRef.value) return
    if (!chartInstance) {
      chartInstance = echarts.init(chartRef.value)
    }
    const times = list.map(i => i.created_at)
    const values = list.map(i => {
      const v = i.value
      return isNaN(Number(v)) ? v : Number(v)
    })
    const propInfo = allProperties.value.find(p => p.identifier === chartProperty.value)
    chartInstance.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 50, right: 20, top: 20, bottom: 40 },
      xAxis: { type: 'category', data: times, axisLabel: { fontSize: 10, rotate: 30 } },
      yAxis: { type: 'value', name: propInfo?.name || chartProperty.value },
      series: [{ type: 'line', data: values, smooth: true, areaStyle: { opacity: 0.15 }, symbol: 'none' }]
    }, true)
  } catch { /* ignore */ }
}

const handleResize = () => chartInstance?.resize()
onMounted(() => window.addEventListener('resize', handleResize))
onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  // Clear MQTT auto-refresh timer
  if (mqttAutoRefreshTimer) {
    clearInterval(mqttAutoRefreshTimer)
  }
})

// ===== Debug Tab Methods =====

// 初始化属性上报表单
const initPropertyReportForm = () => {
  propertyReportForm = {}
  writableProperties.value.forEach(p => {
    if (p._dataType === 'bool') propertyReportForm[p.identifier] = false
    else if (p._dataType === 'int' || p._dataType === 'float') propertyReportForm[p.identifier] = p._specs?.min || 0
    else if (p._dataType === 'enum') propertyReportForm[p.identifier] = p._specs?.enumList?.[0]?.value || ''
    else propertyReportForm[p.identifier] = ''
  })
}

// 重置属性上报表单
const resetPropertyForm = () => {
  initPropertyReportForm()
  lastPropertyReport.value = null
}

// 模拟上报属性
const reportProperties = async () => {
  const payload = { ...propertyReportForm }
  // 过滤掉未填写的字段
  Object.keys(payload).forEach(k => {
    if (payload[k] === '' || payload[k] === null || payload[k] === undefined) {
      delete payload[k]
    }
  })
  if (Object.keys(payload).length === 0) {
    ElMessage.warning('请至少填写一个属性值')
    return
  }
  propertyReportLoading.value = true
  const startTime = Date.now()
  try {
    const sn = device.value.sn || device.value.device_sn || device.value.deviceKey
    const res = await request.post('/api/device/report', { sn, properties: payload })
    ElMessage.success('属性上报成功')
    lastPropertyReport.value = {
      time: new Date(),
      success: true,
      response: res?.message || 'OK'
    }
  } catch (e) {
    ElMessage.error('属性上报失败')
    lastPropertyReport.value = {
      time: new Date(),
      success: false,
      response: e?.message || 'Error'
    }
  } finally {
    propertyReportLoading.value = false
  }
}

// 事件选择变化
const onDebugEventSelect = (eventId) => {
  const ev = allEvents.value.find(e => e.identifier === eventId)
  if (!ev) return
  debugEventForm.event_name = ev.name || ''
  const outParams = ev.outputParams || ev.output || []
  const params = Array.isArray(outParams) ? outParams : Object.entries(outParams).map(([k, v]) => ({ identifier: k, ...v }))
  debugEventForm.params = params.map(p => {
    const dt = (typeof p.dataType === 'object' ? p.dataType?.type : p.dataType) || 'string'
    const specs = (typeof p.dataType === 'object' ? p.dataType?.specs : {}) || {}
    return { ...p, _dataType: dt, _specs: specs }
  })
  debugEventForm.values = {}
  debugEventForm.params.forEach(p => {
    if (p._dataType === 'bool') debugEventForm.values[p.identifier] = false
    else if (p._dataType === 'int' || p._dataType === 'float') debugEventForm.values[p.identifier] = p._specs?.min || p.defaultValue || 0
    else if (p._dataType === 'enum') debugEventForm.values[p.identifier] = p._specs?.enumList?.[0]?.value || p.defaultValue || ''
    else debugEventForm.values[p.identifier] = p.defaultValue || ''
  })
  lastDebugEvent.value = null
}

// 重置事件上报表单
const resetDebugEventForm = () => {
  debugEventForm.event_id = ''
  debugEventForm.event_name = ''
  debugEventForm.params = []
  debugEventForm.values = {}
  lastDebugEvent.value = null
}

// 模拟上报事件
const reportDebugEvent = async () => {
  if (!debugEventForm.event_id) {
    ElMessage.warning('请选择事件')
    return
  }
  const output = { ...debugEventForm.values }
  debugEventLoading.value = true
  try {
    await request.post('/device/' + id.value + '/event', {
      event_id: debugEventForm.event_id,
      event_name: debugEventForm.event_name,
      output
    })
    ElMessage.success('事件上报成功')
    lastDebugEvent.value = { time: new Date(), success: true }
    loadEvents()
  } catch {
    lastDebugEvent.value = { time: new Date(), success: false }
  } finally {
    debugEventLoading.value = false
  }
}

// 服务选择变化
const onDebugServiceSelect = (serviceId) => {
  const svc = allServices.value.find(s => s.identifier === serviceId)
  if (!svc) return
  debugServiceForm.service_name = svc.name || ''
  const inParams = svc.inputParams || svc.input || []
  const params = Array.isArray(inParams) ? inParams : Object.entries(inParams).map(([k, v]) => ({ identifier: k, ...v }))
  debugServiceForm.params = params.map(p => {
    const dt = (typeof p.dataType === 'object' ? p.dataType?.type : p.dataType) || 'string'
    const specs = (typeof p.dataType === 'object' ? p.dataType?.specs : {}) || {}
    return { ...p, _dataType: dt, _specs: specs }
  })
  debugServiceForm.values = {}
  debugServiceForm.params.forEach(p => {
    if (p._dataType === 'bool') debugServiceForm.values[p.identifier] = false
    else if (p._dataType === 'int' || p._dataType === 'float') debugServiceForm.values[p.identifier] = p._specs?.min || p.defaultValue || 0
    else if (p._dataType === 'enum') debugServiceForm.values[p.identifier] = p._specs?.enumList?.[0]?.value || p.defaultValue || ''
    else debugServiceForm.values[p.identifier] = p.defaultValue || ''
  })
}

// 重置服务调用表单
const resetDebugServiceForm = () => {
  debugServiceForm.service_id = ''
  debugServiceForm.service_name = ''
  debugServiceForm.params = []
  debugServiceForm.values = {}
}

// 调用服务 (调试版)
const invokeDebugService = async () => {
  if (!debugServiceForm.service_id) {
    ElMessage.warning('请选择服务')
    return
  }
  const input = { ...debugServiceForm.values }
  debugServiceLoading.value = true
  try {
    await request.post('/device/' + id.value + '/service', {
      service_id: debugServiceForm.service_id,
      service_name: debugServiceForm.service_name,
      input
    })
    ElMessage.success('服务调用成功')
    loadDebugServiceHistory()
  } catch { /* handled by interceptor */ }
  finally {
    debugServiceLoading.value = false
  }
}

// 加载服务调用历史
const loadDebugServiceHistory = async () => {
  debugServiceHistoryLoading.value = true
  try {
    const data = await request.get('/device/' + id.value + '/service', {
      params: { page: debugServiceHistoryPage.value, size: debugServiceHistoryPageSize }
    })
    debugServiceHistory.value = data?.list || []
    debugServiceHistoryTotal.value = data?.total || 0
  } catch { /* ignore */ }
  debugServiceHistoryLoading.value = false
}

// 加载 MQTT 报文
const loadMqttMessages = async () => {
  mqttLoading.value = true
  try {
    const sn = device.value.sn || device.value.device_sn || device.value.deviceKey
    if (!sn) {
      mqttMessages.value = []
      return
    }
    const params = { limit: 50 }
    if (mqttDirectionFilter.value) params.direction = mqttDirectionFilter.value
    if (mqttTopicFilter.value) params.topic = mqttTopicFilter.value
    const data = await request.get('/api/device/mqtt/messages/' + sn, { params })
    mqttMessages.value = data?.list || []
  } catch {
    mqttMessages.value = []
  }
  mqttLoading.value = false
}

// 切换 MQTT 自动刷新
const toggleMqttAutoRefresh = (val) => {
  if (val) {
    loadMqttMessages()
    mqttAutoRefreshTimer = setInterval(loadMqttMessages, 5000)
  } else {
    if (mqttAutoRefreshTimer) {
      clearInterval(mqttAutoRefreshTimer)
      mqttAutoRefreshTimer = null
    }
  }
}

// 加载调试日志
const loadDebugLogs = async () => {
  debugLogLoading.value = true
  try {
    const params = { page: debugLogPage.value, size: debugLogPageSize }
    if (logTypeFilter.value) params.type = logTypeFilter.value
    // 使用 operation-log API，target 为 device:deviceId
    const data = await request.get('/api/admin/operation-log', {
      params: { ...params, target: 'device:' + id.value }
    })
    debugLogs.value = data?.list || []
    debugLogTotal.value = data?.total || 0
  } catch {
    debugLogs.value = []
    debugLogTotal.value = 0
  }
  debugLogLoading.value = false
}

// 获取日志类型标签类型
const getLogTypeTagType = (type) => {
  const map = { property: 'info', event: 'warning', service: 'primary', shadow: 'success' }
  return map[type] || 'info'
}

// 格式化时间
const formatTime = (date) => {
  if (!date) return '-'
  const d = new Date(date)
  return d.toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
}

onMounted(load)
</script>

<style scoped>
.title { font-weight: 600; color: #303133; }
.latest-value { font-weight: 600; color: #409eff; font-size: 15px; }
.mono { font-family: monospace; font-size: 12px; color: #606266; word-break: break-all; }
.shadow-label { font-size: 12px; font-weight: 600; color: #606266; margin-bottom: 4px; }
.shadow-json { background: #f5f7fa; border-radius: 4px; padding: 8px; font-size: 12px; max-height: 160px; overflow: auto; margin: 0; white-space: pre-wrap; word-break: break-all; }
</style>
