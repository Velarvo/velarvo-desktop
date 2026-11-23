<script lang="ts" module>
  export type InputSize = 'sm' | 'md' | 'lg'
</script>

<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLInputAttributes } from 'svelte/elements'
  import { Eye, EyeOff } from '@lucide/svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import { t } from '@/lib/i18n.svelte'

  interface Props extends Omit<
    HTMLInputAttributes,
    'size' | 'value' | 'class'
  > {
    value?: string | number | null
    ref?: HTMLInputElement | null
    size?: InputSize
    mono?: boolean
    invalid?: boolean
    class?: string
    leading?: Snippet
    trailing?: Snippet
  }

  let {
    value = $bindable(''),
    ref = $bindable(null),
    type = 'text',
    size = 'md',
    mono = false,
    invalid = false,
    disabled = false,
    class: className = '',
    leading,
    trailing,
    ...rest
  }: Props = $props()

  let revealed = $state(false)

  const heights: Record<InputSize, string> = {
    sm: 'h-8',
    md: 'h-9',
    lg: 'h-11',
  }

  let isSecret = $derived(type === 'password')
  let hasTrailing = $derived(Boolean(trailing) || isSecret)
</script>

<div
  class={[
    'flex w-full min-w-0 items-center rounded-md border bg-fill text-sm transition',
    heights[size],
    invalid
      ? 'border-red-400/60'
      : 'border-border hover:border-line-strong focus-within:border-primary/50',
    'focus-within:bg-fill-hover',
    disabled && 'pointer-events-none opacity-55',
    className,
  ]}
>
  {#if leading}
    <span class="flex shrink-0 items-center pl-3 text-muted-foreground">
      {@render leading()}
    </span>
  {/if}

  <input
    bind:this={ref}
    bind:value
    type={isSecret && revealed ? 'text' : type}
    {disabled}
    aria-invalid={invalid || undefined}
    class={[
      'h-full min-w-0 flex-1 bg-transparent text-white outline-none placeholder:text-muted-foreground/60',
      leading ? 'pl-2.5' : 'pl-3',
      hasTrailing ? 'pr-1' : 'pr-3',
      (mono || isSecret) && 'font-mono',
    ]}
    {...rest}
  />

  {#if hasTrailing}
    <span class="flex shrink-0 items-center gap-1 pr-1">
      {#if isSecret}
        <IconButton
          icon={revealed ? EyeOff : Eye}
          size="sm"
          label={revealed
            ? t('ssh.hidePassword', 'Hide password')
            : t('ssh.showPassword', 'Show password')}
          tabindex={-1}
          onclick={() => (revealed = !revealed)}
        />
      {/if}
      {@render trailing?.()}
    </span>
  {/if}
</div>
