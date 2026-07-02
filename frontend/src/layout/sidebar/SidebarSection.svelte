<script lang="ts">
  import { flip } from 'svelte/animate'
  import { onDestroy, onMount, tick, type ComponentType } from 'svelte'
  import { createEventDispatcher } from 'svelte'
  import {
    ChevronDown,
    ChevronRight,
    Ellipsis,
    GripVertical,
    Pencil,
    Plus,
    Trash2,
  } from 'lucide-svelte'
  import { IconApi, IconBrandGraphql } from '@tabler/icons-svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import { translate } from '@/lib/i18n'
  import {
    DEFAULT_API_METHOD,
    DEFAULT_GRAPHQL_METHOD,
    isApiItem,
  } from '@/lib/sidebar'
  import type {
    ApiMethod,
    SidebarItemDraft,
    SidebarSection,
  } from '@/types/sidebar'
  import {
    sidebarDragStore,
    attachSidebarDragListeners,
    beginSidebarDrag,
    endSidebarDrag,
  } from './sidebarDragStore'

  export let section: SidebarSection
  export let collapsed = false
  export let expanded = false
  export let highlighted = false
  export let icon: ComponentType
  export let getMethodBadgeStyle: (method?: ApiMethod) => string

  const dispatch = createEventDispatcher<{
    toggle: { sectionId: string }
    removeSection: { sectionId: string }
    renameSection: { sectionId: string; label: string }
    addItem: {
      sectionId: string
    } & SidebarItemDraft
    renameItem: { sectionId: string; itemId: string; name: string }
    removeItem: { sectionId: string; itemId: string }
    reorderItem: { sectionId: string; itemId: string; toIndex: number }
    moveItem: {
      fromSectionId: string
      itemId: string
      toSectionId: string
      toIndex: number
    }
  }>()

  let contextMenuVisible = false
  let contextMenuX = 0
  let contextMenuY = 0
  let renameMode = false
  let renameValue = ''
  let renameInputElement: HTMLInputElement | null = null
  let itemRenameId: string | null = null
  let itemRenameValue = ''
  let itemRenameInputElement: HTMLInputElement | null = null
  let pendingNewBlockRename = false
  let blockTypePickerVisible = false
  let detachDragListeners: (() => void) | null = null
  let itemMenuId: string | null = null
  let itemMenuAnchor: HTMLElement | null = null
  let itemMenuContextX: number | null = null
  let itemMenuContextY: number | null = null

  type NewBlockPreset = {
    key: 'rest' | 'graphql'
    label: string
    icon: ComponentType
  } & Omit<SidebarItemDraft, 'name'>

  const newBlockPresets: NewBlockPreset[] = [
    {
      key: 'rest',
      label: $translate('sidebar.blockTypes.rest'),
      icon: IconApi,
      type: 'API',
      method: DEFAULT_API_METHOD,
    },
    {
      key: 'graphql',
      label: $translate('sidebar.blockTypes.graphql'),
      icon: IconBrandGraphql,
      type: 'API',
      method: DEFAULT_GRAPHQL_METHOD,
    },
  ]

  const startItemRename = async (itemId: string, value: string) => {
    itemRenameId = itemId
    itemRenameValue = value

    await tick()
    itemRenameInputElement?.focus()
    itemRenameInputElement?.select()
  }

  $: dragState = $sidebarDragStore
  $: isSourceOfDrag = dragState?.source.fromSectionId === section.id
  $: draggedItemId = isSourceOfDrag ? (dragState?.source.itemId ?? null) : null

  $: dropIndicator =
    dragState &&
    dragState.target &&
    dragState.target.sectionId === section.id &&
    expanded
      ? dragState.target.indicator
      : null

  $: isHeaderDropTarget = !!(
    dragState &&
    dragState.target &&
    dragState.target.sectionId === section.id &&
    dragState.target.indicator.kind === 'header'
  )

  $: if (pendingNewBlockRename) {
    const candidate = [...section.items]
      .reverse()
      .find((item) => item.name === $translate('sidebar.newBlockName'))
    if (candidate) {
      pendingNewBlockRename = false
      void startItemRename(candidate.id, candidate.name)
    }
  }

  const openContextMenu = (event: MouseEvent) => {
    event.preventDefault()

    blockTypePickerVisible = false
    contextMenuVisible = true
    contextMenuX = event.clientX
    contextMenuY = event.clientY
  }

  const closeContextMenu = () => {
    contextMenuVisible = false
    blockTypePickerVisible = false
  }

  const openSectionAddFromButton = (event: MouseEvent) => {
    event.stopPropagation()
    const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
    contextMenuX = rect.left
    contextMenuY = rect.bottom + 4
    blockTypePickerVisible = true
    contextMenuVisible = true
  }

  const openSectionMenuFromButton = (event: MouseEvent) => {
    event.stopPropagation()
    const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
    contextMenuX = Math.max(8, rect.right - 160)
    contextMenuY = rect.bottom + 4
    blockTypePickerVisible = false
    contextMenuVisible = true
  }

  const openItemMenu = (event: MouseEvent, itemId: string) => {
    event.stopPropagation()
    event.preventDefault()
    itemMenuAnchor = event.currentTarget as HTMLElement
    itemMenuId = itemId
    itemMenuContextX = null
    itemMenuContextY = null
  }

  const openItemMenuContext = (event: MouseEvent, itemId: string) => {
    event.stopPropagation()
    event.preventDefault()
    itemMenuAnchor = null
    itemMenuId = itemId
    itemMenuContextX = event.clientX
    itemMenuContextY = event.clientY
  }

  const closeItemMenu = () => {
    itemMenuId = null
    itemMenuAnchor = null
    itemMenuContextX = null
    itemMenuContextY = null
  }

  const startItemRenameFromMenu = () => {
    const id = itemMenuId
    closeItemMenu()
    if (!id) return
    const target = section.items.find((item) => item.id === id)
    if (target) void startItemRename(target.id, target.name)
  }

  const deleteItemFromMenu = () => {
    const id = itemMenuId
    closeItemMenu()
    if (!id) return
    dispatch('removeItem', { sectionId: section.id, itemId: id })
  }

  const openBlockTypePicker = () => {
    blockTypePickerVisible = true
  }

  const addNewBlock = (preset: NewBlockPreset) => {
    dispatch('addItem', {
      sectionId: section.id,
      name: $translate('sidebar.newBlockName'),
      type: preset.type,
      status: preset.status,
      method: preset.method,
    })

    pendingNewBlockRename = true
    closeContextMenu()

    if (!expanded) {
      dispatch('toggle', { sectionId: section.id })
    }
  }

  const startRename = async () => {
    renameValue = section.label
    renameMode = true
    closeContextMenu()

    await tick()
    renameInputElement?.focus()
    renameInputElement?.select()
  }

  const submitRename = () => {
    const label = renameValue.trim()
    if (!label) {
      renameMode = false
      return
    }

    if (label === section.label) {
      renameMode = false
      return
    }

    dispatch('renameSection', { sectionId: section.id, label })
    renameMode = false
  }

  const cancelRename = () => {
    renameMode = false
    renameValue = section.label
  }

  const handleRenameKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Enter') {
      event.preventDefault()
      event.stopPropagation()
      submitRename()
      return
    }

    if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      cancelRename()
    }
  }

  const handleHeaderClick = () => {
    if (renameMode) return
    dispatch('toggle', { sectionId: section.id })
  }

  const handleHeaderKeydown = (event: KeyboardEvent) => {
    if (renameMode) return
    if (event.key !== 'Enter' && event.key !== ' ') return

    event.preventDefault()
    dispatch('toggle', { sectionId: section.id })
  }

  const submitItemRename = (itemId: string) => {
    const name = itemRenameValue.trim()

    if (!name) {
      itemRenameId = null
      itemRenameValue = ''
      return
    }

    dispatch('renameItem', { sectionId: section.id, itemId, name })
    itemRenameId = null
    itemRenameValue = ''
  }

  const cancelItemRename = () => {
    itemRenameId = null
    itemRenameValue = ''
  }

  const handleItemRenameKeydown = (event: KeyboardEvent, itemId: string) => {
    if (event.key === 'Enter') {
      event.preventDefault()
      event.stopPropagation()
      submitItemRename(itemId)
      return
    }

    if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      cancelItemRename()
    }
  }

  const handleItemDragStart = (event: MouseEvent, itemId: string) => {
    if (event.button !== 0 || itemRenameId === itemId) return

    event.preventDefault()
    event.stopPropagation()

    beginSidebarDrag(
      { itemId, fromSectionId: section.id },
      event.clientX,
      event.clientY,
    )
  }

  const handleGlobalMouseUp = (event: MouseEvent) => {
    const state = $sidebarDragStore
    if (!state) return

    if (state.source.fromSectionId !== section.id) return

    event.preventDefault()

    const finalState = endSidebarDrag()
    const target = finalState?.target
    if (!target) return

    const fromSectionId = state.source.fromSectionId
    const itemId = state.source.itemId

    let toIndex = target.toIndex
    const isSameSection = target.sectionId === fromSectionId

    if (isSameSection) {
      const sourceIndex = section.items.findIndex((item) => item.id === itemId)
      if (sourceIndex === -1) return
      if (sourceIndex < toIndex) toIndex -= 1
      if (toIndex === sourceIndex) return
    }

    dispatch('moveItem', {
      fromSectionId,
      itemId,
      toSectionId: target.sectionId,
      toIndex,
    })
  }

  const handleLocalEscape = (event: KeyboardEvent) => {
    if (event.key !== 'Escape') return
    closeContextMenu()
    closeItemMenu()
    renameMode = false
    cancelItemRename()
  }

  onMount(() => {
    detachDragListeners = attachSidebarDragListeners()
    document.addEventListener('keydown', handleLocalEscape)
    document.addEventListener('mouseup', handleGlobalMouseUp)
  })

  onDestroy(() => {
    document.removeEventListener('keydown', handleLocalEscape)
    document.removeEventListener('mouseup', handleGlobalMouseUp)
    detachDragListeners?.()
    detachDragListeners = null
  })
</script>

<div>
  <div
    class="flex items-center gap-1.5 px-2 py-1.5"
    data-sidebar-section-header={!collapsed ? 'true' : null}
    data-sidebar-section-id={section.id}
    data-sidebar-items-count={section.items.length}
  >
    <div
      role="button"
      tabindex="0"
      on:click={handleHeaderClick}
      on:keydown={handleHeaderKeydown}
      on:contextmenu={openContextMenu}
      class="group flex min-w-0 flex-1 items-center gap-2.5 rounded-md h-8 px-3 py-3 text-sm transition-colors duration-150 {collapsed
        ? 'justify-center'
        : 'justify-between'} {expanded || highlighted
        ? 'text-white bg-white/5'
        : 'text-muted-foreground hover:text-white hover:bg-white/3'} {isHeaderDropTarget
        ? 'ring-2 ring-primary/80 bg-primary/10'
        : ''}"
    >
      <div class="flex min-w-0 items-center gap-2.5">
        {#if !collapsed}
          <span class="text-muted-foreground/50 shrink-0">
            {#if expanded}
              <ChevronDown size={13} />
            {:else}
              <ChevronRight size={13} />
            {/if}
          </span>
        {/if}

        <span
          class="shrink-0 transition-colors {expanded || highlighted
            ? 'text-primary'
            : 'text-muted-foreground group-hover:text-primary/70'}"
        >
          <svelte:component this={icon} size={collapsed ? 18 : 15} />
        </span>

        {#if !collapsed}
          {#if renameMode}
            <input
              bind:this={renameInputElement}
              bind:value={renameValue}
              class="h-7 w-full min-w-0 rounded-md border border-primary/40 bg-background px-2 text-[13px] font-medium text-white outline-none"
              placeholder={$translate('sidebar.sectionNamePlaceholder')}
              on:click|stopPropagation
              on:mousedown|stopPropagation
              on:keydown={handleRenameKeydown}
              on:blur={submitRename}
            />
          {:else}
            <span class="truncate text-[13px] font-medium">{section.label}</span
            >
          {/if}
        {/if}
      </div>

      {#if !collapsed}
        <div class="relative flex items-center justify-end shrink-0 w-14 h-6">
          <div class="flex items-center gap-0.5 transition-opacity opacity-100">
            <button
              type="button"
              class="hidden group-hover:inline-flex items-center justify-center w-6 h-6 z-50 rounded text-muted-foreground hover:text-white hover:bg-white/10 transition-colors"
              on:click={openSectionAddFromButton}
              aria-label={$translate('sidebar.ariaLabels.addNewBlock')}
            >
              <Plus size={13} />
            </button>
            <button
              type="button"
              class="hidden group-hover:inline-flex items-center justify-center w-6 h-6 z-50 rounded text-muted-foreground hover:text-white hover:bg-white/10 transition-colors"
              on:click={openSectionMenuFromButton}
              aria-label={$translate('sidebar.ariaLabels.sectionOptions')}
            >
              <Ellipsis size={13} />
            </button>
          </div>
          <span
            class="absolute right-0 inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-mono font-medium border border-border bg-white/5 text-muted-foreground transition-opacity {contextMenuVisible
              ? 'opacity-0'
              : 'opacity-100 group-hover:opacity-0'}"
          >
            {section.items.length}
          </span>
        </div>
      {/if}
    </div>
  </div>

  {#if expanded && !collapsed}
    <div class="overflow-hidden px-2 pb-2">
      <div
        role="list"
        data-sidebar-items-list="true"
        data-section-id={section.id}
        data-items-count={section.items.length}
        class="relative ml-4.5 border-l border-border/50 pl-3 pr-1 py-1 min-h-7 space-y-0.5 {dragState &&
        dragState.target?.sectionId === section.id &&
        dragState.target.indicator.kind !== 'header'
          ? 'bg-primary/4 rounded-md transition-colors'
          : ''}"
      >
        {#if section.items.length === 0}
          <div
            class="rounded-md px-2.5 py-2 text-[12px] text-muted-foreground bg-white/2 border border-border/50"
          >
            {$translate('sidebar.noConnections')}
          </div>
        {/if}

        {#if dropIndicator && dropIndicator.kind === 'empty'}
          <span
            class="pointer-events-none absolute left-2 right-2 top-1/2 -translate-y-1/2 h-0.5 rounded-full bg-primary shadow-[0_0_0_3px_rgba(34,197,94,0.25)]"
          ></span>
        {/if}

        {#each section.items as item, itemIndex (item.id)}
          <div
            role="listitem"
            data-sidebar-item="true"
            data-sidebar-item-id={item.id}
            data-sidebar-section-id={section.id}
            data-sidebar-item-index={itemIndex}
            animate:flip={{ duration: 170 }}
            class="group relative flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[13px] transition-all duration-150 text-white/70 hover:text-white hover:bg-white/5 {draggedItemId ===
            item.id
              ? 'opacity-40 bg-white/6 scale-[0.99] pointer-events-none'
              : ''}"
            on:contextmenu={(event) => openItemMenuContext(event, item.id)}
          >
            {#if dropIndicator && dropIndicator.kind === 'before' && dropIndicator.itemId === item.id}
              <span
                class="pointer-events-none absolute -top-px left-2 right-2 h-0.5 rounded-full bg-primary shadow-[0_0_0_3px_rgba(34,197,94,0.25)]"
              ></span>
            {/if}

            <button
              type="button"
              class="-m-1.5 shrink-0 rounded-md p-1.5 text-muted-foreground/55 transition-colors hover:bg-white/5 group-hover:text-muted-foreground cursor-grab active:cursor-grabbing"
              on:mousedown={(event) => handleItemDragStart(event, item.id)}
              aria-label={$translate('sidebar.ariaLabels.dragBlock')}
            >
              <GripVertical size={13} />
            </button>

            {#if isApiItem(item.type, item.method)}
              <span
                class="px-1.5 py-0.5 text-[9px] rounded font-medium {getMethodBadgeStyle(
                  item.method,
                )}">{item.method}</span
              >
            {:else}
              <span
                class="h-1.5 w-1.5 rounded-full shrink-0 transition-colors bg-white/20 group-hover:bg-primary/60"
              ></span>
            {/if}

            {#if itemRenameId === item.id}
              <input
                bind:this={itemRenameInputElement}
                bind:value={itemRenameValue}
                class="h-7 w-full min-w-0 rounded-md border border-primary/40 bg-background px-2 text-[12px] text-white outline-none"
                placeholder={$translate('sidebar.blockNamePlaceholder')}
                on:click|stopPropagation
                on:mousedown|stopPropagation
                on:keydown={(event) => handleItemRenameKeydown(event, item.id)}
                on:blur={() => submitItemRename(item.id)}
              />
            {:else}
              <span class="truncate flex-1">{item.name}</span>
              <button
                type="button"
                class="inline-flex shrink-0 rounded p-1 text-muted-foreground/60 hover:text-white hover:bg-white/10 transition opacity-100 pointer-events-auto"
                on:click={(event) => openItemMenu(event, item.id)}
                on:contextmenu={(event) => openItemMenuContext(event, item.id)}
                aria-label={$translate('sidebar.ariaLabels.blockOptions')}
              >
                <Ellipsis size={13} />
              </button>
            {/if}

            {#if dropIndicator && dropIndicator.kind === 'after' && dropIndicator.itemId === item.id}
              <span
                class="pointer-events-none absolute -bottom-px left-2 right-2 h-0.5 rounded-full bg-primary shadow-[0_0_0_3px_rgba(34,197,94,0.25)]"
              ></span>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<FloatingDropdown
  open={contextMenuVisible && !collapsed}
  x={contextMenuX}
  y={contextMenuY}
  group="sidebar-context-menu"
  panelClass="min-w-40 rounded-md border border-border bg-popover p-1 shadow-xl shadow-black/40"
  on:close={closeContextMenu}
>
  <div data-sidebar-context-menu="true" role="menu" tabindex="-1">
    {#if blockTypePickerVisible}
      <div
        class="py-3 flex flex-row items-stretch justify-center gap-4 min-w-56"
      >
        {#each newBlockPresets as preset (preset.key)}
          <button
            type="button"
            class="flex flex-col items-center px-3 py-2 rounded-lg hover:bg-primary/10 focus:bg-primary/15 transition-colors min-w-20"
            on:click={() => addNewBlock(preset)}
          >
            <span
              class="flex items-center justify-center w-12 h-12 rounded-full bg-primary/15 text-primary mb-1"
            >
              <svelte:component this={preset.icon} stroke={2} />
            </span>
            <span class="text-xs text-white font-medium">{preset.label}</span>
          </button>
        {/each}
      </div>
    {:else}
      <button
        type="button"
        class="w-full text-sm cursor-pointer text-white/80 hover:text-white focus:text-white rounded-sm px-2 py-2 inline-flex items-center gap-2 text-left"
        on:click={openBlockTypePicker}
      >
        <Plus size={13} />
        {$translate('sidebar.menu.addNewBlock')}
      </button>
      <button
        type="button"
        class="w-full text-sm cursor-pointer text-white/80 hover:text-white focus:text-white rounded-sm px-2 py-2 inline-flex items-center gap-2 text-left"
        on:click={startRename}
      >
        <Pencil size={13} />
        {$translate('sidebar.menu.rename')}
      </button>
      <div class="my-1 h-px bg-border"></div>
      <button
        type="button"
        class="w-full text-sm text-destructive hover:text-destructive cursor-pointer rounded-sm px-2 py-2 inline-flex items-center gap-2 text-left"
        on:click={() => {
          closeContextMenu()
          dispatch('removeSection', { sectionId: section.id })
        }}
      >
        <Trash2 size={13} />
        {$translate('sidebar.menu.deleteSection')}
      </button>
    {/if}
  </div>
</FloatingDropdown>

<FloatingDropdown
  open={itemMenuId !== null}
  anchorElement={itemMenuAnchor}
  x={itemMenuContextX}
  y={itemMenuContextY}
  placement={itemMenuContextX !== null && itemMenuContextY !== null
    ? undefined
    : 'bottom-end'}
  offset={4}
  group="sidebar-item-menu"
  panelClass="min-w-36 rounded-md border border-border bg-popover p-1 shadow-xl shadow-black/40"
  on:close={closeItemMenu}
>
  <div role="menu" tabindex="-1">
    <button
      type="button"
      class="w-full text-sm cursor-pointer text-white/80 hover:text-white focus:text-white rounded-sm px-2 py-2 inline-flex items-center gap-2 text-left"
      on:click={startItemRenameFromMenu}
    >
      <Pencil size={13} />
      {$translate('sidebar.menu.rename')}
    </button>
    <div class="my-1 h-px bg-border"></div>
    <button
      type="button"
      class="w-full text-sm text-destructive hover:text-destructive cursor-pointer rounded-sm px-2 py-2 inline-flex items-center gap-2 text-left"
      on:click={deleteItemFromMenu}
    >
      <Trash2 size={13} />
      {$translate('sidebar.menu.delete')}
    </button>
  </div>
</FloatingDropdown>

<style>
  :global(.velarvo-noselect) {
    user-select: none !important;
  }
</style>
