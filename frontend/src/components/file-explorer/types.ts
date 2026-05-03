export interface ExplorerNode {
  id: string
  name: string
  path: string
  type?: string
  isContainer?: boolean
  size?: number
  modified?: Date
}

export interface ExplorerBreadcrumb {
  name: string
  path: string
}

export type ExplorerColumnBreakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl'

export interface ExplorerColumn {
  id: string
  label: string
  widthClass?: string
  cellClass?: string
  sortable?: boolean
  hideBelow?: ExplorerColumnBreakpoint
  render: (node: ExplorerNode) => string
  sortValue?: (node: ExplorerNode) => string | number | null | undefined
}

export interface ExplorerDataSource {
  getInitialPath: () => Promise<string>
  loadNodes: (path: string) => Promise<ExplorerNode[]>
  normalizePath?: (path: string) => string
  getParentPath?: (path: string) => string
  getBreadcrumbs?: (path: string) => ExplorerBreadcrumb[]
  getCacheKey?: (path: string) => string
  getCopyValue?: (path: string) => string
  isContainer?: (node: ExplorerNode) => boolean
  isHidden?: (node: ExplorerNode) => boolean
}

export interface ExplorerLabels {
  nameColumn: string
  searchPlaceholder: string
  sortBy: string
  ascending: string
  descending: string
  hidden: string
  showHidden: string
  hideHidden: string
  copy: string
  copied: string
  copyCurrentPath: string
  loading: string
  empty: string
  openContainer: string
}

export const DEFAULT_EXPLORER_LABELS: ExplorerLabels = {
  nameColumn: 'Name',
  searchPlaceholder: 'Search...',
  sortBy: 'Sort by',
  ascending: 'Ascending order',
  descending: 'Descending order',
  hidden: 'Hidden',
  showHidden: 'Show hidden items',
  hideHidden: 'Hide hidden items',
  copy: 'Copy',
  copied: 'Copied',
  copyCurrentPath: 'Copy current path',
  loading: 'Loading...',
  empty: 'No items found',
  openContainer: 'Open item',
}

export type SortOrder = 'asc' | 'desc'
