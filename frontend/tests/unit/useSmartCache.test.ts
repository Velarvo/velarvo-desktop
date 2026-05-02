import { describe, expect, it } from 'vitest'
import { createSmartCache } from '@/components/file-explorer/hooks/useSmartCache'

describe('createSmartCache', () => {
  it('returns fresh entries and tracks cache hits', () => {
    let now = 1000
    const cache = createSmartCache<string>({
      maxSize: 2,
      ttl: 100,
      now: () => now,
    })

    cache.set('documents', 'Documents')

    expect(cache.get('documents')).toBe('Documents')
    expect(cache.getStats()).toMatchObject({
      size: 1,
      hits: 1,
      misses: 0,
    })

    now += 101

    expect(cache.get('documents')).toBeNull()
    expect(cache.getStats()).toMatchObject({
      size: 0,
      misses: 1,
      expirations: 1,
    })
  })

  it('can serve stale data inside the stale window', () => {
    let now = 0
    const cache = createSmartCache<string>({
      maxSize: 2,
      ttl: 100,
      staleTtl: 200,
      now: () => now,
    })

    cache.set('projects', 'Projects')
    now = 150

    expect(cache.get('projects')).toBeNull()

    const staleEntry = cache.getEntry('projects', { allowStale: true })

    expect(staleEntry).toMatchObject({
      data: 'Projects',
      isStale: true,
      age: 150,
      expiresAt: 100,
    })
    expect(cache.getStats()).toMatchObject({
      staleHits: 1,
      misses: 1,
    })
  })

  it('evicts the least recently used entry when full', () => {
    const cache = createSmartCache<string>({
      maxSize: 2,
      ttl: 1000,
    })

    cache.set('a', 'A')
    cache.set('b', 'B')
    cache.get('a')
    cache.set('c', 'C')

    expect(cache.get('a')).toBe('A')
    expect(cache.get('b')).toBeNull()
    expect(cache.get('c')).toBe('C')
    expect(cache.getStats()).toMatchObject({
      evictions: 1,
      size: 2,
    })
  })

  it('clones data on write and read to protect cached values', () => {
    type Item = { name: string }

    const cache = createSmartCache<Item[]>({
      maxSize: 2,
      ttl: 1000,
      clone: (items) => items.map((item) => ({ ...item })),
    })

    const items = [{ name: 'Original' }]
    cache.set('items', items)

    items[0].name = 'Changed outside'

    const cached = cache.get('items')
    expect(cached?.[0].name).toBe('Original')

    if (cached) {
      cached[0].name = 'Changed after read'
    }

    expect(cache.get('items')?.[0].name).toBe('Original')
  })

  it('invalidates global regex patterns deterministically', () => {
    const cache = createSmartCache<string>({
      maxSize: 4,
      ttl: 1000,
    })

    cache.set('root:/documents', 'Documents')
    cache.set('root:/downloads', 'Downloads')
    cache.set('other:/documents', 'Other')

    cache.invalidatePattern(/^root:/g)

    expect(cache.get('root:/documents')).toBeNull()
    expect(cache.get('root:/downloads')).toBeNull()
    expect(cache.get('other:/documents')).toBe('Other')
    expect(cache.getStats().size).toBe(1)
  })
})
