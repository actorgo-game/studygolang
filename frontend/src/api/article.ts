import { get, postForm } from './request'
import type { Article, PaginatedData } from '@/types'

export function getArticles(params: { p?: number; tab?: string }) {
  return get<PaginatedData<Article>>('/articles', params)
}

export function getArticleDetail(id: number) {
  return get<{ article: Article }>('/article/detail', { id })
}

export function createArticle(data: Record<string, any>) {
  return postForm('/articles/new', data)
}

export function modifyArticle(data: Record<string, any>) {
  return postForm('/articles/modify', data)
}
