import { get, postForm } from './request'

export function toggleLike(objid: number, data: { objtype: number; flag?: number }) {
  return postForm(`/like/${objid}`, data)
}

export function toggleFavorite(objid: number, data: { objtype: number; collect?: number }) {
  return postForm(`/favorite/${objid}`, data)
}

export function getFavorites(username: string, params: { p?: number }) {
  return get<any>(`/favorites/${username}`, params)
}

export function search(params: { q: string; p?: number }) {
  return get<any>('/search', params)
}

export function uploadImage(file: File) {
  const formData = new FormData()
  formData.append('img', file)
  return postForm('/image/upload', formData as any)
}
