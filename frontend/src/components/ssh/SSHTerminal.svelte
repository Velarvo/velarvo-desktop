<script lang="ts">
  import { onDestroy, onMount } from 'svelte'
  import { RotateCw, TriangleAlert } from 'lucide-svelte'
  import { translate } from '@/lib/i18n'
  import {
    acquireTerminal,
    fitTerminal,
    reconnectTerminal,
    type TerminalEntry,
  } from '@/lib/sshTerminalSessions'
  import { SSHTerminalStatus } from '@/types/ssh'

  export let paneId: string
  export let connectionId: string

  let container: HTMLDivElement
  let entry: TerminalEntry | null = null
  let resizeObserver: ResizeObserver | null = null
  let rafHandle = 0
  let unsubStatus: (() => void) | null = null
  let unsubError: (() => void) | null = null

  interface TerminalStatusView {
    labelKey: string
    fallbackLabel: string
    labelClass: string
    canReconnect: boolean
    showsError: boolean
  }

  const STATUS_VIEW: Record<SSHTerminalStatus, TerminalStatusView> = {
    [SSHTerminalStatus.Connecting]: {
      labelKey: 'ssh.connecting',
      fallbackLabel: 'Connecting…',
      labelClass: 'text-warning',
      canReconnect: false,
      showsError: false,
    },
    [SSHTerminalStatus.Open]: {
      labelKey: 'ssh.connected',
      fallbackLabel: 'Connected',
      labelClass: 'text-primary',
      canReconnect: false,
      showsError: false,
    },
    [SSHTerminalStatus.Closed]: {
      labelKey: 'ssh.closed',
      fallbackLabel: 'Closed',
      labelClass: 'text-muted-foreground',
      canReconnect: true,
      showsError: false,
    },
    [SSHTerminalStatus.Error]: {
      labelKey: 'ssh.connectionError',
      fallbackLabel: 'Connection error',
      labelClass: 'text-red-300',
      canReconnect: true,
      showsError: true,
    },
  }

  let status = SSHTerminalStatus.Connecting
  let errorMessage = ''

  $: statusView = STATUS_VIEW[status]

  const scheduleFit = () => {
    cancelAnimationFrame(rafHandle)
    rafHandle = requestAnimationFrame(() => {
      if (entry) fitTerminal(entry)
    })
  }

  const reconnect = () => {
    void reconnectTerminal(paneId)
  }

  onMount(() => {
    entry = acquireTerminal(paneId, connectionId)
    // eslint-disable-next-line svelte/no-dom-manipulating
    container.appendChild(entry.host)

    unsubStatus = entry.status.subscribe((value) => (status = value))
    unsubError = entry.error.subscribe((value) => (errorMessage = value))

    resizeObserver = new ResizeObserver(() => scheduleFit())
    resizeObserver.observe(container)

    scheduleFit()
    entry.term.focus()
  })

  onDestroy(() => {
    cancelAnimationFrame(rafHandle)
    resizeObserver?.disconnect()
    unsubStatus?.()
    unsubError?.()
    if (entry && entry.host.parentElement === container) {
      // eslint-disable-next-line svelte/no-dom-manipulating
      container.removeChild(entry.host)
    }
    entry = null
  })
</script>

<div class="flex h-full min-h-0 w-full flex-col bg-surface-canvas">
  {#if statusView.canReconnect}
    <div
      class="flex shrink-0 items-center gap-2.5 border-b px-3 py-2 text-xs {statusView.showsError
        ? 'border-destructive/30 bg-destructive/10 text-red-200'
        : 'border-border bg-surface-base text-muted-foreground'}"
    >
      {#if statusView.showsError}
        <TriangleAlert class="mt-0.5 h-3.5 w-3.5 shrink-0" />
      {/if}
      <span class="min-w-0 break-words {statusView.labelClass}">
        {statusView.showsError
          ? errorMessage
          : $translate(statusView.labelKey, statusView.fallbackLabel)}
      </span>
      <button
        type="button"
        class="ml-auto inline-flex shrink-0 items-center gap-1.5 rounded px-2 py-1 text-muted-foreground transition hover:bg-white/5 hover:text-white"
        on:click={reconnect}
      >
        <RotateCw class="h-3.5 w-3.5" />
        {$translate('ssh.reconnect', 'Reconnect')}
      </button>
    </div>
  {/if}

  <div class="relative min-h-0 flex-1 overflow-hidden p-2">
    <div bind:this={container} class="h-full w-full"></div>
  </div>
</div>
