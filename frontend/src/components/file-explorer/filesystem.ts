import { Service as Files, type Entry } from '@bindings/files'
import { t } from '@/lib/i18n.svelte'
import type { ExplorerColumn, ExplorerDataSource, ExplorerNode } from './types'
import {
  formatDate,
  formatFileSize,
  getBreadcrumbs,
  getParentPath,
  isHiddenFile,
  normalizePath,
} from './utils'

const mapExplorerNode = (node: Entry): ExplorerNode => ({
  id: node.path,
  name: node.name,
  path: node.path,
  type: node.type,
  isContainer: node.type === 'directory',
  size: node.size,
  modified: node.modified ? new Date(node.modified) : undefined,
})

export const fileSystemExplorerDataSource: ExplorerDataSource = {
  getInitialPath: () => Files.Home(),

  async loadNodes(path: string) {
    const entries = await Files.List(path)
    return (entries ?? []).map(mapExplorerNode)
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
    widthClass: 'w-24',
    hideBelow: 'md',
    render: (node) => node.type ?? '—',
    sortValue: (node) => node.type ?? '',
  },
  {
    id: 'size',
    label: t('explorer.columns.size'),
    widthClass: 'w-20',
    cellClass: 'tabular-nums',
    hideBelow: 'sm',
    render: (node) => (node.isContainer ? '—' : formatFileSize(node.size ?? 0)),
    sortValue: (node) => (node.isContainer ? -1 : (node.size ?? 0)),
  },
  {
    id: 'modified',
    label: t('explorer.columns.modified'),
    widthClass: 'w-36',
    hideBelow: 'xs',
    render: (node) => (node.modified ? formatDate(node.modified) : '—'),
    sortValue: (node) => node.modified?.getTime() ?? 0,
  },
]
