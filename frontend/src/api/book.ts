import { get, postForm } from './request'
import type { Book, PaginatedData } from '@/types'

export function getBooks(params: { p?: number }) {
  return get<PaginatedData<Book>>('/books', params)
}

export function getBookDetail(id: number) {
  return get<{ book: Book }>(`/book/${id}`)
}

export function createBook(data: Record<string, any>) {
  return postForm('/book/new', data)
}
