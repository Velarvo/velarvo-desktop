<script lang="ts" module>
  export const MENU_PANEL_CLASS =
    'min-w-44 rounded-md border border-border bg-popover p-1 shadow-xl shadow-black/40'
  export const MENU_SEPARATOR_CLASS = 'mx-1 my-1 block h-px bg-line-subtle'
</script>

<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLButtonAttributes } from 'svelte/elements'
  import type { IconComponent } from '@/types/icon'

  interface Props extends Omit<HTMLButtonAttributes, 'class' | 'children'> {
    icon?: IconComponent
    label?: string
    danger?: boolean
    selected?: boolean
    class?: string
    children?: Snippet
    trailing?: Snippet
  }

  let {
    icon: Icon,
    label = '',
    danger = false,
    selected = false,
    type = 'button',
    role = 'menuitem',
    class: className = '',
    children,
    trailing,
    ...rest
  }: Props = $props()
</script>

<button
  {type}
  {role}
  class={[
    'flex h-8 w-full items-center gap-2.5 rounded px-2.5 text-left text-sm transition outline-none disabled:pointer-events-none disabled:opacity-40',
    danger
      ? 'text-red-300 hover:bg-destructive/15 hover:text-red-200 focus-visible:bg-destructive/15'
      : selected
        ? 'bg-primary/10 text-primary'
        : 'text-fg-muted hover:bg-fill-hover hover:text-white focus-visible:bg-fill-hover focus-visible:text-white',
    className,
  ]}
  {...rest}
>
  {#if Icon}
    <Icon class="h-4 w-4 shrink-0 opacity-80" />
  {/if}
  <span class="min-w-0 flex-1 truncate">
    {#if children}{@render children()}{:else}{label}{/if}
  </span>
  {@render trailing?.()}
</button>
