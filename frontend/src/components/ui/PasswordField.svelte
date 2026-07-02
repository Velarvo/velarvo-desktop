<script lang="ts">
  import { Eye, EyeOff } from 'lucide-svelte'

  export let id: string
  export let label: string
  export let value = ''
  export let placeholder = ''
  export let autocomplete: string | undefined = undefined
  export let hint = ''
  export let disabled = false
  export let showLabel = 'Show password'
  export let hideLabel = 'Hide password'

  let visible = false

  const inputClass = [
    'h-11 w-full rounded-md border border-border bg-card px-3 pr-10 text-sm font-mono text-white outline-none transition placeholder:text-white/25 focus:border-primary/35',
  ].join(' ')
</script>

<div class="space-y-1.5">
  <label
    for={id}
    class="text-xs font-medium uppercase tracking-wide text-muted-foreground"
  >
    {label}
  </label>

  <div class="relative">
    {#if visible}
      <input
        {id}
        type="text"
        {placeholder}
        {autocomplete}
        {disabled}
        bind:value
        class={inputClass}
      />
    {:else}
      <input
        {id}
        type="password"
        {placeholder}
        {autocomplete}
        {disabled}
        bind:value
        class={inputClass}
      />
    {/if}

    <button
      type="button"
      class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-white"
      aria-pressed={visible}
      aria-label={visible ? hideLabel : showLabel}
      on:click={() => (visible = !visible)}
    >
      {#if visible}
        <EyeOff class="h-4 w-4" />
      {:else}
        <Eye class="h-4 w-4" />
      {/if}
    </button>
  </div>

  {#if hint}
    <p class="text-xs text-muted-foreground">{hint}</p>
  {/if}
</div>
