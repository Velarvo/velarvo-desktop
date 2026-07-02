import { writable } from 'svelte/store'

const STORAGE_KEY = 'velarvo:route'
const DEFAULT_ROUTE = '/dashboard'

const readStoredRoute = (): string => {
  if (typeof window === 'undefined') {
    return DEFAULT_ROUTE
  }

  try {
    return window.localStorage.getItem(STORAGE_KEY) || DEFAULT_ROUTE
  } catch {
    return DEFAULT_ROUTE
  }
}

const persistRoute = (path: string) => {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.setItem(STORAGE_KEY, path)
  } catch {}
}

export const route = writable<string>(readStoredRoute())

export const navigate = (path: string) => {
  persistRoute(path)
  route.set(path)
}
