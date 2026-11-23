<script lang="ts">
  import OsIcon from '@/components/ui/OsIcon.svelte'
  import { t } from '@/lib/i18n.svelte'
  import {
    formatRelativeTime,
    sshStatusIcon,
    sshStatusLabel,
  } from '@/lib/sshPresentation'
  import { isSSHConnected, type SSHConnection } from '@/types/ssh'

  interface Props {
    connection: SSHConnection
    active?: boolean
  }

  let { connection, active = false }: Props = $props()

  let isConnected = $derived(isSSHConnected(connection))
  let StatusIcon = $derived(sshStatusIcon(connection))
</script>

<div class="flex min-w-0 items-center gap-3">
  <span
    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-fill-hover text-muted-foreground transition group-hover:bg-fill-active {active
      ? 'bg-primary/15 text-primary'
      : ''}"
  >
    <OsIcon os={connection.os} size={18} />
  </span>

  <span class="min-w-0 flex-1">
    <span class="flex min-w-0 items-center gap-2">
      <span class="truncate text-sm font-semibold text-white">
        {connection.name}
      </span>
      <StatusIcon
        size={13}
        strokeWidth={2.2}
        class="shrink-0 {isConnected
          ? 'text-primary'
          : connection.hasPassword
            ? 'text-warning'
            : 'text-muted-foreground'}"
      />
    </span>
    <span
      class="mt-1 flex min-w-0 items-center gap-2 text-xs text-muted-foreground"
    >
      <span class="truncate">
        {connection.username}@{connection.host}
      </span>
      <span class="h-1 w-1 rounded-full bg-muted-foreground/45"></span>
      <span class="shrink-0">:{connection.port}</span>
    </span>
    <span
      class="mt-1 flex min-w-0 items-center gap-2 text-[11px] text-muted-foreground"
    >
      <span>{t(...sshStatusLabel(connection))}</span>
      <span class="h-1 w-1 rounded-full bg-muted-foreground/45"></span>
      <span class="truncate">{formatRelativeTime(connection.lastUsedAt)}</span>
    </span>
  </span>
</div>
