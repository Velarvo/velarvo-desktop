<script lang="ts" context="module">
  export type AlertVariant = 'destructive' | 'info' | 'warning'
</script>

<script lang="ts">
  import { AlertCircle, Info, AlertTriangle } from 'lucide-svelte'

  export let variant: AlertVariant = 'destructive'
  export let message = ''

  const variantClasses: Record<AlertVariant, string> = {
    destructive:
      'border-destructive/30 bg-destructive/10 text-red-200 [&_svg]:text-destructive',
    info: 'border-primary/25 bg-primary/10 text-primary [&_svg]:text-primary',
    warning:
      'border-amber-400/30 bg-amber-400/10 text-amber-200 [&_svg]:text-amber-300',
  }

  const icons = {
    destructive: AlertCircle,
    info: Info,
    warning: AlertTriangle,
  }

  $: Icon = icons[variant]
</script>

<div
  class={`flex items-center gap-2 rounded-xl border px-4 py-3 text-sm ${variantClasses[variant]}`}
  role="alert"
>
  <svelte:component this={Icon} class="h-4 w-4 shrink-0" />
  <span class="min-w-0 flex-1">
    {#if $$slots.default}<slot />{:else}{message}{/if}
  </span>
</div>
