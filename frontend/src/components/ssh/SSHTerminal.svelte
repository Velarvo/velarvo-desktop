<script lang="ts">
  import type { Attachment } from 'svelte/attachments'
  import ConnectionState from '@/components/ssh/ConnectionState.svelte'
  import { sshState } from '@/lib/sshConnections.svelte'
  import {
    acquireTerminal,
    fitTerminal,
    reconnectTerminal,
    type TerminalView,
  } from '@/lib/sshTerminalSessions.svelte'
  import { emptyConnectProgress, SSHTerminalStatus } from '@/types/ssh'

  interface Props {
    paneId: string
    connectionId: string
  }

  let { paneId, connectionId }: Props = $props()

  let view = $state.raw<TerminalView | null>(null)
  const status = $derived(view?.status ?? SSHTerminalStatus.Connecting)
  const errorMessage = $derived(view?.error ?? '')
  const progress = $derived(view?.progress ?? emptyConnectProgress())

  const connection = $derived(
    sshState.connections.find((item) => item.id === connectionId) ?? null,
  )

  const mountedPaneId = $derived(paneId)
  const mountedConnectionId = $derived(connectionId)

  const reconnect = () => {
    void reconnectTerminal(paneId)
  }

  const mountTerminal =
    (pane: string, connectionId: string): Attachment<HTMLDivElement> =>
    (container) => {
      const entry = acquireTerminal(pane, connectionId)
      container.appendChild(entry.host)

      view = entry.view

      let frame = 0
      const scheduleFit = () => {
        cancelAnimationFrame(frame)
        frame = requestAnimationFrame(() => fitTerminal(entry))
      }
      const resizeObserver = new ResizeObserver(scheduleFit)
      resizeObserver.observe(container)

      scheduleFit()
      entry.term.focus()

      return () => {
        cancelAnimationFrame(frame)
        resizeObserver.disconnect()
        view = null
        if (entry.host.parentElement === container) {
          container.removeChild(entry.host)
        }
      }
    }
</script>

<div class="flex h-full min-h-0 w-full flex-col bg-surface-canvas">
  <div class="relative min-h-0 flex-1 overflow-hidden p-2">
    <div
      {@attach mountTerminal(mountedPaneId, mountedConnectionId)}
      class="h-full w-full select-text"
    ></div>

    <ConnectionState
      {status}
      {errorMessage}
      {connection}
      {progress}
      onReconnect={reconnect}
    />
  </div>
</div>
