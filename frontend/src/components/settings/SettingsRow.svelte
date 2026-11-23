<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    title: string
    description?: string
    children?: Snippet
  }

  let { title, description = '', children }: Props = $props()
</script>

<div
  class="flex items-center justify-between gap-6 border-b border-line-subtle py-4 last:border-b-0"
>
  <div class="min-w-0">
    <p class="text-sm font-medium text-white">{title}</p>
    {#if description}
      <p class="mt-0.5 text-sm text-muted-foreground">{description}</p>
    {/if}
  </div>
  {#if children}
    <div class="flex shrink-0 items-center gap-2">{@render children()}</div>
  {/if}
</div>
