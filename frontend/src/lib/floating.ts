export type Side = 'top' | 'bottom' | 'left' | 'right'
export type Align = 'start' | 'center' | 'end'
export type Placement = Side | `${Side}-start` | `${Side}-end`

export interface Size {
  width: number
  height: number
}

export interface PositionOptions {
  placement: Placement
  offset: number
  margin: number
  flip?: boolean
  fitViewport?: boolean
}

export interface Position {
  side: Side
  left: number
  top: number
  maxHeight: number | null
  arrow: number
}

const OPPOSITE: Record<Side, Side> = {
  top: 'bottom',
  bottom: 'top',
  left: 'right',
  right: 'left',
}

export const isVertical = (side: Side): boolean =>
  side === 'top' || side === 'bottom'

export const parsePlacement = (
  placement: Placement,
): { side: Side; align: Align } => {
  const [side, align] = placement.split('-') as [Side, Align?]
  return { side, align: align ?? 'center' }
}

export const viewportSize = (): Size => ({
  width: document.documentElement.clientWidth || window.innerWidth,
  height: document.documentElement.clientHeight || window.innerHeight,
})

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), Math.max(min, max))

const alignOnAxis = (
  align: Align,
  refStart: number,
  refEnd: number,
  size: number,
  viewport: number,
  margin: number,
) => {
  const startAligned = refStart
  const endAligned = refEnd - size
  const fits = (value: number) =>
    value >= margin && value + size <= viewport - margin

  let value =
    align === 'start'
      ? startAligned
      : align === 'end'
        ? endAligned
        : refStart + (refEnd - refStart) / 2 - size / 2

  if (!fits(value) && align !== 'center') {
    const mirrored = align === 'start' ? endAligned : startAligned
    if (fits(mirrored)) value = mirrored
  }

  return clamp(value, margin, viewport - size - margin)
}

const resolveSide = (
  preferred: Side,
  space: Record<Side, number>,
  size: Size,
  flip: boolean,
): Side => {
  const fitsOn = (side: Side) =>
    space[side] >= (isVertical(side) ? size.height : size.width)

  if (!flip || fitsOn(preferred)) return preferred

  const opposite = OPPOSITE[preferred]
  const perpendicular = (
    isVertical(preferred)
      ? (['right', 'left'] as const)
      : (['bottom', 'top'] as const)
  )
    .slice()
    .sort((a, b) => space[b] - space[a])
  const fitting = [opposite, ...perpendicular].find(fitsOn)
  if (fitting) return fitting

  if (
    !isVertical(preferred) &&
    space[preferred] >= size.width &&
    space[preferred] >= space[opposite]
  ) {
    return preferred
  }
  return space.top > space.bottom ? 'top' : 'bottom'
}

export const computePosition = (
  reference: DOMRect,
  floating: Size,
  {
    placement,
    offset,
    margin,
    flip = true,
    fitViewport = false,
  }: PositionOptions,
): Position => {
  const viewport = viewportSize()
  const space: Record<Side, number> = {
    top: reference.top - offset - margin,
    bottom: viewport.height - reference.bottom - offset - margin,
    left: reference.left - offset - margin,
    right: viewport.width - reference.right - offset - margin,
  }

  const { side: preferred, align } = parsePlacement(placement)
  const side = resolveSide(preferred, space, floating, flip)

  const availableHeight = isVertical(side)
    ? Math.max(0, Math.floor(space[side]))
    : Math.max(0, viewport.height - margin * 2)
  const clipped = fitViewport && floating.height > availableHeight
  const height = clipped ? availableHeight : floating.height
  const { width } = floating

  let left: number
  let top: number

  if (isVertical(side)) {
    top =
      side === 'bottom'
        ? reference.bottom + offset
        : reference.top - offset - height
    left = alignOnAxis(
      align,
      reference.left,
      reference.right,
      width,
      viewport.width,
      margin,
    )
  } else {
    left =
      side === 'right'
        ? reference.right + offset
        : reference.left - offset - width
    left = clamp(left, margin, viewport.width - width - margin)
    top = alignOnAxis(
      align,
      reference.top,
      reference.bottom,
      height,
      viewport.height,
      margin,
    )
  }

  top = clamp(top, margin, viewport.height - height - margin)

  const arrow = isVertical(side)
    ? reference.left + reference.width / 2 - left
    : reference.top + reference.height / 2 - top

  return {
    side,
    left: Math.round(left),
    top: Math.round(top),
    maxHeight: clipped ? availableHeight : null,
    arrow: Math.round(arrow),
  }
}
