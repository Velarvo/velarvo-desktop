import type { Attachment } from 'svelte/attachments'
import { computePosition, isVertical } from '@/lib/floating'
import type { Placement, Side } from '@/lib/floating'
import { MOTION } from '@/lib/motion'

export type TooltipPlacement = Placement

export interface TooltipOptions {
  content: string
  shortcut?: string
  placement?: TooltipPlacement
  overflowOnly?: boolean
  delay?: number
  offset?: number
  disabled?: boolean
}

const SHOW_DELAY = 450
const WARM_WINDOW = 300
const VIEWPORT_MARGIN = 8
const ARROW_INSET = 10
const ARROW_HALF = 4

const BUBBLE_CLASS =
  'pointer-events-none fixed z-2147483001 flex max-w-80 items-center gap-2 rounded-md border border-line bg-popover px-2 py-1 text-xs leading-snug font-medium text-popover-foreground shadow-lg shadow-black/40 select-none [overflow-wrap:anywhere]'
const SHORTCUT_CLASS =
  'shrink-0 rounded bg-fill-active px-1.5 py-px font-mono text-[10px] font-semibold text-muted-foreground'
const ARROW_CLASS = 'absolute h-2 w-2 rotate-45 border-line bg-popover'

const ARROW_EDGE: Record<Side, { edge: string; borders: string }> = {
  top: { edge: 'bottom', borders: '0 1px 1px 0' },
  bottom: { edge: 'top', borders: '1px 0 0 1px' },
  left: { edge: 'right', borders: '1px 1px 0 0' },
  right: { edge: 'left', borders: '0 0 1px 1px' },
}

const ENTER_SHIFT: Record<Side, string> = {
  top: 'translate3d(0, 3px, 0)',
  bottom: 'translate3d(0, -3px, 0)',
  left: 'translate3d(3px, 0, 0)',
  right: 'translate3d(-3px, 0, 0)',
}

let lastHiddenAt = 0
let hideActive: (() => void) | null = null
let nextId = 0

const prefersReducedMotion = () =>
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), Math.max(min, max))

const isTruncated = (node: HTMLElement) =>
  node.scrollWidth > node.clientWidth || node.scrollHeight > node.clientHeight

const renderBubble = (id: string, content: string, shortcut: string) => {
  const bubble = document.createElement('div')
  bubble.id = id
  bubble.setAttribute('role', 'tooltip')
  bubble.className = BUBBLE_CLASS
  bubble.style.cssText = 'top: 0; left: 0; visibility: hidden;'

  const text = document.createElement('span')
  text.className = 'min-w-0'
  text.textContent = content
  bubble.append(text)

  if (shortcut) {
    const kbd = document.createElement('kbd')
    kbd.className = SHORTCUT_CLASS
    kbd.textContent = shortcut
    bubble.append(kbd)
  }

  const arrow = document.createElement('span')
  arrow.className = ARROW_CLASS
  bubble.append(arrow)

  return { bubble, arrow }
}

export const tooltip =
  ({
    content,
    shortcut = '',
    placement = 'top',
    overflowOnly = false,
    delay = SHOW_DELAY,
    offset = 8,
    disabled = false,
  }: TooltipOptions): Attachment<HTMLElement> =>
  (node) => {
    if (disabled || !content) return

    const id = `tooltip-${++nextId}`
    let bubble: HTMLDivElement | null = null
    let timer: ReturnType<typeof setTimeout> | undefined
    let dismissListeners: AbortController | null = null

    const place = (element: HTMLDivElement, arrow: HTMLSpanElement) => {
      const { width, height } = element.getBoundingClientRect()
      const position = computePosition(
        node.getBoundingClientRect(),
        { width, height },
        { placement, offset, margin: VIEWPORT_MARGIN },
      )

      element.style.left = `${position.left}px`
      element.style.top = `${position.top}px`
      element.dataset.side = position.side

      const { edge, borders } = ARROW_EDGE[position.side]
      const length = isVertical(position.side) ? width : height
      const along = clamp(position.arrow, ARROW_INSET, length - ARROW_INSET)
      arrow.style.borderWidth = borders
      arrow.style.setProperty(edge, `-${ARROW_HALF + 1}px`)
      arrow.style.setProperty(
        isVertical(position.side) ? 'left' : 'top',
        `${along - ARROW_HALF}px`,
      )

      element.style.visibility = ''
      return position.side
    }

    const listenForDismiss = () => {
      dismissListeners = new AbortController()
      const { signal } = dismissListeners
      const hideOnEscape = (event: KeyboardEvent) => {
        if (event.key === 'Escape') hide()
      }
      document.addEventListener('keydown', hideOnEscape, { signal })
      window.addEventListener('scroll', hide, { signal, capture: true })
      window.addEventListener('resize', hide, { signal })
      window.addEventListener('blur', hide, { signal })
    }

    const show = () => {
      clearTimeout(timer)
      if (bubble || !node.isConnected) return
      if (overflowOnly && !isTruncated(node)) return
      hideActive?.()

      const rendered = renderBubble(id, content, shortcut)
      document.body.append(rendered.bubble)
      const side = place(rendered.bubble, rendered.arrow)
      bubble = rendered.bubble
      hideActive = hide
      listenForDismiss()

      if (node.getAttribute('aria-label') !== content) {
        node.setAttribute('aria-describedby', id)
      }

      if (!prefersReducedMotion()) {
        bubble.animate(
          [
            { opacity: 0, transform: ENTER_SHIFT[side] },
            { opacity: 1, transform: 'none' },
          ],
          {
            duration: MOTION.element,
            easing: 'cubic-bezier(0.33, 1, 0.68, 1)',
          },
        )
      }
    }

    const hide = () => {
      clearTimeout(timer)
      if (!bubble) return
      dismissListeners?.abort()
      dismissListeners = null
      bubble.remove()
      bubble = null
      lastHiddenAt = performance.now()
      if (hideActive === hide) hideActive = null
      if (node.getAttribute('aria-describedby') === id) {
        node.removeAttribute('aria-describedby')
      }
    }

    const scheduleShow = (event: PointerEvent) => {
      if (event.pointerType === 'touch') return
      clearTimeout(timer)
      const warm =
        hideActive !== null || performance.now() - lastHiddenAt < WARM_WINDOW
      if (warm) show()
      else timer = setTimeout(show, delay)
    }

    const showOnKeyboardFocus = () => {
      if (node.matches(':focus-visible')) show()
    }

    const controller = new AbortController()
    const { signal } = controller
    node.addEventListener('pointerenter', scheduleShow, { signal })
    node.addEventListener('pointerleave', hide, { signal })
    node.addEventListener('pointerdown', hide, { signal })
    node.addEventListener('focus', showOnKeyboardFocus, { signal })
    node.addEventListener('blur', hide, { signal })

    return () => {
      controller.abort()
      hide()
    }
  }
