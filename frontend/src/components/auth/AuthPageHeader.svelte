<script lang="ts">
  import type { ComponentType } from 'svelte'

  export let icon: ComponentType | undefined = undefined
  export let title: string
  export let subtitle = ''
</script>

<header class="space-y-2 text-center">
  {#if icon}
    <div
      class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10 text-primary"
    >
      <svelte:component this={icon} class="h-5 w-5" />
    </div>
  {/if}
  <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-white">
    {title}
  </h2>
  {#if subtitle}
    <p class="text-sm text-muted-foreground">{subtitle}</p>
  {/if}
</header>
