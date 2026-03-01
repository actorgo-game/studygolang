function parseDate(dateStr: string): Date {
  if (!dateStr) return new Date(NaN)
  // Backend returns "YYYY-MM-DD HH:mm:ss"; replace space with T for reliable parsing
  const iso = dateStr.includes('T') ? dateStr : dateStr.replace(' ', 'T')
  return new Date(iso)
}

export function timeAgo(dateStr: string): string {
  const date = parseDate(dateStr)
  if (isNaN(date.getTime()) || date.getFullYear() < 2000) return ''

  const now = new Date()
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000)
  if (seconds < 0) return '刚刚'

  if (seconds < 60) return '刚刚'
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟前`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} 小时前`
  if (seconds < 604800) return `${Math.floor(seconds / 86400)} 天前`
  if (seconds < 2592000) return `${Math.floor(seconds / 604800)} 周前`
  if (seconds < 31536000) return `${Math.floor(seconds / 2592000)} 个月前`
  return `${Math.floor(seconds / 31536000)} 年前`
}

export function formatDate(dateStr: string, fmt = 'YYYY-MM-DD'): string {
  const d = parseDate(dateStr)
  const map: Record<string, string> = {
    YYYY: String(d.getFullYear()),
    MM: String(d.getMonth() + 1).padStart(2, '0'),
    DD: String(d.getDate()).padStart(2, '0'),
    HH: String(d.getHours()).padStart(2, '0'),
    mm: String(d.getMinutes()).padStart(2, '0'),
    ss: String(d.getSeconds()).padStart(2, '0'),
  }
  let result = fmt
  for (const [key, value] of Object.entries(map)) {
    result = result.replace(key, value)
  }
  return result
}
