import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const request = axios.create({
  baseURL: '/api',
  timeout: 15000
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('eiot_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    const data = response.data
    // 业务错误（HTTP 200 但 code != 0）
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0 || data.code === 200) {
        return data.data
      }
      // 认证失效跳转登录
      if (data.code === 401 || data.code === 403) {
        localStorage.removeItem('eiot_token')
        if (window.location.hash !== '#/login') router.push('/login')
      }
      ElMessage.error(data.msg || '请求失败')
      return Promise.reject(new Error(data.msg || '请求失败'))
    }
    return data
  },
  (error) => {
    // HTTP 层错误
    const msg = error.response?.data?.msg || error.message || '网络错误'
    ElMessage.error(msg)
    if (error.response?.status === 401 || error.response?.status === 403) {
      localStorage.removeItem('eiot_token')
      if (window.location.hash !== '#/login') router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default request
