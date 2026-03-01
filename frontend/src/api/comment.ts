import { get, postForm } from './request'
import type { Comment, PaginatedData } from '@/types'

export function getComments(params: { objid: number; objtype: number; p?: number }) {
  return get<PaginatedData<Comment>>('/object/comments', params)
}

export function createComment(objid: number, data: { objtype: number; content: string }) {
  return postForm(`/comment/${objid}`, data)
}

export function modifyComment(cid: number, data: { content: string }) {
  return postForm(`/object/comments/${cid}`, data)
}

export function deleteComment(cid: number) {
  return postForm('/comment/delete', { cid })
}

export function getAtUsers(params: { term: string }) {
  return get<string[]>('/at/users', params)
}
