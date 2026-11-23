import type { Component } from 'svelte'

export type IconComponent = Component<{
  class?: string
  size?: number | string
  strokeWidth?: number | string
}>
