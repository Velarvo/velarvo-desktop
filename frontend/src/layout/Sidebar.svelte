<script lang="ts">
  import { onDestroy } from 'svelte'
  import { Plus } from 'lucide-svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import { t } from '@/lib/i18n'

  import SidebarAddSection from './sidebar/SidebarAddSection.svelte'
  import SidebarSection from './sidebar/SidebarSection.svelte'
  import { getMethodBadgeStyle, getSectionIcon } from './sidebar/sectionMeta'
  import { sidebarSectionsStore } from './sidebar/sidebarStore'

  export let collapsed = false

  let expanded: string[] = []
  let hoveredSection: string | null = null
  let hoveredRect: DOMRect | null = null
  let showAddSectionPanel = false
  let addSectionContainer: HTMLDivElement | null = null
  let hoverTimeout: ReturnType<typeof setTimeout> | null = null
  let initializedExpanded = false

  $: sections = $sidebarSectionsStore

  $: if (!initializedExpanded && sections.length > 0) {
    expanded = sections.map((section) => section.id)
    initializedExpanded = true
  }

  $: if (initializedExpanded) {
    const allowedIds = new Set(sections.map((section) => section.id))
    expanded = expanded.filter((sectionId) => allowedIds.has(sectionId))
  }

  $: hoveredData = hoveredSection
    ? (sections.find((section) => section.id === hoveredSection) ?? null)
    : null

  const toggleSection = (sectionId: string) => {
    expanded = expanded.includes(sectionId)
      ? expanded.filter((item) => item !== sectionId)
      : [...expanded, sectionId]
  }

  const handleMouseEnter = (sectionId: string, event: MouseEvent) => {
    if (!collapsed) {
      return
    }

    if (hoverTimeout) {
      clearTimeout(hoverTimeout)
      hoverTimeout = null
    }

    hoveredSection = sectionId
    hoveredRect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  }

  const sectionMouseEnter = (sectionId: string) => (event: MouseEvent) => {
    handleMouseEnter(sectionId, event)
  }

  const handleMouseLeaveSection = () => {
    if (!collapsed) {
      return
    }

    hoverTimeout = setTimeout(() => {
      hoveredSection = null
      hoveredRect = null
    }, 200)
  }

  const handleMenuEnter = (sectionId: string) => {
    if (hoverTimeout) {
      clearTimeout(hoverTimeout)
      hoverTimeout = null
    }
    hoveredSection = sectionId
  }

  const handleMenuLeave = () => {
    hoveredSection = null
    hoveredRect = null
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

    if (hoveredSection === sectionId) {
      hoveredSection = null
      hoveredRect = null
    }
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
    event: CustomEvent<{
      sectionId: string
      name: string
      type: string
      status?: 'connected' | 'degraded' | 'disconnected'
      method?:
        | 'GET'
        | 'POST'
        | 'PUT'
        | 'DELETE'
        | 'PATCH'
        | 'OPTIONS'
        | 'HEAD'
        | 'GRAPHQL'
    }>,
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

  const isApiItem = (type?: string, method?: string) => {
    if (method) return true

    const normalized = (type ?? '').toUpperCase()
    return (
      normalized === 'API' || normalized === 'REST' || normalized === 'GRAPHQL'
    )
  }

  onDestroy(() => {
    if (hoverTimeout) {
      clearTimeout(hoverTimeout)
    }
  })
</script>

<aside
  class="relative z-200 flex flex-col bg-[#0a0a0f] border-r border-border transition-all duration-300 ease-in-out overflow-y-auto overflow-x-visible {collapsed
    ? 'w-14'
    : 'w-75'}"
>
  <div class="relative z-10 py-2">
    {#if sections.length === 0 && !collapsed}
      <div
        class="mx-2 mt-1 rounded-md border border-border bg-white/2 px-3 py-2 text-xs text-muted-foreground"
      >
        {t('sidebar.empty')}
      </div>
    {/if}

    {#each sections as section}
      <div
        role="presentation"
        on:mouseenter={sectionMouseEnter(section.id)}
        on:mouseleave={handleMouseLeaveSection}
      >
        <SidebarSection
          {section}
          {collapsed}
          expanded={expanded.includes(section.id)}
          highlighted={hoveredSection === section.id}
          icon={getSectionIcon(section.kind)}
          {getMethodBadgeStyle}
          on:toggle={() => toggleSection(section.id)}
          on:removeSection={(event) => removeSection(event.detail.sectionId)}
          on:renameSection={renameSection}
          on:addItem={addSectionItem}
          on:renameItem={renameSectionItem}
          on:reorderItem={reorderSectionItem}
          on:moveItem={moveSectionItem}
          on:removeItem={removeSectionItem}
        />
      </div>
    {/each}

    {#if !collapsed}
      <div class="relative px-2 pt-2" bind:this={addSectionContainer}>
        <button
          type="button"
          class="w-full h-9 rounded-md border border-border bg-white/3 text-sm text-white hover:bg-white/5 inline-flex items-center justify-center gap-2"
          on:click={() => (showAddSectionPanel = !showAddSectionPanel)}
        >
          <Plus size={14} class="text-primary" />
          {t('sidebar.addSection')}
        </button>
      </div>

      <FloatingDropdown
        open={showAddSectionPanel}
        anchorElement={addSectionContainer}
        placement="bottom-start"
        offset={8}
        matchAnchorWidth={true}
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
</aside>

{#if collapsed && hoveredSection && hoveredRect && hoveredData}
  <div
    role="menu"
    tabindex="-1"
    class="fixed z-500 min-w-50 rounded-lg border border-border bg-background shadow-2xl shadow-black/60 overflow-hidden"
    style="top: {hoveredRect.top}px; left: {hoveredRect.right + 6}px;"
    on:mouseenter={() => hoveredSection && handleMenuEnter(hoveredSection)}
    on:mouseleave={handleMenuLeave}
  >
    <div class="flex items-center gap-2 px-3 py-2.5 border-b border-border">
      <svelte:component
        this={getSectionIcon(hoveredData.kind)}
        size={14}
        class="text-primary"
      />
      <span
        class="text-xs font-semibold uppercase tracking-wider text-muted-foreground"
        >{hoveredData.label}</span
      >
    </div>

    <div class="py-1">
      {#each hoveredData.items as item}
        <div
          class="flex items-center gap-2.5 px-3 py-2 text-sm text-white/70 hover:text-white hover:bg-white/5 cursor-pointer transition-colors"
        >
          {#if isApiItem(item.type, item.method)}
            <span
              class="px-1.5 py-0.5 text-[9px] rounded font-medium {getMethodBadgeStyle(
                item.method,
              )}">{item.method}</span
            >
          {:else}
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 shrink-0"
            ></span>
          {/if}
          <span class="truncate">{item.name}</span>
        </div>
      {/each}
    </div>
  </div>
{/if}
