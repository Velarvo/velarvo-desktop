<script lang="ts" module>
  export type IconButtonSize = 'xs' | 'sm' | 'md'
  export type IconButtonVariant = 'ghost' | 'secondary' | 'primary'
</script>

<script lang="ts">
  import type { HTMLButtonAttributes } from 'svelte/elements'
  import type { IconComponent } from '@/types/icon'
  import { tooltip as tooltipAttachment } from './tooltip'
  import type { TooltipPlacement } from './tooltip'

  interface Props extends Omit<HTMLButtonAttributes, 'class' | 'children'> {
    icon: IconComponent
    label: string
    size?: IconButtonSize
    variant?: IconButtonVariant
    pressed?: boolean
    tooltip?: string | boolean
    tooltipPlacement?: TooltipPlacement
    danger?: boolean
    ref?: HTMLButtonElement | null
    class?: string
    iconClass?: string
  }

  let {
    icon: Icon,
    label,
    size = 'md',
    variant = 'ghost',
    pressed = undefined,
    tooltip = true,
    tooltipPlacement = 'top',
    danger = false,
    type = 'button',
    ref = $bindable(null),
    class: className = '',
    iconClass = '',
    ...rest
  }: Props = $props()

  const boxes: Record<IconButtonSize, string> = {
    xs: 'h-6 w-6',
    sm: 'h-7 w-7',
    md: 'h-8 w-8',
  }
  const glyphs: Record<IconButtonSize, string> = {
    xs: 'h-3.5 w-3.5',
    sm: 'h-4 w-4',
    md: 'h-4 w-4',
  }
  const variants: Record<IconButtonVariant, string> = {
    ghost: 'text-muted-foreground hover:bg-fill-hover hover:text-white',
    secondary:
      'border border-border bg-fill text-white hover:border-line-strong hover:bg-fill-hover',
    primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
  }
</script>

<button
  bind:this={ref}
  {type}
  aria-label={label}
  {@attach tooltipAttachment({
    content: typeof tooltip === 'string' ? tooltip : label,
    placement: tooltipPlacement,
    disabled: tooltip === false,
  })}
  aria-pressed={pressed}
  class={[
    'inline-flex shrink-0 items-center justify-center rounded-md transition outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:pointer-events-none disabled:opacity-40',
    boxes[size],
    variants[variant],
    pressed && variant === 'ghost' && 'bg-primary/10 text-primary',
    danger && 'hover:bg-destructive/15 hover:text-red-300',
    className,
  ]}
  {...rest}
>
  <Icon class={`${glyphs[size]} ${iconClass}`} />
</button>
