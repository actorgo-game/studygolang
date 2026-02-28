import { get, postForm } from './request'
import type { Mission, Gift } from '@/types'

export function getDailyMission() {
  return get<{ missions: Mission[] }>('/mission/daily')
}

export function redeemDailyMission() {
  return get('/mission/daily/redeem')
}

export function completeMission(id: number) {
  return get(`/mission/complete/${id}`)
}

export function getBalance() {
  return get<{ balance: number; records: any[] }>('/balance')
}

export function getGifts() {
  return get<Gift[]>('/gift')
}

export function exchangeGift(data: { gift_id: number }) {
  return postForm('/gift/exchange', data)
}

export function getMyGifts() {
  return get<any[]>('/gift/mine')
}

export function getDauRank() {
  return get<any[]>('/top/dau')
}

export function getRichRank() {
  return get<any[]>('/top/rich')
}
