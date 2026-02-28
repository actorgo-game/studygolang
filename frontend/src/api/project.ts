import { get, postForm } from './request'
import type { Project, PaginatedData } from '@/types'

export function getProjects(params: { p?: number }) {
  return get<PaginatedData<Project>>('/projects', params)
}

export function getProjectDetail(uri: string) {
  return get<{ project: Project }>('/project/detail', { uri })
}

export function createProject(data: Record<string, any>) {
  return postForm('/project/new', data)
}

export function modifyProject(data: Record<string, any>) {
  return postForm('/project/modify', data)
}
