<template>
  <div class="page-container">
    <div class="page-header">
      <h2>项目管理</h2>
      <el-button type="primary" @click="showCreate = true">
        <el-icon><Plus /></el-icon> 新建项目
      </el-button>
    </div>

    <!-- 项目列表 -->
    <el-table v-loading="loading" :data="list" stripe class="main-table">
      <el-table-column label="ID" prop="id" width="60" />
      <el-table-column label="项目名称" prop="name" min-width="160">
        <template #default="{ row }">
          <div class="project-name">
            <el-icon class="project-icon"><FolderOpened /></el-icon>
            {{ row.name }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="行业" prop="industry" width="120" />
      <el-table-column label="类型" prop="type" width="100">
        <template #default="{ row }">
          <el-tag :type="row.type === 'consumer' ? 'success' : 'warning'" size="small">
            {{ row.type === 'consumer' ? '消费级' : '商用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="产品数" prop="prod_count" width="90" align="center" />
      <el-table-column label="描述" prop="description" min-width="200" show-overflow-tooltip />
      <el-table-column label="创建时间" prop="created_at" width="160" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="goProducts(row)" :icon="View">查看产品</el-button>
          <el-button size="small" type="primary" @click="editRow(row)" :icon="Edit">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteRow(row)" :icon="Delete">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        v-model:current-page="page"
        :page-size="20"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchList"
      />
    </div>

    <!-- 新建/编辑对话框 -->
    <el-dialog v-model="showCreate" :title="editing.id ? '编辑项目' : '新建项目'" width="480px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="项目名称" required>
          <el-input v-model="form.name" placeholder="如：新手智能家居测试项目" />
        </el-form-item>
        <el-form-item label="行业分类">
          <el-select v-model="form.industry" placeholder="选择行业">
            <el-option label="智能家居" value="智能家居" />
            <el-option label="工业控制" value="工业控制" />
            <el-option label="农业大棚" value="农业大棚" />
            <el-option label="智慧城市" value="智慧城市" />
            <el-option label="消费电子" value="消费电子" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="项目类型">
          <el-radio-group v-model="form.type">
            <el-radio value="consumer">消费级智能硬件</el-radio>
            <el-radio value="commercial">商用设备</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="项目简介（选填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveProject">
          {{ editing.id ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, View, Edit, Delete, FolderOpened } from '@element-plus/icons-vue'
import request from '../utils/request.js'
import { useRouter } from 'vue-router'

const router = useRouter()
const loading = ref(false)
const saving = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const showCreate = ref(false)
const editing = ref({})
const form = reactive({
  name: '',
  industry: '智能家居',
  type: 'consumer',
  description: ''
})

onMounted(() => fetchList())

async function fetchList() {
  loading.value = true
  try {
    const res = await request.get('/admin/project', { params: { page: page.value, size: 20 } })
    list.value = res?.list || []
    total.value = res?.total || 0
  } catch (e) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function editRow(row) {
  editing.value = row
  Object.assign(form, { name: row.name, industry: row.industry || '智能家居', type: row.type || 'consumer', description: row.description || '' })
  showCreate.value = true
}

async function saveProject() {
  if (!form.name.trim()) {
    ElMessage.warning('项目名称不可为空')
    return
  }
  saving.value = true
  try {
    if (editing.value.id) {
      await request.put('/admin/project/' + editing.value.id, form)
      ElMessage.success('保存成功')
    } else {
      await request.post('/admin/project', form)
      ElMessage.success('创建成功')
    }
    showCreate.value = false
    fetchList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    saving.value = false
  }
}

async function deleteRow(row) {
  try {
    await ElMessageBox.confirm(`确定删除项目「${row.name}」吗？`, '确认', { type: 'warning' })
    await request.delete('/admin/project/' + row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '删除失败')
  }
}

function goProducts(row) {
  router.push({ path: '/product', query: { project_id: row.id } })
}
</script>

<style scoped>
.project-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
}
.project-icon {
  color: #409eff;
}
</style>
