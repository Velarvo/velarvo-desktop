<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte'
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
    ArrowUpDown,
    ArrowUp,
    ArrowDown,
  } from 'lucide-svelte'
  import {
    DEFAULT_EXPLORER_LABELS,
    type ExplorerColumn,
    type ExplorerColumnBreakpoint,
    type ExplorerDataSource,
    type ExplorerLabels,
    type ExplorerNode,
    type SortOrder,
  } from './types'
  import FileIcon from './FileIcon.svelte'
  import { getBreadcrumbs, getParentPath, normalizePath } from './utils'
  import { createSmartCache } from './hooks/useSmartCache'

  type NavigationOptions = {
    preserveHistory?: boolean
    forceRefresh?: boolean
  }

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
  const pendingLoads = new Map<string, Promise<ExplorerNode[]>>()
  const dispatch = createEventDispatcher<{
    activate: { node: ExplorerNode; navigated: boolean }
    error: { path: string; error: Error }
    load: { path: string; nodes: ExplorerNode[] }
    pathChange: { path: string }
  }>()

  export let dataSource: ExplorerDataSource
  export let columns: ExplorerColumn[] = []
  export let labels: Partial<ExplorerLabels> = {}
  export let initialPath = ''
  export let defaultSortBy = 'name'
  export let defaultSortOrder: SortOrder = 'asc'
  export let showCopyAction = true
  export let showHiddenToggle = true
  export let groupContainersFirst = true
  export let className = ''
  export let viewportClassName = ''
  export let contentClassName = ''
  export let fillHeight = false
  export let viewportHeight = ''
  export let viewportMaxHeight = '32rem'

  let currentPath = ''
  let homePath = ''
  let isLoading = true
  let isRefreshing = false
  let errorMessage = ''
  let showHiddenItems = true
  let searchQuery = ''
  let sortBy = defaultSortBy
  let sortOrder: SortOrder = defaultSortOrder
  let nodes: ExplorerNode[] = []
  let history: string[] = []
  let historyIndex = -1
  let copiedPath = false
  let copiedResetTimeout: ReturnType<typeof setTimeout> | null = null
  let requestId = 0
  let cacheGeneration = 0
  let bootstrapId = 0
  let isMounted = false
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
      throw new Error('Explorer data source is not configured')
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
      dispatch('pathChange', { path })
    } else {
      currentPath = path
    }

    dispatch('load', { path, nodes: data })

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
        error instanceof Error ? error : new Error('Failed to load items')
      errorMessage = resolvedError.message
      if (!hasVisibleData) {
        nodes = []
      }
      dispatch('error', { path: normalizedPath, error: resolvedError })
    } finally {
      if (currentRequest === requestId && generation === cacheGeneration) {
        isLoading = false
        isRefreshing = false
      }
    }
  }

  const bootstrap = async () => {
    if (!dataSource) {
      errorMessage = 'Explorer data source is not configured'
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
        error instanceof Error
          ? error
          : new Error('Failed to initialize explorer')
      errorMessage = resolvedError.message
      nodes = []
      dispatch('error', { path: initialPath, error: resolvedError })
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

  const setSortBy = (nextSortBy: string) => {
    sortBy = nextSortBy
  }

  const toggleSortOrder = () => {
    sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'
  }

  const isSortColumn = (column: string) => {
    return sortBy === column
  }

  const handleSortByChange = (event: Event) => {
    const target = event.currentTarget as HTMLSelectElement
    setSortBy(target.value)
  }

  const toggleHiddenItems = () => {
    showHiddenItems = !showHiddenItems
  }

  const activateNode = async (node: ExplorerNode) => {
    const navigated = isContainerNode(node)
    if (navigated) {
      await navigateToPath(node.path)
    }

    dispatch('activate', { node, navigated })
  }

  const copyCurrentPath = async () => {
    const pathToCopy = getCopyValue(currentPath || '/')

    try {
      if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(pathToCopy)
      } else {
        const textarea = document.createElement('textarea')
        textarea.value = pathToCopy
        textarea.setAttribute('readonly', '')
        textarea.style.position = 'absolute'
        textarea.style.left = '-9999px'
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
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

  $: mergedLabels = { ...DEFAULT_EXPLORER_LABELS, ...labels }
  $: canGoBack = historyIndex > 0
  $: canGoForward = historyIndex >= 0 && historyIndex < history.length - 1
  $: sortOptions = [
    { id: 'name', label: mergedLabels.nameColumn },
    ...columns
      .filter((column) => column.sortable !== false)
      .map((column) => ({ id: column.id, label: column.label })),
  ]
  $: if (!sortOptions.some((option) => option.id === sortBy)) {
    sortBy = 'name'
  }
  $: breadcrumbs = resolveBreadcrumbs(currentPath || '/')

  $: parentPath = currentPath ? resolveParentPath(currentPath) : ''

  $: normalizedSearchQuery = searchQuery.trim().toLowerCase()
  $: viewportStyles = [
    viewportHeight ? `height: ${viewportHeight}` : '',
    !fillHeight && viewportMaxHeight ? `max-height: ${viewportMaxHeight}` : '',
  ]
    .filter(Boolean)
    .join('; ')

  $: filtered = nodes
    .filter((item) =>
      showHiddenToggle && !showHiddenItems ? !isHiddenNode(item) : true,
    )
    .filter((item) =>
      normalizedSearchQuery
        ? item.name.toLowerCase().includes(normalizedSearchQuery)
        : true,
    )
    .sort((left, right) => compareNodes(left, right))

  $: if (
    isMounted &&
    (dataSource !== previousDataSource || initialPath !== previousInitialPath)
  ) {
    previousDataSource = dataSource
    previousInitialPath = initialPath
    void bootstrap()
  }

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
  })
</script>

<div
  class={`@container ${fillHeight ? 'h-full' : ''} min-h-0 min-w-0 w-full max-w-full flex flex-col rounded-xl border border-border overflow-hidden ${className}`.trim()}
  aria-busy={isLoading || isRefreshing}
>
  <div class="border-b border-border p-3 flex flex-wrap items-center gap-2">
    <button
      type="button"
      class="h-8 w-8 rounded border border-border text-muted-foreground hover:text-white disabled:opacity-40 inline-flex items-center justify-center"
      on:click={handleBack}
      disabled={!canGoBack}><ChevronLeft class="h-4 w-4" /></button
    >
    <button
      type="button"
      class="h-8 w-8 rounded border border-border text-muted-foreground hover:text-white disabled:opacity-40 inline-flex items-center justify-center"
      on:click={handleForward}
      disabled={!canGoForward}><ChevronRight class="h-4 w-4" /></button
    >
    <button
      type="button"
      class="h-8 w-8 rounded border border-border text-muted-foreground hover:text-white inline-flex items-center justify-center"
      on:click={goHome}
      disabled={!homePath}><Home class="h-4 w-4" /></button
    >
    <button
      type="button"
      class="h-8 w-8 rounded border border-border text-muted-foreground hover:text-white inline-flex items-center justify-center"
      on:click={goToParent}
      disabled={!parentPath}><FolderUp class="h-4 w-4" /></button
    >
    <button
      type="button"
      class="h-8 w-8 rounded border border-border text-muted-foreground hover:text-white inline-flex items-center justify-center"
      on:click={handleRefresh}
      ><RefreshCw
        class="h-4 w-4 {isRefreshing ? 'animate-spin' : ''}"
      /></button
    >

    <div class="relative flex-1 min-w-0 basis-40 @md:basis-55">
      <Search
        class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"
      />
      <input
        class="h-9 w-full rounded-md border border-border bg-background px-8 text-sm text-white"
        bind:value={searchQuery}
        placeholder={mergedLabels.searchPlaceholder}
      />
    </div>

    <div
      class="h-8 rounded-md border border-border bg-background px-2 shrink-0 inline-flex items-center gap-1.5"
    >
      <ArrowUpDown class="h-3 w-3 text-muted-foreground" />
      <select
        class="h-7 bg-transparent text-xs text-white pr-4 outline-none"
        bind:value={sortBy}
        on:change={handleSortByChange}
        title={mergedLabels.sortBy}
      >
        {#each sortOptions as option}
          <option value={option.id}>{option.label}</option>
        {/each}
      </select>
    </div>

    <button
      type="button"
      class="h-8 w-8 rounded-md border bg-background shrink-0 inline-flex items-center justify-center transition-colors hover:text-white {sortOrder ===
      'asc'
        ? 'text-primary border-primary/40'
        : 'text-orange-300 border-orange-300/40'}"
      on:click={toggleSortOrder}
      title={sortOrder === 'asc'
        ? mergedLabels.ascending
        : mergedLabels.descending}
    >
      {#if sortOrder === 'asc'}
        <ArrowUp class="h-3.5 w-3.5" />
      {:else}
        <ArrowDown class="h-3.5 w-3.5" />
      {/if}
    </button>

    {#if showHiddenToggle}
      <button
        type="button"
        class="h-8 rounded-md border px-2 text-xs shrink-0 inline-flex items-center gap-1.5 transition-colors {showHiddenItems
          ? 'border-primary/50 bg-primary/10 text-primary'
          : 'border-border bg-background text-muted-foreground hover:text-white'}"
        on:click={toggleHiddenItems}
        aria-pressed={showHiddenItems}
        title={showHiddenItems
          ? mergedLabels.hideHidden
          : mergedLabels.showHidden}
      >
        {#if showHiddenItems}
          <Eye class="h-3.5 w-3.5" />
        {:else}
          <EyeOff class="h-3.5 w-3.5" />
        {/if}
        <span class="@max-md:hidden">{mergedLabels.hidden}</span>
      </button>
    {/if}
  </div>

  <div
    class="px-3 py-2 text-xs text-muted-foreground border-b border-border flex items-center gap-2"
  >
    <div class="min-w-0 flex-1 overflow-x-auto">
      <div class="flex items-center gap-1 whitespace-nowrap">
        {#each breadcrumbs as crumb, index}
          {#if index > 0}
            <ChevronRight
              class="h-3.5 w-3.5 text-muted-foreground/70 shrink-0"
            />
          {/if}
          <button
            type="button"
            class="shrink-0 rounded px-1.5 py-0.5 transition-colors hover:bg-white/5 hover:text-white"
            class:text-white={crumb.path === currentPath}
            class:text-muted-foreground={crumb.path !== currentPath}
            on:click={() => navigateToPath(crumb.path)}
          >
            {crumb.name}
          </button>
        {/each}
      </div>
    </div>

    {#if showCopyAction}
      <button
        type="button"
        class="h-7 px-2 rounded border border-border bg-background text-muted-foreground hover:text-white inline-flex items-center gap-1.5 shrink-0"
        on:click={copyCurrentPath}
        title={mergedLabels.copyCurrentPath}
        aria-label={mergedLabels.copyCurrentPath}
      >
        {#if copiedPath}
          <Check class="h-3.5 w-3.5" />
          <span class="@max-sm:hidden">{mergedLabels.copied}</span>
        {:else}
          <Copy class="h-3.5 w-3.5" />
          <span class="@max-sm:hidden">{mergedLabels.copy}</span>
        {/if}
      </button>
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
    class={`flex-1 min-h-0 min-w-0 overflow-auto ${viewportClassName}`.trim()}
    style={viewportStyles}
  >
    {#if isLoading}
      <div
        class="h-full flex items-center justify-center text-sm text-muted-foreground"
      >
        {mergedLabels.loading}
      </div>
    {:else if filtered.length === 0}
      <div
        class="h-full flex items-center justify-center text-sm text-muted-foreground"
      >
        {mergedLabels.empty}
      </div>
    {:else}
      <div class={`min-w-full ${contentClassName}`.trim()}>
        <table class="w-full table-fixed text-sm">
          <thead class="sticky top-0 bg-[#0c0c12] border-b border-border">
            <tr class="text-left text-muted-foreground">
              <th class="px-3 py-0 font-medium">
                <button
                  type="button"
                  class="w-full h-8 px-1 inline-flex items-center gap-1.5 rounded text-left text-xs transition-colors {isSortColumn(
                    'name',
                  )
                    ? 'text-white'
                    : 'text-muted-foreground hover:text-white'}"
                  on:click={() => setSort('name')}
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
              {#each columns as column}
                <th
                  class="px-3 py-0 font-medium {column.widthClass ??
                    ''} {getResponsiveHideClass(column)}"
                >
                  {#if column.sortable === false}
                    <span
                      class="w-full h-8 px-1 inline-flex items-center rounded text-left text-xs"
                      >{column.label}</span
                    >
                  {:else}
                    <button
                      type="button"
                      class="w-full h-8 px-1 inline-flex items-center gap-1.5 rounded text-left text-xs transition-colors {isSortColumn(
                        column.id,
                      )
                        ? 'text-white'
                        : 'text-muted-foreground hover:text-white'}"
                      on:click={() => setSort(column.id)}
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
            {#each filtered as node}
              <tr
                class="border-b border-border/60 hover:bg-white/3 cursor-default"
                on:dblclick={() => void activateNode(node)}
              >
                <td class="px-3 py-2 min-w-0 max-w-0">
                  {#if isContainerNode(node)}
                    <button
                      type="button"
                      class="flex w-full items-center gap-2 text-left text-white hover:text-primary hover:underline cursor-pointer min-w-0"
                      on:click={() => void activateNode(node)}
                      title={node.name}
                    >
                      <FileIcon name={node.name} isContainer />
                      <span class="truncate min-w-0">{node.name}</span>
                    </button>
                  {:else}
                    <div
                      class="flex w-full items-center gap-2 text-left text-white cursor-default min-w-0"
                      title={node.name}
                    >
                      <FileIcon name={node.name} />
                      <span class="truncate min-w-0">{node.name}</span>
                    </div>
                  {/if}
                </td>
                {#each columns as column}
                  <td
                    class="px-3 py-2 text-muted-foreground truncate {column.cellClass ??
                      ''} {getResponsiveHideClass(column)}"
                    title={column.render(node)}
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
