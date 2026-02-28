import { get } from './request'
import type { Reading, PaginatedData } from '@/types'

export function getReadings(params: { p?: number }) {
  return get<PaginatedData<Reading>>('/readings', params)
}

export function getReadingDetail(id: number) {
  return get<{ reading: Reading }>(`/reading/${id}`)
}
