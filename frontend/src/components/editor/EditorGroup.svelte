<script lang="ts">
  import { onDestroy, tick } from 'svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import {
    Plus,
    SplitSquareHorizontal,
    SplitSquareVertical,
    X,
  } from '@lucide/svelte'
  import { flip } from 'svelte/animate'
  import SSHConnectionForm from '@/components/ssh/SSHConnectionForm.svelte'
  import SSHTerminal from '@/components/ssh/SSHTerminal.svelte'
  import SessionLauncher from '@/components/ssh/SessionLauncher.svelte'
  import SFTPBrowser from '@/components/sftp/SFTPBrowser.svelte'
  import { t } from '@/lib/i18n.svelte'
  import { prefersReducedMotion } from 'svelte/motion'
  import { MOTION, motion } from '@/lib/motion'
  import {
    activateTab,
    closeGroup,
    closeTab,
    EditorTabKind,
    moveTab,
    openContentTab,
    openSession,
    setActiveGroup,
    splitGroup,
    splitGroupMoveTab,
    splitGroupWithTab,
    type GroupNode,
    type SplitDirection,
    editor,
  } from '@/lib/editorLayout.svelte'
  import { terminals } from '@/lib/sshTerminalSessions.svelte'
  import { TAB_KIND_ICON, TAB_KIND_LABEL } from '@/lib/editorTabPresentation'
  import { TERMINAL_STATUS_TONE } from '@/types/ssh'

  interface Props {
    group: GroupNode
    canClose?: boolean
  }

  let { group, canClose = true }: Props = $props()

  type DropZone = SplitDirection | 'center'

  let dropZone = $state<DropZone | null>(null)
  let tabDropIndex = $state<number | null>(null)

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

  let dropZoneClass = $derived(
    dropZone === 'left'
      ? 'inset-y-2 left-2 right-1/2'
      : dropZone === 'right'
        ? 'inset-y-2 right-2 left-1/2'
        : dropZone === 'up'
          ? 'inset-x-2 top-2 bottom-1/2'
          : dropZone === 'down'
            ? 'inset-x-2 bottom-2 top-1/2'
            : 'inset-2',
  )

  let isActiveGroup = $derived(editor.layout.activeGroupId === group.id)
  let tabStrip = $state<HTMLDivElement | null>(null)

  $effect(() => {
    const activeId = group.activeTabId
    const strip = tabStrip
    if (!activeId || !strip) return
    void tick().then(() => {
      strip
        .querySelector<HTMLElement>(`[data-tab-id="${CSS.escape(activeId)}"]`)
        ?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
    })
  })

  $effect.pre(() => {
    if (!editor.draggingTab && !editor.draggingConnection) {
      tabDropIndex = null
      dropZone = null
    }
  })

  const handleNewSession = () => {
    setActiveGroup(group.id)
    openSession(t('editor.newSessionTitle', 'Session'))
  }

  let dragImageEl: HTMLElement | null = null
  let sourceCollapsed = $state(false)

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
    editor.draggingTab = {
      fromGroupId: group.id,
      tabId: tab.id,
      title: tab.title,
    }
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData('text/plain', tab.id)
      dragImageEl = buildDragImage(tab.title)
      event.dataTransfer.setDragImage(dragImageEl, 14, 16)
    }
    requestAnimationFrame(() => (sourceCollapsed = true))
  }

  const onTabDragEnd = () => {
    editor.draggingTab = null
    tabDropIndex = null
    dropZone = null
    sourceCollapsed = false
    clearDragImage()
  }

  let dragLabel = $derived(
    editor.draggingConnection?.name ?? editor.draggingTab?.title ?? '',
  )
  let isDragActive = $derived(
    Boolean(editor.draggingTab || editor.draggingConnection),
  )

  let placeholderIndex = $derived(
    isDragActive && tabDropIndex !== null ? tabDropIndex : null,
  )

  const onTabSlotDragOver = (event: DragEvent, index: number) => {
    if (!editor.draggingTab && !editor.draggingConnection) return
    event.preventDefault()
    event.stopPropagation()
    tabDropIndex = index
    dropZone = null
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = editor.draggingConnection
        ? 'copy'
        : 'move'
    }
  }

  const onTabBarDrop = (event: DragEvent, index: number) => {
    event.preventDefault()
    event.stopPropagation()
    tabDropIndex = null

    const connection = editor.draggingConnection
    if (connection) {
      editor.draggingConnection = null
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

    const dragged = editor.draggingTab
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
    if (!editor.draggingTab && !editor.draggingConnection) return
    event.preventDefault()
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = editor.draggingConnection
        ? 'copy'
        : 'move'
    }
    tabDropIndex = null
    dropZone = computeDropZone(event)
  }

  const onBodyDrop = (event: DragEvent) => {
    event.preventDefault()
    const zone = dropZone ?? computeDropZone(event)
    dropZone = null

    const connection = editor.draggingConnection
    if (connection) {
      editor.draggingConnection = null
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

    const dragged = editor.draggingTab
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
  onpointerdowncapture={() => setActiveGroup(group.id)}
>
  <div
    class="flex h-9 min-w-0 shrink-0 items-stretch border-b border-border bg-surface-base {isActiveGroup
      ? ''
      : 'opacity-90'}"
  >
    <div
      bind:this={tabStrip}
      class="tab-strip-scroll flex min-w-0 flex-1 items-stretch overflow-x-auto overflow-y-hidden"
      onwheel={onTabStripWheel}
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
          editor.draggingTab?.tabId === tab.id &&
          editor.draggingTab?.fromGroupId === group.id}
        {@const terminalStatus =
          tab.kind === EditorTabKind.SSHTerminal
            ? terminals.statuses[tab.id]
            : undefined}
        {@const SvelteComponent = TAB_KIND_ICON[tab.kind]}
        <div
          class="flex h-full shrink-0 items-stretch {isDraggedTab &&
          sourceCollapsed
            ? 'hidden'
            : ''}"
          role="presentation"
          animate:flip={{
            duration: motion(MOTION.layout, prefersReducedMotion.current),
          }}
          ondragover={(event) => onTabSlotDragOver(event, index)}
          ondrop={(event) => onTabBarDrop(event, index)}
        >
          {#if placeholderIndex === index}
            {@render dropGhost()}
          {/if}
          <div
            role="tab"
            tabindex="-1"
            data-tab-id={tab.id}
            aria-selected={isActiveTab}
            draggable="true"
            class="group/tab relative flex h-full max-w-[12rem] shrink-0 cursor-pointer items-center gap-2 border-r border-border pl-3 pr-2 text-xs transition-[background-color,opacity] {isActiveTab
              ? 'bg-surface-canvas text-white'
              : 'bg-surface-base text-muted-foreground hover:bg-fill-subtle hover:text-fg-muted'} {isDraggedTab
              ? 'opacity-40'
              : ''}"
            onclick={() => activateTab(group.id, tab.id)}
            onmousedown={(event) => {
              if (event.button === 1) event.preventDefault()
            }}
            onauxclick={(event) => {
              if (event.button === 1) {
                event.preventDefault()
                closeTab(group.id, tab.id)
              }
            }}
            onkeydown={(event) =>
              event.key === 'Enter' && activateTab(group.id, tab.id)}
            ondragstart={(event) => onTabDragStart(event, tab)}
            ondragend={onTabDragEnd}
            title="{tab.title} · {t(...TAB_KIND_LABEL[tab.kind])}"
          >
            {#if isActiveTab && isActiveGroup}
              <span class="absolute inset-x-0 top-0 h-0.5 bg-primary"></span>
            {/if}
            <SvelteComponent
              size={14}
              strokeWidth={2}
              class="shrink-0 {terminalStatus
                ? TERMINAL_STATUS_TONE[terminalStatus]
                : isActiveTab
                  ? 'text-primary'
                  : 'text-muted-foreground/70'}"
            />
            <span class="truncate">{tab.title}</span>
            <IconButton
              icon={X}
              size="xs"
              label={t('editor.closeTab', 'Close tab')}
              class={`ml-0.5 h-5! w-5! group-hover/tab:opacity-100 ${
                isActiveTab ? 'opacity-100' : 'opacity-0'
              }`}
              onclick={(event) => {
                event.stopPropagation()
                closeTab(group.id, tab.id)
              }}
            />
          </div>
        </div>
      {/each}

      <div
        class="flex min-w-6 flex-1 items-stretch"
        role="presentation"
        ondragover={(event) => onTabSlotDragOver(event, group.tabs.length)}
        ondrop={(event) => onTabBarDrop(event, group.tabs.length)}
      >
        {#if placeholderIndex === group.tabs.length}
          {@render dropGhost()}
        {/if}

        <IconButton
          icon={Plus}
          size="sm"
          class="ml-1 self-center"
          label={t('editor.newSession', 'New session')}
          onclick={handleNewSession}
        />
      </div>
    </div>

    <div class="flex shrink-0 items-center gap-0.5 border-l border-border px-1">
      <IconButton
        icon={SplitSquareHorizontal}
        size="sm"
        label={t('editor.splitRight', 'Split right')}
        onclick={() => splitGroup(group.id, 'right')}
      />
      <IconButton
        icon={SplitSquareVertical}
        size="sm"
        label={t('editor.splitDown', 'Split down')}
        onclick={() => splitGroup(group.id, 'down')}
      />
      {#if canClose}
        <IconButton
          icon={X}
          size="sm"
          danger
          label={t('editor.closePane', 'Close pane')}
          onclick={() => closeGroup(group.id)}
        />
      {/if}
    </div>
  </div>

  <div
    class="relative flex min-h-0 flex-1 flex-col overflow-hidden"
    role="presentation"
    ondragover={onBodyDragOver}
    ondragleave={() => (dropZone = null)}
    ondrop={onBodyDrop}
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
          {:else if tab.kind === EditorTabKind.SFTP}
            <SFTPBrowser connectionId={tab.data.connectionId} />
          {:else}
            <SessionLauncher tabId={tab.id} />
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

<style>
  .tab-strip-scroll::-webkit-scrollbar {
    height: 3px;
  }

  .tab-strip-scroll::-webkit-scrollbar-track {
    background: transparent;
  }

  .tab-strip-scroll::-webkit-scrollbar-thumb {
    border-radius: 9999px;
    background: transparent;
  }

  .tab-strip-scroll:hover::-webkit-scrollbar-thumb {
    background: var(--color-line-strong);
  }
</style>
