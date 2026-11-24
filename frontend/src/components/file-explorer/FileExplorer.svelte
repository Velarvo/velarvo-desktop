<script lang="ts">
  import { tooltip } from '@/components/ui/tooltip'
  import { onDestroy, onMount, tick } from 'svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import Input from '@/components/ui/Input.svelte'
  import {
    RefreshCw,
    ChevronLeft,
    ChevronRight,
    Home,
    FolderUp,
    Search,
    Copy,
    Check,
    Eye,
    EyeOff,
    ArrowUp,
    ArrowDown,
  } from '@lucide/svelte'
  import {
    type ExplorerBreadcrumb,
    type ExplorerColumn,
    type ExplorerColumnBreakpoint,
    type ExplorerDataSource,
    type ExplorerLabels,
    type ExplorerNode,
    type SortOrder,
  } from './types'
  import FileIcon from './FileIcon.svelte'
  import {
    fileStemLength,
    getBreadcrumbs,
    getParentPath,
    normalizePath,
  } from './utils'
  import { createSmartCache } from './hooks/useSmartCache'
  import { copyText } from '@/lib/clipboard'
  import { t } from '@/lib/i18n.svelte'
  import type { IconComponent } from '@/types/icon'

  type NavigationOptions = {
    preserveHistory?: boolean
    forceRefresh?: boolean
  }

  const nameTooltip = (content: string) =>
    tooltip({ content, placement: 'top-start', overflowOnly: true })

  const cloneExplorerNodes = (items: ExplorerNode[]) => {
    return items.map((item) => ({
      ...item,
      modified: item.modified ? new Date(item.modified) : undefined,
    }))
  }

  const cache = createSmartCache<ExplorerNode[]>({
    maxSize: 200,
    ttl: 30000,
    staleTtl: 120000,
    clone: cloneExplorerNodes,
  })
  // eslint-disable-next-line svelte/prefer-svelte-reactivity
  const pendingLoads = new Map<string, Promise<ExplorerNode[]>>()
  let selection = $state(new Set<string>())
  let anchorPath: string | null = null
  let focusPath: string | null = null

  interface Props {
    dataSource: ExplorerDataSource
    columns?: ExplorerColumn[]
    labels?: Partial<ExplorerLabels>
    initialPath?: string
    defaultSortBy?: string
    defaultSortOrder?: SortOrder
    showCopyAction?: boolean
    showHiddenToggle?: boolean
    groupContainersFirst?: boolean
    className?: string
    viewportClassName?: string
    contentClassName?: string
    fillHeight?: boolean
    viewportHeight?: string
    viewportMaxHeight?: string
    draggableNodes?: boolean
    contextMenuNodes?: boolean
    containerDropTargets?: boolean
    dropActive?: boolean
    onActivate?: (detail: { node: ExplorerNode; navigated: boolean }) => void
    onSelect?: (detail: {
      node: ExplorerNode | null
      nodes: ExplorerNode[]
    }) => void
    onNodeDragStart?: (detail: {
      node: ExplorerNode
      nodes: ExplorerNode[]
    }) => void
    onNodeDragEnd?: () => void
    onNodeContextMenu?: (detail: {
      node: ExplorerNode
      nodes: ExplorerNode[]
      x: number
      y: number
    }) => void
    onRenameNode?: (detail: { node: ExplorerNode; name: string }) => void
    onCreateFolder?: (detail: { directory: string; name: string }) => void
    onCreateFile?: (detail: { directory: string; name: string }) => void
    onNodeDrop?: (detail: { node: ExplorerNode }) => void
    onNodeDropTarget?: (detail: { node: ExplorerNode | null }) => void
    onDeleteNodes?: (detail: { nodes: ExplorerNode[] }) => void
    onError?: (detail: { path: string; error: Error }) => void
    onLoad?: (detail: { path: string; nodes: ExplorerNode[] }) => void
    onPathChange?: (detail: { path: string }) => void
  }

  let {
    dataSource,
    columns = [],
    labels = {},
    initialPath = '',
    defaultSortBy = 'name',
    defaultSortOrder = 'asc',
    showCopyAction = true,
    showHiddenToggle = true,
    groupContainersFirst = true,
    className = '',
    viewportClassName = '',
    contentClassName = '',
    fillHeight = false,
    viewportHeight = '',
    viewportMaxHeight = '32rem',
    draggableNodes = false,
    contextMenuNodes = false,
    containerDropTargets = false,
    dropActive = false,
    onActivate,
    onSelect,
    onNodeDragStart,
    onNodeDragEnd,
    onNodeContextMenu,
    onRenameNode,
    onCreateFolder,
    onCreateFile,
    onNodeDrop,
    onNodeDropTarget,
    onDeleteNodes,
    onError,
    onLoad,
    onPathChange,
  }: Props = $props()

  let currentPath = $state('')
  let homePath = $state('')
  let isLoading = $state(true)
  let isRefreshing = $state(false)
  let errorMessage = $state('')
  let showHiddenItems = $state(true)
  let searchQuery = $state('')
  let sortBy = $derived(defaultSortBy)
  let sortOrder = $derived<SortOrder>(defaultSortOrder)
  let nodes = $state<ExplorerNode[]>([])
  let history = $state<string[]>([])
  let historyIndex = $state(-1)
  let copiedPath = $state(false)
  let copiedResetTimeout: ReturnType<typeof setTimeout> | null = null
  let requestId = 0
  let cacheGeneration = 0
  let bootstrapId = 0
  let isMounted = $state(false)
  let previousDataSource: ExplorerDataSource | undefined
  let previousInitialPath = ''

  const normalizeCurrentPath = (path: string) => {
    return dataSource?.normalizePath?.(path) ?? normalizePath(path)
  }

  const resolveParentPath = (path: string) => {
    return dataSource?.getParentPath?.(path) ?? getParentPath(path)
  }

  const resolveBreadcrumbs = (path: string) => {
    return dataSource?.getBreadcrumbs?.(path) ?? getBreadcrumbs(path)
  }

  const isContainerNode = (node: ExplorerNode) => {
    return dataSource?.isContainer?.(node) ?? Boolean(node.isContainer)
  }

  const isHiddenNode = (node: ExplorerNode) => {
    return dataSource?.isHidden?.(node) ?? node.name.startsWith('.')
  }

  const getCacheKey = (path: string) => {
    return dataSource?.getCacheKey?.(path) ?? path
  }

  const getScopedCacheKey = (path: string, generation = cacheGeneration) => {
    return `${generation}:${getCacheKey(path)}`
  }

  const getCopyValue = (path: string) => {
    return dataSource?.getCopyValue?.(path) ?? path
  }

  const getColumn = (columnId: string) => {
    return columns.find((column) => column.id === columnId) ?? null
  }

  const HIDE_BELOW_CLASS: Record<ExplorerColumnBreakpoint, string> = {
    xs: '@max-xs:hidden',
    sm: '@max-sm:hidden',
    md: '@max-md:hidden',
    lg: '@max-lg:hidden',
    xl: '@max-xl:hidden',
    '2xl': '@max-2xl:hidden',
  }

  const getResponsiveHideClass = (column: ExplorerColumn) => {
    return column.hideBelow ? HIDE_BELOW_CLASS[column.hideBelow] : ''
  }

  const getSortValue = (node: ExplorerNode, field: string) => {
    if (field === 'name') {
      return node.name
    }

    const column = getColumn(field)
    if (!column) {
      return ''
    }

    return column.sortValue?.(node) ?? column.render(node)
  }

  const compareValues = (
    left: string | number | null | undefined,
    right: string | number | null | undefined,
  ) => {
    if (typeof left === 'number' && typeof right === 'number') {
      return left - right
    }

    return String(left ?? '').localeCompare(String(right ?? ''), undefined, {
      numeric: true,
      sensitivity: 'base',
    })
  }

  const compareNodes = (left: ExplorerNode, right: ExplorerNode) => {
    if (groupContainersFirst) {
      const leftRank = isContainerNode(left) ? 0 : 1
      const rightRank = isContainerNode(right) ? 0 : 1
      if (leftRank !== rightRank) {
        return leftRank - rightRank
      }
    }

    const comparison = compareValues(
      getSortValue(left, sortBy),
      getSortValue(right, sortBy),
    )
    return sortOrder === 'asc' ? comparison : -comparison
  }

  const resetCacheScope = () => {
    cacheGeneration += 1
    requestId += 1
    pendingLoads.clear()
    cache.clear()
  }

  const fetchNodes = (
    path: string,
    cacheKey: string,
    generation: number,
    forceRefresh = false,
  ) => {
    if (!dataSource) {
      throw new Error(t('explorer.dataSourceMissing'))
    }

    if (!forceRefresh) {
      const pendingLoad = pendingLoads.get(cacheKey)
      if (pendingLoad) {
        return pendingLoad
      }
    }

    const source = dataSource
    const load = source
      .loadNodes(path)
      .then((data) => {
        if (generation === cacheGeneration) {
          cache.set(cacheKey, data)
        }
        return data
      })
      .finally(() => {
        if (pendingLoads.get(cacheKey) === load) {
          pendingLoads.delete(cacheKey)
        }
      })

    pendingLoads.set(cacheKey, load)
    return load
  }

  const applyNavigationResult = (
    path: string,
    data: ExplorerNode[],
    updateHistory: boolean,
  ) => {
    nodes = data

    if (currentPath !== path) {
      currentPath = path
      onPathChange?.({ path })
    } else {
      currentPath = path
    }

    selection = new Set()
    anchorPath = null
    onSelect?.({ node: null, nodes: [] })

    onLoad?.({ path, nodes: data })

    if (updateHistory) {
      const base =
        historyIndex >= 0 ? history.slice(0, historyIndex + 1) : history
      if (base[base.length - 1] !== path) {
        history = [...base, path]
        historyIndex = history.length - 1
      }
    }
  }

  const navigateToPath = async (path: string, options?: NavigationOptions) => {
    cancelRename()
    cancelDraft()
    const normalizedPath = normalizeCurrentPath(path)
    const generation = cacheGeneration
    const cacheKey = getScopedCacheKey(normalizedPath, generation)
    const currentRequest = ++requestId
    const forceRefresh = options?.forceRefresh ?? false
    let shouldUpdateHistory = !options?.preserveHistory
    let hasVisibleData = currentPath === normalizedPath && nodes.length > 0
    errorMessage = ''

    if (!forceRefresh) {
      const cached = cache.getEntry(cacheKey, { allowStale: true })
      if (cached) {
        applyNavigationResult(normalizedPath, cached.data, shouldUpdateHistory)
        shouldUpdateHistory = false
        hasVisibleData = true

        if (!cached.isStale) {
          isLoading = false
          isRefreshing = false
          return
        }
      }
    }

    isLoading = !hasVisibleData
    isRefreshing = hasVisibleData

    try {
      const data = await fetchNodes(
        normalizedPath,
        cacheKey,
        generation,
        forceRefresh,
      )
      if (currentRequest !== requestId || generation !== cacheGeneration) {
        return
      }

      applyNavigationResult(normalizedPath, data, shouldUpdateHistory)
    } catch (error) {
      if (currentRequest !== requestId || generation !== cacheGeneration) {
        return
      }

      const resolvedError =
        error instanceof Error ? error : new Error(t('explorer.loadFailed'))
      errorMessage = resolvedError.message
      if (!hasVisibleData) {
        nodes = []
      }
      onError?.({ path: normalizedPath, error: resolvedError })
    } finally {
      if (currentRequest === requestId && generation === cacheGeneration) {
        isLoading = false
        isRefreshing = false
      }
    }
  }

  const bootstrap = async () => {
    if (!dataSource) {
      errorMessage = t('explorer.dataSourceMissing')
      nodes = []
      isLoading = false
      isRefreshing = false
      return
    }

    const source = dataSource
    const currentBootstrap = ++bootstrapId
    resetCacheScope()
    isLoading = true
    isRefreshing = false
    errorMessage = ''

    try {
      const rootPath = initialPath || (await source.getInitialPath())
      const resolvedRoot =
        source.normalizePath?.(rootPath) ?? normalizePath(rootPath)

      if (currentBootstrap !== bootstrapId) {
        return
      }

      homePath = resolvedRoot
      history = [resolvedRoot]
      historyIndex = 0
      await navigateToPath(resolvedRoot, { preserveHistory: true })
    } catch (error) {
      if (currentBootstrap !== bootstrapId) {
        return
      }

      const resolvedError =
        error instanceof Error ? error : new Error(t('explorer.initFailed'))
      errorMessage = resolvedError.message
      nodes = []
      onError?.({ path: initialPath, error: resolvedError })
    } finally {
      if (currentBootstrap === bootstrapId) {
        isLoading = false
        isRefreshing = false
      }
    }
  }

  const handleBack = () => {
    if (!canGoBack) return
    const nextIndex = historyIndex - 1
    historyIndex = nextIndex
    void navigateToPath(history[nextIndex], { preserveHistory: true })
  }

  const handleForward = () => {
    if (!canGoForward) return
    const nextIndex = historyIndex + 1
    historyIndex = nextIndex
    void navigateToPath(history[nextIndex], { preserveHistory: true })
  }

  const handleRefresh = () => {
    if (!currentPath) return
    cache.invalidate(getScopedCacheKey(currentPath))
    void navigateToPath(currentPath, {
      preserveHistory: true,
      forceRefresh: true,
    })
  }

  export const refresh = handleRefresh

  export const invalidatePath = (path: string) => {
    cache.invalidate(getScopedCacheKey(normalizeCurrentPath(path)))
  }

  const isMac =
    typeof navigator !== 'undefined' && /Mac/i.test(navigator.platform)

  const isToggleClick = (event: MouseEvent) => event.metaKey || event.ctrlKey

  const isMacCtrlClick = (event: MouseEvent) =>
    isMac && event.ctrlKey && !event.metaKey

  const selectedNodes = () =>
    filtered.filter((node) => selection.has(node.path))

  const setSelection = (
    paths: Iterable<string>,
    focus: ExplorerNode | null,
  ) => {
    selection = new Set(paths)
    const nodes = selectedNodes()
    const node =
      focus && selection.has(focus.path) ? focus : (nodes.at(-1) ?? null)
    focusPath = node?.path ?? null
    onSelect?.({ node, nodes })
  }

  const clearSelection = () => {
    anchorPath = null
    if (selection.size > 0) setSelection([], null)
  }

  const rangeTo = (node: ExplorerNode): string[] | null => {
    const paths = filtered.map((item) => item.path)
    const from = anchorPath ? paths.indexOf(anchorPath) : -1
    const to = paths.indexOf(node.path)
    if (from < 0 || to < 0) return null
    return paths.slice(Math.min(from, to), Math.max(from, to) + 1)
  }

  const selectNode = (node: ExplorerNode, event?: MouseEvent) => {
    const toggle = event ? isToggleClick(event) : false

    if (event?.shiftKey) {
      const range = rangeTo(node)
      if (range) {
        setSelection(toggle ? [...selection, ...range] : range, node)
        return
      }
    }

    anchorPath = node.path
    if (!toggle) {
      setSelection([node.path], node)
      return
    }

    setSelection(
      selection.has(node.path)
        ? [...selection].filter((path) => path !== node.path)
        : [...selection, node.path],
      node,
    )
  }

  const onNodeClick = (event: MouseEvent, node: ExplorerNode) => {
    if (isMacCtrlClick(event)) return
    selectNode(node, event)
  }

  const hasSelectionModifier = (event: MouseEvent) =>
    event.shiftKey || isToggleClick(event)

  const onViewportClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement | null
    if (target?.closest('[data-explorer-node], thead')) return
    if (hasSelectionModifier(event)) return
    clearSelection()
  }

  const scrollNodeIntoView = async (path: string) => {
    await tick()
    const rows = explorerRoot?.querySelectorAll<HTMLElement>(
      '[data-explorer-node]',
    )
    const row = Array.from(rows ?? []).find(
      (item) =>
        item.dataset.explorerNode === path && item.offsetParent !== null,
    )
    row?.scrollIntoView?.({ block: 'nearest' })
  }

  const stepSelection = (event: KeyboardEvent, step: number) => {
    if (filtered.length === 0) return
    const current = focusPath
      ? filtered.findIndex((node) => node.path === focusPath)
      : -1
    const index =
      current < 0
        ? step > 0
          ? 0
          : filtered.length - 1
        : Math.min(filtered.length - 1, Math.max(0, current + step))
    const node = filtered[index]

    const range = event.shiftKey ? rangeTo(node) : null
    if (range) {
      setSelection(range, node)
    } else {
      anchorPath = node.path
      setSelection([node.path], node)
    }
    void scrollNodeIntoView(node.path)
  }

  const onViewportKeydown = (event: KeyboardEvent) => {
    const target = event.target as HTMLElement | null
    if (event.defaultPrevented || target?.closest('input, textarea, select')) {
      return
    }

    const key = event.key
    if (key === 'ArrowDown' || key === 'ArrowUp') {
      event.preventDefault()
      stepSelection(event, key === 'ArrowDown' ? 1 : -1)
      return
    }
    if (key === 'Escape') {
      if (selection.size === 0) return
      event.preventDefault()
      clearSelection()
      return
    }
    if (key.toLowerCase() === 'a' && (event.metaKey || event.ctrlKey)) {
      event.preventDefault()
      anchorPath = filtered[0]?.path ?? null
      setSelection(
        filtered.map((node) => node.path),
        null,
      )
      return
    }
    if (key === 'Enter') {
      const nodes = selectedNodes()
      if (nodes.length !== 1) return
      event.preventDefault()
      void activateNode(nodes[0])
      return
    }
    if (key === 'Delete' || key === 'Backspace') {
      const nodes = selectedNodes()
      if (nodes.length === 0) return
      event.preventDefault()
      onDeleteNodes?.({ nodes })
    }
  }

  const pruneSelection = (visible: ExplorerNode[]) => {
    if (selection.size === 0) return
    const shown = new Set(visible.map((node) => node.path))
    const kept = [...selection].filter((path) => shown.has(path))
    if (kept.length !== selection.size) setSelection(kept, null)
  }

  const openNodeMenu = (event: MouseEvent, node: ExplorerNode) => {
    if (isMacCtrlClick(event) && event.button === 0) {
      event.preventDefault()
      selectNode(node, event)
      return
    }
    if (!contextMenuNodes) return
    event.preventDefault()
    event.stopPropagation()
    if (!selection.has(node.path)) selectNode(node)
    onNodeContextMenu?.({
      node,
      nodes: selectedNodes(),
      x: event.clientX,
      y: event.clientY,
    })
  }

  let renamingPath = $state<string | null>(null)
  let renameValue = $state('')
  let explorerRoot = $state<HTMLDivElement | null>(null)

  const isRenaming = (node: ExplorerNode) => renamingPath === node.path

  const visibleRenameField = (): HTMLInputElement | null => {
    const fields = Array.from(
      explorerRoot?.querySelectorAll<HTMLInputElement>('[data-rename-field]') ??
        [],
    )
    return (
      fields.find((field) => field.offsetParent !== null) ?? fields[0] ?? null
    )
  }

  export const startRename = async (node: ExplorerNode) => {
    selectNode(node)
    renamingPath = node.path
    renameValue = node.name
    await tick()

    const field = visibleRenameField()
    field?.focus()
    field?.setSelectionRange(0, fileStemLength(renameValue))
  }

  const cancelRename = () => {
    renamingPath = null
    renameValue = ''
  }

  const commitRename = (node: ExplorerNode) => {
    const next = renameValue.trim()
    cancelRename()
    if (!next || next === node.name) return
    onRenameNode?.({ node, name: next })
  }

  const handleRenameKeydown = (event: KeyboardEvent, node: ExplorerNode) => {
    if (event.key === 'Enter') {
      event.preventDefault()
      commitRename(node)
      return
    }
    if (event.key === 'Escape') {
      event.preventDefault()
      cancelRename()
    }
  }

  const handleRenameBlur = (node: ExplorerNode) => {
    if (isRenaming(node)) commitRename(node)
  }

  type DraftKind = 'folder' | 'file'
  type Draft = { kind: DraftKind; directory: string; name: string }

  let draft = $state<Draft | null>(null)
  let draftField = $state<HTMLInputElement | null>(null)

  const nameTaken = (name: string) =>
    nodes.some((node) => node.name.toLowerCase() === name.toLowerCase())

  let draftInvalid = $derived(draft ? nameTaken(draft.name.trim()) : false)

  const uniqueName = (base: string) => {
    const stem = fileStemLength(base)
    const head = base.slice(0, stem)
    const extension = base.slice(stem)
    let candidate = base
    for (let counter = 2; nameTaken(candidate); counter += 1) {
      candidate = `${head} ${counter}${extension}`
    }
    return candidate
  }

  const startDraft = async (kind: DraftKind) => {
    if (!currentPath || isLoading) return
    cancelRename()
    clearSelection()
    searchQuery = ''
    const name = uniqueName(
      kind === 'folder' ? mergedLabels.newFolder : mergedLabels.newFile,
    )
    draft = { kind, directory: currentPath, name }
    await tick()

    draftField?.scrollIntoView({ block: 'nearest' })
    draftField?.focus()
    draftField?.setSelectionRange(
      0,
      kind === 'file' ? fileStemLength(name) : name.length,
    )
  }

  export const startCreateFolder = () => startDraft('folder')

  export const startCreateFile = () => startDraft('file')

  const cancelDraft = () => {
    draft = null
  }

  const commitDraft = () => {
    const current = draft
    if (!current) return
    const name = current.name.trim()
    if (name && nameTaken(name)) return
    cancelDraft()
    if (!name || current.directory !== currentPath) return
    const detail = { directory: current.directory, name }
    if (current.kind === 'folder') {
      onCreateFolder?.(detail)
    } else {
      onCreateFile?.(detail)
    }
  }

  const handleDraftKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Enter') {
      event.preventDefault()
      commitDraft()
      return
    }
    if (event.key === 'Escape') {
      event.preventDefault()
      cancelDraft()
    }
  }

  const handleDraftBlur = () => {
    if (draftInvalid) {
      cancelDraft()
      return
    }
    commitDraft()
  }

  let dragChip: HTMLElement | null = null
  let draggedNodes: ExplorerNode[] | null = null

  const clearDragChip = () => {
    dragChip?.remove()
    dragChip = null
  }

  const setDragChip = (event: DragEvent, node: ExplorerNode, count: number) => {
    const row = event.currentTarget as HTMLElement | null
    if (!event.dataTransfer?.setDragImage || !row) return

    const chip = document.createElement('div')
    chip.style.cssText = [
      'position:absolute',
      'top:-1000px',
      'left:0',
      'display:inline-flex',
      'align-items:center',
      'gap:8px',
      'max-width:18rem',
      'padding:6px 10px',
      'border:1px solid var(--color-line)',
      'border-radius:10px',
      'background:var(--color-surface-raised)',
      'color:#fff',
      'font-size:12px',
      'line-height:1.2',
      'white-space:nowrap',
      'box-shadow:0 12px 32px rgb(0 0 0 / 0.5)',
      'pointer-events:none',
    ].join(';')

    const icon = row.querySelector('[data-file-icon]')
    if (icon) chip.appendChild(icon.cloneNode(true))

    const label = document.createElement('span')
    label.textContent =
      count > 1
        ? mergedLabels.selectedCount.replace('{count}', String(count))
        : node.name
    label.style.cssText = 'overflow:hidden;text-overflow:ellipsis'
    chip.appendChild(label)

    document.body.appendChild(chip)
    dragChip = chip
    event.dataTransfer.setDragImage(chip, 14, 14)
    setTimeout(clearDragChip, 0)
  }

  const startNodeDrag = (event: DragEvent, node: ExplorerNode) => {
    if (!draggableNodes) return
    if (!selection.has(node.path)) selectNode(node)
    const nodes = selectedNodes()
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'copy'
      event.dataTransfer.setData(
        'text/plain',
        nodes.map((item) => item.path).join('\n'),
      )
    }
    setDragChip(event, node, nodes.length)
    draggedNodes = nodes
    onNodeDragStart?.({ node, nodes })
  }

  const endNodeDrag = () => {
    clearDragChip()
    draggedNodes = null
    onNodeDragEnd?.()
  }

  let dropTargetPath = $state<string | null>(null)

  const isWithin = (folder: string, target: string): boolean => {
    if (target === folder) return true
    const separator = folder.includes('\\') ? '\\' : '/'
    const base = folder.endsWith(separator) ? folder : folder + separator
    return target.startsWith(base)
  }

  const acceptsDropOn = (node: ExplorerNode) => {
    if (!containerDropTargets || !dropActive || !isContainerNode(node)) {
      return false
    }
    if (!draggedNodes) return true

    if (draggedNodes.some((item) => isWithin(item.path, node.path))) {
      return false
    }
    return draggedNodes.some(
      (item) => resolveParentPath(item.path) !== node.path,
    )
  }

  const crumbAsNode = (crumb: ExplorerBreadcrumb): ExplorerNode => ({
    id: crumb.path,
    name: crumb.name,
    path: crumb.path,
    type: 'directory',
    isContainer: true,
  })

  const setDropTarget = (node: ExplorerNode | null) => {
    const next = node?.path ?? null
    if (next === dropTargetPath) return
    dropTargetPath = next
    onNodeDropTarget?.({ node })
  }

  const onNodeDragOver = (event: DragEvent, node: ExplorerNode) => {
    if (!acceptsDropOn(node)) return
    event.preventDefault()
    event.stopPropagation()
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy'
    setDropTarget(node)
  }

  const onNodeDragLeave = (event: DragEvent, node: ExplorerNode) => {
    const next = event.relatedTarget as Node | null
    if (next && (event.currentTarget as HTMLElement).contains(next)) return
    if (dropTargetPath === node.path) setDropTarget(null)
  }

  const handleNodeDrop = (event: DragEvent, node: ExplorerNode) => {
    if (!acceptsDropOn(node)) return
    event.preventDefault()
    event.stopPropagation()
    setDropTarget(null)
    onNodeDrop?.({ node })
    draggedNodes = null
  }

  $effect.pre(() => {
    if (!dropActive) setDropTarget(null)
  })

  const goHome = () => {
    if (!homePath) return
    void navigateToPath(homePath)
  }

  const goToParent = () => {
    if (!parentPath) return
    void navigateToPath(parentPath)
  }

  const setSort = (nextSortBy: string) => {
    if (sortBy === nextSortBy) {
      sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'
      return
    }

    sortBy = nextSortBy
    sortOrder = 'asc'
  }

  const isSortColumn = (column: string) => {
    return sortBy === column
  }

  const toggleHiddenItems = () => {
    showHiddenItems = !showHiddenItems
  }

  const activateNode = async (node: ExplorerNode) => {
    const navigated = isContainerNode(node)
    if (navigated) {
      await navigateToPath(node.path)
    }

    onActivate?.({ node, navigated })
  }

  const copyCurrentPath = async () => {
    const pathToCopy = getCopyValue(currentPath || '/')

    try {
      if (!(await copyText(pathToCopy))) {
        copiedPath = false
        return
      }

      copiedPath = true
      if (copiedResetTimeout) {
        clearTimeout(copiedResetTimeout)
      }
      copiedResetTimeout = setTimeout(() => {
        copiedPath = false
      }, 1400)
    } catch {
      copiedPath = false
    }
  }

  let defaultExplorerLabels = $derived({
    nameColumn: t('explorer.columns.name'),
    searchPlaceholder: t('explorer.searchPlaceholder'),
    back: t('explorer.back'),
    forward: t('explorer.forward'),
    home: t('explorer.home'),
    parent: t('explorer.parent'),
    refresh: t('explorer.refresh'),
    showHidden: t('explorer.showHidden'),
    hideHidden: t('explorer.hideHidden'),
    copy: t('explorer.copy'),
    copied: t('explorer.copied'),
    copyCurrentPath: t('explorer.copyCurrentPath'),
    loading: t('explorer.loading'),
    empty: t('explorer.empty'),
    openContainer: t('explorer.openContainer'),
    rename: t('explorer.rename'),
    newFolder: t('explorer.newFolder'),
    newFolderName: t('explorer.newFolderName'),
    newFile: t('explorer.newFile'),
    newFileName: t('explorer.newFileName'),
    selectedCount: t('explorer.selectedCount'),
  })
  let mergedLabels = $derived({ ...defaultExplorerLabels, ...labels })
  let canGoBack = $derived(historyIndex > 0)
  let canGoForward = $derived(
    historyIndex >= 0 && historyIndex < history.length - 1,
  )
  $effect.pre(() => {
    if (sortBy !== 'name' && !getColumn(sortBy)) sortBy = 'name'
  })
  let breadcrumbs = $derived(resolveBreadcrumbs(currentPath || '/'))

  let parentPath = $derived(currentPath ? resolveParentPath(currentPath) : '')

  let normalizedSearchQuery = $derived(searchQuery.trim().toLowerCase())
  let viewportStyles = $derived(
    [
      viewportHeight ? `height: ${viewportHeight}` : '',
      !fillHeight && viewportMaxHeight
        ? `max-height: ${viewportMaxHeight}`
        : '',
    ]
      .filter(Boolean)
      .join('; '),
  )

  let filtered = $derived(
    nodes
      .filter((item) =>
        showHiddenToggle && !showHiddenItems ? !isHiddenNode(item) : true,
      )
      .filter((item) =>
        normalizedSearchQuery
          ? item.name.toLowerCase().includes(normalizedSearchQuery)
          : true,
      )
      .sort((left, right) => compareNodes(left, right)),
  )

  $effect.pre(() => {
    pruneSelection(filtered)
  })

  $effect.pre(() => {
    if (
      isMounted &&
      (dataSource !== previousDataSource || initialPath !== previousInitialPath)
    ) {
      previousDataSource = dataSource
      previousInitialPath = initialPath
      void bootstrap()
    }
  })

  onMount(() => {
    isMounted = true
    previousDataSource = dataSource
    previousInitialPath = initialPath
    void bootstrap()
  })

  onDestroy(() => {
    if (copiedResetTimeout) {
      clearTimeout(copiedResetTimeout)
    }
    clearDragChip()
  })
</script>

{#snippet toolButton(
  icon: IconComponent,
  label: string,
  onclick: () => void,
  disabled = false,
  extraClass = '',
)}
  <IconButton
    {icon}
    {label}
    size="sm"
    tooltipPlacement="bottom"
    class={extraClass}
    {disabled}
    {onclick}
  />
{/snippet}

{#snippet renameField(node: ExplorerNode)}
  <input
    bind:value={renameValue}
    data-rename-field
    type="text"
    autocomplete="off"
    autocapitalize="off"
    autocorrect="off"
    spellcheck="false"
    aria-label={mergedLabels.rename}
    class="h-6 min-w-0 flex-1 rounded border border-primary/50 bg-fill-hover px-1.5 text-sm leading-none text-white outline-none"
    onclick={(event) => event.stopPropagation()}
    ondblclick={(event) => event.stopPropagation()}
    onkeydown={(event) => handleRenameKeydown(event, node)}
    onblur={() => handleRenameBlur(node)}
  />
{/snippet}

<div
  bind:this={explorerRoot}
  class={`@container ${fillHeight ? 'h-full' : ''} flex min-h-0 w-full min-w-0 flex-col overflow-hidden ${className}`.trim()}
  aria-busy={isLoading || isRefreshing}
>
  <div
    class="flex shrink-0 items-center gap-0.5 border-b border-border px-2 py-1.5"
  >
    {@render toolButton(ChevronLeft, mergedLabels.back, handleBack, !canGoBack)}
    {@render toolButton(
      ChevronRight,
      mergedLabels.forward,
      handleForward,
      !canGoForward,
      '@max-sm:hidden',
    )}
    {@render toolButton(FolderUp, mergedLabels.parent, goToParent, !parentPath)}
    {@render toolButton(
      Home,
      mergedLabels.home,
      goHome,
      !homePath,
      '@max-xs:hidden',
    )}
    <IconButton
      icon={RefreshCw}
      size="sm"
      tooltipPlacement="bottom"
      label={mergedLabels.refresh}
      iconClass={isRefreshing ? 'animate-spin' : ''}
      onclick={handleRefresh}
    />

    <Input
      size="sm"
      class="ml-1 flex-1"
      bind:value={searchQuery}
      aria-label={mergedLabels.searchPlaceholder}
      placeholder={mergedLabels.searchPlaceholder}
      spellcheck="false"
      autocomplete="off"
    >
      {#snippet leading()}
        <Search class="h-3.5 w-3.5" />
      {/snippet}
    </Input>

    {#if showHiddenToggle}
      <IconButton
        icon={showHiddenItems ? Eye : EyeOff}
        size="sm"
        tooltipPlacement="bottom-end"
        class="ml-0.5"
        pressed={showHiddenItems}
        label={showHiddenItems
          ? mergedLabels.hideHidden
          : mergedLabels.showHidden}
        onclick={toggleHiddenItems}
      />
    {/if}
  </div>

  <div
    class="flex shrink-0 items-center gap-1 border-b border-border px-2 py-1 text-xs text-muted-foreground"
  >
    <div class="min-w-0 flex-1 overflow-x-auto [scrollbar-width:none]">
      <div class="flex items-center gap-0.5 whitespace-nowrap">
        {#each breadcrumbs as crumb, index (crumb.path)}
          {#if index > 0}
            <ChevronRight class="h-3 w-3 shrink-0 text-muted-foreground/60" />
          {/if}
          <button
            type="button"
            class={[
              'inline-flex h-6 shrink-0 items-center rounded px-1.5 transition-colors',
              dropTargetPath === crumb.path
                ? 'bg-primary/15 text-white outline-2 -outline-offset-2 outline-primary/55'
                : 'hover:bg-fill-hover hover:text-white',
              crumb.path === currentPath
                ? 'text-white'
                : 'text-muted-foreground',
            ]}
            onclick={() => navigateToPath(crumb.path)}
            ondragover={(event) => onNodeDragOver(event, crumbAsNode(crumb))}
            ondragleave={(event) => onNodeDragLeave(event, crumbAsNode(crumb))}
            ondrop={(event) => handleNodeDrop(event, crumbAsNode(crumb))}
          >
            {crumb.name}
          </button>
        {/each}
      </div>
    </div>

    {#if showCopyAction}
      <IconButton
        icon={copiedPath ? Check : Copy}
        size="xs"
        tooltipPlacement="bottom-end"
        label={copiedPath ? mergedLabels.copied : mergedLabels.copyCurrentPath}
        iconClass={copiedPath ? 'text-primary' : ''}
        onclick={copyCurrentPath}
      />
    {/if}
  </div>

  {#if errorMessage}
    <div
      class="m-3 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive"
    >
      {errorMessage}
    </div>
  {/if}

  <div
    class={`flex-1 min-h-0 min-w-0 overflow-auto outline-none ${viewportClassName}`.trim()}
    style={viewportStyles}
    role="grid"
    aria-multiselectable="true"
    tabindex="0"
    onclick={onViewportClick}
    onkeydown={onViewportKeydown}
  >
    {#if isLoading}
      <div
        class="h-full flex items-center justify-center text-sm text-muted-foreground"
      >
        {mergedLabels.loading}
      </div>
    {:else if filtered.length === 0 && !draft}
      <div
        class="h-full flex items-center justify-center text-sm text-muted-foreground"
      >
        {mergedLabels.empty}
      </div>
    {:else}
      <div class={`min-w-full ${contentClassName}`.trim()}>
        <table class="w-full table-fixed text-sm">
          <thead
            class="sticky top-0 bg-surface-table-header border-b border-border"
          >
            <tr class="text-left text-muted-foreground">
              <th class="px-3 py-0 font-medium">
                <button
                  type="button"
                  class="w-full h-7 px-1 inline-flex items-center gap-1.5 rounded text-left text-xs transition-colors {isSortColumn(
                    'name',
                  )
                    ? 'text-white'
                    : 'text-muted-foreground hover:text-white'}"
                  onclick={() => setSort('name')}
                >
                  <span>{mergedLabels.nameColumn}</span>
                  {#if isSortColumn('name')}
                    {#if sortOrder === 'asc'}
                      <ArrowUp class="h-3 w-3 text-primary" />
                    {:else}
                      <ArrowDown class="h-3 w-3 text-primary" />
                    {/if}
                  {/if}
                </button>
              </th>
              {#each columns as column (column.id)}
                <th
                  class="px-3 py-0 font-medium {column.widthClass ??
                    ''} {getResponsiveHideClass(column)}"
                >
                  {#if column.sortable === false}
                    <span
                      class="w-full h-7 px-1 inline-flex items-center rounded text-left text-xs"
                    >
                      {column.label}
                    </span>
                  {:else}
                    <button
                      type="button"
                      class="w-full h-7 px-1 inline-flex items-center gap-1.5 rounded text-left text-xs transition-colors {isSortColumn(
                        column.id,
                      )
                        ? 'text-white'
                        : 'text-muted-foreground hover:text-white'}"
                      onclick={() => setSort(column.id)}
                    >
                      <span class="truncate">{column.label}</span>
                      {#if isSortColumn(column.id)}
                        {#if sortOrder === 'asc'}
                          <ArrowUp class="h-3 w-3 text-primary shrink-0" />
                        {:else}
                          <ArrowDown class="h-3 w-3 text-primary shrink-0" />
                        {/if}
                      {/if}
                    </button>
                  {/if}
                </th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#if draft}
              <tr class="border-b border-border/60 bg-primary/10 select-none">
                <td class="px-3 py-1.5 min-w-0 max-w-0">
                  <div class="flex w-full min-w-0 items-center gap-2">
                    <FileIcon
                      name={draft.name}
                      isContainer={draft.kind === 'folder'}
                    />
                    <input
                      bind:this={draftField}
                      bind:value={draft.name}
                      type="text"
                      autocomplete="off"
                      autocapitalize="off"
                      autocorrect="off"
                      spellcheck="false"
                      aria-label={draft.kind === 'folder'
                        ? mergedLabels.newFolderName
                        : mergedLabels.newFileName}
                      aria-invalid={draftInvalid}
                      class={[
                        'h-6 min-w-0 flex-1 rounded border bg-fill-hover px-1.5 text-sm leading-none text-white outline-none',
                        draftInvalid
                          ? 'border-destructive/70'
                          : 'border-primary/50',
                      ]}
                      onclick={(event) => event.stopPropagation()}
                      ondblclick={(event) => event.stopPropagation()}
                      onkeydown={handleDraftKeydown}
                      onblur={handleDraftBlur}
                    />
                  </div>
                </td>
                {#each columns as column (column.id)}
                  <td
                    class="px-3 py-1.5 {column.cellClass ??
                      ''} {getResponsiveHideClass(column)}"
                  ></td>
                {/each}
              </tr>
            {/if}
            {#each filtered as node (node.id)}
              <tr
                class="border-b border-border/60 {draggableNodes
                  ? 'cursor-grab active:cursor-grabbing'
                  : 'cursor-default'} {dropTargetPath === node.path
                  ? 'bg-primary/15 outline outline-2 -outline-offset-2 outline-primary/55'
                  : selection.has(node.path)
                    ? 'bg-primary/10'
                    : 'hover:bg-fill-subtle'} select-none"
                aria-selected={selection.has(node.path)}
                data-explorer-node={node.path}
                draggable={draggableNodes && renamingPath !== node.path}
                onclick={(event) => onNodeClick(event, node)}
                ondblclick={() => void activateNode(node)}
                ondragstart={(event) => startNodeDrag(event, node)}
                ondragend={endNodeDrag}
                ondragover={(event) => onNodeDragOver(event, node)}
                ondragleave={(event) => onNodeDragLeave(event, node)}
                ondrop={(event) => handleNodeDrop(event, node)}
                oncontextmenu={(event) => openNodeMenu(event, node)}
              >
                <td class="px-3 py-1.5 min-w-0 max-w-0">
                  {#if renamingPath === node.path}
                    <div class="flex w-full min-w-0 items-center gap-2">
                      <FileIcon
                        name={node.name}
                        isContainer={isContainerNode(node)}
                      />
                      {@render renameField(node)}
                    </div>
                  {:else if isContainerNode(node)}
                    <button
                      type="button"
                      class="flex w-full items-center gap-2 text-left text-white hover:text-primary hover:underline cursor-pointer min-w-0"
                      onclick={(event) => {
                        if (!hasSelectionModifier(event))
                          void activateNode(node)
                      }}
                    >
                      <FileIcon name={node.name} isContainer />
                      <span
                        class="truncate min-w-0"
                        {@attach nameTooltip(node.name)}
                      >
                        {node.name}
                      </span>
                    </button>
                  {:else}
                    <div
                      class="flex w-full items-center gap-2 text-left text-white cursor-default min-w-0"
                    >
                      <FileIcon name={node.name} />
                      <span
                        class="truncate min-w-0"
                        {@attach nameTooltip(node.name)}
                      >
                        {node.name}
                      </span>
                    </div>
                  {/if}
                </td>
                {#each columns as column (column.id)}
                  <td
                    class="px-3 py-1.5 text-xs text-muted-foreground truncate {column.cellClass ??
                      ''} {getResponsiveHideClass(column)}"
                    {@attach nameTooltip(column.render(node))}
                  >
                    {column.render(node)}
                  </td>
                {/each}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>
