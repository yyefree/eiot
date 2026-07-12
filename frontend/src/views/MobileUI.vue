<template>
  <div class="container-page">
    <div class="page-head">
      <h2 class="page-title" style="margin:0">移动端 UI 配置</h2>
      <div class="page-toolbar">
        <span style="margin-right:8px;color:#666">产品：</span>
        <el-select v-model="selectedProduct" placeholder="请选择产品" size="default"
                   filterable style="width:320px" @change="onProductChange">
          <el-option v-for="p in products" :key="p.id" :label="p.name + ' (#' + p.id + ')'" :value="p.id" />
        </el-select>
        <el-button type="primary" size="default" :disabled="!selectedProduct || !hasChanged" @click="saveLayout">
          <el-icon><Check /></el-icon>
          保存配置
        </el-button>
        <el-button size="default" :disabled="!selectedProduct" @click="clearCanvas">
          <el-icon><Delete /></el-icon>
          清空
        </el-button>
      </div>
    </div>

    <el-alert v-if="!selectedProduct" type="warning" :closable="false" style="margin:12px 0">
      请先在上方选择一个产品，每个产品各自拥有独立的移动端 UI 配置。
    </el-alert>

    <div v-if="productInfo" style="margin:10px 0 14px;color:#666;font-size:13px">
      关联物模型：<b>{{ productInfo.thing_model_id || '(无)' }}</b>
      ｜ 可用物理量：<b>{{ (productInfo.properties || []).length }}</b> 个
      ｜ 上次保存：{{ productInfo.mobile_ui_json ? '已有配置' : '空白' }}
    </div>

    <div class="mobile-grid" :class="{ disabled: !selectedProduct }">
      <!-- 左侧：可用组件池（从物模型生成） -->
      <el-card class="panel components">
        <template #header>
          <b>可用组件</b>
          <span style="margin-left:6px;color:#999;font-weight:normal;font-size:12px">拖入右侧画布</span>
        </template>

        <div v-if="availableComponents.length===0" style="color:#999;font-size:13px;padding:10px 0">
          该产品未配置物模型物理量。
        </div>
        <div class="comp-list">
          <div v-for="c in availableComponents" :key="c.key" class="comp-item"
            draggable="true" @dragstart="onDragStart($event, c)">
            <el-icon><component :is="c.icon" /></el-icon>
            <div style="margin-left:10px;flex:1;line-height:1.3">
              <div style="font-size:14px;color:#333;font-weight:500">{{ c.name }}</div>
              <div style="font-size:12px;color:#999">{{ c.identifier }} · {{ c.dataType }}</div>
            </div>
          </div>
          <div class="comp-item" draggable="true"
               @dragstart="onDragStart($event, { type: 'text', name: '文本信息', icon: 'Document', isStatic: true })">
            <el-icon><Document /></el-icon>
            <div style="margin-left:10px"><div style="font-size:14px;color:#333;font-weight:500">文本信息</div>
              <div style="font-size:12px;color:#999">静态文本</div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 中间：画布（移动端预览） -->
      <el-card class="panel canvas">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <b>画布（移动端预览）</b>
            <span style="color:#999;font-weight:normal;font-size:12px">共 {{ canvas.length }} 个组件</span>
          </div>
        </template>
        <div class="phone-frame" @dragover.prevent @drop="onDrop">
          <div class="phone-title">{{ currentProductName }}</div>
          <div class="phone-content">
            <div v-if="canvas.length===0" class="phone-empty">拖拽左侧组件到这里</div>
            <template v-for="(it,idx) in canvas" :key="it.key || idx">
              <div class="phone-item" :class="it.type">
                <div class="phone-item-head">
                  <el-input v-model="it.label" size="small"
                    style="background:transparent;border:none;color:#fff" />
                  <div>
                    <el-button size="small" link :disabled="idx===0" @click="moveUp(idx)">↑</el-button>
                    <el-button size="small" link :disabled="idx===canvas.length-1" @click="moveDown(idx)">↓</el-button>
                    <el-button size="small" type="danger" link @click="removeItem(idx)">×</el-button>
                  </div>
                </div>
                <div class="phone-item-body">
                  <el-switch v-if="it.type==='switch'" />
                  <div v-if="it.type==='slider'" class="slider-wrap">
                    <el-slider :model-value="50" style="width:100%" />
                  </div>
                  <el-input-number v-if="it.type==='number'" :model-value="25" size="small" />
                  <span v-if="it.type==='sensor'" style="font-size:22px;font-weight:600;color:#409eff">
                    {{ it.sample || '---' }} {{ it.unit || '' }}
                  </span>
                  <span v-if="it.type==='text'" style="color:#ddd">{{ it.content || '文本信息' }}</span>
                </div>
              </div>
            </template>
          </div>
        </div>
      </el-card>

      <!-- 右侧：布局 JSON -->
      <el-card class="panel json-view">
        <template #header><b>布局 JSON</b></template>
        <el-input type="textarea" :rows="22" v-model="layoutJson" @change="onJsonChange" />
        <div style="margin-top:8px;color:#999;font-size:12px">
          提示：直接在 JSON 中编辑也会同步到左侧画布。
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import { useRoute } from 'vue-router'

const route = useRoute()

const products = ref([])
const selectedProduct = ref(null)
const productInfo = ref(null)
const canvas = ref([])
const savedLayout = ref('[]')
const hasChanged = computed(() => JSON.stringify(canvas.value) !== savedLayout.value)

const currentProductName = computed(() => {
  if (productInfo.value) return productInfo.value.product_name || '移动端控制'
  return '移动端控制'
})

// 从物模型 properties 映射到 UI 组件类型
function pickCompType(prop) {
  const dt = (typeof prop.dataType === 'object' ? prop.dataType?.type : prop.dataType) || 'string'
  const specs = (typeof prop.dataType === 'object' ? prop.dataType?.specs : {}) || {}
  if (dt === 'bool') return { type: 'switch', icon: 'Lightning' }
  const dtLower = dt.toLowerCase()
  if (dtLower === 'float' || dtLower === 'double' || dtLower === 'int' || dtLower === 'number') {
    if (specs.min !== undefined && specs.max !== undefined) {
      return { type: 'slider', icon: 'Tools' }
    }
    return { type: 'number', icon: 'Edit' }
  }
  if (dtLower === 'string' || dtLower === 'enum') return { type: 'text', icon: 'Document' }
  return { type: 'sensor', icon: 'DataBoard' }
}

const availableComponents = computed(() => {
  const list = []
  const props = (productInfo.value && productInfo.value.properties) || []
  props.forEach((p, idx) => {
    const map = pickCompType(p)
    const specs = (p.dataType && typeof p.dataType === 'object') ? (p.dataType.specs || {}) : {}
    // 为数值/浮点额外加一个「传感器展示」类型
    const baseName = p.name || p.identifier || ('物理量' + idx)
    list.push({
      key: p.identifier + '-' + map.type,
      type: map.type,
      icon: map.icon,
      name: baseName + '（' + componentLabel(map.type) + '）',
      identifier: p.identifier,
      dataType: p.dataType?.type || p.dataType || '-',
      unit: specs.unit || '',
      sample: map.type === 'sensor' ? '--' : undefined
    })
    if (map.type === 'slider' || map.type === 'number') {
      list.push({
        key: p.identifier + '-sensor',
        type: 'sensor',
        icon: 'DataBoard',
        name: baseName + '（传感器展示）',
        identifier: p.identifier,
        dataType: p.dataType?.type || p.dataType || '-',
        unit: specs.unit || '',
        sample: '--'
      })
    }
  })
  return list
})

function componentLabel(t) {
  if (t === 'switch') return '开关'
  if (t === 'slider') return '滑块'
  if (t === 'number') return '数值'
  if (t === 'sensor') return '传感器展示'
  if (t === 'text') return '文本'
  return t
}

// 布局 JSON
const layoutJson = computed({
  get() {
    return JSON.stringify(canvas.value, null, 2)
  },
  set(v) {
    try {
      const obj = JSON.parse(v)
      if (Array.isArray(obj)) canvas.value = obj
    } catch (_) {}
  }
})

function onJsonChange() {} // 由 computed set 处理

function onDragStart(e, c) {
  e.dataTransfer.setData('application/json', JSON.stringify(c))
}
function onDrop(e) {
  try {
    const data = JSON.parse(e.dataTransfer.getData('application/json'))
    const item = {
      type: data.type,
      key: data.key || ('k_' + Date.now() + '_' + Math.random().toString(36).slice(2, 6)),
      label: data.name || '未命名',
      identifier: data.identifier || '',
      dataType: data.dataType || '',
      unit: data.unit || '',
      content: data.isStatic ? '示例文本' : ''
    }
    if (data.type === 'sensor') {
      item.sample = '--'
    }
    canvas.value.push(item)
  } catch (_) {}
}
function removeItem(idx) {
  canvas.value.splice(idx, 1)
}
function moveUp(idx) {
  if (idx > 0) {
    const t = canvas.value[idx]
    canvas.value[idx] = canvas.value[idx - 1]
    canvas.value[idx - 1] = t
  }
}
function moveDown(idx) {
  if (idx < canvas.value.length - 1) {
    const t = canvas.value[idx]
    canvas.value[idx] = canvas.value[idx + 1]
    canvas.value[idx + 1] = t
  }
}
function clearCanvas() {
  ElMessageBox.confirm('确定清空画布？', '提示', { type: 'warning' }).then(() => {
    canvas.value = []
  }).catch(() => {})
}

// 产品加载
async function loadProducts() {
  try {
    const data = await request.get('/admin/product', { params: { page: 1, size: 500 } })
    products.value = (data && data.list) || []
  } catch (e) {}
}

async function onProductChange() {
  if (!selectedProduct.value) {
    productInfo.value = null
    canvas.value = []
    savedLayout.value = '[]'
    return
  }
  try {
    const info = await request.get('/admin/product/' + selectedProduct.value + '/mobile-ui')
    productInfo.value = info
    try {
      const arr = JSON.parse(info.mobile_ui_json || '[]')
      canvas.value = Array.isArray(arr) ? arr : []
    } catch (_) {
      canvas.value = []
    }
    savedLayout.value = JSON.stringify(canvas.value)
  } catch (e) {
    ElMessage.error('加载产品 UI 配置失败')
  }
}

async function saveLayout() {
  if (!selectedProduct.value) return
  try {
    await request.put('/admin/product/' + selectedProduct.value + '/mobile-ui', {
      mobile_ui_json: JSON.stringify(canvas.value)
    })
    savedLayout.value = JSON.stringify(canvas.value)
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error('保存失败')
  }
}

onMounted(async () => {
  await loadProducts()
  const q = route.query.productId
  if (q && products.value.length > 0) {
    const id = Number(q)
    if (products.value.some(p => p.id === id)) selectedProduct.value = id
  }
  if (!selectedProduct.value && products.value.length > 0) {
    selectedProduct.value = products.value[0].id
  }
  if (selectedProduct.value) {
    await onProductChange()
  }
})
</script>

<style scoped>
.container-page { padding: 16px; }
.page-head { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 10px; }
.page-toolbar { display: flex; align-items: center; }
.mobile-grid {
  display: grid;
  grid-template-columns: 280px 1fr 360px;
  gap: 12px;
  margin-top: 12px;
}
.mobile-grid.disabled { opacity: 0.5; pointer-events: none; }
.panel { min-height: 0; }
.comp-list { display: flex; flex-direction: column; gap: 8px; }
.comp-item {
  padding: 10px 12px; background: #f5f7fa; border-radius: 6px; cursor: grab;
  display: flex; align-items: center;
}
.comp-item:hover { background: #ecf5ff; }
.phone-frame { background: #222; border-radius: 24px; padding: 14px; min-height: 560px; color: #fff; }
.phone-title { text-align: center; font-weight: bold; margin-bottom: 10px; color: #fff; }
.phone-content { background: #333; border-radius: 12px; padding: 12px; min-height: 480px; }
.phone-empty { text-align: center; color: #888; margin-top: 80px; font-size: 14px; }
.phone-item { background: #444; border-radius: 10px; padding: 10px; margin-bottom: 10px; color: #fff; }
.phone-item-head { display: flex; justify-content: space-between; align-items: center; font-size: 13px; color: #ddd; margin-bottom: 6px; }
.phone-item-head :deep(.el-input__wrapper) {
  background: transparent; box-shadow: none; color: #fff;
}
.phone-item-head :deep(.el-input__inner) { color: #fff; }
.slider-wrap { display: flex; align-items: center; }
@media (max-width: 1200px) {
  .mobile-grid { grid-template-columns: 1fr; }
  .phone-frame { min-height: 420px; }
}
</style>
