import { describe, expect, it } from 'vitest'
import {
  getFileExtension,
  getFileIconDescriptor,
} from '@/components/file-explorer/fileIcons'

describe('file icon resolver', () => {
  it('uses a folder icon for containers', () => {
    expect(getFileIconDescriptor('src', true)).toMatchObject({
      key: 'folder',
      label: 'Folder',
    })
  })

  it('matches technology icons by exact filenames', () => {
    expect(getFileIconDescriptor('package.json')).toMatchObject({ key: 'npm' })
    expect(getFileIconDescriptor('Dockerfile')).toMatchObject({ key: 'docker' })
    expect(getFileIconDescriptor('Dockerfile.dev')).toMatchObject({
      key: 'docker',
    })
    expect(getFileIconDescriptor('.env.example')).toMatchObject({ key: 'key' })
    expect(getFileIconDescriptor('vite.config.ts')).toMatchObject({
      key: 'vite',
    })
    expect(getFileIconDescriptor('.gitignore')).toMatchObject({ key: 'git' })
  })

  it('matches technology icons by extension', () => {
    expect(getFileIconDescriptor('App.svelte')).toMatchObject({ key: 'svelte' })
    expect(getFileIconDescriptor('main.go')).toMatchObject({ key: 'go' })
    expect(getFileIconDescriptor('Component.tsx')).toMatchObject({
      key: 'react',
    })
    expect(getFileIconDescriptor('styles.scss')).toMatchObject({ key: 'sass' })
  })

  it('extracts extensions without treating dotfiles as extensions', () => {
    expect(getFileExtension('README.md')).toBe('md')
    expect(getFileExtension('.env')).toBe('')
    expect(getFileExtension('Dockerfile')).toBe('')
  })

  it('falls back to a generic file icon', () => {
    expect(getFileIconDescriptor('unknown.custom-ext')).toMatchObject({
      key: 'file',
      label: 'File',
    })
  })
})
