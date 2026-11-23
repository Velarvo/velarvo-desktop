<script lang="ts" module>
  export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'destructive'
  export type ButtonSize = 'xs' | 'sm' | 'md' | 'lg'
</script>

<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'
  import { LoaderCircle } from '@lucide/svelte'

  interface Props extends Omit<HTMLButtonAttributes, 'class'> {
    variant?: ButtonVariant
    size?: ButtonSize
    loading?: boolean
    fullWidth?: boolean
    shortcut?: string
    ref?: HTMLButtonElement | null
    class?: string
    children?: Snippet
  }

  let {
    type = 'button',
    variant = 'primary',
    size = 'md',
    disabled = false,
    loading = false,
    fullWidth = false,
    shortcut = '',
    ref = $bindable(null),
    class: className = '',
    children,
    ...rest
  }: Props = $props()

  const variants: Record<ButtonVariant, string> = {
    primary:
      'bg-primary font-semibold text-primary-foreground hover:bg-primary/90',
    secondary:
      'border border-border bg-fill font-medium text-white hover:border-line-strong hover:bg-fill-hover',
    ghost:
      'font-medium text-muted-foreground hover:bg-fill-hover hover:text-white',
    destructive:
      'border border-destructive/45 bg-destructive/15 font-semibold text-red-200 hover:bg-destructive/25',
  }

  const sizes: Record<ButtonSize, string> = {
    xs: 'h-6 gap-1.5 px-2 text-xs',
    sm: 'h-8 px-3 text-sm',
    md: 'h-9 px-3.5 text-sm',
    lg: 'h-11 px-4 text-sm',
  }

  let isDisabled = $derived(Boolean(disabled) || loading)
</script>

<button
  bind:this={ref}
  {type}
  disabled={isDisabled}
  aria-busy={loading || undefined}
  class={[
    'inline-flex shrink-0 items-center justify-center gap-2 rounded-md whitespace-nowrap transition outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:cursor-not-allowed disabled:opacity-45',
    variants[variant],
    sizes[size],
    fullWidth && 'w-full',
    shortcut && 'pr-2',
    className,
  ]}
  {...rest}
>
  {#if loading}
    <LoaderCircle class="h-4 w-4 animate-spin" />
  {/if}
  {@render children?.()}
  {#if shortcut}
    <kbd
      class={[
        'rounded px-1.5 py-0.5 font-mono text-[10px] font-semibold',
        variant === 'primary'
          ? 'bg-black/15'
          : 'bg-fill-active text-muted-foreground',
      ]}
    >
      {shortcut}
    </kbd>
  {/if}
</button>
