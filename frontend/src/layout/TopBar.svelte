<script lang="ts">
  import { onMount } from 'svelte'
  import {
    BriefcaseBusiness,
    Check,
    ChevronDown,
    Command,
    FolderPlus,
    Layers,
    PanelLeft,
    Plus,
    Settings,
  } from 'lucide-svelte'
  import ProjectCreateModal from '@/components/projects/ProjectCreateModal.svelte'
  import SecuritySettingsModal from '@/components/settings/SecuritySettingsModal.svelte'
  import WorkspaceCreateModal from '@/components/workspaces/WorkspaceCreateModal.svelte'
  import { translate } from '@/lib/i18n'
  import { logger } from '@/lib/logger'
  import {
    bootstrapProjects,
    isLoadingProjects,
    loadProjectIcons,
    projectIconUrls,
    projects,
    projectsError,
    selectProject,
    selectedProject,
  } from '@/lib/projects'
  import {
    isLoadingWorkspaces,
    selectWorkspace,
    selectedWorkspace,
    workspaces,
    workspacesError,
  } from '@/lib/workspaces'

  export let onToggle: () => void

  const fallbackProjectColor = 'var(--color-primary)'
  const fallbackWorkspaceColor = 'var(--color-workspace)'
  const swatchForegroundColor = 'var(--color-swatch-foreground)'

  let openMenu: 'project' | 'workspace' | null = null
  let showCreateProjectModal = false
  let showCreateWorkspaceModal = false
  let showSecuritySettingsModal = false

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

  const openSecuritySettingsModal = () => {
    showSecuritySettingsModal = true
    closeMenus()
  }

  const closeSecuritySettingsModal = () => {
    showSecuritySettingsModal = false
  }

  const openCreateWorkspaceModal = () => {
    if (!$selectedProject) {
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

  const handleDocumentClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement | null
    if (!target) {
      return
    }

    if (!target.closest('[data-topbar-menu]')) {
      closeMenus()
    }
  }

  const chooseProject = (projectId: string) => {
    selectProject(projectId)
    closeMenus()
  }

  $: if ($projects.length > 0) {
    void loadProjectIcons($projects)
  }

  onMount(() => {
    document.addEventListener('click', handleDocumentClick)

    void bootstrapProjects().catch((error) => {
      logger.error('Failed to bootstrap projects in topbar', error)
    })

    return () => {
      document.removeEventListener('click', handleDocumentClick)
    }
  })
</script>

<header
  class="pl-25 relative z-300 flex h-14 items-center gap-2 border-b border-border bg-surface-base/95 px-3 backdrop-blur-sm select-none"
  style="--wails-draggable: drag;"
>
  <button
    type="button"
    class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-white/5 hover:text-white"
    on:click={onToggle}
  >
    <PanelLeft class="h-4 w-4" />
  </button>

  <div class="h-5 w-px shrink-0 bg-border"></div>

  <div class="relative" data-topbar-menu>
    <button
      type="button"
      on:click={() => toggleMenu('project')}
      class="inline-flex h-8 items-center gap-2 rounded-md px-2 text-sm font-medium text-white/80 hover:bg-white/5 hover:text-white"
    >
      <div
        class="flex h-5 w-5 shrink-0 items-center justify-center overflow-hidden rounded-md bg-surface-elevated ring-1 ring-white/10"
        style={`background:${$selectedProject?.color || 'var(--color-surface-elevated)'}; color:${$selectedProject ? swatchForegroundColor : 'var(--color-muted-foreground)'};`}
      >
        {#if $selectedProject && $projectIconUrls[$selectedProject.id]}
          <img
            class="h-full w-full object-cover"
            src={$projectIconUrls[$selectedProject.id]}
            alt=""
          />
        {:else if $selectedProject}
          <span class="text-[11px] font-black">
            {$selectedProject.name.charAt(0).toUpperCase()}
          </span>
        {:else}
          <BriefcaseBusiness class="h-3.5 w-3.5" />
        {/if}
      </div>

      <span
        >{$selectedProject?.name || $translate('topbar.projectsFallback')}</span
      >
      <ChevronDown class="h-3 w-3 opacity-40" />
    </button>

    {#if openMenu === 'project'}
      <div
        class="absolute left-0 top-10 z-310 w-[21rem] overflow-hidden rounded-2xl border border-white/10 bg-[linear-gradient(180deg,_var(--color-menu-start),_var(--color-menu-end))] p-1.5 shadow-[0_25px_80px_rgba(0,0,0,1)]"
      >
        <div class="rounded-[1rem] p-3">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p
                class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
              >
                {$translate('topbar.projectSwitcherTitle')}
              </p>
              <p class="mt-1 text-sm text-white">
                {$translate('topbar.projectSwitcherSubtitle')}
              </p>
            </div>

            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-primary/20 bg-primary/10 text-primary transition hover:border-primary/30 hover:bg-primary/15"
              aria-label={$translate('topbar.createProject')}
              on:click={openCreateProjectModal}
            >
              <FolderPlus class="h-4 w-4" />
            </button>
          </div>

          <div class="mt-3 space-y-1.5">
            {#if $isLoadingProjects}
              <div
                class="rounded-xl border border-white/8 bg-white/[0.03] px-3 py-3 text-sm text-muted-foreground"
              >
                {$translate('topbar.projectsLoading')}
              </div>
            {:else if $projectsError}
              <div
                class="rounded-xl border border-destructive/25 bg-destructive/10 px-3 py-3 text-sm text-red-200"
              >
                {$projectsError}
              </div>
            {:else if $projects.length === 0}
              <div
                class="rounded-xl border border-dashed border-white/10 bg-white/[0.02] px-3 py-4 text-sm text-muted-foreground"
              >
                {$translate('topbar.projectsEmpty')}
              </div>
            {:else}
              {#each $projects as project (project.id)}
                <button
                  type="button"
                  class="group flex w-full items-center gap-3 rounded-[1rem] px-3 py-3 text-left transition {project.id ===
                  $selectedProject?.id
                    ? 'bg-primary/10'
                    : 'bg-white/2 hover:bg-white/4'}"
                  on:click={() => chooseProject(project.id)}
                >
                  <span
                    class="inline-flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg text-sm font-black shadow-[inset_0_1px_0_rgba(255,255,255,0.18)]"
                    style={`background:${project.color || fallbackProjectColor}; color:${swatchForegroundColor};`}
                  >
                    {#if $projectIconUrls[project.id]}
                      <img
                        class="h-full w-full object-cover"
                        src={$projectIconUrls[project.id]}
                        alt=""
                      />
                    {:else}
                      {project.name.charAt(0).toUpperCase()}
                    {/if}
                  </span>

                  <span class="min-w-0 flex-1">
                    <span class="flex items-center gap-2">
                      <span class="truncate text-sm font-semibold text-white">
                        {project.name}
                      </span>
                      {#if project.id === $selectedProject?.id}
                        <span
                          class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-primary/15 text-primary"
                        >
                          <Check class="h-3.5 w-3.5" />
                        </span>
                      {/if}
                    </span>
                  </span>
                </button>
              {/each}
            {/if}
          </div>

          <div class="mt-3 border-t border-white/6 pt-3">
            <button
              type="button"
              class="inline-flex w-full items-center justify-center gap-2 rounded-[1rem] border border-white/8 bg-white/[0.03] px-3 py-2.5 text-sm font-medium text-white/80 transition hover:border-primary/20 hover:bg-primary/10 hover:text-white"
              on:click={openCreateProjectModal}
            >
              <FolderPlus class="h-4 w-4" />
              {$translate('topbar.addProject')}
            </button>
          </div>
        </div>
      </div>
    {/if}
  </div>

  <span class="shrink-0 text-base font-light text-border">/</span>

  <div class="relative" data-topbar-menu>
    <button
      type="button"
      on:click={() => toggleMenu('workspace')}
      disabled={!$selectedProject}
      class="inline-flex h-8 items-center gap-2 rounded-md px-2 text-sm font-medium text-white/80 hover:bg-white/5 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
    >
      <span
        class="h-1.5 w-1.5 shrink-0 rounded-full"
        style={`background:${$selectedWorkspace?.color || fallbackWorkspaceColor};`}
      ></span>
      <span
        >{$selectedWorkspace?.name ||
          $translate('topbar.workspaceFallback')}</span
      >
      <ChevronDown class="h-3 w-3 opacity-40" />
    </button>

    {#if openMenu === 'workspace'}
      <div
        class="absolute left-0 top-10 z-310 w-[21rem] overflow-hidden rounded-2xl border border-white/10 bg-[linear-gradient(180deg,_var(--color-menu-start),_var(--color-menu-end))] p-1.5 shadow-[0_25px_80px_rgba(0,0,0,1)]"
      >
        <div class="rounded-[1rem] p-3">
          <div class="flex items-center justify-between gap-3">
            <div>
              <p
                class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
              >
                {$translate('topbar.workspaceSwitcherTitle')}
              </p>
              <p class="mt-1 text-sm text-white">
                {$selectedProject?.name ||
                  $translate('topbar.workspaceSwitcherSubtitle')}
              </p>
            </div>

            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-sky-400/20 bg-sky-400/10 text-sky-300 transition hover:border-sky-400/30 hover:bg-sky-400/15 disabled:cursor-not-allowed disabled:opacity-50"
              aria-label={$translate('topbar.createWorkspace')}
              disabled={!$selectedProject}
              on:click={openCreateWorkspaceModal}
            >
              <Plus class="h-4 w-4" />
            </button>
          </div>

          <div class="mt-3 space-y-1.5">
            {#if !$selectedProject}
              <div
                class="rounded-xl border border-dashed border-white/10 bg-white/[0.02] px-3 py-4 text-sm text-muted-foreground"
              >
                {$translate('topbar.workspaceNoProject')}
              </div>
            {:else if $isLoadingWorkspaces}
              <div
                class="rounded-xl border border-white/8 bg-white/[0.03] px-3 py-3 text-sm text-muted-foreground"
              >
                {$translate('topbar.workspacesLoading')}
              </div>
            {:else if $workspacesError}
              <div
                class="rounded-xl border border-destructive/25 bg-destructive/10 px-3 py-3 text-sm text-red-200"
              >
                {$workspacesError}
              </div>
            {:else if $workspaces.length === 0}
              <div
                class="rounded-xl border border-dashed border-white/10 bg-white/[0.02] px-3 py-4 text-sm text-muted-foreground"
              >
                {$translate('topbar.workspacesEmpty')}
              </div>
            {:else}
              {#each $workspaces as workspace (workspace.id)}
                <button
                  type="button"
                  class="group flex w-full items-center gap-3 rounded-[1rem] px-3 py-3 text-left transition {workspace.id ===
                  $selectedWorkspace?.id
                    ? 'bg-sky-400/10'
                    : 'bg-white/2 hover:bg-white/4'}"
                  on:click={() => chooseWorkspace(workspace.id)}
                >
                  <span
                    class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-sm font-black shadow-[inset_0_1px_0_rgba(255,255,255,0.18)]"
                    style={`background:${workspace.color || fallbackWorkspaceColor}; color:${swatchForegroundColor};`}
                  >
                    <Layers class="h-4 w-4" />
                  </span>

                  <span class="min-w-0 flex-1">
                    <span class="flex items-center gap-2">
                      <span class="truncate text-sm font-semibold text-white">
                        {workspace.name}
                      </span>
                      {#if workspace.id === $selectedWorkspace?.id}
                        <span
                          class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-sky-400/15 text-sky-300"
                        >
                          <Check class="h-3.5 w-3.5" />
                        </span>
                      {/if}
                    </span>
                  </span>
                </button>
              {/each}
            {/if}
          </div>

          <div class="mt-3 border-t border-white/6 pt-3">
            <button
              type="button"
              class="inline-flex w-full items-center justify-center gap-2 rounded-[1rem] border border-white/8 bg-white/[0.03] px-3 py-2.5 text-sm font-medium text-white/80 transition hover:border-sky-400/20 hover:bg-sky-400/10 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!$selectedProject}
              on:click={openCreateWorkspaceModal}
            >
              <Plus class="h-4 w-4" />
              {$translate('topbar.addWorkspace')}
            </button>
          </div>
        </div>
      </div>
    {/if}
  </div>

  <div class="mx-4 flex max-w-md flex-1 justify-center">
    <button
      type="button"
      class="flex h-8 w-full items-center gap-2.5 rounded-lg bg-white/4 px-3 text-muted-foreground transition-all duration-150 hover:bg-white/6 hover:text-white/60 focus:outline-none focus:ring-1 focus:ring-primary/25"
    >
      <Command class="h-3.5 w-3.5 shrink-0" />
      <span class="flex-1 text-left text-sm">{$translate('topbar.search')}</span
      >
      <div class="flex shrink-0 items-center gap-1 text-[10px] font-mono">
        <kbd
          class="rounded border border-border bg-white/5 px-1.5 py-0.5 text-white/40"
        >
          Ctrl
        </kbd>
        <kbd
          class="rounded border border-border bg-white/5 px-1.5 py-0.5 text-white/40"
        >
          K
        </kbd>
      </div>
    </button>
  </div>

  <div class="ml-auto flex shrink-0 items-center gap-1.5">
    <button
      type="button"
      class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-white/5 hover:text-white"
      aria-label={$translate('topbar.openSecuritySettings')}
      on:click={openSecuritySettingsModal}
    >
      <Settings class="h-4 w-4" />
    </button>
  </div>
</header>

<ProjectCreateModal
  open={showCreateProjectModal}
  on:close={closeCreateProjectModal}
  on:created={handleProjectCreated}
/>

<WorkspaceCreateModal
  open={showCreateWorkspaceModal}
  projectId={$selectedProject?.id ?? null}
  projectName={$selectedProject?.name ?? ''}
  on:close={closeCreateWorkspaceModal}
  on:created={handleWorkspaceCreated}
/>

<SecuritySettingsModal
  open={showSecuritySettingsModal}
  on:close={closeSecuritySettingsModal}
/>
