<script lang="ts" module>
  export interface SelectOption {
    value: string
    label: string
    description?: string
  }
</script>

<script lang="ts">
  import { Check, ChevronDown } from '@lucide/svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'

  interface Props {
    value?: string
    options?: SelectOption[]
    placeholder?: string
    disabled?: boolean
    triggerClass?: string
    panelClass?: string
    onChange?: (detail: string) => void
  }

  let {
    value = $bindable(''),
    options = [],
    placeholder = 'Select',
    disabled = false,
    triggerClass = '',
    panelClass = '',
    onChange,
  }: Props = $props()

  let open = $state(false)
  let anchorElement = $state<HTMLButtonElement | null>(null)

  let selected = $derived(
    options.find((option) => option.value === value) ?? null,
  )

  const choose = (nextValue: string) => {
    value = nextValue
    open = false
    onChange?.(value)
  }
</script>

<button
  type="button"
  bind:this={anchorElement}
  {disabled}
  class={`inline-flex h-9 w-full items-center justify-between gap-3 rounded-md border bg-fill px-3 text-left text-sm text-white outline-none transition hover:border-line-strong focus-visible:border-primary/50 disabled:cursor-not-allowed disabled:opacity-55 ${
    open ? 'border-primary/50 bg-fill-hover' : 'border-border'
  } ${triggerClass}`}
  aria-haspopup="listbox"
  aria-expanded={open}
  onclick={() => (open = !open)}
>
  <span class="min-w-0 flex-1 truncate">
    {selected?.label ?? placeholder}
  </span>
  <ChevronDown
    class={`h-4 w-4 shrink-0 text-muted-foreground transition ${open ? 'rotate-180' : ''}`}
  />
</button>

<FloatingDropdown
  {open}
  {anchorElement}
  matchAnchorWidth
  offset={6}
  group="select"
  panelClass={`flex flex-col overflow-hidden rounded-md border border-border bg-popover shadow-xl shadow-black/40 ${panelClass}`}
  onClose={() => (open = false)}
>
  <div class="max-h-72 min-h-0 flex-1 overflow-y-auto p-1" role="listbox">
    {#each options as option (option.value)}
      <button
        type="button"
        class={`flex w-full items-center gap-2 rounded px-2.5 py-1.5 text-left transition ${
          option.value === value
            ? 'text-primary'
            : 'text-white hover:bg-fill-hover'
        }`}
        role="option"
        aria-selected={option.value === value}
        onclick={() => choose(option.value)}
      >
        <span class="min-w-0 flex-1">
          <span class="block truncate text-sm">
            {option.label}
          </span>
          {#if option.description}
            <span class="mt-0.5 block truncate text-xs text-muted-foreground">
              {option.description}
            </span>
          {/if}
        </span>
        {#if option.value === value}
          <Check class="h-4 w-4 shrink-0" />
        {/if}
      </button>
    {/each}
  </div>
</FloatingDropdown>
