import { FolderKanban, type Icon as LucideIcon } from 'lucide-svelte'

import type { ApiMethod, SectionKind } from './types'

export const getSectionIcon = (_kind: SectionKind): typeof LucideIcon => {
  return FolderKanban
}

export const getMethodBadgeStyle = (method?: ApiMethod) => {
  if (!method) {
    return 'bg-gray-900 text-gray-300'
  }

  switch (method.toUpperCase()) {
    case 'GET':
      return 'bg-blue-900 text-blue-300'
    case 'POST':
      return 'bg-green-900 text-green-300'
    case 'PUT':
      return 'bg-yellow-900 text-yellow-300'
    case 'DELETE':
      return 'bg-red-900 text-red-300'
    case 'PATCH':
      return 'bg-purple-900 text-purple-300'
    case 'OPTIONS':
      return 'bg-indigo-900 text-indigo-300'
    case 'HEAD':
      return 'bg-gray-900 text-gray-300'
    case 'GRAPHQL':
      return 'bg-pink-900 text-pink-300'
    default:
      return 'bg-gray-900 text-gray-300'
  }
}
