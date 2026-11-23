export type DragSource = {
  itemId: string
  fromSectionId: string
}

export type DropTarget = {
  sectionId: string
  toIndex: number
  indicator:
    | { kind: 'before' | 'after'; itemId: string }
    | { kind: 'empty' }
    | { kind: 'header' }
}

export type DragState = {
  source: DragSource
  clientX: number
  clientY: number
  target: DropTarget | null
}

class SidebarDrag {
  state = $state.raw<DragState | null>(null)
}

export const sidebarDrag = new SidebarDrag()

let globalListenersRef = 0
let rafId: number | null = null

const HORIZONTAL_TOLERANCE = 12

const pointerInside = (rect: DOMRect, clientX: number, clientY: number) => {
  return (
    clientY >= rect.top &&
    clientY <= rect.bottom &&
    clientX >= rect.left - HORIZONTAL_TOLERANCE &&
    clientX <= rect.right + HORIZONTAL_TOLERANCE
  )
}

const resolveDropTarget = (
  clientX: number,
  clientY: number,
  source: DragSource,
): DropTarget | null => {
  if (typeof document === 'undefined') return null

  const sectionHeaders = Array.from(
    document.querySelectorAll<HTMLElement>(
      '[data-sidebar-section-header="true"]',
    ),
  )

  for (const header of sectionHeaders) {
    if (!pointerInside(header.getBoundingClientRect(), clientX, clientY))
      continue

    const sectionId = header.dataset.sidebarSectionId
    if (!sectionId) continue

    const itemCount = Number(header.dataset.sidebarItemsCount ?? 0)

    return {
      sectionId,
      toIndex: itemCount,
      indicator: { kind: 'header' },
    }
  }

  const lists = Array.from(
    document.querySelectorAll<HTMLElement>('[data-sidebar-items-list="true"]'),
  )
  if (lists.length === 0) return null

  let activeList: HTMLElement | null = null
  for (const list of lists) {
    if (pointerInside(list.getBoundingClientRect(), clientX, clientY)) {
      activeList = list
      break
    }
  }
  if (!activeList) return null

  const sectionId = activeList.dataset.sectionId
  if (!sectionId) return null

  const itemNodes = Array.from(
    activeList.querySelectorAll<HTMLElement>('[data-sidebar-item="true"]'),
  )

  if (itemNodes.length === 0) {
    return {
      sectionId,
      toIndex: 0,
      indicator: { kind: 'empty' },
    }
  }

  for (let i = 0; i < itemNodes.length; i++) {
    const node = itemNodes[i]
    const rect = node.getBoundingClientRect()
    const midpoint = rect.top + rect.height / 2

    if (clientY < midpoint) {
      const itemId = node.dataset.sidebarItemId ?? ''
      const itemIndex = Number(node.dataset.sidebarItemIndex ?? i)

      if (
        sectionId === source.fromSectionId &&
        (itemId === source.itemId ||
          (i > 0 && itemNodes[i - 1].dataset.sidebarItemId === source.itemId))
      ) {
        return null
      }

      return {
        sectionId,
        toIndex: itemIndex,
        indicator: { kind: 'before', itemId },
      }
    }
  }

  const last = itemNodes[itemNodes.length - 1]
  const lastItemId = last.dataset.sidebarItemId ?? ''
  const lastIndex = Number(
    last.dataset.sidebarItemIndex ?? itemNodes.length - 1,
  )

  if (sectionId === source.fromSectionId && lastItemId === source.itemId) {
    return null
  }

  return {
    sectionId,
    toIndex: lastIndex + 1,
    indicator: { kind: 'after', itemId: lastItemId },
  }
}

const scheduleResolve = () => {
  if (typeof window === 'undefined' || rafId !== null) return
  rafId = window.requestAnimationFrame(() => {
    rafId = null
    const state = sidebarDrag.state
    if (!state) return

    const target = resolveDropTarget(state.clientX, state.clientY, state.source)
    sidebarDrag.state = { ...state, target }
  })
}

const handleGlobalMouseMove = (event: MouseEvent) => {
  const state = sidebarDrag.state
  if (!state) return

  sidebarDrag.state = {
    ...state,
    clientX: event.clientX,
    clientY: event.clientY,
  }

  scheduleResolve()
}

const cancelDrag = () => {
  sidebarDrag.state = null
  if (typeof window !== 'undefined' && rafId !== null) {
    window.cancelAnimationFrame(rafId)
    rafId = null
  }
  if (typeof document !== 'undefined') {
    document.body.style.cursor = ''
  }
}

const handleGlobalBlur = () => {
  if (sidebarDrag.state) cancelDrag()
}

const handleGlobalKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && sidebarDrag.state) cancelDrag()
}

export const attachSidebarDragListeners = () => {
  if (typeof document === 'undefined') return () => {}

  globalListenersRef += 1
  if (globalListenersRef === 1) {
    document.addEventListener('mousemove', handleGlobalMouseMove)
    document.addEventListener('keydown', handleGlobalKeydown)
    window.addEventListener('blur', handleGlobalBlur)
  }

  return () => {
    globalListenersRef = Math.max(0, globalListenersRef - 1)
    if (globalListenersRef === 0) {
      document.removeEventListener('mousemove', handleGlobalMouseMove)
      document.removeEventListener('keydown', handleGlobalKeydown)
      window.removeEventListener('blur', handleGlobalBlur)
      cancelDrag()
    }
  }
}

export const beginSidebarDrag = (
  source: DragSource,
  clientX: number,
  clientY: number,
) => {
  if (typeof document !== 'undefined') {
    document.body.style.cursor = 'grabbing'
  }
  sidebarDrag.state = { source, clientX, clientY, target: null }
  scheduleResolve()
}

export const endSidebarDrag = (): DragState | null => {
  const state = sidebarDrag.state
  cancelDrag()
  return state
}
