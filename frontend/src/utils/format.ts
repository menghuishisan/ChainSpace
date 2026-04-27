/**
 * 格式化工具函数
 */
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import { DATE_FORMAT } from './constants'

// 启用相对时间插件
dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

/**
 * 格式化日期时间
 */
export function formatDateTime(
  date: string | Date | undefined | null,
  format: string = 'YYYY-MM-DD HH:mm:ss'
): string {
  if (!date) return ''
  return dayjs(date).format(format)
}

/**
 * 格式化相对时间（如：5分钟前）
 */
export function formatRelativeTime(date: string | Date | undefined | null): string {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

/**
 * 格式化日期
 */
export function formatDate(date: string | Date | undefined | null): string {
  return formatDateTime(date, DATE_FORMAT.DATE)
}

/**
 * 格式化时间
 */
export function formatTime(date: string | Date | undefined | null): string {
  return formatDateTime(date, DATE_FORMAT.TIME)
}

/**
 * 格式化时长（秒 -> 时:分:秒）
 */
export function formatDuration(seconds: number | undefined | null): string {
  if (!seconds || seconds < 0) return '00:00:00'
  
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  
  return [hours, minutes, secs]
    .map(v => v.toString().padStart(2, '0'))
    .join(':')
}

/**
 * 格式化时长为中文（分钟 -> X小时X分钟）
 */
export function formatDurationCN(minutes: number | undefined | null): string {
  if (!minutes || minutes <= 0) return '-'
  
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  
  if (hours > 0 && mins > 0) {
    return `${hours}小时${mins}分钟`
  } else if (hours > 0) {
    return `${hours}小时`
  } else {
    return `${mins}分钟`
  }
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number | undefined | null): string {
  if (!bytes || bytes < 0) return '0 B'
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  let size = bytes
  
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index++
  }
  
  return `${size.toFixed(index > 0 ? 2 : 0)} ${units[index]}`
}

/**
 * 格式化数字（千分位）
 */
export function formatNumber(num: number | undefined | null): string {
  if (num === null || num === undefined) return '-'
  return num.toLocaleString('zh-CN')
}

/**
 * 格式化百分比
 */
export function formatPercent(value: number | undefined | null, decimals: number = 0): string {
  if (value === null || value === undefined) return '-'
  return `${value.toFixed(decimals)}%`
}

/**
 * 截断字符串
 */
export function truncateString(str: string | undefined | null, maxLength: number): string {
  if (!str) return ''
  if (str.length <= maxLength) return str
  return str.slice(0, maxLength) + '...'
}

/**
 * 格式化地址（区块链地址缩写）
 */
export function formatAddress(address: string | undefined | null, prefixLen: number = 6, suffixLen: number = 4): string {
  if (!address) return '-'
  if (address.length <= prefixLen + suffixLen) return address
  return `${address.slice(0, prefixLen)}...${address.slice(-suffixLen)}`
}

/**
 * 格式化交易哈希
 */
export function formatTxHash(hash: string | undefined | null): string {
  return formatAddress(hash, 10, 8)
}

/**
 * 解析 Markdown 中的纯文本（去除标记）
 */
export function stripMarkdown(markdown: string | undefined | null): string {
  if (!markdown) return ''
  return markdown
    .replace(/[#*_~`>\[\]()!]/g, '')
    .replace(/\n+/g, ' ')
    .trim()
}
