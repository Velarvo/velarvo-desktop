<script lang="ts">
  import type { ComponentType } from 'svelte'
  import {
    Database,
    Globe2,
    KeyRound,
    Plus,
    Search,
    Terminal,
  } from 'lucide-svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import { t } from '@/lib/i18n'
  import type { SidebarItemDraft } from '@/types/sidebar'

  import SidebarAddSection from './sidebar/SidebarAddSection.svelte'
  import SidebarSection from './sidebar/SidebarSection.svelte'
  import { getMethodBadgeStyle, getSectionIcon } from './sidebar/sectionMeta'
  import { sidebarSectionsStore } from './sidebar/sidebarStore'

  export let collapsed = false

  type SidebarMode = 'creds' | 'ssh' | 'db' | 'api'

  type RailItem = {
    id: SidebarMode
    label: string
    shortLabel: string
    description: string
    icon: ComponentType
  }

  type ResourceItem = {
    id: string
    name: string
    meta: string
    detail: string
    badge?: string
    icon: ComponentType
  }

  type ResourcePanel = {
    eyebrow: string
    title: string
    searchPlaceholder: string
    empty: string
    items: ResourceItem[]
  }

  let activeMode: SidebarMode = 'api'
  let resourceSearch = ''
  let expanded: string[] = []
  let showAddSectionPanel = false
  let addSectionContainer: HTMLDivElement | null = null
  let initializedExpanded = false

  const railItems: RailItem[] = [
    {
      id: 'creds',
      label: 'Credentials',
      shortLabel: 'Credentials',
      description: 'Secrets, tokens and passwords',
      icon: KeyRound,
    },
    {
      id: 'ssh',
      label: 'SSH',
      shortLabel: 'SSH',
      description: 'Servers and terminal access',
      icon: Terminal,
    },
    {
      id: 'db',
      label: 'Databases',
      shortLabel: 'DB',
      description: 'Database connections',
      icon: Database,
    },
    {
      id: 'api',
      label: 'API',
      shortLabel: 'API',
      description: 'Collections and requests',
      icon: Globe2,
    },
  ]

  const resourcePanels: Record<Exclude<SidebarMode, 'api'>, ResourcePanel> = {
    creds: {
      eyebrow: 'Vault',
      title: 'Credentials',
      searchPlaceholder: 'Search credentials...',
      empty: 'No credentials yet.',
      items: [],
    },
    ssh: {
      eyebrow: 'Access',
      title: 'SSH Connections',
      searchPlaceholder: 'Search SSH hosts...',
      empty: 'No SSH connections yet.',
      items: [],
    },
    db: {
      eyebrow: 'Data',
      title: 'Databases',
      searchPlaceholder: 'Search connections...',
      empty: 'No database connections yet.',
      items: [],
    },
  }

  $: sections = $sidebarSectionsStore
  $: activeRailItem =
    railItems.find((item) => item.id === activeMode) ?? railItems[0]
  $: activeResourcePanel =
    activeMode === 'api' ? null : resourcePanels[activeMode]
  $: normalizedResourceSearch = resourceSearch.trim().toLowerCase()
  $: filteredResourceItems = activeResourcePanel
    ? activeResourcePanel.items.filter((item) => {
        if (!normalizedResourceSearch) return true

        return [item.name, item.meta, item.detail, item.badge ?? ''].some(
          (value) => value.toLowerCase().includes(normalizedResourceSearch),
        )
      })
    : []

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
  class="relative z-200 flex h-full shrink-0 overflow-visible border-r border-border bg-[#0a0a0f] transition-all duration-300 ease-in-out {collapsed
    ? 'w-[4.5rem]'
    : 'w-[23rem]'}"
>
  <nav
    class="flex w-[4.5rem] shrink-0 flex-col items-center gap-2 border-r border-border bg-[#08080c] px-2 py-4"
    aria-label="Resource types"
  >
    <div
      class="mb-3 flex h-10 w-10 items-center justify-center rounded-xl border border-primary/20 bg-primary/15 text-sm font-black text-primary shadow-[0_0_24px_rgba(18,206,144,0.12)]"
    >
      V
    </div>

    {#each railItems as item}
      <button
        type="button"
        class="group relative flex h-11 w-11 items-center justify-center rounded-xl border transition-all duration-150 {activeMode ===
        item.id
          ? 'border-primary/30 bg-primary/12 text-primary shadow-[inset_0_0_0_1px_rgba(18,206,144,0.08)]'
          : 'border-transparent text-muted-foreground hover:border-border hover:bg-white/5 hover:text-white'}"
        aria-label={item.label}
        aria-pressed={activeMode === item.id}
        on:click={() => selectMode(item.id)}
      >
        {#if activeMode === item.id}
          <span
            class="absolute -left-2 h-6 w-0.5 rounded-r-full bg-primary shadow-[0_0_12px_rgba(18,206,144,0.65)]"
          ></span>
        {/if}

        <svelte:component this={item.icon} size={19} strokeWidth={2.1} />

        <span
          class="pointer-events-none absolute left-[3.35rem] z-500 hidden min-w-max rounded-lg border border-border bg-[#101016] px-2.5 py-1.5 text-xs font-medium text-white shadow-xl shadow-black/40 group-hover:block"
        >
          {item.shortLabel}
        </span>
      </button>
    {/each}
  </nav>

  {#if !collapsed}
    <section class="flex min-w-0 flex-1 flex-col bg-[#0a0a0f]">
      <div class="border-b border-border px-4 py-4">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p
              class="text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground"
            >
              {activeMode === 'api'
                ? 'Collections'
                : activeResourcePanel?.eyebrow}
            </p>
            <div class="mt-1 flex items-center gap-2">
              <svelte:component
                this={activeRailItem.icon}
                size={17}
                class="shrink-0 text-primary"
              />
              <h2 class="truncate text-sm font-semibold text-white">
                {activeMode === 'api' ? 'API' : activeResourcePanel?.title}
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
                class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-primary/25 bg-primary text-background shadow-[0_10px_28px_rgba(18,206,144,0.18)] transition hover:bg-primary/90"
                on:click={() => (showAddSectionPanel = !showAddSectionPanel)}
                aria-label={t('sidebar.addSection')}
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
              panelClass="rounded-lg border border-border bg-[#11111a] shadow-2xl shadow-black/60 p-2.5"
              on:close={() => (showAddSectionPanel = false)}
            >
              <SidebarAddSection
                on:addCustom={(event) => addCustomSection(event.detail.label)}
              />
            </FloatingDropdown>
          {/if}
        </div>

        {#if activeMode !== 'api'}
          <label
            class="mt-4 flex h-10 items-center gap-2 rounded-xl border border-border bg-white/3 px-3 text-muted-foreground transition focus-within:border-primary/35 focus-within:bg-white/5"
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
              class="mx-2 rounded-xl border border-dashed border-border bg-white/2 px-4 py-5 text-sm text-muted-foreground"
            >
              {t('sidebar.empty')}
            </div>
          {/if}

          <div class="space-y-1">
            {#each sections as section}
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
        {:else if activeResourcePanel}
          {#if filteredResourceItems.length === 0}
            <div
              class="mx-2 rounded-xl border border-border bg-white/2 px-4 py-5 text-sm text-muted-foreground"
            >
              {activeResourcePanel.empty}
            </div>
          {/if}

          <div class="space-y-1.5">
            {#each filteredResourceItems as item, index}
              <button
                type="button"
                class="group relative w-full rounded-xl border px-3 py-3 text-left transition-all duration-150 {index ===
                0
                  ? 'border-primary/30 bg-primary/10 shadow-[inset_0_0_0_1px_rgba(18,206,144,0.04)]'
                  : 'border-transparent bg-transparent hover:border-border hover:bg-white/4'}"
              >
                {#if index === 0}
                  <span
                    class="absolute left-0 top-3 bottom-3 w-0.5 rounded-r-full bg-primary"
                  ></span>
                {/if}

                <div class="flex min-w-0 items-center gap-3">
                  <span
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-white/5 text-primary/90 transition group-hover:border-primary/25 group-hover:bg-primary/10"
                  >
                    <svelte:component
                      this={item.icon}
                      size={18}
                      strokeWidth={2}
                    />
                  </span>

                  <span class="min-w-0 flex-1">
                    <span
                      class="flex min-w-0 items-center gap-2 text-sm font-semibold text-white"
                    >
                      <span class="truncate">{item.name}</span>
                      {#if item.badge}
                        <span
                          class="shrink-0 rounded-full border border-border bg-white/5 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground"
                        >
                          {item.badge}
                        </span>
                      {/if}
                    </span>
                    <span
                      class="mt-1 flex min-w-0 items-center gap-2 text-xs text-muted-foreground"
                    >
                      <span class="truncate">{item.meta}</span>
                      <span
                        class="h-1 w-1 shrink-0 rounded-full bg-muted-foreground/40"
                      ></span>
                      <span class="truncate">{item.detail}</span>
                    </span>
                  </span>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </section>
  {/if}
</aside>
