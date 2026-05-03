export const API_METHODS = [
  'GET',
  'POST',
  'PUT',
  'DELETE',
  'PATCH',
  'OPTIONS',
  'HEAD',
  'GRAPHQL',
] as const

export type ApiMethod = (typeof API_METHODS)[number]

export const SIDEBAR_ITEM_STATUSES = [
  'connected',
  'degraded',
  'disconnected',
] as const

export type SidebarItemStatus = (typeof SIDEBAR_ITEM_STATUSES)[number]

export const SIDEBAR_SECTION_KINDS = ['custom'] as const

export type SectionKind = (typeof SIDEBAR_SECTION_KINDS)[number]

export interface SidebarItemDraft {
  name: string
  type: string
  status?: SidebarItemStatus
  method?: ApiMethod
}

export interface SidebarItem extends SidebarItemDraft {
  id: string
}

export interface SidebarSection {
  id: string
  label: string
  kind: SectionKind
  items: SidebarItem[]
}
