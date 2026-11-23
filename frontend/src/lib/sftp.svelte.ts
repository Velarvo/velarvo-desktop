import { Events, type CancellablePromise } from '@wailsio/runtime'
import { SFTP } from '@bindings/sshconn'
import { SftpTransferProgress } from '@events'
import type {
  ExplorerBreadcrumb,
  ExplorerDataSource,
  ExplorerNode,
} from '@/components/file-explorer/types'
import { describeError, isCancelled } from '@/lib/errors'

export type TransferDirection = 'upload' | 'download'
export type TransferStatus = 'running' | 'done' | 'failed' | 'cancelled'

export interface Transfer {
  id: string
  name: string
  direction: TransferDirection
  status: TransferStatus
  done: number
  total: number
  error?: string
}

class TransfersStore {
  items = $state.raw<Transfer[]>([])
}

export const transfersState = new TransfersStore()

// eslint-disable-next-line svelte/prefer-svelte-reactivity -- bookkeeping, never rendered
const running = new Map<string, CancellablePromise<void>>()

const patch = (id: string, changes: Partial<Transfer>) => {
  transfersState.items = transfersState.items.map((item) =>
    item.id === id ? { ...item, ...changes } : item,
  )
}

export const remoteJoin = (base: string, name: string): string =>
  base === '/' ? `/${name}` : `${base.replace(/\/+$/, '')}/${name}`

export const remoteParent = (path: string): string => {
  const trimmed = path.replace(/\/+$/, '')
  const index = trimmed.lastIndexOf('/')
  return index <= 0 ? '/' : trimmed.slice(0, index)
}

export const remoteBaseName = (path: string): string => {
  const trimmed = path.replace(/\/+$/, '')
  if (!trimmed) return '/'
  return trimmed.slice(trimmed.lastIndexOf('/') + 1)
}

const remoteNormalize = (path: string): string => {
  if (!path) return '/'
  const collapsed = path.replace(/\/{2,}/g, '/')
  return collapsed.length > 1 ? collapsed.replace(/\/+$/, '') : collapsed
}

const remoteBreadcrumbs = (path: string): ExplorerBreadcrumb[] => {
  const normalized = remoteNormalize(path)
  const crumbs: ExplorerBreadcrumb[] = [{ name: '/', path: '/' }]
  if (normalized === '/') return crumbs

  let current = ''
  for (const segment of normalized.split('/').filter(Boolean)) {
    current = `${current}/${segment}`
    crumbs.push({ name: segment, path: current })
  }
  return crumbs
}

export const openRemoteHome = (connectionId: string): Promise<string> =>
  SFTP.Home(connectionId)

export const createSFTPDataSource = (
  connectionId: string,
): ExplorerDataSource => ({
  getInitialPath: () => openRemoteHome(connectionId),

  async loadNodes(path: string) {
    const entries = await SFTP.List(connectionId, path)

    return (entries ?? []).map((entry): ExplorerNode => ({
      id: entry.path,
      name: entry.name,
      path: entry.path,
      type: entry.type,
      isContainer: entry.type === 'directory',
      size: entry.size,
      // eslint-disable-next-line svelte/prefer-svelte-reactivity
      modified: entry.modified ? new Date(entry.modified) : undefined,
    }))
  },

  normalizePath: remoteNormalize,
  getParentPath: remoteParent,
  getBreadcrumbs: remoteBreadcrumbs,
  isContainer: (node) => Boolean(node.isContainer),
  isHidden: (node) => node.name.startsWith('.'),
})

const runTransfer = async (
  direction: TransferDirection,
  name: string,
  start: (transferId: string) => CancellablePromise<void>,
): Promise<boolean> => {
  const id = crypto.randomUUID()
  transfersState.items = [
    ...transfersState.items,
    { id, name, direction, status: 'running', done: 0, total: 0 },
  ]

  const stopListening = Events.On(SftpTransferProgress, ({ data }) => {
    if (data.transferId === id)
      patch(id, { done: data.done, total: data.total })
  })

  const call = start(id)
  running.set(id, call)
  try {
    await call
    patch(id, { status: 'done' })
    return true
  } catch (error) {
    patch(
      id,
      isCancelled(error)
        ? { status: 'cancelled' }
        : { status: 'failed', error: describeError(error) },
    )
    return false
  } finally {
    running.delete(id)
    stopListening()
  }
}

export const uploadFile = (
  connectionId: string,
  localPath: string,
  remotePath: string,
) =>
  runTransfer('upload', remoteBaseName(remotePath), (transferId) =>
    SFTP.Upload(connectionId, localPath, remotePath, transferId),
  )

export const downloadFile = (
  connectionId: string,
  remotePath: string,
  localPath: string,
) =>
  runTransfer('download', remoteBaseName(remotePath), (transferId) =>
    SFTP.Download(connectionId, remotePath, localPath, transferId),
  )

export const cancelTransfer = (id: string): void => {
  running.get(id)?.cancel()
}

export const clearSettledTransfers = () => {
  transfersState.items = transfersState.items.filter(
    (item) => item.status === 'running',
  )
}

export const createRemoteDirectory = (connectionId: string, path: string) =>
  SFTP.Mkdir(connectionId, path)

export const removeRemotePath = (connectionId: string, path: string) =>
  SFTP.Remove(connectionId, path)

export const copyRemotePath = (
  connectionId: string,
  from: string,
  to: string,
) => SFTP.Copy(connectionId, from, to)

export const renameRemotePath = (
  connectionId: string,
  from: string,
  to: string,
) => SFTP.Rename(connectionId, from, to)
