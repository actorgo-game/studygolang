import axios from 'axios'
import type { ApiResponse } from '@/types'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  withCredentials: true,
})

request.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse
    if (data.code !== 0) {
      return Promise.reject(new Error(data.msg || '请求失败'))
    }
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export async function get<T = any>(url: string, params?: any): Promise<T> {
  const resp = await request.get<ApiResponse<T>>(url, { params })
  return resp.data.data
}

export async function post<T = any>(url: string, data?: any): Promise<T> {
  const resp = await request.post<ApiResponse<T>>(url, data)
  return resp.data.data
}

export async function put<T = any>(url: string, data?: any): Promise<T> {
  const resp = await request.put<ApiResponse<T>>(url, data)
  return resp.data.data
}

export async function del<T = any>(url: string, params?: any): Promise<T> {
  const resp = await request.delete<ApiResponse<T>>(url, { params })
  return resp.data.data
}

export async function postForm<T = any>(url: string, data: Record<string, any>): Promise<T> {
  const formData = new URLSearchParams()
  for (const [key, value] of Object.entries(data)) {
    if (value !== undefined && value !== null) {
      formData.append(key, String(value))
    }
  }
  const resp = await request.post<ApiResponse<T>>(url, formData, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  })
  return resp.data.data
}

export default request
