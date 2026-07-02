<script lang="ts">
  export let id: string
  export let label: string
  export let value = ''
  export let type: 'text' | 'email' | 'password' = 'text'
  export let placeholder = ''
  export let autocomplete: string | undefined = undefined
  export let mono = false
  export let hint = ''
  export let disabled = false
  export let invalid = false

  $: hintId = hint ? `${id}-hint` : undefined

  $: inputClass = [
    'h-11 w-full rounded-md border border-border bg-card px-3 text-sm text-white outline-none transition placeholder:text-white/25 focus:border-primary/35',
    mono ? 'font-mono' : '',
    $$slots.trailing ? 'pr-10' : '',
  ]
    .filter(Boolean)
    .join(' ')
</script>

<div class="space-y-1.5">
  <label
    for={id}
    class="text-xs font-medium uppercase tracking-wide text-muted-foreground"
  >
    {label}
  </label>

  <div class="relative">
    {#if type === 'email'}
      <input
        {id}
        type="email"
        {placeholder}
        {autocomplete}
        {disabled}
        aria-invalid={invalid}
        aria-describedby={hintId}
        bind:value
        class={inputClass}
      />
    {:else if type === 'password'}
      <input
        {id}
        type="password"
        {placeholder}
        {autocomplete}
        {disabled}
        aria-invalid={invalid}
        aria-describedby={hintId}
        bind:value
        class={inputClass}
      />
    {:else}
      <input
        {id}
        type="text"
        {placeholder}
        {autocomplete}
        {disabled}
        aria-invalid={invalid}
        aria-describedby={hintId}
        bind:value
        class={inputClass}
      />
    {/if}

    {#if $$slots.trailing}
      <div class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center">
        <slot name="trailing" />
      </div>
    {/if}
  </div>

  {#if hint}
    <p id={hintId} class="text-xs text-muted-foreground">{hint}</p>
  {/if}
</div>
