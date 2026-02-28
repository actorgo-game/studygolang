import { get, postForm } from './request'
import type { Wiki, PaginatedData } from '@/types'

export function getWikiList(params: { p?: number }) {
  return get<PaginatedData<Wiki>>('/wiki', params)
}

export function getWikiDetail(uri: string) {
  return get<{ wiki: Wiki }>(`/wiki/${uri}`)
}

export function createWiki(data: Record<string, any>) {
  return postForm('/wiki/new', data)
}

export function modifyWiki(data: Record<string, any>) {
  return postForm('/wiki/modify', data)
}
