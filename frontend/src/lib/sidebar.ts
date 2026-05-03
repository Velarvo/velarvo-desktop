import {
  API_METHODS,
  SIDEBAR_ITEM_STATUSES,
  type ApiMethod,
  type SectionKind,
  type SidebarItemStatus,
} from '@/types/sidebar'

export const DEFAULT_API_METHOD: ApiMethod = 'GET'
export const DEFAULT_GRAPHQL_METHOD: ApiMethod = 'GRAPHQL'
export const DEFAULT_SECTION_KIND: SectionKind = 'custom'

export const SIDEBAR_API_ITEM_TYPES = ['API', 'REST', 'GRAPHQL'] as const

const apiMethodSet = new Set<string>(API_METHODS)
const sidebarItemStatusSet = new Set<string>(SIDEBAR_ITEM_STATUSES)
const sidebarApiItemTypeSet = new Set<string>(SIDEBAR_API_ITEM_TYPES)

export const isApiMethod = (value: unknown): value is ApiMethod => {
  return typeof value === 'string' && apiMethodSet.has(value)
}

export const isSidebarItemStatus = (
  value: unknown,
): value is SidebarItemStatus => {
  return typeof value === 'string' && sidebarItemStatusSet.has(value)
}

export const isApiItem = (type?: string, method?: ApiMethod) => {
  if (method) return true
  if (!type) return false

  return sidebarApiItemTypeSet.has(type.toUpperCase())
}
