<script lang="ts" context="module">
  export type StatusBadgeTone = 'active' | 'muted'
</script>

<script lang="ts">
  export let tone: StatusBadgeTone = 'muted'
  export let label = ''

  const toneClasses: Record<StatusBadgeTone, string> = {
    active: 'border-primary/25 bg-primary/10 text-primary',
    muted: 'border-white/10 bg-white/[0.04] text-muted-foreground',
  }
</script>

<span
  class={`rounded-full border px-2.5 py-1 text-[11px] font-semibold ${toneClasses[tone]}`}
>
  {#if $$slots.default}<slot />{:else}{label}{/if}
</span>
