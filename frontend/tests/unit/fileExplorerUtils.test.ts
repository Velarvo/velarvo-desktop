import { describe, expect, it } from 'vitest'
import {
  formatFileSize,
  getBreadcrumbs,
  getParentPath,
  isHiddenFile,
  normalizePath,
} from '@/components/file-explorer/utils'

describe('file explorer utils', () => {
  it('normalizes paths consistently', () => {
    expect(normalizePath('')).toBe('/')
    expect(normalizePath('/')).toBe('/')
    expect(normalizePath('/fixture-root//workspace///project/')).toBe(
      '/fixture-root/workspace/project',
    )
  })

  it('returns parent paths without escaping root', () => {
    expect(getParentPath('/')).toBe('')
    expect(getParentPath('/fixture-root')).toBe('/')
    expect(getParentPath('/fixture-root/workspace/project')).toBe(
      '/fixture-root/workspace',
    )
  })

  it('builds breadcrumbs from root to the selected path', () => {
    expect(getBreadcrumbs('/fixture-root/workspace/project')).toEqual([
      { name: '/', path: '/' },
      { name: 'fixture-root', path: '/fixture-root' },
      { name: 'workspace', path: '/fixture-root/workspace' },
      { name: 'project', path: '/fixture-root/workspace/project' },
    ])
  })

  it('formats file sizes with stable units', () => {
    expect(formatFileSize(0)).toBe('0 B')
    expect(formatFileSize(512)).toBe('512 B')
    expect(formatFileSize(1536)).toBe('1.5 KB')
    expect(formatFileSize(1024 * 1024 * 3)).toBe('3.0 MB')
  })

  it('detects hidden dotfiles', () => {
    expect(isHiddenFile('.env')).toBe(true)
    expect(isHiddenFile('README.md')).toBe(false)
  })
})
