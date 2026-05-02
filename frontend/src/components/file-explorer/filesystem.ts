import { GetHomeDirectory, ListDirectory } from '../../../wailsjs/go/app/App'
import type { app } from '../../../wailsjs/go/models'
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
      throw new Error('Wails runtime is not available')
    }

    return GetHomeDirectory()
  },

  async loadNodes(path: string) {
    if (!isWailsRuntimeAvailable()) {
      throw new Error('Wails runtime is not available')
    }

    const nodes = await ListDirectory(path)
    return nodes.map(mapExplorerNode)
  },

  normalizePath,
  getParentPath,
  getBreadcrumbs,
  isContainer: (node) => Boolean(node.isContainer),
  isHidden: (node) => isHiddenFile(node.name),
}

export const fileSystemExplorerColumns: ExplorerColumn[] = [
  {
    id: 'type',
    label: 'Type',
    widthClass: 'w-28',
    render: (node) => node.type ?? '—',
    sortValue: (node) => node.type ?? '',
  },
  {
    id: 'size',
    label: 'Size',
    widthClass: 'w-32',
    cellClass: 'tabular-nums',
    render: (node) => (node.isContainer ? '—' : formatFileSize(node.size ?? 0)),
    sortValue: (node) => (node.isContainer ? -1 : (node.size ?? 0)),
  },
  {
    id: 'modified',
    label: 'Modified',
    widthClass: 'w-40',
    render: (node) => (node.modified ? formatDate(node.modified) : '—'),
    sortValue: (node) => node.modified?.getTime() ?? 0,
  },
]

export const fileSystemExplorerLabels: Partial<ExplorerLabels> = {
  searchPlaceholder: 'Search files...',
  loading: 'Loading files...',
  empty: 'No files found',
  openContainer: 'Open folder',
  showHidden: 'Show hidden files',
  hideHidden: 'Hide hidden files',
  copyCurrentPath: 'Copy current path',
}
