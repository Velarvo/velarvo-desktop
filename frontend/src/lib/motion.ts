import { readable } from 'svelte/store'

export const prefersReducedMotion = readable(false, (set) => {
  if (typeof window === 'undefined' || !window.matchMedia) {
    return
  }

  const mql = window.matchMedia('(prefers-reduced-motion: reduce)')
  set(mql.matches)

  const onChange = () => set(mql.matches)
  mql.addEventListener('change', onChange)
  return () => mql.removeEventListener('change', onChange)
})

export const MOTION = {
  layout: 150,
  element: 120,
} as const

export const motion = (ms: number, reduced: boolean): number =>
  reduced ? 0 : ms
