import { writable } from 'svelte/store'

export const route = writable(window.location.pathname || '/')

window.addEventListener('popstate', () => {
  route.set(window.location.pathname || '/')
})

export const navigate = (path: string) => {
  if (window.location.pathname === path) {
    return
  }

  window.history.pushState({}, '', path)
  route.set(path)
}
