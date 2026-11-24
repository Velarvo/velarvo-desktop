<script lang="ts">
  import { onMount } from 'svelte'
  import {
    BriefcaseBusiness,
    Check,
    ChevronDown,
    PanelLeft,
    Plus,
    Search,
    Settings,
  } from '@lucide/svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import MenuItem, {
    MENU_PANEL_CLASS,
    MENU_SEPARATOR_CLASS,
  } from '@/components/ui/MenuItem.svelte'
  import ProjectCreateModal from '@/components/projects/ProjectCreateModal.svelte'
  import SettingsModal from '@/components/settings/SettingsModal.svelte'
  import WorkspaceCreateModal from '@/components/workspaces/WorkspaceCreateModal.svelte'
  import { t } from '@/lib/i18n.svelte'
  import { logger } from '@/lib/logger'
  import {
    bootstrapProjects,
    selectProject,
    projectsState,
    type Project,
  } from '@/lib/projects.svelte'
  import { selectWorkspace, workspacesState } from '@/lib/workspaces.svelte'

  interface Props {
    onToggle: () => void
  }

  let { onToggle }: Props = $props()

  const fallbackProjectColor = 'var(--color-primary)'
  const fallbackWorkspaceColor = 'var(--color-workspace)'
  const swatchForegroundColor = 'var(--color-swatch-foreground)'

  let openMenu = $state<'project' | 'workspace' | null>(null)
  let projectTrigger = $state<HTMLButtonElement | null>(null)
  let workspaceTrigger = $state<HTMLButtonElement | null>(null)

  const isMac =
    typeof navigator !== 'undefined' && /Mac/i.test(navigator.platform)
  let showCreateProjectModal = $state(false)
  let showCreateWorkspaceModal = $state(false)
  let showSettingsModal = $state(false)

  const toggleMenu = (menu: 'project' | 'workspace') => {
    openMenu = openMenu === menu ? null : menu
  }

  const closeMenus = () => {
    openMenu = null
  }

  const openCreateProjectModal = () => {
    showCreateProjectModal = true
    closeMenus()
  }

  const closeCreateProjectModal = () => {
    showCreateProjectModal = false
  }

  const openSettingsModal = () => {
    showSettingsModal = true
    closeMenus()
  }

  const closeSettingsModal = () => {
    showSettingsModal = false
  }

  const openCreateWorkspaceModal = () => {
    if (!projectsState.selected) {
      return
    }
    showCreateWorkspaceModal = true
    closeMenus()
  }

  const closeCreateWorkspaceModal = () => {
    showCreateWorkspaceModal = false
  }

  const handleProjectCreated = async () => {
    closeCreateProjectModal()
  }

  const handleWorkspaceCreated = async () => {
    closeCreateWorkspaceModal()
  }

  const chooseWorkspace = (workspaceId: string) => {
    selectWorkspace(workspaceId)
    closeMenus()
  }

  const chooseProject = (projectId: string) => {
    selectProject(projectId)
    closeMenus()
  }

  onMount(() => {
    void bootstrapProjects().catch((error) => {
      logger.error('Failed to bootstrap projects in topbar', error)
    })
  })
</script>

{#snippet projectAvatar(project: Project | null, size: 'sm' | 'md')}
  <span
    class={[
      'inline-flex shrink-0 items-center justify-center overflow-hidden rounded font-black',
      size === 'sm' ? 'h-5 w-5 text-[11px]' : 'h-6 w-6 text-xs',
    ]}
    style={project
      ? `background:${project.color || fallbackProjectColor}; color:${swatchForegroundColor};`
      : 'background:var(--color-fill-active); color:var(--color-muted-foreground);'}
  >
    {#if project && projectsState.iconUrls[project.id]}
      <img
        class="h-full w-full object-cover"
        src={projectsState.iconUrls[project.id]}
        alt=""
      />
    {:else if project}
      {project.name.charAt(0).toUpperCase()}
    {:else}
      <BriefcaseBusiness class="h-3.5 w-3.5" />
    {/if}
  </span>
{/snippet}

{#snippet menuHeading(text: string)}
  <p
    class="px-2.5 pb-1 pt-1.5 text-[11px] font-semibold uppercase tracking-[0.14em] text-muted-foreground/70"
  >
    {text}
  </p>
{/snippet}

{#snippet menuNote(text: string, tone: 'muted' | 'error' = 'muted')}
  <p
    class={[
      'px-2.5 py-2 text-sm',
      tone === 'error' ? 'text-red-300' : 'text-muted-foreground',
    ]}
  >
    {text}
  </p>
{/snippet}

<header
  class="titlebar pl-25 relative z-300 flex h-14 items-center gap-1 border-b border-border bg-surface-base/95 px-3 backdrop-blur-sm"
>
  <IconButton
    tooltipPlacement="bottom"
    icon={PanelLeft}
    label={t('topbar.toggleSidebar', 'Toggle sidebar')}
    onclick={onToggle}
  />

  <div class="mx-1 h-5 w-px shrink-0 bg-border"></div>

  <button
    type="button"
    bind:this={projectTrigger}
    aria-haspopup="menu"
    aria-expanded={openMenu === 'project'}
    onclick={() => toggleMenu('project')}
    class={[
      'inline-flex h-8 items-center gap-2 rounded-md px-2 text-sm font-medium transition hover:bg-fill-hover hover:text-white',
      openMenu === 'project' ? 'bg-fill-hover text-white' : 'text-fg-muted',
    ]}
  >
    {@render projectAvatar(projectsState.selected, 'sm')}
    <span>
      {projectsState.selected?.name || t('topbar.projectsFallback')}
    </span>
    <ChevronDown class="h-3 w-3 opacity-40" />
  </button>

  <FloatingDropdown
    open={openMenu === 'project'}
    anchorElement={projectTrigger}
    offset={6}
    group="topbar"
    panelClass={`${MENU_PANEL_CLASS} w-72`}
    onClose={closeMenus}
  >
    <div role="menu">
      {@render menuHeading(t('topbar.projectSwitcherTitle'))}

      {#if projectsState.loading}
        {@render menuNote(t('topbar.projectsLoading'))}
      {:else if projectsState.error}
        {@render menuNote(projectsState.error, 'error')}
      {:else if projectsState.items.length === 0}
        {@render menuNote(t('topbar.projectsEmpty'))}
      {:else}
        {#each projectsState.items as project (project.id)}
          <MenuItem
            role="menuitemradio"
            aria-checked={project.id === projectsState.selected?.id}
            class="h-9"
            onclick={() => chooseProject(project.id)}
          >
            <span class="flex items-center gap-2.5">
              {@render projectAvatar(project, 'md')}
              <span class="truncate text-white">{project.name}</span>
            </span>
            {#snippet trailing()}
              {#if project.id === projectsState.selected?.id}
                <Check class="h-4 w-4 shrink-0 text-primary" />
              {/if}
            {/snippet}
          </MenuItem>
        {/each}
      {/if}

      <span class={MENU_SEPARATOR_CLASS}></span>
      <MenuItem
        icon={Plus}
        label={t('topbar.addProject')}
        onclick={openCreateProjectModal}
      />
    </div>
  </FloatingDropdown>

  <span class="text-sm text-muted-foreground/40">/</span>

  <button
    type="button"
    bind:this={workspaceTrigger}
    aria-haspopup="menu"
    aria-expanded={openMenu === 'workspace'}
    disabled={!projectsState.selected}
    onclick={() => toggleMenu('workspace')}
    class={[
      'inline-flex h-8 items-center gap-2 rounded-md px-2 text-sm font-medium transition hover:bg-fill-hover hover:text-white disabled:cursor-not-allowed disabled:opacity-50',
      openMenu === 'workspace' ? 'bg-fill-hover text-white' : 'text-fg-muted',
    ]}
  >
    <span
      class="h-2 w-2 shrink-0 rounded-full"
      style={`background:${workspacesState.selected?.color || fallbackWorkspaceColor};`}
    ></span>
    <span>
      {workspacesState.selected?.name || t('topbar.workspaceFallback')}
    </span>
    <ChevronDown class="h-3 w-3 opacity-40" />
  </button>

  <FloatingDropdown
    open={openMenu === 'workspace'}
    anchorElement={workspaceTrigger}
    offset={6}
    group="topbar"
    panelClass={`${MENU_PANEL_CLASS} w-64`}
    onClose={closeMenus}
  >
    <div role="menu">
      {@render menuHeading(t('topbar.workspaceSwitcherTitle'))}

      {#if !projectsState.selected}
        {@render menuNote(t('topbar.workspaceNoProject'))}
      {:else if workspacesState.loading}
        {@render menuNote(t('topbar.workspacesLoading'))}
      {:else if workspacesState.error}
        {@render menuNote(workspacesState.error, 'error')}
      {:else if workspacesState.items.length === 0}
        {@render menuNote(t('topbar.workspacesEmpty'))}
      {:else}
        {#each workspacesState.items as workspace (workspace.id)}
          <MenuItem
            role="menuitemradio"
            aria-checked={workspace.id === workspacesState.selected?.id}
            onclick={() => chooseWorkspace(workspace.id)}
          >
            <span class="flex items-center gap-2.5">
              <span
                class="h-2 w-2 shrink-0 rounded-full"
                style={`background:${workspace.color || fallbackWorkspaceColor};`}
              ></span>
              <span class="truncate text-white">{workspace.name}</span>
            </span>
            {#snippet trailing()}
              {#if workspace.id === workspacesState.selected?.id}
                <Check class="h-4 w-4 shrink-0 text-primary" />
              {/if}
            {/snippet}
          </MenuItem>
        {/each}
      {/if}

      <span class={MENU_SEPARATOR_CLASS}></span>
      <MenuItem
        icon={Plus}
        label={t('topbar.addWorkspace')}
        disabled={!projectsState.selected}
        onclick={openCreateWorkspaceModal}
      />
    </div>
  </FloatingDropdown>

  <div class="mx-4 flex max-w-md flex-1 justify-center">
    <button
      type="button"
      class="flex h-8 w-full items-center gap-2.5 rounded-md border border-border bg-fill px-3 text-muted-foreground transition hover:border-line-strong hover:text-fg-muted outline-none focus-visible:border-primary/50"
    >
      <Search class="h-3.5 w-3.5 shrink-0" />
      <span class="flex-1 text-left text-sm">
        {t('topbar.search')}
      </span>
      <kbd
        class="shrink-0 rounded bg-fill-active px-1.5 py-0.5 font-mono text-[10px] font-semibold text-muted-foreground"
      >
        {isMac ? '⌘K' : 'Ctrl K'}
      </kbd>
    </button>
  </div>

  <div class="ml-auto flex shrink-0 items-center">
    <IconButton
      tooltipPlacement="bottom"
      icon={Settings}
      label={t('topbar.openSecuritySettings')}
      onclick={openSettingsModal}
    />
  </div>
</header>

<ProjectCreateModal
  open={showCreateProjectModal}
  onClose={closeCreateProjectModal}
  onCreated={handleProjectCreated}
/>

<WorkspaceCreateModal
  open={showCreateWorkspaceModal}
  projectId={projectsState.selected?.id ?? null}
  projectName={projectsState.selected?.name ?? ''}
  onClose={closeCreateWorkspaceModal}
  onCreated={handleWorkspaceCreated}
/>

<SettingsModal open={showSettingsModal} onClose={closeSettingsModal} />
