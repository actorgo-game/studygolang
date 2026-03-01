import { get, postForm } from './request'
import type { Resource, PaginatedData } from '@/types'

export function getResources(params: { p?: number; catid?: number }) {
  return get<PaginatedData<Resource>>('/resources', params)
}

export function getResourceDetail(id: number) {
  return get<{ resource: Resource }>('/resource/detail', { id })
}

export function createResource(data: Record<string, any>) {
  return postForm('/resources/new', data)
}

export function modifyResource(data: Record<string, any>) {
  return postForm('/resources/modify', data)
}

export function deleteResource(id: number) {
  return postForm('/resources/delete', { id })
}
