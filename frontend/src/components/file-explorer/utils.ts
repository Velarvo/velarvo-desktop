import type { ExplorerBreadcrumb } from './types'

const KB = 1024
const MB = KB * 1024
const GB = MB * 1024
const TB = GB * 1024

export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  if (bytes < KB) return `${bytes} B`
  if (bytes < MB) return `${(bytes / KB).toFixed(1)} KB`
  if (bytes < GB) return `${(bytes / MB).toFixed(1)} MB`
  if (bytes < TB) return `${(bytes / GB).toFixed(2)} GB`
  return `${(bytes / TB).toFixed(2)} TB`
}

export const formatDate = (date: Date): string => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  const formatTime = (value: Date) =>
    value.toLocaleTimeString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
    })

  if (days === 0) {
    if (hours === 0) {
      if (minutes === 0) {
        return 'Just now'
      }
      if (minutes < 60) {
        return `${minutes} min ago`
      }
    }
    return `Today ${formatTime(date)}`
  }

  if (days === 1) {
    return `Yesterday ${formatTime(date)}`
  }

  if (days < 7) {
    const weekday = new Intl.DateTimeFormat(undefined, {
      weekday: 'short',
    }).format(date)
    return `${weekday} ${formatTime(date)}`
  }

  return date.toLocaleDateString(undefined, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

export const isHiddenFile = (filename: string): boolean => {
  return filename.startsWith('.')
}

export const normalizePath = (path: string): string => {
  const normalized = path.replace(/\/+/g, '/').replace(/\/$/, '')
  return normalized || '/'
}

export const getParentPath = (path: string): string => {
  const normalized = normalizePath(path)
  if (normalized === '/') {
    return ''
  }

  const parts = normalized.split('/').filter(Boolean)
  if (parts.length <= 1) {
    return '/'
  }

  return `/${parts.slice(0, -1).join('/')}`
}

export const getBreadcrumbs = (path: string): ExplorerBreadcrumb[] => {
  const normalized = normalizePath(path)
  const parts = normalized.split('/').filter(Boolean)
  const breadcrumbs: ExplorerBreadcrumb[] = [{ name: '/', path: '/' }]

  let accumulatedPath = ''
  for (const part of parts) {
    accumulatedPath += `/${part}`
    breadcrumbs.push({
      name: part,
      path: accumulatedPath,
    })
  }

  return breadcrumbs
}
