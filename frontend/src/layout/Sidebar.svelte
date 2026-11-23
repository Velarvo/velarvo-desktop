<script lang="ts">
  import { tooltip } from '@/components/ui/tooltip'
  import type { IconComponent } from '@/types/icon'
  import IconButton from '@/components/ui/IconButton.svelte'
  import Input from '@/components/ui/Input.svelte'
  import { fly } from 'svelte/transition'
  import { cubicOut } from 'svelte/easing'
  import {
    Database,
    FolderTree,
    Globe,
    Plus,
    Search,
    Terminal,
  } from '@lucide/svelte'
  import ConnectionRow from '@/components/ssh/ConnectionRow.svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import { t } from '@/lib/i18n.svelte'
  import { navigate, router } from '@/lib/router.svelte'
  import { selectSSHConnection, sshState } from '@/lib/sshConnections.svelte'
  import { matchesSSHQuery } from '@/lib/sshPresentation'
  import {
    EditorResourceKind,
    EditorTabKind,
    openContentTab,
    openSFTPForConnection,
    editor,
  } from '@/lib/editorLayout.svelte'
  import type { SSHConnection } from '@/types/ssh'
  import type { SidebarItemDraft } from '@/types/sidebar'

  import SidebarAddSection from './sidebar/SidebarAddSection.svelte'
  import SidebarSection from './sidebar/SidebarSection.svelte'
  import { getMethodBadgeStyle, getSectionIcon } from './sidebar/sectionMeta'
  import { sidebarSections } from './sidebar/sidebarStore.svelte'

  interface Props {
    collapsed?: boolean
  }

  let { collapsed = false }: Props = $props()

  type SidebarMode = 'ssh' | 'db' | 'api'

  type RailItem = {
    id: SidebarMode
    label: string
    shortLabel: string
    icon: IconComponent
  }

  type ResourcePanel = {
    eyebrow: string
    title: string
    searchPlaceholder: string
    empty: string
  }

  let activeMode = $state<SidebarMode>('api')
  let resourceSearch = $state('')
  let expanded = $state<string[]>([])
  let showAddSectionPanel = $state(false)
  let addSectionContainer = $state<HTMLDivElement | null>(null)
  let initializedExpanded = $state(false)

  let flyoutOpen = $state(false)
  let flyoutTimer = $state<ReturnType<typeof setTimeout> | null>(null)
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

  $effect.pre(() => {
    if (!collapsed && (flyoutOpen || flyoutTimer)) {
      clearFlyoutTimer()
      flyoutOpen = false
    }
  })

  let resourcePanels: Record<
    Exclude<SidebarMode, 'api'>,
    ResourcePanel
  > = $derived({
    ssh: {
      eyebrow: t('sidebar.panels.ssh.eyebrow'),
      title: t('sidebar.panels.ssh.title'),
      searchPlaceholder: t('sidebar.panels.ssh.searchPlaceholder'),
      empty: t('sidebar.panels.ssh.empty'),
    },
    db: {
      eyebrow: t('sidebar.panels.databases.eyebrow'),
      title: t('sidebar.panels.databases.title'),
      searchPlaceholder: t('sidebar.panels.databases.searchPlaceholder'),
      empty: t('sidebar.panels.databases.empty'),
    },
  })

  const railItems: RailItem[] = $derived([
    {
      id: 'ssh',
      label: t('sidebar.rail.ssh.label'),
      shortLabel: t('sidebar.rail.ssh.shortLabel'),
      icon: Terminal,
    },
    {
      id: 'db',
      label: t('sidebar.rail.databases.label'),
      shortLabel: t('sidebar.rail.databases.shortLabel'),
      icon: Database,
    },
    {
      id: 'api',
      label: t('sidebar.rail.api.label'),
      shortLabel: t('sidebar.rail.api.shortLabel'),
      icon: Globe,
    },
  ])

  let sections = $derived(sidebarSections.sections)
  let activeRailItem = $derived(
    railItems.find((item) => item.id === activeMode) ?? railItems[0],
  )
  let activeResourcePanel = $derived(
    activeMode === 'api' ? null : resourcePanels[activeMode],
  )
  let normalizedResourceSearch = $derived(resourceSearch.trim().toLowerCase())
  let sshConnectionList = $derived(
    sshState.connections.filter((connection) =>
      matchesSSHQuery(connection, normalizedResourceSearch),
    ),
  )
  let hasSSHMatches = $derived(sshConnectionList.length > 0)

  $effect.pre(() => {
    if (!initializedExpanded && sections.length > 0) {
      expanded = sections.map((section) => section.id)
      initializedExpanded = true
    }
  })

  $effect.pre(() => {
    if (initializedExpanded) {
      const allowedIds = new Set(sections.map((section) => section.id))
      expanded = expanded.filter((sectionId) => allowedIds.has(sectionId))
    }
  })

  const selectMode = (mode: SidebarMode) => {
    activeMode = mode
    resourceSearch = ''
    showAddSectionPanel = false

    if (router.path !== '/dashboard') {
      navigate('/dashboard')
    }
  }

  const goToDashboard = () => {
    if (router.path !== '/dashboard') navigate('/dashboard')
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

    const connection = sshState.connections.find((item) => item.id === id)
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

    const connection = sshState.connections.find((item) => item.id === id)
    if (connection) {
      openContentTab({
        title: `${connection.name} · ${t('ssh.terminal', 'Terminal')}`,
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

  const openSFTP = (connection: SSHConnection) => {
    activeMode = 'ssh'
    selectSSHConnection(connection.id)
    openSFTPForConnection(
      connection.id,
      `${connection.name} · ${t('sftp.files', 'Files')}`,
    )
    goToDashboard()
  }

  const onConnectionDragStart = (
    event: DragEvent,
    connection: SSHConnection,
  ) => {
    editor.draggingConnection = {
      kind: EditorResourceKind.SSHConnection,
      id: connection.id,
      name: connection.name,
    }
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'copy'
      event.dataTransfer.setData('text/plain', connection.id)
    }
    dismissFlyoutForDrag()
  }

  const onConnectionDragEnd = () => {
    editor.draggingConnection = null
    flyoutSuppressed = false
  }

  const toggleSection = (sectionId: string) => {
    expanded = expanded.includes(sectionId)
      ? expanded.filter((item) => item !== sectionId)
      : [...expanded, sectionId]
  }

  const addCustomSection = (label: string) => {
    const created = sidebarSections.addSection(label)
    if (!created) return
    expanded = [...expanded, created.id]
    showAddSectionPanel = false
  }

  const removeSection = (sectionId: string) => {
    sidebarSections.removeSection(sectionId)
    expanded = expanded.filter((item) => item !== sectionId)
  }

  const renameSection = (detail: { sectionId: string; label: string }) => {
    sidebarSections.renameSection(detail.sectionId, detail.label)
  }

  const addSectionItem = (detail: { sectionId: string } & SidebarItemDraft) => {
    sidebarSections.addItem(detail.sectionId, {
      name: detail.name,
      type: detail.type,
      status: detail.status,
      method: detail.method,
    })
  }

  const removeSectionItem = (detail: { sectionId: string; itemId: string }) => {
    sidebarSections.removeItem(detail.sectionId, detail.itemId)
  }

  const renameSectionItem = (detail: {
    sectionId: string
    itemId: string
    name: string
  }) => {
    sidebarSections.renameItem(detail.sectionId, detail.itemId, detail.name)
  }

  const moveSectionItem = (detail: {
    fromSectionId: string
    itemId: string
    toSectionId: string
    toIndex: number
  }) => {
    sidebarSections.moveItem(
      detail.fromSectionId,
      detail.itemId,
      detail.toSectionId,
      detail.toIndex,
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
    aria-label={t('sidebar.resourceTypes')}
    onmouseenter={openFlyout}
    onmouseleave={scheduleFlyoutClose}
  >
    {#each railItems as item (item.id)}
      <button
        type="button"
        class="group relative flex h-10 w-10 items-center justify-center rounded-md transition-all duration-150 {activeMode ===
        item.id
          ? 'bg-primary/12 text-primary'
          : 'text-muted-foreground hover:bg-fill-hover hover:text-white'}"
        aria-label={item.label}
        aria-pressed={activeMode === item.id}
        onclick={() => selectMode(item.id)}
      >
        {#if activeMode === item.id}
          <span
            class="absolute -left-2 h-6 w-0.5 rounded-r-full bg-primary shadow-[0_0_12px_rgb(var(--color-primary-rgb)_/_0.65)]"
          ></span>
        {/if}

        <item.icon size={19} strokeWidth={2.1} />

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
              ? t('sidebar.collections')
              : activeResourcePanel?.eyebrow}
          </p>
          <div class="mt-1 flex items-center gap-2">
            <activeRailItem.icon size={17} class="shrink-0 text-primary" />
            <h2 class="truncate text-sm font-semibold text-white">
              {activeMode === 'api'
                ? t('sidebar.rail.api.label')
                : activeResourcePanel?.title}
            </h2>
          </div>
        </div>

        {#if activeMode === 'api'}
          <div class="relative shrink-0" bind:this={addSectionContainer}>
            <IconButton
              icon={Plus}
              variant="primary"
              label={t('sidebar.addSection')}
              pressed={showAddSectionPanel}
              onclick={() => (showAddSectionPanel = !showAddSectionPanel)}
            />
          </div>

          <FloatingDropdown
            open={showAddSectionPanel}
            anchorElement={addSectionContainer}
            placement="bottom-end"
            offset={8}
            group="sidebar-add-section"
            panelClass="w-64 rounded-md border border-border bg-popover p-2.5 shadow-xl shadow-black/40"
            onClose={() => (showAddSectionPanel = false)}
          >
            <SidebarAddSection
              onAddCustom={(detail) => addCustomSection(detail.label)}
            />
          </FloatingDropdown>
        {:else if activeMode === 'ssh'}
          <IconButton
            icon={Plus}
            variant="primary"
            label={t('sidebar.createSSHConnection', 'New SSH connection')}
            onclick={handleCreateSSHConnection}
          />
        {/if}
      </div>

      {#if activeMode !== 'api'}
        <Input
          class="mt-4"
          bind:value={resourceSearch}
          aria-label={activeResourcePanel?.searchPlaceholder}
          placeholder={activeResourcePanel?.searchPlaceholder}
          autocomplete="off"
          spellcheck="false"
        >
          {#snippet leading()}
            <Search class="h-4 w-4" />
          {/snippet}
        </Input>
      {/if}
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto overflow-x-hidden px-2 py-3">
      {#if activeMode === 'api'}
        {#if sections.length === 0}
          <div
            class="mx-2 rounded-md border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground"
          >
            {t('sidebar.empty')}
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
              onToggle={() => toggleSection(section.id)}
              onRemoveSection={(detail) => removeSection(detail.sectionId)}
              onRenameSection={renameSection}
              onAddItem={addSectionItem}
              onRenameItem={renameSectionItem}
              onMoveItem={moveSectionItem}
              onRemoveItem={removeSectionItem}
            />
          {/each}
        </div>
      {:else if activeMode === 'ssh'}
        {#if !hasSSHMatches && !normalizedResourceSearch}
          <div
            class="mx-2 rounded-md border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground"
          >
            {activeResourcePanel?.empty}
          </div>
        {/if}

        {#if !hasSSHMatches && normalizedResourceSearch}
          <div
            class="mx-2 rounded-md border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground"
          >
            No SSH connections match your search.
          </div>
        {/if}

        <div class="space-y-1.5">
          {#each sshConnectionList as connection (connection.id)}
            {@const isActive = sshState.selectedId === connection.id}
            <div class="group relative">
              <button
                type="button"
                draggable="true"
                class="w-full rounded-lg px-3 py-3 text-left transition-all duration-150 {isActive
                  ? 'bg-primary/10'
                  : 'bg-transparent group-hover:bg-fill'}"
                onclick={() => onConnectionClick(connection.id)}
                ondblclick={() => onConnectionDblClick(connection.id)}
                ondragstart={(event) =>
                  onConnectionDragStart(event, connection)}
                ondragend={onConnectionDragEnd}
              >
                <ConnectionRow {connection} active={isActive} />
              </button>

              <button
                type="button"
                class="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground opacity-0 transition hover:bg-fill-active hover:text-white focus-visible:opacity-100 group-hover:opacity-100"
                aria-label={t('sftp.openFiles', 'Open files (SFTP)')}
                {@attach tooltip({
                  content: t('sftp.openFiles', 'Open files (SFTP)'),
                  placement: 'right',
                })}
                onclick={(event) => {
                  event.stopPropagation()
                  openSFTP(connection)
                }}
              >
                <FolderTree size={14} strokeWidth={2.1} />
              </button>
            </div>
          {/each}
        </div>
      {:else if activeResourcePanel}
        <div
          class="mx-2 rounded-md border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground"
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
      onmouseenter={openFlyout}
      onmouseleave={scheduleFlyoutClose}
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
