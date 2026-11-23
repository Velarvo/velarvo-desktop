<script lang="ts">
  import { tick } from 'svelte'
  import { CornerDownLeft, Plus, Search, Server } from '@lucide/svelte'
  import Button from '@/components/ui/Button.svelte'
  import Input from '@/components/ui/Input.svelte'
  import OsIcon from '@/components/ui/OsIcon.svelte'
  import { t } from '@/lib/i18n.svelte'
  import {
    EditorTabKind,
    openContentTab,
    replaceTab,
    type EditorTabContent,
  } from '@/lib/editorLayout.svelte'
  import { selectSSHConnection, sshState } from '@/lib/sshConnections.svelte'
  import { formatRelativeTime, matchesSSHQuery } from '@/lib/sshPresentation'
  import { isSSHConnected, type SSHConnection } from '@/types/ssh'

  interface Props {
    tabId?: string | null
    groupId?: string | null
  }

  let { tabId = null, groupId = null }: Props = $props()

  let query = $state('')
  let activeIndex = $state(0)
  let searchInput = $state<HTMLInputElement | null>(null)

  const byRelevance = (a: SSHConnection, b: SSHConnection): number =>
    Number(isSSHConnected(b)) - Number(isSSHConnected(a)) ||
    (b.lastUsedAt ?? 0) - (a.lastUsedAt ?? 0) ||
    a.name.localeCompare(b.name)

  let connections = $derived(
    sshState.connections
      .filter((c) => matchesSSHQuery(c, query.trim().toLowerCase()))
      .sort(byRelevance),
  )
  let hasAny = $derived(sshState.connections.length > 0)
  let canCreate = $derived(sshState.workspaceId !== null)

  $effect(() => {
    if (activeIndex >= connections.length) activeIndex = 0
  })

  $effect(() => {
    if (hasAny) void tick().then(() => searchInput?.focus())
  })

  const open = (content: EditorTabContent) => {
    if (tabId) replaceTab(tabId, content)
    else openContentTab(content, { groupId: groupId ?? undefined })
  }

  const launch = (connection: SSHConnection) => {
    selectSSHConnection(connection.id)
    open({
      title: connection.name,
      kind: EditorTabKind.SSHTerminal,
      data: { connectionId: connection.id },
    })
  }

  const createConnection = () =>
    open({
      title: t('launcher.newConnection'),
      kind: EditorTabKind.SSHConnectionForm,
      data: {},
    })

  const onKeydown = (event: KeyboardEvent) => {
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        activeIndex = (activeIndex + 1) % Math.max(connections.length, 1)
        break
      case 'ArrowUp':
        event.preventDefault()
        activeIndex =
          (activeIndex - 1 + connections.length) %
          Math.max(connections.length, 1)
        break
      case 'Enter': {
        const connection = connections[activeIndex]
        if (connection) launch(connection)
        break
      }
      case 'Escape':
        query = ''
        break
    }
  }
</script>

<div class="h-full overflow-y-auto bg-surface-canvas">
  <div class="mx-auto flex w-full max-w-2xl flex-col px-6 pt-[12vh] pb-10">
    <header class="flex items-end justify-between gap-4">
      <div>
        <h2 class="text-lg font-semibold tracking-tight text-white">
          {t('launcher.title')}
        </h2>
        <p class="mt-1 text-sm text-muted-foreground">
          {t('launcher.subtitle')}
        </p>
      </div>
      {#if hasAny}
        <Button
          variant="secondary"
          size="sm"
          disabled={!canCreate}
          onclick={createConnection}
        >
          <Plus class="h-4 w-4" />
          {t('launcher.newConnection')}
        </Button>
      {/if}
    </header>

    {#if hasAny}
      <Input
        size="lg"
        class="mt-6"
        bind:ref={searchInput}
        bind:value={query}
        onkeydown={onKeydown}
        autocomplete="off"
        spellcheck="false"
        aria-label={t('launcher.searchPlaceholder')}
        placeholder={t('launcher.searchPlaceholder')}
      >
        {#snippet leading()}
          <Search class="h-4 w-4" />
        {/snippet}
      </Input>

      <ul
        class="mt-3 overflow-hidden rounded-md border border-border bg-fill-subtle"
      >
        {#each connections as connection, index (connection.id)}
          {@const active = index === activeIndex}
          <li class="border-b border-border last:border-b-0">
            <button
              type="button"
              class="flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors {active
                ? 'bg-fill-hover'
                : 'hover:bg-fill'}"
              onmouseenter={() => (activeIndex = index)}
              onclick={() => launch(connection)}
            >
              <span
                class="relative flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-line-subtle bg-fill text-muted-foreground"
              >
                {#if connection.os}
                  <OsIcon os={connection.os} size={15} />
                {:else}
                  <Server class="h-4 w-4" />
                {/if}
                {#if isSSHConnected(connection)}
                  <span
                    class="absolute -right-0.5 -bottom-0.5 h-2.5 w-2.5 rounded-full border-2 border-surface-canvas bg-primary"
                    title={t('ssh.connected')}
                  ></span>
                {/if}
              </span>

              <span class="min-w-0 flex-1">
                <span class="block truncate text-sm font-medium text-white">
                  {connection.name}
                </span>
                <span class="block truncate text-xs text-muted-foreground">
                  {connection.username}@{connection.host}{connection.port !== 22
                    ? `:${connection.port}`
                    : ''}
                </span>
              </span>

              <span class="shrink-0 text-xs text-muted-foreground tabular-nums">
                {formatRelativeTime(connection.lastUsedAt)}
              </span>
              <CornerDownLeft
                class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-opacity {active
                  ? 'opacity-100'
                  : 'opacity-0'}"
              />
            </button>
          </li>
        {:else}
          <li class="px-4 py-8 text-center text-sm text-muted-foreground">
            {t('launcher.noMatches')}
          </li>
        {/each}
      </ul>
    {:else}
      <div
        class="mt-6 flex flex-col items-center rounded-md border border-dashed border-border px-6 py-12 text-center"
      >
        <span
          class="flex h-10 w-10 items-center justify-center rounded-lg border border-line-subtle bg-fill text-muted-foreground"
        >
          <Server class="h-5 w-5" />
        </span>
        <p class="mt-4 text-sm font-medium text-white">
          {t('launcher.empty')}
        </p>
        <p class="mt-1 max-w-xs text-sm text-muted-foreground">
          {t('launcher.emptyHint')}
        </p>
        <Button
          class="mt-5"
          size="sm"
          disabled={!canCreate}
          onclick={createConnection}
        >
          <Plus class="h-4 w-4" />
          {t('launcher.newConnection')}
        </Button>
      </div>
    {/if}
  </div>
</div>
