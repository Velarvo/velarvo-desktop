<script lang="ts">
  import { tick } from 'svelte'
  import { AlertTriangle } from '@lucide/svelte'
  import Button from '@/components/ui/Button.svelte'
  import Modal from '@/components/ui/Modal.svelte'
  import { t } from '@/lib/i18n.svelte'

  interface Props {
    open?: boolean
    title?: string
    message?: string
    confirmLabel?: string
    cancelLabel?: string
    destructive?: boolean
    onConfirm?: () => void
    onClose?: () => void
  }

  let {
    open = false,
    title = '',
    message = '',
    confirmLabel = '',
    cancelLabel = '',
    destructive = false,
    onConfirm,
    onClose,
  }: Props = $props()

  let actions = $state<HTMLDivElement | null>(null)

  const focusConfirm = async () => {
    await tick()
    const buttons = actions?.querySelectorAll('button')
    buttons?.[buttons.length - 1]?.focus()
  }

  $effect(() => {
    if (open) void focusConfirm()
  })
</script>

<Modal {open} {title} size="sm" onClose={() => onClose?.()}>
  {#snippet icon()}
    <AlertTriangle
      class="h-4 w-4 {destructive ? 'text-destructive' : 'text-primary'}"
      aria-hidden="true"
    />
  {/snippet}

  {#if message}
    <p class="text-sm leading-relaxed text-muted-foreground">{message}</p>
  {/if}

  {#snippet footer()}
    <div class="contents" bind:this={actions}>
      <Button variant="ghost" size="sm" onclick={() => onClose?.()}>
        {cancelLabel || t('common.cancel')}
      </Button>
      <Button
        variant={destructive ? 'destructive' : 'primary'}
        size="sm"
        onclick={() => onConfirm?.()}
      >
        {confirmLabel || t('common.confirm')}
      </Button>
    </div>
  {/snippet}
</Modal>
