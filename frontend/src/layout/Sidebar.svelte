<script lang="ts">
  import type { ComponentType } from 'svelte'
  import { fly } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import { Database, Globe, Plus, Search, Terminal } from 'lucide-svelte'
  import ConnectionRow from '@/components/ssh/ConnectionRow.svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import { translate } from '@/lib/i18n'
  import { navigate, route } from '@/lib/router'
  import {
    selectSSHConnection,
    selectedSSHConnectionId,
    sshConnections,
  } from '@/lib/sshConnections'
  import { matchesSSHQuery } from '@/lib/sshPresentation'
  import {
    draggingConnection,
    EditorResourceKind,
    EditorTabKind,
    openContentTab,
  } from '@/lib/editorLayout'
  import type { SSHConnection } from '@/types/ssh'
  import type { SidebarItemDraft } from '@/types/sidebar'

  import SidebarAddSection from './sidebar/SidebarAddSection.svelte'
  import SidebarSection from './sidebar/SidebarSection.svelte'
  import { getMethodBadgeStyle, getSectionIcon } from './sidebar/sectionMeta'
  import { sidebarSectionsStore } from './sidebar/sidebarStore'

  export let collapsed = false

  type SidebarMode = 'ssh' | 'db' | 'api'

  type RailItem = {
    id: SidebarMode
    label: string
    shortLabel: string
    description: string
    icon: ComponentType
  }

  type ResourcePanel = {
    eyebrow: string
    title: string
    searchPlaceholder: string
    empty: string
  }

  let activeMode: SidebarMode = 'api'
  let resourceSearch = ''
  let expanded: string[] = []
  let showAddSectionPanel = false
  let addSectionContainer: HTMLDivElement | null = null
  let initializedExpanded = false

  let flyoutOpen = false
  let flyoutTimer: ReturnType<typeof setTimeout> | null = null
  let flyoutSuppressed = false
  const FLYOUT_CLOSE_DELAY = 160

  const clearFlyoutTimer = () => {
    if (flyoutTimer) {
      clearTimeout(flyoutTimer)
      flyoutTimer = null
    }
  }

  const openFlyout = () => {
    if (!collapsed || flyoutSuppressed) return
    clearFlyoutTimer()
    flyoutOpen = true
  }

  const dismissFlyoutForDrag = () => {
    if (!flyoutOpen && !flyoutTimer) return
    flyoutSuppressed = true
    clearFlyoutTimer()
    flyoutOpen = false
  }

  const scheduleFlyoutClose = () => {
    clearFlyoutTimer()
    flyoutTimer = setTimeout(() => {
      flyoutOpen = false
      flyoutTimer = null
    }, FLYOUT_CLOSE_DELAY)
  }

  $: if (!collapsed && (flyoutOpen || flyoutTimer)) {
    clearFlyoutTimer()
    flyoutOpen = false
  }

  let railItems: RailItem[] = []
  let resourcePanels: Record<Exclude<SidebarMode, 'api'>, ResourcePanel>

  $: railItems = [
    {
      id: 'ssh',
      label: $translate('sidebar.rail.ssh.label'),
      shortLabel: $translate('sidebar.rail.ssh.shortLabel'),
      description: $translate('sidebar.rail.ssh.description'),
      icon: Terminal,
    },
    {
      id: 'db',
      label: $translate('sidebar.rail.databases.label'),
      shortLabel: $translate('sidebar.rail.databases.shortLabel'),
      description: $translate('sidebar.rail.databases.description'),
      icon: Database,
    },
    {
      id: 'api',
      label: $translate('sidebar.rail.api.label'),
      shortLabel: $translate('sidebar.rail.api.shortLabel'),
      description: $translate('sidebar.rail.api.description'),
      icon: Globe,
    },
  ]

  $: resourcePanels = {
    ssh: {
      eyebrow: $translate('sidebar.panels.ssh.eyebrow'),
      title: $translate('sidebar.panels.ssh.title'),
      searchPlaceholder: $translate('sidebar.panels.ssh.searchPlaceholder'),
      empty: $translate('sidebar.panels.ssh.empty'),
    },
    db: {
      eyebrow: $translate('sidebar.panels.databases.eyebrow'),
      title: $translate('sidebar.panels.databases.title'),
      searchPlaceholder: $translate(
        'sidebar.panels.databases.searchPlaceholder',
      ),
      empty: $translate('sidebar.panels.databases.empty'),
    },
  }

  $: sections = $sidebarSectionsStore
  $: activeRailItem =
    railItems.find((item) => item.id === activeMode) ?? railItems[0]
  $: activeResourcePanel =
    activeMode === 'api' ? null : resourcePanels[activeMode]
  $: normalizedResourceSearch = resourceSearch.trim().toLowerCase()
  $: sshConnectionList = $sshConnections.filter((connection) =>
    matchesSSHQuery(connection, normalizedResourceSearch),
  )
  $: hasSSHMatches = sshConnectionList.length > 0

  $: if (!initializedExpanded && sections.length > 0) {
    expanded = sections.map((section) => section.id)
    initializedExpanded = true
  }

  $: if (initializedExpanded) {
    const allowedIds = new Set(sections.map((section) => section.id))
    expanded = expanded.filter((sectionId) => allowedIds.has(sectionId))
  }

  const selectMode = (mode: SidebarMode) => {
    activeMode = mode
    resourceSearch = ''
    showAddSectionPanel = false

    if ($route !== '/dashboard') {
      navigate('/dashboard')
    }
  }

  const goToDashboard = () => {
    if ($route !== '/dashboard') navigate('/dashboard')
  }

  const handleCreateSSHConnection = () => {
    activeMode = 'ssh'
    resourceSearch = ''
    openContentTab(
      {
        title: 'New SSH connection',
        kind: EditorTabKind.SSHConnectionForm,
        data: {},
      },
      {
        isSame: (tab) =>
          tab.kind === EditorTabKind.SSHConnectionForm &&
          tab.data.connectionId === undefined,
      },
    )
    goToDashboard()
  }

  const handleSelectSSHConnection = (id: string) => {
    activeMode = 'ssh'
    selectSSHConnection(id)

    const connection = $sshConnections.find((item) => item.id === id)
    if (connection) {
      openContentTab(
        {
          title: connection.name,
          kind: EditorTabKind.SSHConnectionForm,
          data: { connectionId: id },
        },
        {
          isSame: (tab) =>
            tab.kind === EditorTabKind.SSHConnectionForm &&
            tab.data.connectionId === id,
        },
      )
    }
    goToDashboard()
  }

  const handleOpenSSHTerminal = (id: string) => {
    activeMode = 'ssh'
    selectSSHConnection(id)

    const connection = $sshConnections.find((item) => item.id === id)
    if (connection) {
      openContentTab({
        title: `${connection.name} · ${$translate('ssh.terminal', 'Terminal')}`,
        kind: EditorTabKind.SSHTerminal,
        data: { connectionId: id },
      })
    }
    goToDashboard()
  }

  let clickTimer: ReturnType<typeof setTimeout> | null = null
  const CLICK_DELAY = 220

  const onConnectionClick = (id: string) => {
    if (clickTimer) clearTimeout(clickTimer)
    clickTimer = setTimeout(() => {
      clickTimer = null
      handleSelectSSHConnection(id)
    }, CLICK_DELAY)
  }

  const onConnectionDblClick = (id: string) => {
    if (clickTimer) {
      clearTimeout(clickTimer)
      clickTimer = null
    }
    handleOpenSSHTerminal(id)
  }

  const onConnectionDragStart = (
    event: DragEvent,
    connection: SSHConnection,
  ) => {
    draggingConnection.set({
      kind: EditorResourceKind.SSHConnection,
      id: connection.id,
      name: connection.name,
    })
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'copy'
      event.dataTransfer.setData('text/plain', connection.id)
    }
    dismissFlyoutForDrag()
  }

  const onConnectionDragEnd = () => {
    draggingConnection.set(null)
    flyoutSuppressed = false
  }

  const toggleSection = (sectionId: string) => {
    expanded = expanded.includes(sectionId)
      ? expanded.filter((item) => item !== sectionId)
      : [...expanded, sectionId]
  }

  const addCustomSection = (label: string) => {
    const created = sidebarSectionsStore.addSection(label)
    if (!created) return
    expanded = [...expanded, created.id]
    showAddSectionPanel = false
  }

  const removeSection = (sectionId: string) => {
    sidebarSectionsStore.removeSection(sectionId)
    expanded = expanded.filter((item) => item !== sectionId)
  }

  const renameSection = (
    event: CustomEvent<{ sectionId: string; label: string }>,
  ) => {
    sidebarSectionsStore.renameSection(
      event.detail.sectionId,
      event.detail.label,
    )
  }

  const addSectionItem = (
    event: CustomEvent<{ sectionId: string } & SidebarItemDraft>,
  ) => {
    sidebarSectionsStore.addItem(event.detail.sectionId, {
      name: event.detail.name,
      type: event.detail.type,
      status: event.detail.status,
      method: event.detail.method,
    })
  }

  const removeSectionItem = (
    event: CustomEvent<{ sectionId: string; itemId: string }>,
  ) => {
    sidebarSectionsStore.removeItem(event.detail.sectionId, event.detail.itemId)
  }

  const renameSectionItem = (
    event: CustomEvent<{ sectionId: string; itemId: string; name: string }>,
  ) => {
    sidebarSectionsStore.renameItem(
      event.detail.sectionId,
      event.detail.itemId,
      event.detail.name,
    )
  }

  const reorderSectionItem = (
    event: CustomEvent<{ sectionId: string; itemId: string; toIndex: number }>,
  ) => {
    sidebarSectionsStore.reorderItem(
      event.detail.sectionId,
      event.detail.itemId,
      event.detail.toIndex,
    )
  }

  const moveSectionItem = (
    event: CustomEvent<{
      fromSectionId: string
      itemId: string
      toSectionId: string
      toIndex: number
    }>,
  ) => {
    sidebarSectionsStore.moveItem(
      event.detail.fromSectionId,
      event.detail.itemId,
      event.detail.toSectionId,
      event.detail.toIndex,
    )
  }
</script>

<aside
  class="relative z-200 flex h-full shrink-0 overflow-visible border-r border-sidebar-border bg-sidebar-background transition-all duration-300 ease-in-out {collapsed
    ? 'w-[4.5rem]'
    : 'w-[23rem]'}"
>
  <nav
    class="flex w-[4.5rem] shrink-0 flex-col items-center gap-2 border-r border-sidebar-border bg-sidebar-rail px-2 py-4"
    aria-label={$translate('sidebar.resourceTypes')}
    on:mouseenter={openFlyout}
    on:mouseleave={scheduleFlyoutClose}
  >
    {#each railItems as item (item.id)}
      <button
        type="button"
        class="group relative flex h-11 w-11 items-center justify-center rounded-xl transition-all duration-150 {activeMode ===
        item.id
          ? 'bg-primary/12 text-primary'
          : 'text-muted-foreground hover:bg-white/5 hover:text-white'}"
        aria-label={item.label}
        aria-pressed={activeMode === item.id}
        on:click={() => selectMode(item.id)}
      >
        {#if activeMode === item.id}
          <span
            class="absolute -left-2 h-6 w-0.5 rounded-r-full bg-primary shadow-[0_0_12px_rgb(var(--color-primary-rgb)_/_0.65)]"
          ></span>
        {/if}

        <svelte:component this={item.icon} size={19} strokeWidth={2.1} />

        <span
          class="pointer-events-none absolute left-[3.2rem] z-500 hidden min-w-max rounded-lg border border-border bg-sidebar-tooltip px-2.5 py-1.5 text-xs font-medium text-white shadow-xl shadow-black/40 group-hover:block"
        >
          {item.shortLabel}
        </span>
      </button>
    {/each}
  </nav>

  {#snippet panelBody()}
    <div class="border-b border-border px-4 py-4">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p
            class="text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground"
          >
            {activeMode === 'api'
              ? $translate('sidebar.collections')
              : activeResourcePanel?.eyebrow}
          </p>
          <div class="mt-1 flex items-center gap-2">
            <svelte:component
              this={activeRailItem.icon}
              size={17}
              class="shrink-0 text-primary"
            />
            <h2 class="truncate text-sm font-semibold text-white">
              {activeMode === 'api'
                ? $translate('sidebar.rail.api.label')
                : activeResourcePanel?.title}
            </h2>
          </div>
          <p class="mt-1 truncate text-xs text-muted-foreground">
            {activeRailItem.description}
          </p>
        </div>

        {#if activeMode === 'api'}
          <div class="relative shrink-0" bind:this={addSectionContainer}>
            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-primary/25 bg-primary text-background shadow-[0_10px_28px_rgb(var(--color-primary-rgb)_/_0.18)] transition hover:bg-primary/90"
              on:click={() => (showAddSectionPanel = !showAddSectionPanel)}
              aria-label={$translate('sidebar.addSection')}
            >
              <Plus size={16} strokeWidth={2.2} />
            </button>
          </div>

          <FloatingDropdown
            open={showAddSectionPanel}
            anchorElement={addSectionContainer}
            placement="bottom-end"
            offset={8}
            group="sidebar-add-section"
            panelClass="rounded-lg border border-border bg-sidebar-popover shadow-2xl shadow-black/60 p-2.5"
            on:close={() => (showAddSectionPanel = false)}
          >
            <SidebarAddSection
              on:addCustom={(event) => addCustomSection(event.detail.label)}
            />
          </FloatingDropdown>
        {:else if activeMode === 'ssh'}
          <button
            type="button"
            class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-primary/25 bg-primary text-background shadow-[0_10px_28px_rgb(var(--color-primary-rgb)_/_0.18)] transition hover:bg-primary/90"
            aria-label="Create SSH connection"
            on:click={handleCreateSSHConnection}
          >
            <Plus size={16} strokeWidth={2.2} />
          </button>
        {/if}
      </div>

      {#if activeMode !== 'api'}
        <label
          class="mt-4 flex h-10 items-center gap-2 rounded-xl bg-white/4 px-3 text-muted-foreground transition focus-within:bg-white/6 focus-within:ring-1 focus-within:ring-primary/30"
        >
          <Search size={15} class="shrink-0" />
          <input
            bind:value={resourceSearch}
            class="h-full min-w-0 flex-1 bg-transparent text-sm text-white outline-none placeholder:text-muted-foreground"
            placeholder={activeResourcePanel?.searchPlaceholder}
          />
        </label>
      {/if}
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto overflow-x-hidden px-2 py-3">
      {#if activeMode === 'api'}
        {#if sections.length === 0}
          <div
            class="mx-2 rounded-xl bg-white/2 px-4 py-6 text-center text-sm text-muted-foreground"
          >
            {$translate('sidebar.empty')}
          </div>
        {/if}

        <div class="space-y-1">
          {#each sections as section (section.id)}
            <SidebarSection
              {section}
              collapsed={false}
              expanded={expanded.includes(section.id)}
              icon={getSectionIcon(section.kind)}
              {getMethodBadgeStyle}
              on:toggle={() => toggleSection(section.id)}
              on:removeSection={(event) =>
                removeSection(event.detail.sectionId)}
              on:renameSection={renameSection}
              on:addItem={addSectionItem}
              on:renameItem={renameSectionItem}
              on:reorderItem={reorderSectionItem}
              on:moveItem={moveSectionItem}
              on:removeItem={removeSectionItem}
            />
          {/each}
        </div>
      {:else if activeMode === 'ssh'}
        {#if !hasSSHMatches && !normalizedResourceSearch}
          <div
            class="mx-2 rounded-xl bg-white/2 px-4 py-6 text-center text-sm text-muted-foreground"
          >
            {activeResourcePanel?.empty}
          </div>
        {/if}

        {#if !hasSSHMatches && normalizedResourceSearch}
          <div
            class="mx-2 rounded-xl bg-white/2 px-4 py-6 text-center text-sm text-muted-foreground"
          >
            No SSH connections match your search.
          </div>
        {/if}

        <div class="space-y-1.5">
          {#each sshConnectionList as connection (connection.id)}
            {@const isActive = $selectedSSHConnectionId === connection.id}
            <button
              type="button"
              draggable="true"
              class="group relative w-full rounded-lg px-3 py-3 text-left transition-all duration-150 {isActive
                ? 'bg-primary/10'
                : 'bg-transparent hover:bg-white/4'}"
              on:click={() => onConnectionClick(connection.id)}
              on:dblclick={() => onConnectionDblClick(connection.id)}
              on:dragstart={(event) => onConnectionDragStart(event, connection)}
              on:dragend={onConnectionDragEnd}
            >
              <ConnectionRow {connection} active={isActive} />
            </button>
          {/each}
        </div>
      {:else if activeResourcePanel}
        <div
          class="mx-2 rounded-xl bg-white/2 px-4 py-6 text-center text-sm text-muted-foreground"
        >
          {activeResourcePanel.empty}
        </div>
      {/if}
    </div>
  {/snippet}

  {#if !collapsed}
    <section class="flex min-w-0 flex-1 flex-col bg-sidebar-panel">
      {@render panelBody()}
    </section>
  {/if}

  {#if collapsed && flyoutOpen}
    <div
      role="presentation"
      on:mouseenter={openFlyout}
      on:mouseleave={scheduleFlyoutClose}
      transition:fly={{ x: -16, duration: 200, easing: cubicOut }}
      class="absolute left-[4.5rem] top-3 z-500 flex h-[calc(100%-4.5rem)] w-[min(21rem,calc(100vw-6rem))] flex-col overflow-hidden rounded-r-2xl border border-l-0 border-sidebar-border bg-sidebar-panel shadow-[0_30px_80px_-12px_rgba(0,0,0,0.85)]"
    >
      <section
        class="flex min-w-0 flex-1 flex-col overflow-hidden bg-sidebar-panel"
      >
        {@render panelBody()}
      </section>
    </div>
  {/if}
</aside>
