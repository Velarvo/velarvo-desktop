<script lang="ts" module>
  export type AlertVariant = 'destructive' | 'info' | 'warning'
</script>

<script lang="ts">
  import type { Snippet } from 'svelte'
  import { AlertCircle, Info, AlertTriangle } from '@lucide/svelte'

  interface Props {
    variant?: AlertVariant
    message?: string
    children?: Snippet
  }

  let { variant = 'destructive', message = '', children }: Props = $props()

  const variantClasses: Record<AlertVariant, string> = {
    destructive:
      'border-destructive/30 bg-destructive/10 text-red-200 [&_svg]:text-destructive',
    info: 'border-primary/25 bg-primary/10 text-primary/90 [&_svg]:text-primary',
    warning:
      'border-warning/30 bg-warning/10 text-amber-100 [&_svg]:text-warning',
  }

  const icons = {
    destructive: AlertCircle,
    info: Info,
    warning: AlertTriangle,
  }

  let Icon = $derived(icons[variant])
</script>

<div
  class={`flex items-start gap-2.5 rounded-md border px-3 py-2.5 text-sm ${variantClasses[variant]}`}
  role="alert"
>
  <Icon class="mt-0.5 h-4 w-4 shrink-0" />
  <span class="min-w-0 flex-1 break-words">
    {#if children}{@render children?.()}{:else}{message}{/if}
  </span>
</div>
