<script lang="ts">
  import type { ComponentType } from 'svelte'

  export let icon: ComponentType
  export let title: string
  export let description = ''
</script>

<div
  class="rounded-2xl border border-white/10 bg-[radial-gradient(circle_at_top_right,_rgb(var(--color-primary-rgb)_/_0.12),_transparent_42%),rgba(255,255,255,0.035)] p-5"
>
  <div class="flex items-start gap-4">
    <div
      class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10 text-primary"
    >
      <svelte:component this={icon} class="h-5 w-5" />
    </div>

    <div class="min-w-0 flex-1">
      <h4 class="text-base font-semibold text-white">{title}</h4>
      {#if description}
        <p class="mt-1 text-sm leading-6 text-muted-foreground">
          {description}
        </p>
      {/if}
      {#if $$slots.default}
        <div class="mt-4"><slot /></div>
      {/if}
    </div>
  </div>
</div>
