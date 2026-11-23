<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    title: string
    class?: string
    children: Snippet
  }

  let { title, class: className = '', children }: Props = $props()
</script>

<section class={className}>
  <h3
    class="mb-3 text-[11px] font-semibold uppercase tracking-[0.14em] text-muted-foreground/70"
  >
    {title}
  </h3>
  <div class="space-y-3">
    {@render children()}
  </div>
</section>
