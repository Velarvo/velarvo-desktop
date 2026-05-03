import { FolderKanban, type Icon as LucideIcon } from 'lucide-svelte'

import type { ApiMethod, SectionKind } from '@/types/sidebar'

const DEFAULT_METHOD_BADGE_STYLE = 'bg-gray-900 text-gray-300'

const methodBadgeStyles: Record<ApiMethod, string> = {
  GET: 'bg-blue-900 text-blue-300',
  POST: 'bg-green-900 text-green-300',
  PUT: 'bg-yellow-900 text-yellow-300',
  DELETE: 'bg-red-900 text-red-300',
  PATCH: 'bg-purple-900 text-purple-300',
  OPTIONS: 'bg-indigo-900 text-indigo-300',
  HEAD: 'bg-gray-900 text-gray-300',
  GRAPHQL: 'bg-pink-900 text-pink-300',
}

export const getSectionIcon = (_kind: SectionKind): typeof LucideIcon => {
  return FolderKanban
}

export const getMethodBadgeStyle = (method?: ApiMethod) => {
  if (!method) {
    return DEFAULT_METHOD_BADGE_STYLE
  }

  return methodBadgeStyles[method] ?? DEFAULT_METHOD_BADGE_STYLE
}
