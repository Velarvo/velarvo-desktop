import { GetHomeDirectory, ListDirectory } from '../../../wailsjs/go/app/App'
import type { app } from '../../../wailsjs/go/models'
import { getResponseMessage } from '@/lib/errors'
import { t } from '@/lib/i18n'
import type {
  ExplorerColumn,
  ExplorerDataSource,
  ExplorerLabels,
  ExplorerNode,
} from './types'
import {
  formatDate,
  formatFileSize,
  getBreadcrumbs,
  getParentPath,
  isHiddenFile,
  normalizePath,
} from './utils'

const isWailsRuntimeAvailable = () => {
  if (typeof window === 'undefined') {
    return false
  }

  return Boolean((window as Window & { go?: unknown }).go)
}

const mapExplorerNode = (node: app.ExplorerNode): ExplorerNode => ({
  id: node.id,
  name: node.name,
  path: node.path,
  type: node.type,
  isContainer: node.type === 'directory',
  size: node.size,
  modified: node.modified ? new Date(node.modified) : undefined,
})

export const fileSystemExplorerDataSource: ExplorerDataSource = {
  async getInitialPath() {
    if (!isWailsRuntimeAvailable()) {
      throw new Error(t('errors.FILESYSTEM_UNAVAILABLE'))
    }

    const resp = await GetHomeDirectory()
    if (!resp.success || !resp.data) {
      throw new Error(
        getResponseMessage(resp, t('errors.FILESYSTEM_READ_HOME_FAILED')),
      )
    }

    return resp.data
  },

  async loadNodes(path: string) {
    if (!isWailsRuntimeAvailable()) {
      throw new Error(t('errors.FILESYSTEM_UNAVAILABLE'))
    }

    const resp = await ListDirectory(path)
    if (!resp.success || !resp.data) {
      throw new Error(
        getResponseMessage(resp, t('errors.FILESYSTEM_LIST_DIRECTORY_FAILED')),
      )
    }

    return resp.data.map(mapExplorerNode)
  },

  normalizePath,
  getParentPath,
  getBreadcrumbs,
  isContainer: (node) => Boolean(node.isContainer),
  isHidden: (node) => isHiddenFile(node.name),
}

export const getFileSystemExplorerColumns = (): ExplorerColumn[] => [
  {
    id: 'type',
    label: t('explorer.columns.type'),
    widthClass: 'w-28',
    hideBelow: 'lg',
    render: (node) => node.type ?? '—',
    sortValue: (node) => node.type ?? '',
  },
  {
    id: 'size',
    label: t('explorer.columns.size'),
    widthClass: 'w-32',
    cellClass: 'tabular-nums',
    hideBelow: 'md',
    render: (node) => (node.isContainer ? '—' : formatFileSize(node.size ?? 0)),
    sortValue: (node) => (node.isContainer ? -1 : (node.size ?? 0)),
  },
  {
    id: 'modified',
    label: t('explorer.columns.modified'),
    widthClass: 'w-40',
    hideBelow: 'sm',
    render: (node) => (node.modified ? formatDate(node.modified) : '—'),
    sortValue: (node) => node.modified?.getTime() ?? 0,
  },
]

export const getFileSystemExplorerLabels = (): Partial<ExplorerLabels> => ({
  nameColumn: t('explorer.columns.name'),
  searchPlaceholder: t('explorer.searchPlaceholder'),
  sortBy: t('explorer.sortBy'),
  ascending: t('explorer.ascending'),
  descending: t('explorer.descending'),
  hidden: t('explorer.hidden'),
  loading: t('explorer.loading'),
  empty: t('explorer.empty'),
  openContainer: t('explorer.openContainer'),
  showHidden: t('explorer.showHidden'),
  hideHidden: t('explorer.hideHidden'),
  copy: t('explorer.copy'),
  copied: t('explorer.copied'),
  copyCurrentPath: t('explorer.copyCurrentPath'),
})
