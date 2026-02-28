import { get, postForm } from './request'
import type { Topic, TopicNode, PaginatedData } from '@/types'

export function getTopics(params: { p?: number; tab?: string }) {
  return get<PaginatedData<Topic>>('/topics', params)
}

export function getTopicDetail(tid: number) {
  return get<{ topic: Topic }>('/topic/detail', { tid })
}

export function getNoReplyTopics(params: { p?: number }) {
  return get<PaginatedData<Topic>>('/topics/no_reply', params)
}

export function getLastTopics(params: { p?: number }) {
  return get<PaginatedData<Topic>>('/topics/last', params)
}

export function getNodeTopics(nid: number, params: { p?: number }) {
  return get<PaginatedData<Topic>>(`/topics/node/${nid}`, params)
}

export function createTopic(data: { title: string; content: string; nid: number; tags?: string }) {
  return postForm('/topics/new', data)
}

export function modifyTopic(data: { tid: number; title: string; content: string; nid: number; tags?: string }) {
  return postForm('/topics/modify', data)
}

export function getNodes() {
  return get<TopicNode[]>('/nodes')
}

export function setTopicTop(data: { tid: number; top: number }) {
  return postForm('/topic/set_top', data)
}
