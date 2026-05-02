export type ApiMethod =
  | 'GET'
  | 'POST'
  | 'PUT'
  | 'DELETE'
  | 'PATCH'
  | 'OPTIONS'
  | 'HEAD'
  | 'GRAPHQL'

export type SectionKind = 'custom'

export interface SidebarItem {
  id: string
  name: string
  type: string
  status?: 'connected' | 'degraded' | 'disconnected'
  method?: ApiMethod
}

export interface SidebarSection {
  id: string
  label: string
  kind: SectionKind
  items: SidebarItem[]
}
