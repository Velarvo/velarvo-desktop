const STORAGE_KEY = 'velarvo:route'
const DEFAULT_ROUTE = '/dashboard'

const readStoredRoute = (): string => {
  try {
    return window.localStorage.getItem(STORAGE_KEY) || DEFAULT_ROUTE
  } catch {
    return DEFAULT_ROUTE
  }
}

class Router {
  path = $state(readStoredRoute())

  navigate = (path: string) => {
    try {
      window.localStorage.setItem(STORAGE_KEY, path)
    } catch {}
    this.path = path
  }
}

export const router = new Router()
export const navigate = router.navigate
