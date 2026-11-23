<script lang="ts">
  import type { Snippet } from 'svelte'
  import { t } from '@/lib/i18n.svelte'

  interface Props {
    label: string
    for?: string
    layout?: 'stacked' | 'inline'
    hint?: string
    error?: string
    invalid?: boolean
    optional?: boolean
    class?: string
    children: Snippet
  }

  let {
    label,
    for: forId = undefined,
    layout = 'stacked',
    hint = '',
    error = '',
    invalid = false,
    optional = false,
    class: className = '',
    children,
  }: Props = $props()

  let isInvalid = $derived(invalid || Boolean(error))
  let labelClass = $derived([
    'flex items-baseline justify-between gap-2 text-[13px] transition-colors',
    isInvalid ? 'text-red-300' : 'text-muted-foreground',
    layout === 'inline' && '@md:h-9 @md:items-center',
  ])
</script>

<div
  class={[
    layout === 'inline'
      ? 'grid grid-cols-1 items-start gap-1.5 @md:grid-cols-[7.5rem_minmax(0,1fr)] @md:gap-x-4'
      : 'flex flex-col gap-1.5',
    className,
  ]}
>
  {#if forId}
    <label for={forId} class={labelClass}>
      {label}
      {#if optional}
        <span class="text-xs text-muted-foreground/60">
          {t('common.optional', 'Optional')}
        </span>
      {/if}
    </label>
  {:else}
    <span class={labelClass}>
      {label}
      {#if optional}
        <span class="text-xs text-muted-foreground/60">
          {t('common.optional', 'Optional')}
        </span>
      {/if}
    </span>
  {/if}

  <div class="min-w-0">
    {@render children()}
    {#if error}
      <p class="mt-1.5 text-xs text-red-300" role="alert">{error}</p>
    {:else if hint}
      <p class="mt-1.5 text-xs text-muted-foreground">{hint}</p>
    {/if}
  </div>
</div>
