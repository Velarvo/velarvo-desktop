<script lang="ts">
  import { onMount } from 'svelte'
  import { ChevronDown, Settings, Command, PanelLeft } from 'lucide-svelte'
  import { user, logout } from '@/lib/auth'
  import { navigate } from '@/lib/router'

  export let onToggle: () => void

  const organizations = [{ id: 'personal', name: 'Personal', code: 'prs' }]

  const workspaces = [{ id: 'local', name: 'Local', env: 'local' }]

  const envColor: Record<string, string> = {
    local: 'bg-sky-400',
    prod: 'bg-emerald-500',
    stg: 'bg-amber-400',
    dev: 'bg-muted-foreground',
  }

  let openMenu: 'org' | 'workspace' | 'user' | null = null

  const toggleMenu = (menu: 'org' | 'workspace' | 'user') => {
    openMenu = openMenu === menu ? null : menu
  }

  const closeMenus = () => {
    openMenu = null
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

  const handleLogout = async () => {
    await logout()
    closeMenus()
    navigate('/login')
  }

  onMount(() => {
    document.addEventListener('click', handleDocumentClick)
    return () => {
      document.removeEventListener('click', handleDocumentClick)
    }
  })
</script>

<header
  class="pl-25 relative z-300 flex h-14 items-center gap-2 border-b border-border bg-[#0a0a0f]/95 backdrop-blur-sm px-3 select-none"
  style="--wails-draggable: drag;"
>
  <button
    type="button"
    class="h-8 w-8 shrink-0 text-muted-foreground hover:text-white hover:bg-white/5 inline-flex items-center justify-center rounded-md"
    on:click={onToggle}
  >
    <PanelLeft class="h-4 w-4" />
  </button>

  <div class="h-5 w-px bg-border shrink-0"></div>

  <div class="relative" data-topbar-menu>
    <button
      type="button"
      on:click={() => toggleMenu('org')}
      class="h-8 gap-2 px-2 text-sm text-white/80 hover:text-white hover:bg-white/5 font-medium inline-flex items-center rounded-md"
    >
      <div
        class="h-5 w-5 overflow-hidden rounded-md ring-1 ring-white/10 shrink-0"
      ></div>
      <span>Personal</span>
      <ChevronDown class="h-3 w-3 opacity-40" />
    </button>

    {#if openMenu === 'org'}
      <div
        class="absolute left-0 top-10 z-310 w-52 rounded-md border border-border bg-[#0e0e14] shadow-xl shadow-black/40 p-1"
      >
        {#each organizations as org}
          <button
            type="button"
            class="w-full gap-2.5 text-sm cursor-pointer text-white/80 hover:text-white hover:bg-white/5 focus:text-white rounded-sm px-2 py-2 inline-flex items-center text-left"
          >
            <div
              class="h-5 w-5 overflow-hidden rounded-md ring-1 ring-white/10 shrink-0"
            ></div>
            {org.name}
          </button>
        {/each}
        <div class="my-1 h-px bg-border"></div>
        <button
          type="button"
          class="w-full text-left text-sm text-primary/80 hover:text-primary cursor-pointer focus:text-primary rounded-sm px-2 py-2"
          >New organization</button
        >
      </div>
    {/if}
  </div>

  <span class="text-border text-base font-light shrink-0">/</span>

  <div class="relative" data-topbar-menu>
    <button
      type="button"
      on:click={() => toggleMenu('workspace')}
      class="h-8 gap-2 px-2 text-sm text-white/80 hover:text-white hover:bg-white/5 font-medium inline-flex items-center rounded-md"
    >
      <span class="h-1.5 w-1.5 rounded-full bg-sky-400 shrink-0"></span>
      <span>Local</span>
      <ChevronDown class="h-3 w-3 opacity-40" />
    </button>

    {#if openMenu === 'workspace'}
      <div
        class="absolute left-0 top-10 z-310 w-48 rounded-md border border-border bg-[#0e0e14] shadow-xl shadow-black/40 p-1"
      >
        {#each workspaces as ws}
          <button
            type="button"
            class="w-full text-sm cursor-pointer text-white/80 hover:text-white focus:text-white rounded-sm px-2 py-2 inline-flex items-center text-left"
          >
            <span
              class="mr-2.5 h-1.5 w-1.5 rounded-full shrink-0 {envColor[
                ws.env
              ]}"
            ></span>
            {ws.name}
            <span class="ml-auto font-mono text-[10px] text-muted-foreground"
              >{ws.env}</span
            >
          </button>
        {/each}
        <div class="my-1 h-px bg-border"></div>
        <button
          type="button"
          class="w-full text-left text-sm text-primary/80 hover:text-primary cursor-pointer focus:text-primary rounded-sm px-2 py-2"
          >New workspace</button
        >
      </div>
    {/if}
  </div>

  <div class="flex justify-center mx-4 flex-1 max-w-md">
    <button
      type="button"
      class="flex h-8 w-full items-center gap-2.5 rounded-lg border border-border bg-white/3 px-3 text-muted-foreground transition-all duration-150 hover:border-primary/30 hover:bg-white/5 hover:text-white/60 focus:outline-none focus:ring-1 focus:ring-primary/20"
    >
      <Command class="h-3.5 w-3.5 shrink-0" />
      <span class="flex-1 text-left text-sm">Search...</span>
      <div class="flex items-center gap-1 text-[10px] font-mono shrink-0">
        <kbd
          class="rounded border border-border bg-white/5 px-1.5 py-0.5 text-white/40"
          >Ctrl</kbd
        >
        <kbd
          class="rounded border border-border bg-white/5 px-1.5 py-0.5 text-white/40"
          >K</kbd
        >
      </div>
    </button>
  </div>

  <div class="ml-auto flex items-center gap-1.5 shrink-0">
    <button
      type="button"
      class="h-8 w-8 text-muted-foreground hover:text-white hover:bg-white/5 inline-flex items-center justify-center rounded-md"
    >
      <Settings class="h-4 w-4" />
    </button>

    <div class="relative" data-topbar-menu>
      <button
        type="button"
        on:click={() => toggleMenu('user')}
        class="h-8 gap-2 px-2 text-sm text-white/80 hover:text-white hover:bg-white/5 font-medium inline-flex items-center rounded-md"
      >
        <div
          class="flex justify-center items-center h-8 w-8 shrink-0 rounded-full bg-primary/20 text-[13px] font-semibold text-primary"
        >
          {$user?.firstName?.[0] ?? ''}{$user?.lastName?.[0] ?? ''}
        </div>
        <span class="hidden sm:inline"
          >{$user?.firstName} {$user?.lastName}</span
        >
        <ChevronDown class="h-3 w-3 opacity-40" />
      </button>

      {#if openMenu === 'user'}
        <div
          class="absolute right-0 top-10 z-310 w-48 rounded-md border border-border bg-[#0e0e14] shadow-xl shadow-black/40 p-1"
        >
          <div class="px-3 py-2 border-b border-border mb-1">
            <p class="text-sm font-medium text-white">
              {$user?.firstName}
              {$user?.lastName}
            </p>
            <p class="text-xs text-muted-foreground">{$user?.email}</p>
          </div>
          <button
            type="button"
            class="w-full text-left text-sm text-white/80 hover:text-white cursor-pointer rounded-sm px-2 py-2"
            >Profile</button
          >
          <div class="my-1 h-px bg-border"></div>
          <button
            type="button"
            on:click={handleLogout}
            class="w-full text-left text-sm text-destructive hover:text-destructive cursor-pointer rounded-sm px-2 py-2"
            >Sign out</button
          >
        </div>
      {/if}
    </div>
  </div>
</header>
