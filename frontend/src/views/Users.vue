<template>
  <div class="users">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="title">用户列表</span>
        </div>
      </template>

      <el-table :data="list" stripe class="desktop-only" style="width:100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="phone" label="手机号" width="160" />
        <el-table-column prop="nickname" label="昵称" width="160" />
        <el-table-column prop="email" label="邮箱" min-width="180" show-overflow-tooltip />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'warning' : 'info'" size="small">{{ row.role === 'admin' ? '管理员' : '普通用户' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="注册时间" width="170" />
      </el-table>

      <div class="mobile-list mobile-only">
        <el-card v-for="row in list" :key="row.id" class="m-item" shadow="hover">
          <div class="m-top">
            <span class="m-name">{{ row.nickname || row.phone }}</span>
            <el-tag :type="row.role === 'admin' ? 'warning' : 'info'" size="small">{{ row.role === 'admin' ? '管理员' : '普通用户' }}</el-tag>
          </div>
          <div class="m-sub">手机号: {{ row.phone }}</div>
          <div class="m-sub">邮箱: {{ row.email || '-' }}</div>
          <div class="m-sub">状态: {{ row.status === 1 ? '正常' : '禁用' }} · {{ row.createdAt }}</div>
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request from '@/utils/request'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const load = async () => {
  try {
    const data = await request.get('/admin/user', { params: { page: page.value, size: pageSize.value } })
    list.value = Array.isArray(data?.list) ? data.list : Array.isArray(data?.items) ? data.items : Array.isArray(data) ? data : []
    total.value = data?.total ?? list.value.length
  } catch (e) {
    list.value = []
    total.value = 0
  }
}

onMounted(load)
</script>

<style scoped>
.title { font-weight: 600; color: #303133; }
.m-item { margin-bottom: 10px; }
.m-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.m-name { font-weight: 600; color: #303133; }
.m-sub { color: #909399; font-size: 13px; margin-top: 4px; word-break: break-all; }
.desktop-only { display: table; }
.mobile-only { display: none; }
@media (max-width: 768px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
}
</style>
