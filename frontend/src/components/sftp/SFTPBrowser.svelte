<script lang="ts">
  import { tooltip } from '@/components/ui/tooltip'
  import { onMount } from 'svelte'
  import {
    ArrowDownToLine,
    ArrowLeftRight,
    ArrowUpFromLine,
    ClipboardCopy,
    Copy,
    FolderPlus,
    HardDrive,
    ListEnd,
    PencilLine,
    Redo2,
    Server,
    Trash2,
    Undo2,
  } from '@lucide/svelte'
  import FileExplorer from '@/components/file-explorer/FileExplorer.svelte'
  import {
    fileSystemExplorerDataSource,
    getFileSystemExplorerColumns,
  } from '@/components/file-explorer/filesystem'
  import type { ExplorerNode } from '@/components/file-explorer/types'
  import type { IconComponent } from '@/types/icon'
  import { fileStemLength } from '@/components/file-explorer/utils'
  import FilePreviewModal from '@/components/sftp/FilePreviewModal.svelte'
  import TransferQueue from '@/components/sftp/TransferQueue.svelte'
  import ConnectionState from '@/components/ssh/ConnectionState.svelte'
  import ConfirmDialog from '@/components/ui/ConfirmDialog.svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import MenuItem, {
    MENU_PANEL_CLASS,
    MENU_SEPARATOR_CLASS,
  } from '@/components/ui/MenuItem.svelte'
  import PromptDialog from '@/components/ui/PromptDialog.svelte'
  import { copyText } from '@/lib/clipboard'
  import {
    localBaseName,
    localJoin,
    localParent,
    removeLocalPath,
    renameLocalPath,
  } from '@/lib/localFiles'
  import {
    previewLocalFile,
    previewRemoteFile,
    type FilePreview,
  } from '@/lib/filePreview'
  import { describeError } from '@/lib/errors'
  import { t } from '@/lib/i18n.svelte'
  import { logger } from '@/lib/logger'
  import { Transient } from '@/lib/transient.svelte'
  import {
    copyRemotePath,
    createRemoteDirectory,
    createSFTPDataSource,
    downloadFile,
    openRemoteHome,
    remoteBaseName,
    remoteJoin,
    remoteParent,
    removeRemotePath,
    renameRemotePath,
    uploadFile,
    transfersState,
  } from '@/lib/sftp.svelte'
  import { SFTPHistory } from '@/lib/sftpHistory.svelte'
  import {
    createPhaseTracker,
    ensureSSHConnected,
  } from '@/lib/sshConnectProgress'
  import { sshState } from '@/lib/sshConnections.svelte'
  import {
    emptyConnectProgress,
    SSHConnectPhase,
    SSHTerminalStatus,
    type SSHConnectProgress,
  } from '@/types/ssh'

  interface Props {
    connectionId: string
  }

  let { connectionId }: Props = $props()

  type Side = 'local' | 'remote'

  let localExplorer = $state<FileExplorer>()
  let remoteExplorer = $state<FileExplorer>()

  let localPath = $state('')
  let remotePath = $state('/')
  let remoteSelection = $state<ExplorerNode[]>([])
  let remoteNames = new Set<string>()

  let activeTransfers = $derived(
    transfersState.items.filter((item) => item.status === 'running').length,
  )

  type Notice = { text: string; isError: boolean } | null
  const notice = new Transient<Notice>(4000, null)

  let status = $state<SSHTerminalStatus>(SSHTerminalStatus.Connecting)
  let connectError = $state('')
  let progress = $state<SSHConnectProgress>(emptyConnectProgress())
  let remoteHome = $state('')

  let ready = $derived(status === SSHTerminalStatus.Open)

  const SFTP_PHASE_LABELS: Partial<Record<SSHConnectPhase, [string, string]>> =
    {
      [SSHConnectPhase.Session]: ['sftp.phaseSession', 'SFTP session'],
    }

  const connect = async () => {
    status = SSHTerminalStatus.Connecting
    connectError = ''

    const phases = createPhaseTracker((value) => (progress = value))

    try {
      phases.begin(SSHConnectPhase.Channel)
      await ensureSSHConnected(connectionId)
      phases.settle()

      phases.begin(SSHConnectPhase.Session)
      remoteHome = await openRemoteHome(connectionId)
      phases.settle()

      status = SSHTerminalStatus.Open
    } catch (error) {
      phases.fail()
      connectError = describeError(error)
      status = SSHTerminalStatus.Error
    }
  }

  onMount(connect)

  let connection = $derived(
    sshState.connections.find((item) => item.id === connectionId) ?? null,
  )
  const sourceConnectionId = $derived(connectionId)
  const remoteSource = $derived(createSFTPDataSource(sourceConnectionId))
  const columns = getFileSystemExplorerColumns()
  let remoteLabel = $derived(
    connection
      ? `${connection.username}@${connection.host}`
      : t('sftp.remote', 'Remote'),
  )

  const flash = (text: string, isError = false) =>
    notice.show({ text, isError })

  const fail = (error: unknown, action = '') => {
    logger.error(`SFTP ${action || 'action'} failed`, error)
    const message = describeError(error)
    flash(action ? `${action}: ${message}` : message, true)
  }

  interface SideOps {
    explorer: () => FileExplorer | undefined
    currentPath: () => string
    join: (directory: string, name: string) => string
    parent: (path: string) => string
    baseName: (path: string) => string
    rename: (from: string, to: string) => Promise<void>
    remove: (path: string) => Promise<void>
  }

  const sides: Record<Side, SideOps> = {
    local: {
      explorer: () => localExplorer,
      currentPath: () => localPath,
      join: localJoin,
      parent: localParent,
      baseName: localBaseName,
      rename: renameLocalPath,
      remove: removeLocalPath,
    },
    remote: {
      explorer: () => remoteExplorer,
      currentPath: () => remotePath,
      join: remoteJoin,
      parent: remoteParent,
      baseName: remoteBaseName,
      rename: (from, to) => renameRemotePath(connectionId, from, to),
      remove: (path) => removeRemotePath(connectionId, path),
    },
  }

  const settleMoves = (side: Side, moves: [string, string][]) => {
    const ops = sides[side]
    const explorer = ops.explorer()
    const current = ops.currentPath()

    for (const [from, to] of moves) {
      for (const directory of [ops.parent(from), ops.parent(to)]) {
        if (directory !== current) explorer?.invalidatePath(directory)
      }
    }
    explorer?.refresh()
  }

  const describe = (nodes: ExplorerNode[]) =>
    nodes.length === 1
      ? nodes[0].name
      : t('sftp.itemCount', '{count} items').replace(
          '{count}',
          String(nodes.length),
        )

  const filesOf = (nodes: ExplorerNode[]) =>
    nodes.filter((node) => !node.isContainer)

  const history = new SFTPHistory()

  const runHistory = async (step: () => Promise<void>) => {
    try {
      await step()
      remoteExplorer?.refresh()
      localExplorer?.refresh()
    } catch (error) {
      fail(error, t('sftp.undo', 'Undo'))
    }
  }

  const upload = async (nodes: ExplorerNode[], directory = '') => {
    const files = filesOf(nodes)
    if (files.length === 0) return

    const into = directory || remotePath
    let undoable = into === remotePath
    const uploaded: { node: ExplorerNode; target: string }[] = []

    for (const node of files) {
      const target = remoteJoin(into, node.name)
      if (remoteNames.has(node.name)) undoable = false
      if (await uploadFile(connectionId, node.path, target)) {
        uploaded.push({ node, target })
      }
    }
    if (uploaded.length === 0) return

    if (into !== remotePath) remoteExplorer?.invalidatePath(into)
    remoteExplorer?.refresh()

    const done = uploaded.map((item) => item.node)
    const landed = into === remotePath ? '' : ` → ${remoteBaseName(into)}`
    flash(`${t('sftp.uploaded', 'Uploaded')} ${describe(done)}${landed}`)

    if (!undoable) return
    history.push({
      label: `${t('sftp.upload', 'Upload')} ${describe(done)}`,
      undo: async () => {
        for (const { target } of uploaded) {
          await removeRemotePath(connectionId, target)
        }
      },
      redo: async () => {
        for (const { node, target } of uploaded) {
          await uploadFile(connectionId, node.path, target)
        }
      },
    })
  }

  const download = async (nodes: ExplorerNode[], directory = '') => {
    const files = filesOf(nodes)
    if (files.length === 0) return

    const into = directory || localPath
    const done: ExplorerNode[] = []
    for (const node of files) {
      if (
        await downloadFile(connectionId, node.path, localJoin(into, node.name))
      ) {
        done.push(node)
      }
    }
    if (done.length === 0) return

    if (into !== localPath) localExplorer?.invalidatePath(into)
    localExplorer?.refresh()

    const landed = into === localPath ? '' : ` → ${localBaseName(into)}`
    flash(`${t('sftp.downloaded', 'Downloaded')} ${describe(done)}${landed}`)
  }

  type PreviewState = {
    name: string
    origin: string
    loading: boolean
    error: string
    item: FilePreview | null
  }

  let previewState = $state<PreviewState | null>(null)
  let previewRequest = 0

  const openPreview = async (
    node: ExplorerNode | null,
    origin: string,
    read: (path: string) => Promise<FilePreview>,
  ) => {
    if (!node || node.isContainer) return

    const request = ++previewRequest
    previewState = {
      name: node.name,
      origin,
      loading: true,
      error: '',
      item: null,
    }

    try {
      const item = await read(node.path)
      if (request !== previewRequest) return
      previewState = { ...previewState, loading: false, item }
    } catch (error) {
      if (request !== previewRequest) return
      logger.error('SFTP preview failed', error)
      previewState = {
        ...previewState,
        loading: false,
        error: describeError(error),
      }
    }
  }

  const closePreview = () => {
    previewRequest += 1
    previewState = null
  }

  const onLocalActivate = (detail: { node: ExplorerNode }) =>
    openPreview(
      detail.node,
      t('sftp.thisMac', 'This computer'),
      previewLocalFile,
    )

  const onRemoteActivate = (detail: { node: ExplorerNode }) =>
    openPreview(detail.node, remoteLabel, (path) =>
      previewRemoteFile(connectionId, path),
    )

  let dragging = $state<{ side: Side; nodes: ExplorerNode[] } | null>(null)
  let dropSide = $state<Side | null>(null)
  let dropNodeSide = $state<Side | null>(null)

  let transfersOpen = $state(false)
  let transfersAnchor = $state<HTMLButtonElement | null>(null)

  const onNodeDragStart = (
    side: Side,
    detail: { node: ExplorerNode; nodes: ExplorerNode[] },
  ) => {
    dragging = { side, nodes: detail.nodes }
  }

  const endDrag = () => {
    dragging = null
    dropSide = null
    dropNodeSide = null
  }

  const onNodeDropTarget = (
    side: Side,
    detail: { node: ExplorerNode | null },
  ) => {
    dropNodeSide = detail.node ? side : null
  }

  const accepts = (side: Side): boolean =>
    Boolean(
      dragging && dragging.side !== side && filesOf(dragging.nodes).length > 0,
    )

  const acceptsInto = (side: Side): boolean =>
    Boolean(dragging) && (dragging?.side === side || accepts(side))

  let acceptsLocalDrop = $derived(dragging ? acceptsInto('local') : false)
  let acceptsRemoteDrop = $derived(dragging ? acceptsInto('remote') : false)

  const onPaneDragOver = (side: Side, event: DragEvent) => {
    if (!accepts(side)) return
    event.preventDefault()
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy'
    dropSide = side
  }

  const onPaneDragLeave = (side: Side, event: DragEvent) => {
    const next = event.relatedTarget as Node | null
    if (next && (event.currentTarget as HTMLElement).contains(next)) return
    if (dropSide === side) dropSide = null
  }

  const onPaneDrop = (side: Side, event: DragEvent) => {
    if (!accepts(side)) return
    event.preventDefault()
    transferTo(side, '')
  }

  const transferTo = (side: Side, directory: string) => {
    if (!accepts(side)) return

    const nodes = dragging?.nodes ?? []
    endDrag()
    void (side === 'remote'
      ? upload(nodes, directory)
      : download(nodes, directory))
  }

  const onNodeDrop = (side: Side, detail: { node: ExplorerNode }) => {
    const directory = detail.node.path
    if (dragging?.side !== side) {
      transferTo(side, directory)
      return
    }

    const nodes = dragging.nodes
    endDrag()
    void moveInto(side, nodes, directory)
  }

  const moveInto = async (
    side: Side,
    nodes: ExplorerNode[],
    directory: string,
  ) => {
    const ops = sides[side]
    const pending = nodes
      .map((node) => ({
        node,
        to: ops.join(directory, ops.baseName(node.path)),
      }))
      .filter(({ node, to }) => node.path !== to)
    if (pending.length === 0) return

    type Pair = [from: string, to: string]
    const moveAll = async (pairs: Pair[]) => {
      try {
        for (const [from, to] of pairs) await ops.rename(from, to)
      } finally {
        settleMoves(side, pairs)
      }
    }

    const moved: typeof pending = []
    try {
      for (const item of pending) {
        await ops.rename(item.node.path, item.to)
        moved.push(item)
      }
    } catch (error) {
      fail(error, t('sftp.move', 'Move'))
    } finally {
      settleMoves(
        side,
        pending.map(({ node, to }) => [node.path, to]),
      )
    }
    if (moved.length === 0) return

    const what = describe(moved.map(({ node }) => node))
    if (moved.length === pending.length) {
      flash(`${t('sftp.moved', 'Moved')} ${what} → ${ops.baseName(directory)}`)
    }

    const forward: Pair[] = moved.map(({ node, to }) => [node.path, to])
    const backward: Pair[] = forward
      .map(([from, to]): Pair => [to, from])
      .reverse()
    history.push({
      label: `${t('sftp.move', 'Move')} ${what}`,
      undo: () => moveAll(backward),
      redo: () => moveAll(forward),
    })
  }

  const newRemoteFolder = () =>
    askName({
      title: t('sftp.newFolder', 'New folder'),
      label: t('sftp.folderName', 'Folder name'),
      confirmLabel: t('sftp.create', 'Create'),
      value: '',
      onSubmit: async (name) => {
        const target = remoteJoin(remotePath, name)
        await createRemoteDirectory(connectionId, target)
        remoteExplorer?.refresh()
        history.push({
          label: `${t('sftp.newFolder', 'New folder')} ${name}`,
          undo: () => removeRemotePath(connectionId, target),
          redo: () => createRemoteDirectory(connectionId, target),
        })
      },
    })

  let pendingDelete = $state<{ side: Side; nodes: ExplorerNode[] } | null>(null)

  const askDelete = (side: Side, nodes: ExplorerNode[]) => {
    if (nodes.length > 0) pendingDelete = { side, nodes }
  }

  const confirmDelete = async () => {
    const request = pendingDelete
    pendingDelete = null
    if (!request) return

    const { side, nodes } = request
    const ops = sides[side]
    const removed: ExplorerNode[] = []
    try {
      for (const node of nodes) {
        await ops.remove(node.path)
        removed.push(node)
      }
      flash(`${t('sftp.deleted', 'Deleted')} ${describe(removed)}`)
    } catch (error) {
      fail(error, t('sftp.delete', 'Delete'))
    } finally {
      if (removed.length > 0) ops.explorer()?.refresh()
    }
  }

  const rememberRemoteNames = (nodes: ExplorerNode[]) => {
    remoteNames = new Set(nodes.map((node) => node.name))
  }

  let deleteMessage = $derived(
    (() => {
      if (!pendingDelete) return ''
      const { side, nodes } = pendingDelete
      if (nodes.length > 1) {
        return (
          side === 'local'
            ? t(
                'sftp.deleteLocalManyConfirm',
                'Delete {count} items from this computer? This cannot be undone.',
              )
            : t(
                'sftp.deleteManyConfirm',
                'Delete {count} items from the server? This cannot be undone.',
              )
        ).replace('{count}', String(nodes.length))
      }
      return (
        side === 'local'
          ? t(
              'sftp.deleteLocalConfirm',
              'Delete {name} from this computer? This cannot be undone.',
            )
          : t(
              'sftp.deleteConfirm',
              'Delete {name} from the server? This cannot be undone.',
            )
      ).replace('{name}', sides[side].baseName(nodes[0].path))
    })(),
  )

  type Prompt = {
    title: string
    label: string
    confirmLabel: string
    value: string
    selectionEnd: number | null
    onSubmit: (value: string) => Promise<void>
  }

  let prompt = $state<Prompt | null>(null)

  const askName = (ask: Omit<Prompt, 'selectionEnd'> & { stem?: boolean }) => {
    prompt = {
      ...ask,
      selectionEnd: ask.stem ? fileStemLength(ask.value) : null,
    }
  }

  const submitPrompt = async (detail: { value: string }) => {
    const current = prompt
    prompt = null
    if (!current) return

    try {
      await current.onSubmit(detail.value)
    } catch (error) {
      fail(error, current.title)
    }
  }

  type Menu = {
    side: Side
    node: ExplorerNode
    nodes: ExplorerNode[]
    x: number
    y: number
  }

  let menu = $state<Menu | null>(null)

  const openMenu = (side: Side, detail: Omit<Menu, 'side'>) => {
    menu = { side, ...detail }
  }

  const closeMenu = () => (menu = null)

  let menuMulti = $derived((menu?.nodes.length ?? 0) > 1)
  let menuHasFiles = $derived(menu ? filesOf(menu.nodes).length > 0 : false)

  type MenuAction = (
    node: ExplorerNode,
    side: Side,
    nodes: ExplorerNode[],
  ) => unknown

  const fromMenu = (label: string, action: MenuAction) => () => {
    if (!menu) return
    const { node, side, nodes } = menu
    closeMenu()
    void (async () => {
      try {
        await action(node, side, nodes)
      } catch (error) {
        fail(error, label)
      }
    })()
  }

  const copyPaths = async (nodes: ExplorerNode[]) => {
    const text = nodes.map((node) => node.path).join('\n')
    if (await copyText(text)) {
      flash(t('sftp.pathCopied', 'Path copied'))
      return
    }
    flash(t('sftp.copyFailed', 'Could not copy the path'), true)
  }

  const renameNode = async (side: Side, node: ExplorerNode, next: string) => {
    const ops = sides[side]
    const name = ops.baseName(node.path)
    if (next === name) return

    const from = node.path
    const target = ops.join(ops.parent(from), next)

    try {
      await ops.rename(from, target)
      ops.explorer()?.refresh()
      history.push({
        label: `${t('sftp.rename', 'Rename')} ${name}`,
        undo: () => ops.rename(target, from),
        redo: () => ops.rename(from, target),
      })
    } catch (error) {
      fail(error, t('sftp.rename', 'Rename'))
    }
  }

  const duplicateName = (name: string): string => {
    const stem = fileStemLength(name)
    const base = name.slice(0, stem)
    const extension = name.slice(stem)
    const suffix = t('sftp.copySuffix', 'copy')

    let candidate = `${base} ${suffix}${extension}`
    let counter = 2
    while (remoteNames.has(candidate)) {
      candidate = `${base} ${suffix} ${counter}${extension}`
      counter += 1
    }
    return candidate
  }

  const duplicateRemote = (node: ExplorerNode) => {
    const name = remoteBaseName(node.path)
    askName({
      title: t('sftp.duplicate', 'Duplicate'),
      label: t('sftp.newName', 'New name'),
      confirmLabel: t('sftp.duplicate', 'Duplicate'),
      value: duplicateName(name),
      stem: true,
      onSubmit: async (next) => {
        const target = remoteJoin(remotePath, next)
        await copyRemotePath(connectionId, node.path, target)
        remoteExplorer?.refresh()
        history.push({
          label: `${t('sftp.duplicate', 'Duplicate')} ${name}`,
          undo: () => removeRemotePath(connectionId, target),
          redo: () => copyRemotePath(connectionId, node.path, target),
        })
      },
    })
  }
</script>

{#snippet menuItem(
  icon: IconComponent,
  label: string,
  action: MenuAction,
  danger = false,
)}
  <MenuItem {icon} {label} {danger} onclick={fromMenu(label, action)} />
{/snippet}

{#snippet paneHeader(icon: IconComponent, label: string)}
  {@const SvelteComponent_1 = icon}
  <div class="flex h-9 shrink-0 items-center gap-2 border-b border-border px-3">
    <SvelteComponent_1 size={13} class="shrink-0 text-primary" />
    <span
      class="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase"
    >
      {label}
    </span>
  </div>
{/snippet}

{#snippet historyButton(
  icon: IconComponent,
  label: string,
  enabled: boolean,
  detail: string,
  run: () => Promise<void>,
)}
  <IconButton
    {icon}
    {label}
    size="sm"
    tooltip={detail ? `${label}: ${detail}` : label}
    disabled={!enabled}
    onclick={() => void runHistory(run)}
  />
{/snippet}

<div
  class="@container relative flex h-full min-h-0 w-full flex-col bg-surface-canvas"
>
  {#if ready}
    <header
      class="flex h-11 shrink-0 items-center gap-3 border-b border-border bg-surface-base px-3"
    >
      <span
        class="flex shrink-0 items-center gap-2 text-xs font-semibold text-white"
      >
        <ArrowLeftRight size={14} class="text-primary" />
        {t('sftp.title', 'File transfer')}
      </span>

      <span class="truncate font-mono text-[11px] text-muted-foreground">
        {remoteLabel}
      </span>

      {#if notice.current}
        <span
          class="min-w-0 text-[11px] {notice.current.isError
            ? 'line-clamp-2 select-text text-red-400'
            : 'truncate text-primary'}"
          {@attach tooltip({
            content: notice.current.text,
            overflowOnly: true,
          })}
        >
          {notice.current.text}
        </span>
      {/if}

      <div class="ml-auto flex shrink-0 items-center gap-1">
        <span class="hidden pr-1 text-[11px] text-muted-foreground @2xl:inline">
          {t('sftp.dragHint', 'Drag files between panes')}
        </span>

        {@render historyButton(
          Undo2,
          t('sftp.undo', 'Undo'),
          history.canUndo,
          history.undoLabel,
          history.undo,
        )}
        {@render historyButton(
          Redo2,
          t('sftp.redo', 'Redo'),
          history.canRedo,
          history.redoLabel,
          history.redo,
        )}

        <span class="mx-1 h-4 w-px bg-border"></span>

        <span class="relative inline-flex">
          <IconButton
            icon={ListEnd}
            size="sm"
            label={t('sftp.transfers', 'Transfers')}
            pressed={transfersOpen}
            bind:ref={transfersAnchor}
            onclick={() => (transfersOpen = !transfersOpen)}
          />
          {#if activeTransfers > 0}
            <span class="badge">{activeTransfers}</span>
          {/if}
        </span>

        <IconButton
          icon={FolderPlus}
          size="sm"
          label={t('sftp.newFolder', 'New folder')}
          onclick={newRemoteFolder}
        />
        <IconButton
          icon={Trash2}
          size="sm"
          danger
          label={t('sftp.delete', 'Delete')}
          disabled={remoteSelection.length === 0}
          onclick={() => askDelete('remote', remoteSelection)}
        />
      </div>
    </header>

    <div class="flex min-h-0 flex-1 flex-col @2xl:flex-row">
      <section
        class="pane {dropSide === 'local' && dropNodeSide !== 'local'
          ? 'pane-drop'
          : ''}"
        aria-label={t('sftp.thisMac', 'This computer')}
        ondragover={(event) => onPaneDragOver('local', event)}
        ondragleave={(event) => onPaneDragLeave('local', event)}
        ondrop={(event) => onPaneDrop('local', event)}
      >
        {@render paneHeader(HardDrive, t('sftp.thisMac', 'This computer'))}

        <FileExplorer
          bind:this={localExplorer}
          dataSource={fileSystemExplorerDataSource}
          {columns}
          draggableNodes
          contextMenuNodes
          fillHeight
          viewportMaxHeight=""
          className="min-h-0 flex-1"
          onPathChange={(detail) => (localPath = detail.path)}
          onLoad={(detail) => (localPath = detail.path)}
          onActivate={onLocalActivate}
          containerDropTargets
          dropActive={acceptsLocalDrop}
          onNodeDragStart={(event) => onNodeDragStart('local', event)}
          onNodeDragEnd={endDrag}
          onNodeDrop={(event) => onNodeDrop('local', event)}
          onNodeDropTarget={(event) => onNodeDropTarget('local', event)}
          onRenameNode={(detail) =>
            void renameNode('local', detail.node, detail.name)}
          onNodeContextMenu={(event) => openMenu('local', event)}
          onDeleteNodes={(detail) => askDelete('local', detail.nodes)}
        />
      </section>

      <section
        class="pane border-t border-border @2xl:border-t-0 @2xl:border-l {dropSide ===
          'remote' && dropNodeSide !== 'remote'
          ? 'pane-drop'
          : ''}"
        aria-label={t('sftp.server', 'Server')}
        ondragover={(event) => onPaneDragOver('remote', event)}
        ondragleave={(event) => onPaneDragLeave('remote', event)}
        ondrop={(event) => onPaneDrop('remote', event)}
      >
        {@render paneHeader(Server, t('sftp.server', 'Server'))}

        <FileExplorer
          bind:this={remoteExplorer}
          dataSource={remoteSource}
          initialPath={remoteHome}
          {columns}
          draggableNodes
          contextMenuNodes
          fillHeight
          viewportMaxHeight=""
          className="min-h-0 flex-1"
          onPathChange={(detail) => (remotePath = detail.path)}
          onLoad={(detail) => {
            remotePath = detail.path
            rememberRemoteNames(detail.nodes)
          }}
          onSelect={(detail) => (remoteSelection = detail.nodes)}
          onRenameNode={(detail) =>
            void renameNode('remote', detail.node, detail.name)}
          onActivate={onRemoteActivate}
          containerDropTargets
          dropActive={acceptsRemoteDrop}
          onNodeDragStart={(event) => onNodeDragStart('remote', event)}
          onNodeDragEnd={endDrag}
          onNodeDrop={(event) => onNodeDrop('remote', event)}
          onNodeDropTarget={(event) => onNodeDropTarget('remote', event)}
          onNodeContextMenu={(event) => openMenu('remote', event)}
          onDeleteNodes={(detail) => askDelete('remote', detail.nodes)}
        />
      </section>
    </div>
  {/if}

  <FloatingDropdown
    open={transfersOpen}
    anchorElement={transfersAnchor}
    placement="bottom-end"
    panelClass="overflow-hidden rounded-md border border-border bg-popover shadow-xl shadow-black/40"
    onClose={() => (transfersOpen = false)}
  >
    <TransferQueue />
  </FloatingDropdown>

  <FloatingDropdown
    open={Boolean(menu)}
    x={menu?.x ?? 0}
    y={menu?.y ?? 0}
    panelClass={MENU_PANEL_CLASS}
    onClose={closeMenu}
  >
    {#if menu}
      {#if menuHasFiles}
        {#if menu.side === 'remote'}
          {@render menuItem(
            ArrowDownToLine,
            t('sftp.download', 'Download'),
            (_target, _side, targets) => download(targets),
          )}
          {#if !menuMulti}
            {@render menuItem(
              Copy,
              t('sftp.duplicate', 'Duplicate'),
              duplicateRemote,
            )}
          {/if}
        {:else}
          {@render menuItem(
            ArrowUpFromLine,
            t('sftp.upload', 'Upload'),
            (_target, _side, targets) => upload(targets),
          )}
        {/if}
      {/if}
      {#if !menuMulti}
        {@render menuItem(
          PencilLine,
          t('sftp.rename', 'Rename'),
          (target, side) => sides[side].explorer()?.startRename(target),
        )}
      {/if}
      {@render menuItem(
        ClipboardCopy,
        menuMulti
          ? t('sftp.copyPaths', 'Copy paths')
          : t('sftp.copyPath', 'Copy path'),
        (_target, _side, targets) => copyPaths(targets),
      )}
      <span class={MENU_SEPARATOR_CLASS}></span>
      {@render menuItem(
        Trash2,
        t('sftp.delete', 'Delete'),
        (_target, side, targets) => askDelete(side, targets),
        true,
      )}
    {/if}
  </FloatingDropdown>

  <FilePreviewModal
    open={Boolean(previewState)}
    name={previewState?.name ?? ''}
    origin={previewState?.origin ?? ''}
    loading={previewState?.loading ?? false}
    error={previewState?.error ?? ''}
    item={previewState?.item ?? null}
    onClose={closePreview}
  />

  <ConfirmDialog
    open={Boolean(pendingDelete)}
    title={pendingDelete?.side === 'local'
      ? t('sftp.deleteLocalTitle', 'Delete from this computer?')
      : t('sftp.deleteTitle', 'Delete from the server?')}
    message={deleteMessage}
    confirmLabel={t('sftp.delete', 'Delete')}
    destructive
    onConfirm={() => void confirmDelete()}
    onClose={() => (pendingDelete = null)}
  />

  <PromptDialog
    open={Boolean(prompt)}
    title={prompt?.title ?? ''}
    label={prompt?.label ?? ''}
    confirmLabel={prompt?.confirmLabel ?? ''}
    value={prompt?.value ?? ''}
    selectionEnd={prompt?.selectionEnd ?? null}
    onSubmit={submitPrompt}
    onClose={() => (prompt = null)}
  />

  <ConnectionState
    {status}
    {connection}
    {progress}
    errorMessage={connectError}
    phaseLabels={SFTP_PHASE_LABELS}
    dismissible={false}
    onReconnect={connect}
  />
</div>

<style>
  .pane {
    position: relative;
    display: flex;
    flex: 1 1 0;
    min-width: 0;
    min-height: 0;
    flex-direction: column;
  }

  .pane-drop::after {
    content: '';
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 5;
    border: 2px solid rgb(var(--color-primary-rgb) / 0.55);
    background: rgb(var(--color-primary-rgb) / 0.06);
  }

  .badge {
    position: absolute;
    top: -0.15rem;
    right: -0.15rem;
    display: grid;
    place-items: center;
    min-width: 0.95rem;
    height: 0.95rem;
    padding: 0 0.2rem;
    border-radius: 9999px;
    background: var(--color-primary);
    color: var(--color-primary-foreground);
    font-size: 0.6rem;
    font-weight: 700;
    line-height: 1;
  }
</style>
