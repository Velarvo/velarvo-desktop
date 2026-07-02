import { derived, get, writable } from 'svelte/store'
import {
  ConnectSSHConnection,
  CreateSSHConnection,
  DeleteSSHConnection,
  DisconnectSSHConnection,
  ListSSHConnections,
  UpdateSSHConnection,
} from '../../wailsjs/go/app/App'
import type { APIResponse } from '@/types/api'
import { unwrapResponse } from '@/lib/errors'
import { t } from '@/lib/i18n'
import { logger } from '@/lib/logger'
import { selectedWorkspaceId } from '@/lib/workspaces'
import {
  SSHConnectionStatus,
  SSH_DEFAULT_PORT,
  type CreateSSHConnectionInput,
  type SSHConnection,
  type UpdateSSHConnectionInput,
} from '@/types/ssh'

export type {
  CreateSSHConnectionInput,
  SSHConnection,
  UpdateSSHConnectionInput,
} from '@/types/ssh'

export const sshConnections = writable<SSHConnection[]>([])
export const isLoadingSSH = writable(false)
export const sshError = writable<string | null>(null)
export const selectedSSHConnectionId = writable<string | null>(null)
export const connectedSSHIds = derived(sshConnections, ($connections) => {
  const ids = new Set<string>()
  for (const connection of $connections) {
    if (connection.status === SSHConnectionStatus.Connected) {
      ids.add(connection.id)
    }
  }
  return ids
})

export const selectedSSHConnection = derived(
  [sshConnections, selectedSSHConnectionId],
  ([$sshConnections, $id]) =>
    $sshConnections.find((connection) => connection.id === $id) ?? null,
)

const normalizeConnection = (raw: unknown): SSHConnection => {
  const dto = (raw ?? {}) as Record<string, unknown>
  return {
    id: String(dto.id ?? ''),
    workspaceId: String(dto.workspaceId ?? ''),
    name: String(dto.name ?? ''),
    host: String(dto.host ?? ''),
    port: Number(dto.port ?? SSH_DEFAULT_PORT),
    username: String(dto.username ?? ''),
    hasPassword: Boolean(dto.hasPassword),
    os: String(dto.os ?? ''),
    status:
      dto.connected === true
        ? SSHConnectionStatus.Connected
        : SSHConnectionStatus.Disconnected,
    sortOrder: Number(dto.sortOrder ?? 0),
    lastUsedAt: dto.lastUsedAt != null ? Number(dto.lastUsedAt) : undefined,
    createdAt: Number(dto.createdAt ?? 0),
    updatedAt: Number(dto.updatedAt ?? 0),
    revision: String(dto.revision ?? ''),
  }
}

const upsertConnection = (connection: SSHConnection): void => {
  sshConnections.update((list) => {
    const index = list.findIndex((item) => item.id === connection.id)
    if (index < 0) return [connection, ...list]
    const next = [...list]
    next[index] = connection
    return next
  })
}

const setConnectionStatus = (id: string, status: SSHConnectionStatus): void => {
  sshConnections.update((list) =>
    list.map((item) => (item.id === id ? { ...item, status } : item)),
  )
}

export const refreshSSHConnections = async (): Promise<
  APIResponse<SSHConnection[]>
> => {
  const workspaceId = get(selectedWorkspaceId)
  if (!workspaceId) {
    sshConnections.set([])
    selectedSSHConnectionId.set(null)
    sshError.set(null)
    return { success: true, code: 'OK', data: [] }
  }

  isLoadingSSH.set(true)
  sshError.set(null)
  try {
    const resp = (await ListSSHConnections(workspaceId)) as APIResponse<
      unknown[]
    >
    const data = unwrapResponse(resp, t('errors.SSH_LIST_FAILED'))
    const connections = (data ?? []).map(normalizeConnection)
    sshConnections.set(connections)
    return { success: true, code: 'OK', data: connections }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    sshError.set(message)
    logger.error('Failed to load SSH connections', error)
    return { success: false, code: 'SSH_LIST_FAILED', error: message }
  } finally {
    isLoadingSSH.set(false)
  }
}

export const createSSHConnection = async (
  input: CreateSSHConnectionInput,
): Promise<SSHConnection> => {
  const workspaceId = get(selectedWorkspaceId)
  if (!workspaceId) {
    throw new Error(t('errors.SSH_NO_WORKSPACE'))
  }

  const resp = (await CreateSSHConnection({
    ...input,
    workspaceId,
  })) as APIResponse<unknown>
  const connection = normalizeConnection(
    unwrapResponse(resp, t('errors.SSH_CREATE_FAILED')),
  )
  upsertConnection(connection)
  selectedSSHConnectionId.set(connection.id)
  return connection
}

export const updateSSHConnection = async (
  input: UpdateSSHConnectionInput,
): Promise<SSHConnection> => {
  const resp = (await UpdateSSHConnection(input)) as APIResponse<unknown>
  const connection = normalizeConnection(
    unwrapResponse(resp, t('errors.SSH_UPDATE_FAILED')),
  )
  upsertConnection(connection)
  return connection
}

export const deleteSSHConnection = async (id: string): Promise<void> => {
  const resp = (await DeleteSSHConnection(id)) as APIResponse<unknown>
  unwrapResponse(resp, t('errors.SSH_DELETE_FAILED'))
  sshConnections.update((list) => list.filter((item) => item.id !== id))
  if (get(selectedSSHConnectionId) === id) {
    selectedSSHConnectionId.set(null)
  }
}

export const connectSSH = async (id: string): Promise<void> => {
  const resp = (await ConnectSSHConnection(id)) as APIResponse<{ os?: string }>
  const state = unwrapResponse(resp, t('errors.SSH_CONNECT_FAILED'))
  sshConnections.update((list) =>
    list.map((item) =>
      item.id === id
        ? {
            ...item,
            status: SSHConnectionStatus.Connected,
            os: state?.os || item.os,
          }
        : item,
    ),
  )
}

export const disconnectSSH = async (id: string): Promise<void> => {
  const resp = (await DisconnectSSHConnection(id)) as APIResponse<unknown>
  unwrapResponse(resp, t('errors.SSH_DISCONNECT_FAILED'))
  setConnectionStatus(id, SSHConnectionStatus.Disconnected)
}

export const selectSSHConnection = (id: string | null): void => {
  selectedSSHConnectionId.set(id)
}

let activeWorkspaceId: string | null = null

selectedWorkspaceId.subscribe((workspaceId) => {
  if (workspaceId === activeWorkspaceId) {
    return
  }

  activeWorkspaceId = workspaceId
  selectedSSHConnectionId.set(null)

  void refreshSSHConnections().catch((error) => {
    logger.error(
      'Failed to refresh SSH connections after workspace change',
      error,
    )
  })
})
