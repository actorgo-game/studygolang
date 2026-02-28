import { get, postForm } from './request'
import type { Me, User, PaginatedData, Topic, Article, Resource, Project, Comment } from '@/types'

export function getCurrentUser() {
  return get<Me>('/user/current')
}

export function login(data: { username: string; passwd: string; remember_me?: string }) {
  return postForm('/account/login', data)
}

export function register(data: { username: string; passwd: string; email: string; captcha?: string }) {
  return postForm('/account/register', data)
}

export function logout() {
  return get('/account/logout')
}

export function getUserProfile(username: string) {
  return get<{ user: User }>(`/user/${username}`)
}

export function getUserTopics(username: string, params: { p?: number }) {
  return get<PaginatedData<Topic>>(`/user/${username}/topics`, params)
}

export function getUserArticles(username: string, params: { p?: number }) {
  return get<PaginatedData<Article>>(`/user/${username}/articles`, params)
}

export function getUserResources(username: string, params: { p?: number }) {
  return get<PaginatedData<Resource>>(`/user/${username}/resources`, params)
}

export function getUserProjects(username: string, params: { p?: number }) {
  return get<PaginatedData<Project>>(`/user/${username}/projects`, params)
}

export function getUserComments(username: string, params: { p?: number }) {
  return get<PaginatedData<Comment>>(`/user/${username}/comments`, params)
}

export function modifyUser(data: Record<string, any>) {
  return postForm('/user/modify', data)
}

export function changePassword(data: { cur_passwd: string; passwd: string }) {
  return postForm('/account/changepwd', data)
}

export function changeAvatar(file: File) {
  const formData = new FormData()
  formData.append('avatar', file)
  return postForm('/account/change_avatar', formData as any)
}

export function getUsers(params: { p?: number }) {
  return get<PaginatedData<User>>('/users', params)
}
