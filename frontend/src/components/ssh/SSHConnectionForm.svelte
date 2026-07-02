<script lang="ts">
  import {
    CircleCheck,
    Eye,
    EyeOff,
    Plug,
    Terminal,
    TriangleAlert,
  } from 'lucide-svelte'
  import { translate } from '@/lib/i18n'
  import {
    createSSHConnection,
    selectSSHConnection,
    sshConnections,
    updateSSHConnection,
  } from '@/lib/sshConnections'
  import {
    closeTab,
    EditorTabKind,
    replaceTab,
    updateTab,
  } from '@/lib/editorLayout'
  import {
    isSSHConnected,
    SSH_DEFAULT_PORT,
    type SSHConnection,
  } from '@/types/ssh'

  export let tabId: string
  export let groupId: string
  export let connectionId: string | null = null

  let name = ''
  let host = ''
  let port = SSH_DEFAULT_PORT
  let username = ''
  let password = ''
  let clearPassword = false
  let showPassword = false

  let justSaved = false
  let savedTimer: ReturnType<typeof setTimeout> | null = null
  let busy = false
  let errorMessage = ''

  let hydratedFor: string | null | undefined = undefined
  $: hydrate(connectionId)

  const hydrate = (id: string | null) => {
    if (id === hydratedFor) return
    hydratedFor = id
    if (!id) return

    const connection = $sshConnections.find((item) => item.id === id)
    if (!connection) return

    name = connection.name
    host = connection.host
    port = connection.port
    username = connection.username
  }

  $: storeConnection =
    connectionId !== null
      ? ($sshConnections.find((item) => item.id === connectionId) ?? null)
      : null
  $: isEditing = connectionId !== null
  $: hasStoredPassword = storeConnection?.hasPassword ?? false
  $: isConnected = storeConnection ? isSSHConnected(storeConnection) : false
  $: primaryDisabled = !host.trim() || !username.trim() || busy

  const flagSaved = () => {
    justSaved = true
    if (savedTimer) clearTimeout(savedTimer)
    savedTimer = setTimeout(() => (justSaved = false), 2500)
  }

  const buildInput = () => ({
    name,
    host,
    port: Number(port) || SSH_DEFAULT_PORT,
    username,
    password,
  })

  const persist = async (): Promise<SSHConnection | null> => {
    const input = buildInput()

    if (connectionId) {
      const updated = await updateSSHConnection({
        ...input,
        id: connectionId,
        clearPassword,
      })
      selectSSHConnection(updated.id)
      updateTab(tabId, { title: updated.name })
      password = ''
      clearPassword = false
      return updated
    }

    const created = await createSSHConnection(input)
    selectSSHConnection(created.id)
    connectionId = created.id
    updateTab(tabId, {
      title: created.name,
      data: { connectionId: created.id },
    })
    password = ''
    return created
  }

  const save = async () => {
    if (primaryDisabled) return
    busy = true
    errorMessage = ''
    try {
      await persist()
      flagSaved()
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : String(error)
    } finally {
      busy = false
    }
  }

  const connect = async () => {
    if (primaryDisabled) return
    busy = true
    errorMessage = ''
    try {
      const connection = await persist()
      if (!connection) return
      replaceTab(tabId, {
        title: `${connection.name || connection.host} · ${$translate('ssh.terminal', 'Terminal')}`,
        kind: EditorTabKind.SSHTerminal,
        data: { connectionId: connection.id },
      })
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : String(error)
    } finally {
      busy = false
    }
  }

  const cancel = () => closeTab(groupId, tabId)
</script>

<div class="@container flex h-full min-h-0 w-full flex-col bg-surface-canvas">
  <div class="min-h-0 flex-1 overflow-y-auto">
    <div class="mx-auto w-full max-w-5xl px-4 py-6 @2xl:px-8 @2xl:py-8">
      <header class="flex flex-wrap items-start justify-between gap-4">
        <div class="flex min-w-0 items-center gap-3">
          <span
            class="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-primary/20 bg-primary/10 text-primary"
          >
            <Terminal class="h-5 w-5" />
          </span>
          <div class="min-w-0">
            <p
              class="text-[10px] font-semibold uppercase tracking-[0.18em] text-primary"
            >
              {$translate('ssh.eyebrow', 'SSH connection')}
            </p>
            <h2 class="truncate text-xl font-semibold text-white">
              {isEditing
                ? name || $translate('ssh.editTitle', 'Edit connection')
                : $translate('ssh.createTitle', 'New SSH connection')}
            </h2>
            <p class="mt-1 text-sm text-muted-foreground">
              {$translate(
                'ssh.subtitle',
                'Configure the target and password, then save.',
              )}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          {#if isConnected}
            <span
              class="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-3 py-1.5 text-xs font-medium text-primary"
            >
              <CircleCheck class="h-4 w-4" />
              {$translate('ssh.connected', 'Connected')}
            </span>
          {:else if justSaved}
            <span
              class="inline-flex items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-3 py-1.5 text-xs font-medium text-primary"
            >
              <CircleCheck class="h-4 w-4" />
              {$translate('ssh.saved', 'Saved')}
            </span>
          {/if}
        </div>
      </header>

      {#if errorMessage}
        <div
          class="mt-5 flex items-start gap-2.5 rounded-lg border border-destructive/30 bg-destructive/10 px-3.5 py-3 text-sm text-red-200"
        >
          <TriangleAlert class="mt-0.5 h-4 w-4 shrink-0" />
          <span class="min-w-0 break-words">{errorMessage}</span>
        </div>
      {/if}

      <div class="mt-6">
        <div class="space-y-5">
          <label class="space-y-1.5">
            <span class="text-sm font-semibold text-white">
              {$translate('ssh.displayName', 'Display name')}
            </span>
            <input
              bind:value={name}
              placeholder="Production API"
              class="h-11 w-full rounded-md border border-border bg-white/5 px-3.5 text-sm text-white outline-none transition placeholder:text-muted-foreground focus:border-primary/40 focus:bg-white/[0.07]"
            />
          </label>

          <div class="grid gap-4 @sm:grid-cols-[minmax(0,1fr)_6rem]">
            <label class="space-y-1.5">
              <span class="text-xs font-semibold text-muted-foreground">
                {$translate('ssh.host', 'Host')}
              </span>
              <input
                bind:value={host}
                placeholder="api-01.production.internal"
                class="h-10 w-full rounded-md border border-border bg-white/[0.035] px-3 text-sm text-white outline-none transition placeholder:text-muted-foreground focus:border-primary/40 focus:bg-white/[0.05]"
              />
            </label>

            <label class="space-y-1.5">
              <span class="text-xs font-semibold text-muted-foreground">
                {$translate('ssh.port', 'Port')}
              </span>
              <input
                bind:value={port}
                type="number"
                min="1"
                max="65535"
                class="h-10 w-full rounded-md border border-border bg-white/[0.035] px-3 text-sm text-white outline-none transition focus:border-primary/40 focus:bg-white/[0.05]"
              />
            </label>
          </div>

          <label class="space-y-1.5">
            <span class="text-xs font-semibold text-muted-foreground">
              {$translate('ssh.username', 'Username')}
            </span>
            <input
              bind:value={username}
              placeholder="deploy"
              autocomplete="off"
              class="h-10 w-full rounded-md border border-border bg-white/[0.035] px-3 text-sm text-white outline-none transition placeholder:text-muted-foreground focus:border-primary/40 focus:bg-white/[0.05]"
            />
          </label>

          <div>
            <p class="text-xs font-semibold text-muted-foreground">
              {$translate('ssh.password', 'Password')}
            </p>
            <div class="relative mt-2">
              {#if showPassword}
                <input
                  bind:value={password}
                  type="text"
                  autocomplete="off"
                  disabled={clearPassword}
                  placeholder={hasStoredPassword
                    ? $translate('ssh.passwordStored', '•••••••• (stored)')
                    : $translate(
                        'ssh.passwordPlaceholder',
                        'Connection password',
                      )}
                  class="h-10 w-full rounded-md border border-border bg-white/[0.035] px-3 pr-10 font-mono text-sm text-white outline-none transition placeholder:text-muted-foreground focus:border-primary/40 focus:bg-white/[0.05] disabled:opacity-50"
                />
              {:else}
                <input
                  bind:value={password}
                  type="password"
                  autocomplete="off"
                  disabled={clearPassword}
                  placeholder={hasStoredPassword
                    ? $translate('ssh.passwordStored', '•••••••• (stored)')
                    : $translate(
                        'ssh.passwordPlaceholder',
                        'Connection password',
                      )}
                  class="h-10 w-full rounded-md border border-border bg-white/[0.035] px-3 pr-10 font-mono text-sm text-white outline-none transition placeholder:text-muted-foreground focus:border-primary/40 focus:bg-white/[0.05] disabled:opacity-50"
                />
              {/if}
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-white"
                aria-label={showPassword
                  ? $translate('ssh.hidePassword', 'Hide password')
                  : $translate('ssh.showPassword', 'Show password')}
                on:click={() => (showPassword = !showPassword)}
              >
                {#if showPassword}
                  <EyeOff class="h-4 w-4" />
                {:else}
                  <Eye class="h-4 w-4" />
                {/if}
              </button>
            </div>
            {#if isEditing && hasStoredPassword}
              <label
                class="mt-2 flex items-center gap-2 text-xs text-muted-foreground"
              >
                <input
                  type="checkbox"
                  bind:checked={clearPassword}
                  class="h-3.5 w-3.5 accent-primary"
                />
                {$translate('ssh.clearPassword', 'Remove stored password')}
              </label>
            {:else if isEditing}
              <p class="mt-2 text-xs text-muted-foreground">
                {$translate(
                  'ssh.passwordEditHint',
                  'Leave blank to keep the current password.',
                )}
              </p>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </div>

  <footer class="w-full flex items-center justify-center">
    <div
      class="flex w-full shrink-0 flex-wrap items-center justify-end gap-3 border-t border-border bg-surface-base px-4 py-3 @3xl:my-2 @3xl:w-2/3 @3xl:rounded-md @3xl:border @3xl:px-8"
    >
      <button
        type="button"
        class="mr-auto inline-flex h-10 items-center rounded-md border border-border bg-white/[0.035] px-4 text-sm font-semibold text-white transition hover:border-white/20 hover:bg-white/[0.06]"
        on:click={cancel}
      >
        {$translate('common.cancel', 'Cancel')}
      </button>

      <button
        type="button"
        disabled={primaryDisabled}
        class="inline-flex h-10 items-center rounded-md border border-border bg-white/[0.035] px-4 text-sm font-semibold text-white transition hover:border-white/20 hover:bg-white/[0.06] disabled:cursor-not-allowed disabled:opacity-45"
        on:click={save}
      >
        {isEditing
          ? $translate('ssh.saveChanges', 'Save changes')
          : $translate('ssh.save', 'Save connection')}
      </button>

      <button
        type="button"
        disabled={primaryDisabled}
        class="inline-flex h-10 items-center gap-2 rounded-md bg-primary px-4 text-sm font-semibold text-background transition hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-45"
        on:click={connect}
      >
        <Plug class="h-4 w-4" />
        {$translate('ssh.saveAndConnect', 'Save & connect')}
      </button>
    </div>
  </footer>
</div>
