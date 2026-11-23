export const MOTION = {
  layout: 150,
  element: 120,
} as const

export const motion = (ms: number, reduced: boolean): number =>
  reduced ? 0 : ms
