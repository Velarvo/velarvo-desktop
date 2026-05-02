<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte'
  import { createEventDispatcher } from 'svelte'

  export let open = false
  export let x: number | null = null
  export let y: number | null = null
  export let anchorElement: HTMLElement | null = null
  export let placement:
    | 'bottom-start'
    | 'bottom-end'
    | 'top-start'
    | 'top-end' = 'bottom-start'
  export let offset = 8
  export let panelClass = ''
  export let group = 'global'
  export let closeOnOutsideClick = true
  export let closeOnEscape = true
  export let matchAnchorWidth = false

  const dispatch = createEventDispatcher<{ close: void }>()
  const DROPDOWN_OPEN_EVENT = 'floating-dropdown-opened'

  let dropdownElement: HTMLDivElement | null = null
  let position = { left: 0, top: 0 }
  let widthStyle = ''
  let wasOpen = false
  const dropdownId = `dropdown-${Math.random().toString(36).slice(2, 10)}`

  const portalToBody = (node: HTMLElement) => {
    if (typeof document === 'undefined') {
      return {
        destroy() {
          return
        },
      }
    }

    document.body.appendChild(node)

    return {
      destroy() {
        if (node.parentNode === document.body) {
          document.body.removeChild(node)
        }
      },
    }
  }

  const requestClose = () => {
    dispatch('close')
  }

  const getAnchorRect = () => {
    if (!anchorElement) return null
    return anchorElement.getBoundingClientRect()
  }

  const calculatePosition = () => {
    const anchorRect = getAnchorRect()

    let nextLeft = 0
    let nextTop = 0

    if (anchorRect) {
      widthStyle = matchAnchorWidth ? `${Math.round(anchorRect.width)}px` : ''

      if (placement.startsWith('bottom')) {
        nextTop = anchorRect.bottom + offset
      } else {
        nextTop = anchorRect.top - offset
      }

      if (placement.endsWith('end')) {
        nextLeft = anchorRect.right
      } else {
        nextLeft = anchorRect.left
      }
    } else {
      widthStyle = ''
      nextLeft = x ?? 0
      nextTop = y ?? 0
    }

    if (!dropdownElement) {
      position = { left: Math.round(nextLeft), top: Math.round(nextTop) }
      return
    }

    const dropdownRect = dropdownElement.getBoundingClientRect()
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight

    if (anchorRect) {
      if (placement.endsWith('end')) {
        nextLeft = nextLeft - dropdownRect.width
      }

      if (placement.startsWith('top')) {
        nextTop = nextTop - dropdownRect.height
      }
    }

    const minX = 8
    const minY = 8
    const maxX = viewportWidth - dropdownRect.width - 8
    const maxY = viewportHeight - dropdownRect.height - 8

    position = {
      left: Math.round(
        Math.min(Math.max(nextLeft, minX), Math.max(minX, maxX)),
      ),
      top: Math.round(Math.min(Math.max(nextTop, minY), Math.max(minY, maxY))),
    }
  }

  const handleDocumentMouseDown = (event: MouseEvent) => {
    if (!open || !closeOnOutsideClick || !dropdownElement) return

    const target = event.target as Node | null
    if (!target) return
    if (dropdownElement.contains(target)) return

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

  const handleViewportChange = () => {
    if (!open) return
    calculatePosition()
  }

  $: if (open && !wasOpen) {
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

    void tick().then(() => {
      if (!open) return
      calculatePosition()
    })
  }

  $: if (!open && wasOpen) {
    wasOpen = false
  }

  $: if (open && (x !== null || y !== null || anchorElement)) {
    void tick().then(() => {
      if (!open) return
      calculatePosition()
    })
  }

  onMount(() => {
    document.addEventListener('mousedown', handleDocumentMouseDown)
    document.addEventListener('keydown', handleDocumentKeydown)
    window.addEventListener('resize', handleViewportChange)
    window.addEventListener('scroll', handleViewportChange, true)
    window.addEventListener(DROPDOWN_OPEN_EVENT, handleDropdownOpened)
  })

  onDestroy(() => {
    document.removeEventListener('mousedown', handleDocumentMouseDown)
    document.removeEventListener('keydown', handleDocumentKeydown)
    window.removeEventListener('resize', handleViewportChange)
    window.removeEventListener('scroll', handleViewportChange, true)
    window.removeEventListener(DROPDOWN_OPEN_EVENT, handleDropdownOpened)
  })
</script>

{#if open}
  <div
    use:portalToBody
    bind:this={dropdownElement}
    class={`fixed z-2147483000 ${panelClass}`}
    style={`top: ${position.top}px; left: ${position.left}px;${widthStyle ? ` width: ${widthStyle};` : ''}`}
  >
    <slot />
  </div>
{/if}
