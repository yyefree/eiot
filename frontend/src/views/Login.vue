<template>
  <div class="login-page">
    <el-card class="login-card">
      <div class="login-header">
        <div class="logo">
          <el-icon :size="36"><Connection /></el-icon>
        </div>
        <h2>EIOT 物联网平台</h2>
        <p class="subtitle">欢迎登录，请选择登录方式</p>
      </div>

      <el-tabs v-model="loginType" class="login-tabs">
        <el-tab-pane label="密码登录" name="password">
          <el-form :model="passwordForm" ref="passwordFormRef" label-position="top" size="large">
            <el-form-item label="手机号" prop="phone" :rules="[{ required: true, message: '请输入手机号' }]">
              <el-input v-model="passwordForm.phone" placeholder="请输入手机号（如 13800000000）" :prefix-icon="User" maxlength="20" />
            </el-form-item>
            <el-form-item label="密码" prop="password" :rules="[{ required: true, message: '请输入密码' }]">
              <el-input v-model="passwordForm.password" placeholder="请输入密码（如 admin123）" type="password" show-password :prefix-icon="Lock" />
            </el-form-item>
            <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="submitPassword">登 录</el-button>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="验证码登录" name="code">
          <el-form :model="codeForm" ref="codeFormRef" label-position="top" size="large">
            <el-form-item label="手机号" prop="phone" :rules="[{ required: true, message: '请输入手机号' }]">
              <el-input v-model="codeForm.phone" placeholder="请输入手机号" :prefix-icon="User" maxlength="20" />
            </el-form-item>
            <el-form-item label="验证码" prop="code" :rules="[{ required: true, message: '请输入验证码' }]">
              <div style="display:flex; gap:8px; width:100%">
                <el-input v-model="codeForm.code" placeholder="请输入验证码" :prefix-icon="Key" style="flex:1" />
                <el-button :disabled="codeCountdown > 0" @click="sendCode" style="min-width:110px">
                  {{ codeCountdown > 0 ? codeCountdown + 's' : '获取验证码' }}
                </el-button>
              </div>
            </el-form-item>
            <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="submitCode">登 录</el-button>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="login-footer">
        <span>© EIOT 物联网平台</span>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Key, Connection } from '@element-plus/icons-vue'
import request from '@/utils/request'

const router = useRouter()
const loginType = ref('password')
const loading = ref(false)
const passwordFormRef = ref()
const codeFormRef = ref()

const passwordForm = ref({ phone: '', password: '' })
const codeForm = ref({ phone: '', code: '' })
const codeCountdown = ref(0)
let countdownTimer = null

onUnmounted(() => { if (countdownTimer) clearInterval(countdownTimer) })

const submitPassword = async () => {
  try {
    await passwordFormRef.value.validate()
    loading.value = true
    const data = await request.post('/auth/login', passwordForm.value)
    const token = typeof data === 'string' ? data : data?.token || data?.accessToken
    if (!token) { ElMessage.error('登录失败: 未获取到令牌'); return }
    localStorage.setItem('eiot_token', token)
    if (data?.user) localStorage.setItem('eiot_user', JSON.stringify(data.user))
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e) {
    ElMessage.error(e?.message || '登录失败，请重试')
  } finally {
    loading.value = false
  }
}

const sendCode = async () => {
  if (!codeForm.value.phone) {
    ElMessage.warning('请先输入手机号')
    return
  }
  try {
    await request.post('/auth/send-code', { phone: codeForm.value.phone })
    ElMessage.success('验证码已发送')
    codeCountdown.value = 60
    if (countdownTimer) clearInterval(countdownTimer)
    countdownTimer = setInterval(() => {
      codeCountdown.value--
      if (codeCountdown.value <= 0) { clearInterval(countdownTimer); countdownTimer = null }
    }, 1000)
  } catch (e) {
    ElMessage.error(e?.message || '发送失败，请重试')
  }
}

const submitCode = async () => {
  try {
    await codeFormRef.value.validate()
    loading.value = true
    const data = await request.post('/auth/login-code', codeForm.value)
    const token = typeof data === 'string' ? data : data?.token || data?.accessToken
    if (!token) { ElMessage.error('登录失败: 未获取到令牌'); return }
    localStorage.setItem('eiot_token', token)
    if (data?.user) localStorage.setItem('eiot_user', JSON.stringify(data.user))
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e) {
    ElMessage.error(e?.message || '登录失败，请重试')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #409eff22 0%, #52c41a15 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.login-card {
  width: 420px;
  max-width: 100%;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
}
.login-header {
  text-align: center;
  padding: 10px 0 10px;
}
.login-header .logo {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #409eff, #52c41a);
  border-radius: 16px;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 14px;
}
.login-header h2 {
  margin: 0 0 6px;
  font-size: 22px;
  color: #303133;
}
.subtitle {
  color: #909399;
  font-size: 13px;
  margin: 0 0 4px;
}
.login-tabs {
  margin-top: 8px;
}
.login-footer {
  text-align: center;
  color: #909399;
  font-size: 12px;
  margin-top: 16px;
}
@media (max-width: 768px) {
  .login-page { padding: 0; }
  .login-card {
    width: 100%;
    height: 100vh;
    border-radius: 0;
    box-shadow: none;
  }
}
</style>
