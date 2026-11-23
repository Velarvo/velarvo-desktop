import { Service as Files } from '@bindings/files'

const separatorOf = (path: string): string => (path.includes('\\') ? '\\' : '/')

const stripTrailingSeparators = (path: string): string =>
  path.replace(/[\\/]+$/, '')

const lastSeparatorIndex = (path: string): number =>
  Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'))

export const localJoin = (base: string, name: string): string => {
  if (!base) return name
  const separator = separatorOf(base)
  return base.endsWith(separator)
    ? `${base}${name}`
    : `${base}${separator}${name}`
}

export const localBaseName = (path: string): string => {
  const trimmed = stripTrailingSeparators(path)
  return trimmed.slice(lastSeparatorIndex(trimmed) + 1)
}

export const localParent = (path: string): string => {
  const trimmed = stripTrailingSeparators(path)
  const index = lastSeparatorIndex(trimmed)
  if (index < 0) return trimmed
  return index === 0 ? trimmed.slice(0, 1) : trimmed.slice(0, index)
}

export const renameLocalPath = (from: string, to: string) =>
  Files.Rename(from, to)

export const removeLocalPath = (path: string) => Files.Remove(path)
