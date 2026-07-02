<script lang="ts">
  import EditorGroup from './EditorGroup.svelte'
  import {
    MIN_PANE_FRACTION,
    resizeSplit,
    type LayoutNode,
  } from '@/lib/editorLayout'

  import { onDestroy } from 'svelte'

  export let node: LayoutNode
  export let canClose = true

  let containerEl: HTMLDivElement | null = null
  let childEls: HTMLDivElement[] = []

  type DragState = {
    index: number
    startPos: number
    extent: number
    startSizes: number[]
    sizes: number[]
  }
  let drag: DragState | null = null

  const stopResizing = () => {
    document.body.classList.remove('velarvo-resizing')
    document.body.style.cursor = ''
  }

  const onHandlePointerDown = (event: PointerEvent, index: number) => {
    if (node.type !== 'split' || !containerEl) return
    event.preventDefault()
    ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)

    const rect = containerEl.getBoundingClientRect()
    drag = {
      index,
      startPos: node.orientation === 'row' ? event.clientX : event.clientY,
      extent: node.orientation === 'row' ? rect.width : rect.height,
      startSizes: [...node.sizes],
      sizes: [...node.sizes],
    }

    document.body.classList.add('velarvo-resizing')
    document.body.style.cursor =
      node.orientation === 'row' ? 'col-resize' : 'row-resize'
  }

  const onHandlePointerMove = (event: PointerEvent) => {
    if (!drag || node.type !== 'split' || drag.extent === 0) return

    const current = node.orientation === 'row' ? event.clientX : event.clientY
    const deltaFraction = (current - drag.startPos) / drag.extent

    const { index, startSizes } = drag
    const pairTotal = startSizes[index] + startSizes[index + 1]

    let first = startSizes[index] + deltaFraction
    first = Math.max(
      MIN_PANE_FRACTION,
      Math.min(first, pairTotal - MIN_PANE_FRACTION),
    )
    const second = pairTotal - first

    drag.sizes[index] = first
    drag.sizes[index + 1] = second

    const a = childEls[index]
    const b = childEls[index + 1]
    if (a) a.style.flexGrow = String(first)
    if (b) b.style.flexGrow = String(second)
  }

  const onHandlePointerUp = (event: PointerEvent) => {
    if (!drag) return
    ;(event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId)
    resizeSplit(node.id, drag.sizes)
    drag = null
    stopResizing()
  }

  onDestroy(stopResizing)
</script>

{#if node.type === 'group'}
  <EditorGroup group={node} {canClose} />
{:else}
  <div
    bind:this={containerEl}
    class="flex min-h-0 min-w-0 flex-1 {node.orientation === 'row'
      ? 'flex-row'
      : 'flex-col'}"
  >
    {#each node.children as child, index (child.id)}
      <div
        bind:this={childEls[index]}
        class="flex min-h-0 min-w-0"
        style={`flex: ${node.sizes[index]} 1 0%;`}
      >
        <svelte:self node={child} canClose={true} />
      </div>

      {#if index < node.children.length - 1}
        <div
          role="separator"
          aria-orientation={node.orientation === 'row'
            ? 'vertical'
            : 'horizontal'}
          tabindex="-1"
          class="group/handle relative z-10 shrink-0 bg-border transition-colors hover:bg-primary/60 {node.orientation ===
          'row'
            ? 'w-px cursor-col-resize'
            : 'h-px cursor-row-resize'}"
          on:pointerdown={(event) => onHandlePointerDown(event, index)}
          on:pointermove={onHandlePointerMove}
          on:pointerup={onHandlePointerUp}
        >
          <span
            class="absolute {node.orientation === 'row'
              ? '-inset-x-1.5 inset-y-0'
              : '-inset-y-1.5 inset-x-0'}"
          ></span>
        </div>
      {/if}
    {/each}
  </div>
{/if}
