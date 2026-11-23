<script lang="ts">
  import { flip } from 'svelte/animate'
  import { onMount, tick } from 'svelte'
  import type { IconComponent } from '@/types/icon'
  import {
    ChevronDown,
    ChevronRight,
    Ellipsis,
    GripVertical,
    Pencil,
    Plus,
    Trash2,
  } from '@lucide/svelte'
  import {
    IconApi,
    IconBrandGraphql,
    type Icon as TablerIcon,
  } from '@tabler/icons-svelte-runes'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import MenuItem, {
    MENU_PANEL_CLASS,
    MENU_SEPARATOR_CLASS,
  } from '@/components/ui/MenuItem.svelte'
  import { t } from '@/lib/i18n.svelte'
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
    sidebarDrag,
    attachSidebarDragListeners,
    beginSidebarDrag,
    endSidebarDrag,
  } from './sidebarDragStore.svelte'

  interface Props {
    section: SidebarSection
    collapsed?: boolean
    expanded?: boolean
    highlighted?: boolean
    icon: IconComponent
    getMethodBadgeStyle: (method?: ApiMethod) => string
    onToggle?: (detail: { sectionId: string }) => void
    onRemoveSection?: (detail: { sectionId: string }) => void
    onRenameSection?: (detail: { sectionId: string; label: string }) => void
    onAddItem?: (detail: { sectionId: string } & SidebarItemDraft) => void
    onRenameItem?: (detail: {
      sectionId: string
      itemId: string
      name: string
    }) => void
    onRemoveItem?: (detail: { sectionId: string; itemId: string }) => void
    onMoveItem?: (detail: {
      fromSectionId: string
      itemId: string
      toSectionId: string
      toIndex: number
    }) => void
  }

  let {
    section,
    collapsed = false,
    expanded = false,
    highlighted = false,
    icon,
    getMethodBadgeStyle,
    onToggle,
    onRemoveSection,
    onRenameSection,
    onAddItem,
    onRenameItem,
    onRemoveItem,
    onMoveItem,
  }: Props = $props()

  let contextMenuVisible = $state(false)
  let contextMenuX = $state(0)
  let contextMenuY = $state(0)
  let renameMode = $state(false)
  let renameValue = $state('')
  let renameInputElement = $state<HTMLInputElement | null>(null)
  let itemRenameId = $state<string | null>(null)
  let itemRenameValue = $state('')
  let itemRenameInputElement = $state<HTMLInputElement | null>(null)
  let pendingNewBlockRename = $state(false)
  let blockTypePickerVisible = $state(false)
  let itemMenuId = $state<string | null>(null)
  let itemMenuAnchor = $state<HTMLElement | null>(null)
  let contextMenuTrigger = $state<HTMLElement | null>(null)
  let itemMenuContextX = $state<number | null>(null)
  let itemMenuContextY = $state<number | null>(null)

  type NewBlockPreset = {
    key: 'rest' | 'graphql'
    label: string
    icon: TablerIcon
  } & Omit<SidebarItemDraft, 'name'>

  const newBlockPresets: NewBlockPreset[] = [
    {
      key: 'rest',
      label: t('sidebar.blockTypes.rest'),
      icon: IconApi,
      type: 'API',
      method: DEFAULT_API_METHOD,
    },
    {
      key: 'graphql',
      label: t('sidebar.blockTypes.graphql'),
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

  let dragState = $derived(sidebarDrag.state)
  let isSourceOfDrag = $derived(dragState?.source.fromSectionId === section.id)
  let draggedItemId = $derived(
    isSourceOfDrag ? (dragState?.source.itemId ?? null) : null,
  )

  let dropIndicator = $derived(
    dragState &&
      dragState.target &&
      dragState.target.sectionId === section.id &&
      expanded
      ? dragState.target.indicator
      : null,
  )

  let isHeaderDropTarget = $derived(
    !!(
      dragState &&
      dragState.target &&
      dragState.target.sectionId === section.id &&
      dragState.target.indicator.kind === 'header'
    ),
  )

  $effect.pre(() => {
    if (pendingNewBlockRename) {
      const candidate = [...section.items]
        .reverse()
        .find((item) => item.name === t('sidebar.newBlockName'))
      if (candidate) {
        pendingNewBlockRename = false
        void startItemRename(candidate.id, candidate.name)
      }
    }
  })

  const openContextMenu = (event: MouseEvent) => {
    event.preventDefault()

    contextMenuTrigger = null
    blockTypePickerVisible = false
    contextMenuVisible = true
    contextMenuX = event.clientX
    contextMenuY = event.clientY
  }

  const closeContextMenu = () => {
    contextMenuVisible = false
    blockTypePickerVisible = false
    contextMenuTrigger = null
  }

  const openSectionAddFromButton = (event: MouseEvent) => {
    event.stopPropagation()
    const trigger = event.currentTarget as HTMLElement

    if (
      contextMenuVisible &&
      blockTypePickerVisible &&
      contextMenuTrigger === trigger
    ) {
      closeContextMenu()
      return
    }

    const rect = trigger.getBoundingClientRect()
    contextMenuTrigger = trigger
    contextMenuX = rect.left
    contextMenuY = rect.bottom + 4
    blockTypePickerVisible = true
    contextMenuVisible = true
  }

  const openSectionMenuFromButton = (event: MouseEvent) => {
    event.stopPropagation()
    const trigger = event.currentTarget as HTMLElement

    if (
      contextMenuVisible &&
      !blockTypePickerVisible &&
      contextMenuTrigger === trigger
    ) {
      closeContextMenu()
      return
    }

    const rect = trigger.getBoundingClientRect()
    contextMenuTrigger = trigger
    contextMenuX = Math.max(8, rect.right - 160)
    contextMenuY = rect.bottom + 4
    blockTypePickerVisible = false
    contextMenuVisible = true
  }

  const openItemMenu = (event: MouseEvent, itemId: string) => {
    event.stopPropagation()
    event.preventDefault()
    const trigger = event.currentTarget as HTMLElement

    if (itemMenuId === itemId && itemMenuAnchor === trigger) {
      closeItemMenu()
      return
    }

    itemMenuAnchor = trigger
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
    onRemoveItem?.({ sectionId: section.id, itemId: id })
  }

  const openBlockTypePicker = () => {
    blockTypePickerVisible = true
  }

  const addNewBlock = (preset: NewBlockPreset) => {
    onAddItem?.({
      sectionId: section.id,
      name: t('sidebar.newBlockName'),
      type: preset.type,
      status: preset.status,
      method: preset.method,
    })

    pendingNewBlockRename = true
    closeContextMenu()

    if (!expanded) {
      onToggle?.({ sectionId: section.id })
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

    onRenameSection?.({ sectionId: section.id, label })
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
    onToggle?.({ sectionId: section.id })
  }

  const handleHeaderKeydown = (event: KeyboardEvent) => {
    if (renameMode) return
    if (event.key !== 'Enter' && event.key !== ' ') return

    event.preventDefault()
    onToggle?.({ sectionId: section.id })
  }

  const submitItemRename = (itemId: string) => {
    const name = itemRenameValue.trim()

    if (!name) {
      itemRenameId = null
      itemRenameValue = ''
      return
    }

    onRenameItem?.({ sectionId: section.id, itemId, name })
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
    const state = sidebarDrag.state
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

    onMoveItem?.({
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

  onMount(() => attachSidebarDragListeners())

  const SvelteComponent = $derived(icon)
</script>

<svelte:document
  onkeydown={handleLocalEscape}
  onmouseup={handleGlobalMouseUp}
/>

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
      onclick={handleHeaderClick}
      onkeydown={handleHeaderKeydown}
      oncontextmenu={openContextMenu}
      class="group flex min-w-0 flex-1 items-center gap-2.5 rounded-md h-8 px-3 py-3 text-sm transition-colors duration-150 {collapsed
        ? 'justify-center'
        : 'justify-between'} {expanded || highlighted
        ? 'text-white bg-fill-hover'
        : 'text-muted-foreground hover:text-white hover:bg-fill-subtle'} {isHeaderDropTarget
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
          <SvelteComponent size={collapsed ? 18 : 15} />
        </span>

        {#if !collapsed}
          {#if renameMode}
            <input
              bind:this={renameInputElement}
              bind:value={renameValue}
              class="h-7 w-full min-w-0 rounded border border-primary/50 bg-fill-hover px-2 text-[13px] font-medium text-white outline-none"
              placeholder={t('sidebar.sectionNamePlaceholder')}
              onclick={(event) => event.stopPropagation()}
              onmousedown={(event) => event.stopPropagation()}
              onkeydown={handleRenameKeydown}
              onblur={submitRename}
            />
          {:else}
            <span class="truncate text-[13px] font-medium">
              {section.label}
            </span>
          {/if}
        {/if}
      </div>

      {#if !collapsed}
        <div class="relative flex items-center justify-end shrink-0 w-14 h-6">
          <div class="flex items-center gap-0.5 transition-opacity opacity-100">
            <IconButton
              icon={Plus}
              size="xs"
              class="z-50 opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
              label={t('sidebar.ariaLabels.addNewBlock')}
              onclick={openSectionAddFromButton}
            />
            <IconButton
              icon={Ellipsis}
              size="xs"
              class="z-50 opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
              label={t('sidebar.ariaLabels.sectionOptions')}
              onclick={openSectionMenuFromButton}
            />
          </div>
          <span
            class="absolute right-0 inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-mono font-medium border border-border bg-fill-hover text-muted-foreground transition-opacity {contextMenuVisible
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
            class="rounded-md px-2.5 py-2 text-[12px] text-muted-foreground bg-fill-subtle border border-border/50"
          >
            {t('sidebar.noConnections')}
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
            class="group relative flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[13px] transition-all duration-150 text-fg-muted hover:text-white hover:bg-fill-hover {draggedItemId ===
            item.id
              ? 'opacity-40 bg-fill-hover scale-[0.99] pointer-events-none'
              : ''}"
            oncontextmenu={(event) => openItemMenuContext(event, item.id)}
          >
            {#if dropIndicator && dropIndicator.kind === 'before' && dropIndicator.itemId === item.id}
              <span
                class="pointer-events-none absolute -top-px left-2 right-2 h-0.5 rounded-full bg-primary shadow-[0_0_0_3px_rgba(34,197,94,0.25)]"
              ></span>
            {/if}

            <button
              type="button"
              class="-m-1.5 shrink-0 rounded-md p-1.5 text-muted-foreground/55 transition-colors hover:bg-fill-hover group-hover:text-muted-foreground cursor-grab active:cursor-grabbing"
              onmousedown={(event) => handleItemDragStart(event, item.id)}
              aria-label={t('sidebar.ariaLabels.dragBlock')}
            >
              <GripVertical size={13} />
            </button>

            {#if isApiItem(item.type, item.method)}
              <span
                class="px-1.5 py-0.5 text-[9px] rounded font-medium {getMethodBadgeStyle(
                  item.method,
                )}"
              >
                {item.method}
              </span>
            {:else}
              <span
                class="h-1.5 w-1.5 rounded-full shrink-0 transition-colors bg-fill-strong group-hover:bg-primary/60"
              ></span>
            {/if}

            {#if itemRenameId === item.id}
              <input
                bind:this={itemRenameInputElement}
                bind:value={itemRenameValue}
                class="h-7 w-full min-w-0 rounded border border-primary/50 bg-fill-hover px-2 text-[12px] text-white outline-none"
                placeholder={t('sidebar.blockNamePlaceholder')}
                onclick={(event) => event.stopPropagation()}
                onmousedown={(event) => event.stopPropagation()}
                onkeydown={(event) => handleItemRenameKeydown(event, item.id)}
                onblur={() => submitItemRename(item.id)}
              />
            {:else}
              <span class="truncate flex-1">{item.name}</span>
              <IconButton
                icon={Ellipsis}
                size="xs"
                class="pointer-events-auto"
                label={t('sidebar.ariaLabels.blockOptions')}
                onclick={(event) => openItemMenu(event, item.id)}
                oncontextmenu={(event) => openItemMenuContext(event, item.id)}
              />
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
  triggerElement={contextMenuTrigger}
  group="sidebar-context-menu"
  panelClass={MENU_PANEL_CLASS}
  onClose={closeContextMenu}
>
  <div data-sidebar-context-menu="true" role="menu" tabindex="-1">
    {#if blockTypePickerVisible}
      <div
        class="py-3 flex flex-row items-stretch justify-center gap-4 min-w-56"
      >
        {#each newBlockPresets as preset (preset.key)}
          <button
            type="button"
            class="flex min-w-20 flex-col items-center gap-1.5 rounded-md px-3 py-2 outline-none transition-colors hover:bg-fill-hover focus-visible:bg-fill-hover"
            onclick={() => addNewBlock(preset)}
          >
            <span
              class="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10 text-primary"
            >
              <preset.icon stroke={2} />
            </span>
            <span class="text-xs text-white font-medium">{preset.label}</span>
          </button>
        {/each}
      </div>
    {:else}
      <MenuItem
        icon={Plus}
        label={t('sidebar.menu.addNewBlock')}
        onclick={openBlockTypePicker}
      />
      <MenuItem
        icon={Pencil}
        label={t('sidebar.menu.rename')}
        onclick={startRename}
      />
      <span class={MENU_SEPARATOR_CLASS}></span>
      <MenuItem
        icon={Trash2}
        danger
        label={t('sidebar.menu.deleteSection')}
        onclick={() => {
          closeContextMenu()
          onRemoveSection?.({ sectionId: section.id })
        }}
      />
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
  panelClass={MENU_PANEL_CLASS}
  onClose={closeItemMenu}
>
  <div role="menu" tabindex="-1">
    <MenuItem
      icon={Pencil}
      label={t('sidebar.menu.rename')}
      onclick={startItemRenameFromMenu}
    />
    <span class={MENU_SEPARATOR_CLASS}></span>
    <MenuItem
      icon={Trash2}
      label={t('sidebar.menu.delete')}
      danger
      onclick={deleteItemFromMenu}
    />
  </div>
</FloatingDropdown>

<style>
  :global(.velarvo-noselect) {
    user-select: none !important;
  }
</style>
