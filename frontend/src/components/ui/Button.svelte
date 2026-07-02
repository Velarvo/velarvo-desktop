<script lang="ts" context="module">
  export type ButtonVariant =
    | 'primary'
    | 'primary-glow'
    | 'secondary'
    | 'ghost'
    | 'outline'
    | 'destructive'
  export type ButtonSize = 'sm' | 'md' | 'lg'
  export type ButtonRadius = 'md' | 'xl'
</script>

<script lang="ts">
  import { LoaderCircle } from 'lucide-svelte'

  export let type: 'button' | 'submit' | 'reset' = 'button'
  export let variant: ButtonVariant = 'primary'
  export let size: ButtonSize = 'md'
  export let radius: ButtonRadius = 'xl'
  export let disabled = false
  export let loading = false
  export let fullWidth = false
  export let href: string | undefined = undefined
  export let ariaLabel: string | undefined = undefined
  let extraClass = ''
  export { extraClass as class }

  const variantClasses: Record<ButtonVariant, string> = {
    primary: 'bg-primary text-primary-foreground hover:brightness-105',
    'primary-glow':
      'border border-primary/30 bg-primary text-primary-foreground shadow-[0_10px_30px_rgb(var(--color-primary-rgb)_/_0.22)] hover:brightness-105',
    secondary:
      'border border-white/10 bg-white/[0.04] text-white hover:border-white/15 hover:bg-white/[0.07]',
    ghost:
      'border border-transparent text-white/80 hover:bg-white/[0.05] hover:text-white',
    outline:
      'border border-white/10 text-white/80 hover:border-white/15 hover:bg-white/[0.05] hover:text-white',
    destructive:
      'border border-destructive/40 bg-destructive/10 text-destructive hover:bg-destructive/15',
  }

  const sizeClasses: Record<ButtonSize, string> = {
    sm: 'h-9 px-3 text-xs',
    md: 'h-11 px-4 text-sm',
    lg: 'h-12 px-5 text-sm',
  }

  const radiusClasses: Record<ButtonRadius, string> = {
    md: 'rounded-md',
    xl: 'rounded-xl',
  }

  $: classes = [
    'inline-flex items-center justify-center gap-2 font-semibold transition disabled:cursor-not-allowed disabled:opacity-60',
    radiusClasses[radius],
    variantClasses[variant],
    sizeClasses[size],
    fullWidth ? 'w-full' : '',
    isDisabled ? 'pointer-events-none opacity-60' : '',
    extraClass,
  ]
    .filter(Boolean)
    .join(' ')

  $: isDisabled = disabled || loading

  const handleDisabledClick = (event: MouseEvent) => {
    if (!isDisabled) return
    event.preventDefault()
    event.stopPropagation()
  }
</script>

{#if href}
  <a
    {href}
    class={classes}
    aria-label={ariaLabel}
    aria-disabled={isDisabled}
    aria-busy={loading}
    tabindex={isDisabled ? -1 : undefined}
    on:click={handleDisabledClick}
  >
    {#if loading}
      <LoaderCircle class="h-4 w-4 animate-spin" />
    {/if}
    <slot />
  </a>
{:else}
  <button
    {type}
    class={classes}
    disabled={isDisabled}
    aria-label={ariaLabel}
    aria-busy={loading}
    on:click
  >
    {#if loading}
      <LoaderCircle class="h-4 w-4 animate-spin" />
    {/if}
    <slot />
  </button>
{/if}
