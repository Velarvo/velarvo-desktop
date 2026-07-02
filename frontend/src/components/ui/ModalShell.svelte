<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte'
  import { X } from 'lucide-svelte'

  export let open = false
  export let title = ''
  export let subtitle = ''
  export let closeLabel = 'Close'
  export let titleId = 'modal-title'
  export let descriptionId = 'modal-description'
  export let maxWidthClass = 'max-w-3xl'
  export let headerClass = 'px-6 py-5'
  export let bodyClass = ''
  export let closeDisabled = false

  const dispatch = createEventDispatcher<{ close: void }>()

  const close = () => {
    if (closeDisabled) return
    dispatch('close')
  }

  const handleKeydown = (event: KeyboardEvent) => {
    if (!open || event.key !== 'Escape') {
      return
    }

    event.preventDefault()
    close()
  }

  onMount(() => {
    if (typeof window === 'undefined') {
      return
    }

    window.addEventListener('keydown', handleKeydown)
  })

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('keydown', handleKeydown)
    }
  })
</script>

{#if open}
  <div class="fixed inset-0 z-[1200]">
    <button
      type="button"
      class="absolute inset-0 bg-surface-overlay/78 backdrop-blur-md"
      aria-label={closeLabel}
      on:click={close}
    ></button>

    <div class="relative flex h-full items-center justify-center p-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={subtitle ? descriptionId : undefined}
        class={`relative w-full overflow-hidden rounded-2xl border border-white/10 bg-[linear-gradient(180deg,_var(--color-modal-start),_var(--color-modal-end))] shadow-[0_30px_120px_rgba(0,0,0,0.58)] ${maxWidthClass}`}
      >
        <div
          class="absolute inset-x-8 top-0 h-px bg-gradient-to-r from-transparent via-primary/50 to-transparent"
        ></div>
        <slot name="backdrop"></slot>

        <header
          class={`relative flex items-start justify-between gap-5 border-b border-white/8 ${headerClass}`}
        >
          <div class="flex min-w-0 items-start gap-4">
            <slot name="icon"></slot>
            <div class="min-w-0">
              <h2 id={titleId} class="text-lg font-semibold text-white">
                {title}
              </h2>
              {#if subtitle}
                <p
                  id={descriptionId}
                  class="mt-1 max-w-xl text-sm text-muted-foreground"
                >
                  {subtitle}
                </p>
              {/if}
            </div>
          </div>

          <button
            type="button"
            class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04] text-muted-foreground transition hover:border-white/15 hover:bg-white/[0.07] hover:text-white disabled:cursor-not-allowed disabled:opacity-45"
            aria-label={closeLabel}
            disabled={closeDisabled}
            on:click={close}
          >
            <X class="h-4 w-4" />
          </button>
        </header>

        <div class={bodyClass}>
          <slot></slot>
        </div>
      </div>
    </div>
  </div>
{/if}
