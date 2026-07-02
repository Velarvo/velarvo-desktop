<script lang="ts">
  import { tick } from 'svelte'
  import { Plug, Search, Terminal } from 'lucide-svelte'
  import OsIcon from '@/components/ui/OsIcon.svelte'
  import { translate } from '@/lib/i18n'
  import { EditorTabKind, openContentTab, replaceTab } from '@/lib/editorLayout'
  import { selectSSHConnection, sshConnections } from '@/lib/sshConnections'
  import { formatRelativeTime, matchesSSHQuery } from '@/lib/sshPresentation'
  import { isSSHConnected, type SSHConnection } from '@/types/ssh'

  export let tabId: string | null = null
  export let groupId: string | null = null

  let query = ''
  let searchInput: HTMLInputElement | null = null

  $: normalizedQuery = query.trim().toLowerCase()

  const byRelevance = (a: SSHConnection, b: SSHConnection): number => {
    const connected = Number(isSSHConnected(b)) - Number(isSSHConnected(a))
    if (connected !== 0) return connected
    const recency = (b.lastUsedAt ?? 0) - (a.lastUsedAt ?? 0)
    if (recency !== 0) return recency
    return a.name.localeCompare(b.name)
  }

  $: connections = $sshConnections
    .filter((connection) => matchesSSHQuery(connection, normalizedQuery))
    .sort(byRelevance)

  $: hasAny = $sshConnections.length > 0

  const connectionMeta = (connection: SSHConnection): string =>
    `${connection.username}@${connection.host}:${connection.port}`

  const focusSearch = async () => {
    await tick()
    searchInput?.focus()
  }
  void focusSearch()

  const launch = (connection: SSHConnection) => {
    selectSSHConnection(connection.id)
    const content = {
      title: `${connection.name} · ${$translate('ssh.terminal', 'Terminal')}`,
      kind: EditorTabKind.SSHTerminal,
      data: { connectionId: connection.id },
    } as const
    if (tabId) {
      replaceTab(tabId, content)
    } else {
      openContentTab(content, { groupId: groupId ?? undefined })
    }
  }

  const onKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Escape' && query) query = ''
  }
</script>

<div class="flex h-full max-h-[500px] flex-col bg-surface-canvas px-4 py-8">
  <div
    class="mx-auto flex h-full w-full max-w-3xl flex-col overflow-hidden rounded-2xl border border-border bg-surface-subtle"
  >
    <div
      class="relative flex h-16 shrink-0 items-center border-b border-border px-4 sm:h-18 sm:px-5"
    >
      <Search
        class="pointer-events-none h-[18px] w-[18px] shrink-0 text-muted-foreground"
      />
      <input
        bind:this={searchInput}
        bind:value={query}
        on:keydown={onKeydown}
        type="text"
        autocomplete="off"
        spellcheck="false"
        placeholder={$translate(
          'launcher.searchPlaceholder',
          'Search connections…',
        )}
        class="h-full min-w-0 flex-1 bg-transparent px-3 text-base font-medium text-white outline-none placeholder:text-muted-foreground sm:text-lg"
      />
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto p-2 sm:p-3">
      {#if connections.length > 0}
        <div class="space-y-1">
          {#each connections as connection (connection.id)}
            <button
              type="button"
              class="group grid w-full grid-cols-[minmax(0,1fr)] gap-2 rounded-xl border border-transparent px-3 py-3 text-left text-muted-foreground transition-all duration-150 hover:border-white/8 hover:bg-white/[0.04] hover:text-white sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center sm:gap-4 sm:px-4"
              on:click={() => launch(connection)}
            >
              <span class="flex min-w-0 items-center gap-3">
                <span
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-white/8 bg-white/[0.035] text-muted-foreground transition group-hover:border-primary/20 group-hover:text-primary"
                >
                  {#if connection.os}
                    <OsIcon os={connection.os} size={16} />
                  {:else}
                    <Terminal class="h-4 w-4" />
                  {/if}
                </span>

                <span class="min-w-0">
                  <span class="flex min-w-0 items-center gap-3">
                    <span
                      class="truncate text-sm font-semibold text-white/80 group-hover:text-white"
                    >
                      {connection.name}
                    </span>
                    <span
                      class="hidden truncate text-xs text-muted-foreground sm:inline"
                    >
                      {connectionMeta(connection)}
                    </span>
                  </span>
                  <span
                    class="mt-1 flex min-w-0 items-center gap-2 text-xs text-muted-foreground"
                  >
                    <span class="truncate">{connection.host}</span>
                    <span class="h-1 w-1 rounded-full bg-muted-foreground/45"
                    ></span>
                    <span class="shrink-0"
                      >{formatRelativeTime(connection.lastUsedAt)}</span
                    >
                  </span>
                </span>
              </span>
            </button>
          {/each}
        </div>
      {:else}
        <div
          class="flex min-h-52 flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border px-6 py-10 text-center"
        >
          <span
            class="flex h-11 w-11 items-center justify-center rounded-xl border border-white/8 bg-white/[0.03] text-muted-foreground"
          >
            <Plug class="h-5 w-5" />
          </span>
          <p class="text-sm font-medium text-white/70">
            {hasAny
              ? $translate('launcher.noMatches', 'No connections match.')
              : $translate('launcher.empty', 'No connections yet.')}
          </p>
          <p class="text-xs text-muted-foreground">
            {hasAny
              ? $translate('launcher.noMatchesHint', 'Try a different search.')
              : $translate(
                  'launcher.emptyHint',
                  'Add a connection from the sidebar to get started.',
                )}
          </p>
        </div>
      {/if}
    </div>
  </div>
</div>
