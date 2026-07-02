<script lang="ts">
  import { onDestroy } from 'svelte'
  import {
    Plus,
    SplitSquareHorizontal,
    SplitSquareVertical,
    X,
  } from 'lucide-svelte'
  import { flip } from 'svelte/animate'
  import SSHConnectionForm from '@/components/ssh/SSHConnectionForm.svelte'
  import SSHTerminal from '@/components/ssh/SSHTerminal.svelte'
  import SessionLauncher from '@/components/ssh/SessionLauncher.svelte'
  import { translate } from '@/lib/i18n'
  import { MOTION, motion, prefersReducedMotion } from '@/lib/motion'
  import {
    activateTab,
    closeGroup,
    closeTab,
    draggingConnection,
    draggingTab,
    EditorTabKind,
    editorState,
    moveTab,
    openContentTab,
    openSession,
    setActiveGroup,
    splitGroup,
    splitGroupMoveTab,
    splitGroupWithTab,
    type GroupNode,
    type SplitDirection,
  } from '@/lib/editorLayout'
  import { terminalStatuses } from '@/lib/sshTerminalSessions'
  import { TERMINAL_STATUS_INDICATOR } from '@/types/ssh'

  export let group: GroupNode
  export let canClose = true

  type DropZone = SplitDirection | 'center'

  let dropZone: DropZone | null = null
  let tabDropIndex: number | null = null

  const EDGE_FRACTION = 0.22

  const computeDropZone = (event: DragEvent): DropZone => {
    const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
    if (rect.width === 0 || rect.height === 0) return 'center'

    const distances: Record<SplitDirection, number> = {
      left: (event.clientX - rect.left) / rect.width,
      right: (rect.right - event.clientX) / rect.width,
      up: (event.clientY - rect.top) / rect.height,
      down: (rect.bottom - event.clientY) / rect.height,
    }

    let nearest: SplitDirection = 'right'
    let min = Infinity
    for (const key of Object.keys(distances) as SplitDirection[]) {
      if (distances[key] < min) {
        min = distances[key]
        nearest = key
      }
    }

    return min < EDGE_FRACTION ? nearest : 'center'
  }

  $: dropZoneClass =
    dropZone === 'left'
      ? 'inset-y-2 left-2 right-1/2'
      : dropZone === 'right'
        ? 'inset-y-2 right-2 left-1/2'
        : dropZone === 'up'
          ? 'inset-x-2 top-2 bottom-1/2'
          : dropZone === 'down'
            ? 'inset-x-2 bottom-2 top-1/2'
            : 'inset-2'

  $: isActiveGroup = $editorState.activeGroupId === group.id

  $: if (!$draggingTab && !$draggingConnection) {
    tabDropIndex = null
    dropZone = null
  }

  const handleNewSession = () => {
    setActiveGroup(group.id)
    openSession($translate('editor.newSessionTitle', 'Session'))
  }

  let dragImageEl: HTMLElement | null = null
  let sourceCollapsed = false

  const buildDragImage = (title: string): HTMLElement => {
    const el = document.createElement('div')
    el.className = 'velarvo-tab-drag-image'
    el.textContent = title
    document.body.appendChild(el)
    return el
  }

  const clearDragImage = () => {
    dragImageEl?.remove()
    dragImageEl = null
  }

  onDestroy(clearDragImage)

  const onTabDragStart = (
    event: DragEvent,
    tab: { id: string; title: string },
  ) => {
    draggingTab.set({ fromGroupId: group.id, tabId: tab.id, title: tab.title })
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData('text/plain', tab.id)
      dragImageEl = buildDragImage(tab.title)
      event.dataTransfer.setDragImage(dragImageEl, 14, 16)
    }
    requestAnimationFrame(() => (sourceCollapsed = true))
  }

  const onTabDragEnd = () => {
    draggingTab.set(null)
    tabDropIndex = null
    dropZone = null
    sourceCollapsed = false
    clearDragImage()
  }

  $: dragLabel = $draggingConnection?.name ?? $draggingTab?.title ?? ''
  $: isDragActive = Boolean($draggingTab || $draggingConnection)

  $: placeholderIndex =
    isDragActive && tabDropIndex !== null ? tabDropIndex : null

  const onTabSlotDragOver = (event: DragEvent, index: number) => {
    if (!$draggingTab && !$draggingConnection) return
    event.preventDefault()
    event.stopPropagation()
    tabDropIndex = index
    dropZone = null
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = $draggingConnection ? 'copy' : 'move'
    }
  }

  const onTabBarDrop = (event: DragEvent, index: number) => {
    event.preventDefault()
    event.stopPropagation()
    tabDropIndex = null

    const connection = $draggingConnection
    if (connection) {
      draggingConnection.set(null)
      setActiveGroup(group.id)
      openContentTab(
        {
          title: connection.name,
          kind: EditorTabKind.SSHConnectionForm,
          data: { connectionId: connection.id },
        },
        {
          groupId: group.id,
          index,
          isSame: (tab) =>
            tab.kind === EditorTabKind.SSHConnectionForm &&
            tab.data.connectionId === connection.id,
        },
      )
      return
    }

    const dragged = $draggingTab
    if (!dragged) return
    moveTab(dragged.fromGroupId, dragged.tabId, group.id, index)
  }

  const onTabStripWheel = (event: WheelEvent) => {
    if (event.deltaY === 0) return
    const strip = event.currentTarget as HTMLElement
    if (strip.scrollWidth <= strip.clientWidth) return
    event.preventDefault()
    strip.scrollLeft += event.deltaY
  }

  const onBodyDragOver = (event: DragEvent) => {
    if (!$draggingTab && !$draggingConnection) return
    event.preventDefault()
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = $draggingConnection ? 'copy' : 'move'
    }
    tabDropIndex = null
    dropZone = computeDropZone(event)
  }

  const onBodyDrop = (event: DragEvent) => {
    event.preventDefault()
    const zone = dropZone ?? computeDropZone(event)
    dropZone = null

    const connection = $draggingConnection
    if (connection) {
      draggingConnection.set(null)
      const content = {
        title: connection.name,
        kind: EditorTabKind.SSHConnectionForm,
        data: { connectionId: connection.id },
      } as const

      if (zone === 'center') {
        setActiveGroup(group.id)
        openContentTab(content, {
          groupId: group.id,
          isSame: (tab) =>
            tab.kind === EditorTabKind.SSHConnectionForm &&
            tab.data.connectionId === connection.id,
        })
      } else {
        splitGroupWithTab(group.id, zone, content)
      }
      return
    }

    const dragged = $draggingTab
    if (!dragged) return
    if (zone === 'center') {
      moveTab(dragged.fromGroupId, dragged.tabId, group.id, group.tabs.length)
    } else {
      splitGroupMoveTab(dragged.fromGroupId, dragged.tabId, group.id, zone)
    }
  }
</script>

<section
  role="presentation"
  class="flex min-h-0 min-w-0 flex-1 flex-col bg-surface-canvas"
  on:pointerdown|capture={() => setActiveGroup(group.id)}
>
  <div
    class="flex h-9 shrink-0 items-stretch border-b border-border bg-surface-base {isActiveGroup
      ? ''
      : 'opacity-90'}"
  >
    <div
      class="tab-strip-scroll flex min-w-0 flex-1 items-stretch"
      on:wheel={onTabStripWheel}
    >
      {#snippet dropGhost()}
        <div
          class="pointer-events-none relative flex h-full max-w-[12rem] shrink-0 items-center gap-2 border-r border-border bg-surface-canvas pl-3 pr-2 text-xs text-white opacity-55"
        >
          <span class="absolute inset-x-0 top-0 h-0.5 bg-primary"></span>
          <span class="h-1.5 w-1.5 shrink-0 rounded-full bg-primary"></span>
          <span class="truncate">{dragLabel}</span>
          <span
            class="ml-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground"
          >
            <X size={13} strokeWidth={2.4} />
          </span>
        </div>
      {/snippet}

      {#each group.tabs as tab, index (tab.id)}
        {@const isActiveTab = group.activeTabId === tab.id}
        {@const isDraggedTab =
          $draggingTab?.tabId === tab.id &&
          $draggingTab?.fromGroupId === group.id}
        {@const terminalStatus =
          tab.kind === EditorTabKind.SSHTerminal
            ? $terminalStatuses[tab.id]
            : undefined}
        <div
          class="flex h-full shrink-0 items-stretch {isDraggedTab &&
          sourceCollapsed
            ? 'hidden'
            : ''}"
          role="presentation"
          animate:flip={{
            duration: motion(MOTION.layout, $prefersReducedMotion),
          }}
          on:dragover={(event) => onTabSlotDragOver(event, index)}
          on:drop={(event) => onTabBarDrop(event, index)}
        >
          {#if placeholderIndex === index}
            {@render dropGhost()}
          {/if}
          <div
            role="tab"
            tabindex="-1"
            aria-selected={isActiveTab}
            draggable="true"
            class="group/tab relative flex h-full max-w-[12rem] shrink-0 cursor-pointer select-none items-center gap-2 border-r border-border pl-3 pr-2 text-xs transition-[background-color,opacity] {isActiveTab
              ? 'bg-surface-canvas text-white'
              : 'bg-surface-base text-muted-foreground hover:bg-white/[0.03] hover:text-white/80'} {isDraggedTab
              ? 'opacity-40'
              : ''}"
            on:click={() => activateTab(group.id, tab.id)}
            on:mousedown={(event) => {
              if (event.button === 1) event.preventDefault()
            }}
            on:auxclick={(event) => {
              if (event.button === 1) {
                event.preventDefault()
                closeTab(group.id, tab.id)
              }
            }}
            on:keydown={(event) =>
              event.key === 'Enter' && activateTab(group.id, tab.id)}
            on:dragstart={(event) => onTabDragStart(event, tab)}
            on:dragend={onTabDragEnd}
          >
            {#if isActiveTab && isActiveGroup}
              <span class="absolute inset-x-0 top-0 h-0.5 bg-primary"></span>
            {/if}
            <span
              class="h-1.5 w-1.5 shrink-0 rounded-full {terminalStatus
                ? TERMINAL_STATUS_INDICATOR[terminalStatus]
                : isActiveTab
                  ? 'bg-primary'
                  : 'bg-muted-foreground/40'}"
            ></span>
            <span class="truncate">{tab.title}</span>
            <button
              type="button"
              class="ml-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground opacity-0 transition hover:bg-white/10 hover:text-white group-hover/tab:opacity-100 {isActiveTab
                ? 'opacity-100'
                : ''}"
              aria-label={$translate('editor.closeTab', 'Close tab')}
              on:click|stopPropagation={() => closeTab(group.id, tab.id)}
            >
              <X size={13} strokeWidth={2.4} />
            </button>
          </div>
        </div>
      {/each}

      <div
        class="flex min-w-6 flex-1 items-stretch"
        role="presentation"
        on:dragover={(event) => onTabSlotDragOver(event, group.tabs.length)}
        on:drop={(event) => onTabBarDrop(event, group.tabs.length)}
      >
        {#if placeholderIndex === group.tabs.length}
          {@render dropGhost()}
        {/if}

        <button
          type="button"
          class="ml-1 inline-flex h-7 w-7 shrink-0 items-center justify-center self-center rounded text-muted-foreground transition hover:bg-white/5 hover:text-white"
          aria-label={$translate('editor.newSession', 'New session')}
          title={$translate('editor.newSession', 'New session')}
          on:click={handleNewSession}
        >
          <Plus size={15} strokeWidth={2.2} />
        </button>
      </div>
    </div>

    <div class="flex shrink-0 items-center gap-0.5 border-l border-border px-1">
      <button
        type="button"
        class="inline-flex h-7 w-7 items-center justify-center rounded text-muted-foreground transition hover:bg-white/5 hover:text-white"
        aria-label={$translate('editor.splitRight', 'Split right')}
        title={$translate('editor.splitRight', 'Split right')}
        on:click={() => splitGroup(group.id, 'right')}
      >
        <SplitSquareHorizontal size={15} strokeWidth={2.1} />
      </button>
      <button
        type="button"
        class="inline-flex h-7 w-7 items-center justify-center rounded text-muted-foreground transition hover:bg-white/5 hover:text-white"
        aria-label={$translate('editor.splitDown', 'Split down')}
        title={$translate('editor.splitDown', 'Split down')}
        on:click={() => splitGroup(group.id, 'down')}
      >
        <SplitSquareVertical size={15} strokeWidth={2.1} />
      </button>
      {#if canClose}
        <button
          type="button"
          class="inline-flex h-7 w-7 items-center justify-center rounded text-muted-foreground transition hover:bg-destructive/15 hover:text-red-300"
          aria-label={$translate('editor.closePane', 'Close pane')}
          title={$translate('editor.closePane', 'Close pane')}
          on:click={() => closeGroup(group.id)}
        >
          <X size={15} strokeWidth={2.2} />
        </button>
      {/if}
    </div>
  </div>

  <div
    class="relative flex min-h-0 flex-1 flex-col overflow-hidden"
    role="presentation"
    on:dragover={onBodyDragOver}
    on:dragleave={() => (dropZone = null)}
    on:drop={onBodyDrop}
  >
    {#if dropZone}
      <div
        class="pointer-events-none absolute z-20 rounded-lg border-2 border-dashed border-primary/50 bg-primary/10 transition-all duration-100 {dropZoneClass}"
      ></div>
    {/if}

    {#if group.tabs.length === 0}
      <SessionLauncher groupId={group.id} />
    {:else}
      {#each group.tabs as tab (tab.id)}
        <div
          class="absolute inset-0 flex min-h-0 flex-col {tab.id ===
          group.activeTabId
            ? ''
            : 'hidden'}"
        >
          {#if tab.kind === EditorTabKind.SSHConnectionForm}
            <SSHConnectionForm
              tabId={tab.id}
              groupId={group.id}
              connectionId={tab.data.connectionId ?? null}
            />
          {:else if tab.kind === EditorTabKind.SSHTerminal}
            <SSHTerminal paneId={tab.id} connectionId={tab.data.connectionId} />
          {:else}
            <SessionLauncher tabId={tab.id} />
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>
