<script lang="ts">
  import { Check } from '@lucide/svelte'
  import { ACCENT_COLORS } from '@/lib/colors'

  interface Props {
    value: string
    colors?: readonly string[]
    label?: string | undefined
    disabled?: boolean
  }

  let {
    value = $bindable(),
    colors = ACCENT_COLORS,
    label = undefined,
    disabled = false,
  }: Props = $props()

  const select = (color: string) => {
    if (disabled) return
    value = color
  }
</script>

<div class="space-y-2">
  {#if label}
    <span class="block text-[13px] text-muted-foreground">
      {label}
    </span>
  {/if}

  <div role="radiogroup" aria-label={label} class="grid grid-cols-9 gap-2">
    {#each colors as color (color)}
      {@const selected = value === color}
      <button
        type="button"
        role="radio"
        aria-checked={selected}
        aria-label={color}
        {disabled}
        class="relative flex aspect-square items-center justify-center rounded-md outline-none transition hover:scale-110 focus-visible:ring-2 focus-visible:ring-primary/40 disabled:cursor-not-allowed disabled:opacity-50 {selected
          ? 'ring-2 ring-white/80 ring-offset-2 ring-offset-surface-raised'
          : 'ring-1 ring-inset ring-line'}"
        style={`background:${color};`}
        onclick={() => select(color)}
      >
        {#if selected}
          <Check
            class="h-3.5 w-3.5 text-white drop-shadow-[0_1px_1px_rgba(0,0,0,0.5)]"
            strokeWidth={3}
          />
        {/if}
      </button>
    {/each}
  </div>
</div>
