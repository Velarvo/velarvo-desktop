<script lang="ts">
  import type { Attachment } from 'svelte/attachments'
  import type { Snippet } from 'svelte'
  import { onMount, tick, untrack } from 'svelte'
  import { on } from 'svelte/events'
  import { cubicOut } from 'svelte/easing'
  import { prefersReducedMotion } from 'svelte/motion'
  import { computePosition, viewportSize } from '@/lib/floating'
  import type { Placement, Side } from '@/lib/floating'
  import { MOTION, motion } from '@/lib/motion'

  interface Props {
    open?: boolean
    x?: number | null
    y?: number | null
    anchorElement?: HTMLElement | null
    triggerElement?: HTMLElement | null
    placement?: Placement
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

  let dropdownElement = $state<HTMLDivElement | null>(null)
  let wasOpen = $state(false)
  let resolvedSide = $state<Side>('bottom')
  let frame = 0
  const dropdownId = `dropdown-${Math.random().toString(36).slice(2, 10)}`

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

    const viewport = viewportSize()
    const anchorRect = anchorElement?.getBoundingClientRect() ?? null
    const reference = anchorRect ?? new DOMRect(x ?? 0, y ?? 0, 0, 0)

    node.style.maxWidth = `${Math.max(160, viewport.width - viewportMargin * 2)}px`
    node.style.width =
      matchAnchorWidth && anchorRect
        ? `${Math.round(Math.min(anchorRect.width, viewport.width - viewportMargin * 2))}px`
        : ''
    node.style.maxHeight = ''
    node.style.overflowY = ''

    const { width, height } = node.getBoundingClientRect()
    const position = computePosition(
      reference,
      { width, height },
      {
        placement,
        offset: anchorRect ? offset : 0,
        margin: viewportMargin,
        flip,
        fitViewport,
      },
    )

    if (position.maxHeight !== null) {
      node.style.maxHeight = `${position.maxHeight}px`
      node.style.overflowY = 'auto'
    }
    node.style.left = `${position.left}px`
    node.style.top = `${position.top}px`
    node.style.visibility = ''
    node.dataset.side = position.side

    resolvedSide = position.side
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
