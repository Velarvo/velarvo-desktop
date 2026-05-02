interface CacheEntry<T> {
  data: T
  createdAt: number
  expiresAt: number
  staleUntil: number
  accessCount: number
  lastAccessed: number
}

export type CacheEvictionReason = 'expired' | 'lru' | 'manual' | 'clear'

export interface CacheOptions<T> {
  maxSize: number
  ttl: number
  staleTtl?: number
  clone?: (data: T) => T
  now?: () => number
  onEvict?: (key: string, reason: CacheEvictionReason) => void
}

export interface CacheStats {
  size: number
  maxSize: number
  hits: number
  staleHits: number
  misses: number
  evictions: number
  expirations: number
  writes: number
}

export interface CacheHit<T> {
  data: T
  isStale: boolean
  age: number
  expiresAt: number
}

interface CacheReadOptions {
  allowStale?: boolean
}

const identity = <T>(data: T) => data

const normalizeOptions = <T>(options: CacheOptions<T>) => {
  if (!Number.isFinite(options.maxSize) || options.maxSize < 1) {
    throw new Error('Cache maxSize must be greater than zero')
  }

  if (!Number.isFinite(options.ttl) || options.ttl < 0) {
    throw new Error('Cache ttl must be a non-negative number')
  }

  if (
    options.staleTtl !== undefined &&
    (!Number.isFinite(options.staleTtl) || options.staleTtl < 0)
  ) {
    throw new Error('Cache staleTtl must be a non-negative number')
  }

  return {
    ...options,
    staleTtl: options.staleTtl ?? 0,
    clone: options.clone ?? identity,
    now: options.now ?? Date.now,
  }
}

export const createSmartCache = <T>(rawOptions: CacheOptions<T>) => {
  const options = normalizeOptions(rawOptions)
  const cache = new Map<string, CacheEntry<T>>()

  const stats: CacheStats = {
    size: 0,
    maxSize: options.maxSize,
    hits: 0,
    staleHits: 0,
    misses: 0,
    evictions: 0,
    expirations: 0,
    writes: 0,
  }

  const syncSize = () => {
    stats.size = cache.size
  }

  const deleteEntry = (key: string, reason: CacheEvictionReason) => {
    const deleted = cache.delete(key)
    if (!deleted) {
      return false
    }

    if (reason === 'lru') {
      stats.evictions += 1
    } else if (reason === 'expired') {
      stats.expirations += 1
    }

    options.onEvict?.(key, reason)
    syncSize()
    return true
  }

  const evictOldest = () => {
    if (cache.size === 0) {
      return
    }

    const oldestKey = cache.keys().next().value as string | undefined
    if (oldestKey !== undefined) {
      deleteEntry(oldestKey, 'lru')
    }
  }

  const touch = (key: string, entry: CacheEntry<T>, now: number) => {
    entry.accessCount += 1
    entry.lastAccessed = now

    cache.delete(key)
    cache.set(key, entry)
  }

  const prune = () => {
    const now = options.now()
    for (const [key, entry] of cache.entries()) {
      if (now > entry.staleUntil) {
        deleteEntry(key, 'expired')
      }
    }
    syncSize()
  }

  const getEntry = (
    key: string,
    readOptions: CacheReadOptions = {},
  ): CacheHit<T> | null => {
    const entry = cache.get(key)
    const now = options.now()

    if (!entry) {
      stats.misses += 1
      return null
    }

    const isFresh = now <= entry.expiresAt
    const isStale = !isFresh && now <= entry.staleUntil

    if (!isFresh && (!readOptions.allowStale || !isStale)) {
      if (now > entry.staleUntil) {
        deleteEntry(key, 'expired')
      }
      stats.misses += 1
      return null
    }

    touch(key, entry, now)

    if (isStale) {
      stats.staleHits += 1
    } else {
      stats.hits += 1
    }

    return {
      data: options.clone(entry.data),
      isStale,
      age: now - entry.createdAt,
      expiresAt: entry.expiresAt,
    }
  }

  const get = (key: string, readOptions?: CacheReadOptions): T | null => {
    return getEntry(key, readOptions)?.data ?? null
  }

  const set = (key: string, data: T) => {
    prune()

    if (cache.has(key)) {
      cache.delete(key)
    }

    while (cache.size >= options.maxSize) {
      evictOldest()
    }

    const now = options.now()
    cache.set(key, {
      data: options.clone(data),
      createdAt: now,
      expiresAt: now + options.ttl,
      staleUntil: now + options.ttl + options.staleTtl,
      accessCount: 0,
      lastAccessed: now,
    })

    stats.writes += 1
    syncSize()
  }

  const invalidate = (key: string) => {
    deleteEntry(key, 'manual')
  }

  const invalidatePattern = (pattern: RegExp) => {
    for (const key of cache.keys()) {
      pattern.lastIndex = 0
      if (pattern.test(key)) {
        deleteEntry(key, 'manual')
      }
    }
    syncSize()
  }

  const clear = () => {
    if (cache.size > 0) {
      for (const key of cache.keys()) {
        options.onEvict?.(key, 'clear')
      }
    }
    cache.clear()
    syncSize()
  }

  const getStats = (): CacheStats => {
    return { ...stats }
  }

  return {
    get,
    getEntry,
    set,
    invalidate,
    invalidatePattern,
    clear,
    prune,
    stats,
    getStats,
  }
}
