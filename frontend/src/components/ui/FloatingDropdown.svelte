<script lang="ts" module>
  export type DropdownSide = 'top' | 'bottom' | 'left' | 'right'
  export type DropdownAlign = 'start' | 'end' | 'center'
  export type DropdownPlacement =
    DropdownSide | `${DropdownSide}-start` | `${DropdownSide}-end`
</script>

<script lang="ts">
  import type { Attachment } from 'svelte/attachments'
  import type { Snippet } from 'svelte'
  import { onMount, tick, untrack } from 'svelte'
  import { on } from 'svelte/events'
  import { cubicOut } from 'svelte/easing'
  import { prefersReducedMotion } from 'svelte/motion'
  import { MOTION, motion } from '@/lib/motion'

  interface Props {
    open?: boolean
    x?: number | null
    y?: number | null
    anchorElement?: HTMLElement | null
    triggerElement?: HTMLElement | null
    placement?: DropdownPlacement
    offset?: number
    flip?: boolean
    fitViewport?: boolean
    viewportMargin?: number
    panelClass?: string
    group?: string
    closeOnOutsideClick?: boolean
    closeOnEscape?: boolean
    matchAnchorWidth?: boolean
    children?: Snippet
    onClose?: () => void
  }

  let {
    open = false,
    x = null,
    y = null,
    anchorElement = null,
    triggerElement = null,
    placement = 'bottom-start',
    offset = 8,
    flip = true,
    fitViewport = true,
    viewportMargin = 8,
    panelClass = '',
    group = 'global',
    closeOnOutsideClick = true,
    closeOnEscape = true,
    matchAnchorWidth = false,
    children,
    onClose,
  }: Props = $props()

  const DROPDOWN_OPEN_EVENT = 'floating-dropdown-opened'
  const OPPOSITE: Record<DropdownSide, DropdownSide> = {
    top: 'bottom',
    bottom: 'top',
    left: 'right',
    right: 'left',
  }

  let dropdownElement = $state<HTMLDivElement | null>(null)
  let wasOpen = $state(false)
  let resolvedSide = $state<DropdownSide>('bottom')
  let frame = 0
  const dropdownId = `dropdown-${Math.random().toString(36).slice(2, 10)}`

  const parsePlacement = (
    value: DropdownPlacement,
  ): { side: DropdownSide; align: DropdownAlign } => {
    const [side, align] = value.split('-') as [DropdownSide, DropdownAlign?]
    return { side, align: align ?? 'center' }
  }

  const isVertical = (side: DropdownSide) => side === 'top' || side === 'bottom'

  const clamp = (value: number, min: number, max: number) =>
    Math.min(Math.max(value, min), Math.max(min, max))

  const alignOnAxis = (
    align: DropdownAlign,
    refStart: number,
    refEnd: number,
    size: number,
    viewportSize: number,
    margin: number,
  ) => {
    const startAligned = refStart
    const endAligned = refEnd - size
    const fits = (value: number) =>
      value >= margin && value + size <= viewportSize - margin

    let value =
      align === 'start'
        ? startAligned
        : align === 'end'
          ? endAligned
          : refStart + (refEnd - refStart) / 2 - size / 2

    if (!fits(value) && align !== 'center') {
      const mirrored = align === 'start' ? endAligned : startAligned
      if (fits(mirrored)) value = mirrored
    }

    return clamp(value, margin, viewportSize - size - margin)
  }

  const portalToBody: Attachment<HTMLElement> = (node) => {
    document.body.appendChild(node)
    untrack(() => calculatePosition(node as HTMLDivElement))
    return () => node.remove()
  }

  const requestClose = () => {
    onClose?.()
  }

  const calculatePosition = (node: HTMLDivElement | null = dropdownElement) => {
    if (!node || typeof window === 'undefined') return

    const margin = viewportMargin
    const viewportWidth =
      document.documentElement.clientWidth || window.innerWidth
    const viewportHeight =
      document.documentElement.clientHeight || window.innerHeight
    const anchorRect = anchorElement?.getBoundingClientRect() ?? null
    const ref = anchorRect ?? new DOMRect(x ?? 0, y ?? 0, 0, 0)
    const gap = anchorRect ? offset : 0

    node.style.maxWidth = `${Math.max(160, viewportWidth - margin * 2)}px`
    node.style.width =
      matchAnchorWidth && anchorRect
        ? `${Math.round(Math.min(anchorRect.width, viewportWidth - margin * 2))}px`
        : ''

    node.style.maxHeight = ''
    node.style.overflowY = ''
    const { width, height } = node.getBoundingClientRect()

    const space: Record<DropdownSide, number> = {
      top: ref.top - gap - margin,
      bottom: viewportHeight - ref.bottom - gap - margin,
      left: ref.left - gap - margin,
      right: viewportWidth - ref.right - gap - margin,
    }
    const fitsOn = (side: DropdownSide) =>
      space[side] >= (isVertical(side) ? height : width)

    const { side: preferred, align } = parsePlacement(placement)
    const opposite = OPPOSITE[preferred]

    let side = preferred
    if (flip && !fitsOn(preferred)) {
      const perpendicular = (
        isVertical(preferred)
          ? (['right', 'left'] as const)
          : (['bottom', 'top'] as const)
      )
        .slice()
        .sort((a, b) => space[b] - space[a])
      const candidates: DropdownSide[] = [opposite, ...perpendicular]
      const fitting = candidates.find(fitsOn)

      if (fitting) {
        side = fitting
      } else {
        side = space.top > space.bottom ? 'top' : 'bottom'
        if (
          !isVertical(preferred) &&
          space[preferred] >= width &&
          space[preferred] >= space[opposite]
        ) {
          side = preferred
        }
      }
    }

    const maxHeight = isVertical(side)
      ? Math.max(0, Math.floor(space[side]))
      : Math.max(0, viewportHeight - margin * 2)
    const panelHeight = fitViewport ? Math.min(height, maxHeight) : height

    if (fitViewport && panelHeight < height) {
      node.style.maxHeight = `${maxHeight}px`
      node.style.overflowY = 'auto'
    }

    let left: number
    let top: number

    if (isVertical(side)) {
      top = side === 'bottom' ? ref.bottom + gap : ref.top - gap - panelHeight
      left = alignOnAxis(
        align,
        ref.left,
        ref.right,
        width,
        viewportWidth,
        margin,
      )
    } else {
      left = side === 'right' ? ref.right + gap : ref.left - gap - width
      left = clamp(left, margin, viewportWidth - width - margin)
      top = alignOnAxis(
        align,
        ref.top,
        ref.bottom,
        panelHeight,
        viewportHeight,
        margin,
      )
    }

    top = clamp(top, margin, viewportHeight - panelHeight - margin)

    node.style.left = `${Math.round(left)}px`
    node.style.top = `${Math.round(top)}px`
    node.style.visibility = ''
    node.dataset.side = side

    resolvedSide = side
  }

  const scheduleFrame = () => {
    if (frame) cancelAnimationFrame(frame)
    frame = requestAnimationFrame(() => {
      frame = 0
      if (open) calculatePosition()
    })
  }

  const handleDocumentMouseDown = (event: MouseEvent) => {
    if (!open || !closeOnOutsideClick || !dropdownElement) return

    const target = event.target as Node | null
    if (!target) return
    if (dropdownElement.contains(target)) return

    const trigger = triggerElement ?? anchorElement
    if (trigger?.contains(target)) return

    requestClose()
  }

  const handleDocumentKeydown = (event: KeyboardEvent) => {
    if (!open || !closeOnEscape) return
    if (event.key !== 'Escape') return

    requestClose()
  }

  const handleDropdownOpened = (event: Event) => {
    const customEvent = event as CustomEvent<{ id?: string; group?: string }>

    if (customEvent.detail?.group !== group) return
    if (customEvent.detail?.id === dropdownId) return
    if (!open) return

    requestClose()
  }

  const revealMotion = (
    _node: HTMLElement,
    { duration }: { duration: number },
  ) => {
    const shift = 4
    const [dx, dy] =
      resolvedSide === 'bottom'
        ? [0, -shift]
        : resolvedSide === 'top'
          ? [0, shift]
          : resolvedSide === 'right'
            ? [-shift, 0]
            : [shift, 0]

    return {
      duration,
      easing: cubicOut,
      css: (t: number, u: number) =>
        `opacity: ${t}; transform: translate3d(${u * dx}px, ${u * dy}px, 0);`,
    }
  }

  const schedulePosition = (..._dependencies: unknown[]) => {
    void tick().then(() => {
      if (open) calculatePosition()
    })
  }

  const handleResize = () => {
    if (open) calculatePosition()
  }

  const handleScroll = (event: Event) => {
    if (!open) return
    if (dropdownElement?.contains(event.target as Node)) return
    calculatePosition()
  }

  $effect(() => {
    if (open && !wasOpen) {
      wasOpen = true

      if (typeof window !== 'undefined') {
        window.dispatchEvent(
          new CustomEvent(DROPDOWN_OPEN_EVENT, {
            detail: {
              id: dropdownId,
              group,
            },
          }),
        )
      }

      schedulePosition()
    }
  })

  $effect.pre(() => {
    if (!open && wasOpen) {
      wasOpen = false
    }
  })

  $effect(() => {
    if (open) {
      schedulePosition(
        x,
        y,
        anchorElement,
        placement,
        offset,
        matchAnchorWidth,
        flip,
        fitViewport,
      )
    }
  })

  $effect(() => {
    const node = dropdownElement
    if (!node || typeof ResizeObserver === 'undefined') return

    const observer = new ResizeObserver(scheduleFrame)
    observer.observe(node)
    for (const child of node.children) observer.observe(child)
    if (anchorElement) observer.observe(anchorElement)

    return () => {
      observer.disconnect()
      if (frame) cancelAnimationFrame(frame)
      frame = 0
    }
  })

  onMount(() => on(window, DROPDOWN_OPEN_EVENT, handleDropdownOpened))
</script>

<svelte:document
  onmousedown={handleDocumentMouseDown}
  onkeydown={handleDocumentKeydown}
/>
<svelte:window onresize={handleResize} onscrollcapture={handleScroll} />

{#if open}
  <div
    {@attach portalToBody}
    bind:this={dropdownElement}
    class={`fixed z-2147483000 ${panelClass}`}
    style="top: 0; left: 0; visibility: hidden;"
    transition:revealMotion={{
      duration: motion(MOTION.element, prefersReducedMotion.current),
    }}
  >
    {@render children?.()}
  </div>
{/if}
