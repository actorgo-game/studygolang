import { get, postForm } from './request'
import type { Message, PaginatedData } from '@/types'

export function getMessages(msgtype: string, params: { p?: number }) {
  return get<PaginatedData<Message>>(`/message/${msgtype}`, params)
}

export function sendMessage(data: { to: number; content: string }) {
  return postForm('/message/send', data)
}

export function deleteMessage(data: { id: number }) {
  return postForm('/message/delete', data)
}
