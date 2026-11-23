<script lang="ts" generics="T extends string">
  interface Option {
    value: T
    label: string
  }

  interface Props {
    value: T
    options: Option[]
    label: string
    disabled?: boolean
    onChange?: (value: T) => void
  }

  let { value, options, label, disabled = false, onChange }: Props = $props()
</script>

<div
  role="radiogroup"
  aria-label={label}
  class="inline-flex rounded-md border border-border bg-fill p-0.5"
>
  {#each options as option (option.value)}
    {@const active = value === option.value}
    <button
      type="button"
      role="radio"
      aria-checked={active}
      {disabled}
      class={[
        'h-7 rounded px-3 text-sm font-medium transition outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:opacity-50',
        active
          ? 'bg-fill-active text-white shadow-[inset_0_0_0_1px_var(--color-line-subtle)]'
          : 'text-muted-foreground hover:text-white',
      ]}
      onclick={() => onChange?.(option.value)}
    >
      {option.label}
    </button>
  {/each}
</div>
