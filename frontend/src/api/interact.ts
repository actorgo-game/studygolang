import { get, postForm } from './request'

export function toggleLike(objid: number, data: { objtype: number }) {
  return postForm<{ liked: boolean }>(`/like/${objid}`, data)
}

export function hadLike(objid: number, objtype: number) {
  return get<{ liked: boolean }>(`/like/${objid}`, { objtype })
}

export function toggleFavorite(objid: number, data: { objtype: number }) {
  return postForm<{ favorited: boolean }>(`/favorite/${objid}`, data)
}

export function hadFavorite(objid: number, objtype: number) {
  return get<{ favorited: boolean }>(`/favorite/${objid}`, { objtype })
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
