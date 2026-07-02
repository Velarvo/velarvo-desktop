<script lang="ts" context="module">
  export type InfoTileTone = 'neutral' | 'primary'
</script>

<script lang="ts">
  import type { ComponentType } from 'svelte'

  export let icon: ComponentType
  export let title: string
  export let value = ''
  export let tone: InfoTileTone = 'neutral'

  const iconToneClasses: Record<InfoTileTone, string> = {
    primary: 'bg-primary/10 text-primary',
    neutral: 'bg-white/[0.05] text-muted-foreground',
  }
</script>

<div class="rounded-xl border border-white/10 bg-white/[0.035] p-4">
  <div class="flex items-center justify-between gap-3">
    <div class="flex min-w-0 items-center gap-3">
      <div
        class={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${iconToneClasses[tone]}`}
      >
        <svelte:component this={icon} class="h-4 w-4" />
      </div>
      <div class="min-w-0">
        <p class="truncate text-sm font-semibold text-white">{title}</p>
        {#if value}
          <p class="truncate text-xs text-muted-foreground">{value}</p>
        {/if}
      </div>
    </div>

    {#if $$slots.trailing}
      <div class="shrink-0"><slot name="trailing" /></div>
    {/if}
  </div>
</div>
