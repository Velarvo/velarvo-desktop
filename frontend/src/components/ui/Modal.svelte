<script lang="ts" module>
  export type ModalSize = 'sm' | 'md' | 'lg' | 'xl'
</script>

<script lang="ts">
  import { tick, type Snippet } from 'svelte'
  import type { Attachment } from 'svelte/attachments'
  import { X } from '@lucide/svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import { t } from '@/lib/i18n.svelte'

  interface Props {
    open: boolean
    title: string
    description?: string
    size?: ModalSize
    dismissible?: boolean
    onsubmit?: () => void
    flush?: boolean
    onClose: () => void
    icon?: Snippet
    children: Snippet
    footer?: Snippet
  }

  let {
    open,
    title,
    description = '',
    size = 'md',
    dismissible = true,
    onsubmit,
    flush = false,
    onClose,
    icon,
    children,
    footer,
  }: Props = $props()

  const widths: Record<ModalSize, string> = {
    sm: 'max-w-sm',
    md: 'max-w-lg',
    lg: 'max-w-2xl',
    xl: 'max-w-4xl',
  }

  const id = $props.id()
  let dialog = $state<HTMLDialogElement>()

  const close = () => {
    if (dismissible) onClose()
  }

  $effect(() => {
    if (!dialog) return
    if (open && !dialog.open) {
      dialog.showModal()
      void focusFirstField()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  })

  const focusFirstField = async () => {
    await tick()
    dialog
      ?.querySelector<HTMLElement>(
        'input:not([type=hidden]):not([type=file]):not([disabled]), textarea',
      )
      ?.focus()
  }

  const closeOnBackdrop: Attachment<HTMLDialogElement> = (node) => {
    const onClick = (event: MouseEvent) => {
      if (event.target === node) close()
    }
    node.addEventListener('click', onClick)
    return () => node.removeEventListener('click', onClick)
  }

  const submit = (event: SubmitEvent) => {
    event.preventDefault()
    onsubmit?.()
  }
</script>

<dialog
  bind:this={dialog}
  {@attach closeOnBackdrop}
  aria-labelledby="{id}-title"
  aria-describedby={description ? `${id}-description` : undefined}
  oncancel={(event) => {
    event.preventDefault()
    close()
  }}
  class="m-auto max-h-[calc(100vh-2rem)] w-[calc(100%-2rem)] {widths[
    size
  ]} overflow-hidden rounded-lg border border-line bg-surface-raised p-0 text-foreground shadow-[0_30px_120px_rgba(0,0,0,0.58)] backdrop:bg-surface-overlay/78 backdrop:backdrop-blur-md"
>
  {#if open}
    <div class="relative flex max-h-[calc(100vh-2rem)] flex-col">
      <header
        class="flex shrink-0 items-start gap-3 border-b border-line-subtle px-5 py-4"
      >
        {#if icon}
          <span class="mt-0.5 shrink-0">{@render icon()}</span>
        {/if}
        <div class="min-w-0 flex-1">
          <h2
            id="{id}-title"
            class="truncate text-base font-semibold text-white"
          >
            {title}
          </h2>
          {#if description}
            <p
              id="{id}-description"
              class="mt-0.5 text-sm text-muted-foreground"
            >
              {description}
            </p>
          {/if}
        </div>
        <IconButton
          icon={X}
          class="-mr-1"
          label={t('common.close')}
          disabled={!dismissible}
          onclick={close}
        />
      </header>

      {#if onsubmit}
        <form class="flex min-h-0 flex-1 flex-col" onsubmit={submit}>
          {@render body()}
        </form>
      {:else}
        {@render body()}
      {/if}
    </div>
  {/if}
</dialog>

{#snippet body()}
  <div class="min-h-0 flex-1 overflow-y-auto {flush ? '' : 'px-5 py-5'}">
    {@render children()}
  </div>
  {#if footer}
    <footer
      class="flex shrink-0 items-center justify-end gap-2 border-t border-line-subtle bg-fill-subtle px-5 py-3"
    >
      {@render footer()}
    </footer>
  {/if}
{/snippet}
