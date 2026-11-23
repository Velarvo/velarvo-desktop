<script lang="ts">
  import { CheckCircle2, CircleSlash, TriangleAlert, X } from '@lucide/svelte'
  import Button from '@/components/ui/Button.svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import { formatFileSize } from '@/components/file-explorer/utils'
  import { t } from '@/lib/i18n.svelte'
  import {
    cancelTransfer,
    clearSettledTransfers,
    type Transfer,
    transfersState,
  } from '@/lib/sftp.svelte'

  let list = $derived(transfersState.items)
  let running = $derived(list.filter((item) => item.status === 'running'))
  let settled = $derived(list.length - running.length)

  const percent = (item: Transfer) =>
    item.total > 0
      ? Math.min(100, Math.round((item.done / item.total) * 100))
      : 0
</script>

<div class="w-[22rem] max-w-[90vw]">
  <div
    class="flex items-center justify-between gap-3 border-b border-line-subtle px-3 py-2 text-[11px] tracking-wider text-muted-foreground uppercase"
  >
    <span>
      {t('sftp.transfers', 'Transfers')}
      {#if running.length > 0}
        · {running.length}
        {t('sftp.active', 'active')}
      {/if}
    </span>

    {#if settled > 0}
      <Button variant="ghost" size="xs" onclick={clearSettledTransfers}>
        {t('sftp.clearFinished', 'Clear finished')}
      </Button>
    {/if}
  </div>

  {#if list.length === 0}
    <p class="px-3 py-6 text-center text-xs text-muted-foreground">
      {t('sftp.noTransfers', 'No transfers yet.')}
    </p>
  {:else}
    <div class="max-h-64 overflow-y-auto px-3 py-2">
      {#each list as item (item.id)}
        <div class="flex items-center gap-3 py-1">
          <span
            class="w-11 shrink-0 text-center font-mono text-[10px] uppercase {item.direction ===
            'upload'
              ? 'text-primary'
              : 'text-workspace'}"
          >
            {item.direction === 'upload' ? '↑ up' : '↓ down'}
          </span>

          <span class="min-w-0 flex-1">
            <span class="flex items-center gap-2">
              <span class="min-w-0 flex-1 truncate text-xs text-fg-muted">
                {item.name}
              </span>

              {#if item.status === 'running'}
                <span
                  class="shrink-0 font-mono text-[10px] tabular-nums text-muted-foreground"
                >
                  {formatFileSize(item.done)}{#if item.total > 0}
                    / {formatFileSize(item.total)}{/if}
                </span>
                <IconButton
                  icon={X}
                  size="xs"
                  label={t('common.cancel', 'Cancel')}
                  onclick={() => cancelTransfer(item.id)}
                />
              {:else if item.status === 'done'}
                <CheckCircle2 size={13} class="shrink-0 text-primary" />
              {:else if item.status === 'cancelled'}
                <CircleSlash size={13} class="shrink-0 text-muted-foreground" />
              {:else}
                <TriangleAlert size={13} class="shrink-0 text-destructive" />
              {/if}
            </span>

            {#if item.status === 'running'}
              <span
                class="mt-1 block h-0.5 w-full overflow-hidden rounded-full bg-fill-active"
              >
                <span
                  class="block h-full rounded-full bg-primary transition-[width] duration-150"
                  style="width: {percent(item)}%"
                ></span>
              </span>
            {:else if item.status === 'failed' && item.error}
              <span
                class="mt-0.5 block truncate font-mono text-[10px] text-destructive"
              >
                {item.error}
              </span>
            {/if}
          </span>
        </div>
      {/each}
    </div>
  {/if}
</div>
