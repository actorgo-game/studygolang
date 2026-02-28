import { get } from './request'
import type { SiteStat, FriendLink, Topic, Article, Resource, Project, Comment, User, Reading, TopicNode } from '@/types'

export function getSiteStat() {
  return get<SiteStat>('/websites/stat')
}

export function getRecentTopics() {
  return get<Topic[]>('/topics/recent')
}

export function getRecentArticles() {
  return get<Article[]>('/articles/recent')
}

export function getRecentResources() {
  return get<Resource[]>('/resources/recent')
}

export function getRecentProjects() {
  return get<Project[]>('/projects/recent')
}

export function getRecentComments() {
  return get<Comment[]>('/comments/recent')
}

export function getRecentReadings() {
  return get<Reading[]>('/readings/recent')
}

export function getActiveUsers() {
  return get<User[]>('/users/active')
}

export function getNewestUsers() {
  return get<User[]>('/users/newest')
}

export function getHotNodes() {
  return get<TopicNode[]>('/nodes/hot')
}

export function getViewRank(params: { objtype: number; rank_type: string; limit?: number }) {
  return get<any[]>('/rank/view', params)
}

export function getFriendLinks() {
  return get<FriendLink[]>('/friend/links')
}
