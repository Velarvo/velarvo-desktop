import { untrack } from 'svelte'
import { Service as SSH } from '@bindings/sshconn'
import { describeError } from '@/lib/errors'
import { t } from '@/lib/i18n.svelte'
import { logger } from '@/lib/logger'
import { workspacesState } from '@/lib/workspaces.svelte'
import type {
  CreateSSHKeyInput,
  ImportSSHKeyInput,
  SSHConnection,
  SSHConnectionInput,
  SSHKey,
} from '@/types/ssh'

export type {
  CreateSSHKeyInput,
  ImportSSHKeyInput,
  SSHConnection,
  SSHConnectionInput,
  SSHKey,
} from '@/types/ssh'

class SSHStore {
  workspaceId = $state<string | null>(null)
  connections = $state.raw<SSHConnection[]>([])
  keys = $state.raw<SSHKey[]>([])
  loading = $state(false)
  error = $state<string | null>(null)
  selectedId = $state<string | null>(null)
  connectedIds = $derived(
    new Set(this.connections.filter((c) => c.connected).map((c) => c.id)),
  )
  selected = $derived(
    this.connections.find((item) => item.id === this.selectedId) ?? null,
  )
}

export const sshState = new SSHStore()

const requireWorkspace = (): string => {
  const workspaceId = sshState.workspaceId
  if (!workspaceId) throw new Error(t('errors.SSH_NO_WORKSPACE'))
  return workspaceId
}

const upsertConnection = (connection: SSHConnection): void => {
  const exists = sshState.connections.some((c) => c.id === connection.id)
  sshState.connections = exists
    ? sshState.connections.map((c) => (c.id === connection.id ? connection : c))
    : [connection, ...sshState.connections]
}

const addKey = (key: SSHKey): void => {
  sshState.keys = [...sshState.keys, key].sort((a, b) =>
    a.label.localeCompare(b.label, undefined, { sensitivity: 'base' }),
  )
}

const loadWorkspace = async (workspaceId: string | null): Promise<void> => {
  sshState.workspaceId = workspaceId
  sshState.connections = []
  sshState.keys = []
  sshState.selectedId = null
  sshState.error = null
  if (!workspaceId) return

  sshState.loading = true
  try {
    const [connections, keys] = await Promise.all([
      SSH.List(workspaceId),
      SSH.ListKeys(workspaceId),
    ])
    if (sshState.workspaceId !== workspaceId) return
    sshState.connections = connections ?? []
    sshState.keys = keys ?? []
  } catch (error) {
    logger.error('Failed to load SSH connections', error)
    sshState.error = describeError(error)
  } finally {
    sshState.loading = false
  }
}

export const refreshSSHKeys = async (): Promise<void> => {
  const workspaceId = sshState.workspaceId
  if (!workspaceId) return
  try {
    sshState.keys = (await SSH.ListKeys(workspaceId)) ?? []
  } catch (error) {
    logger.error('Failed to load SSH keys', error)
  }
}

export const createSSHKey = async (
  input: CreateSSHKeyInput,
): Promise<SSHKey> => {
  const key = await SSH.GenerateKey(requireWorkspace(), input)
  addKey(key)
  return key
}

export const importSSHKey = async (
  input: ImportSSHKeyInput,
): Promise<SSHKey> => {
  const key = await SSH.ImportKey(requireWorkspace(), input)
  addKey(key)
  return key
}

export const createSSHConnection = async (
  input: SSHConnectionInput,
): Promise<SSHConnection> => {
  const connection = await SSH.Create(requireWorkspace(), input)
  upsertConnection(connection)
  sshState.selectedId = connection.id
  return connection
}

export const updateSSHConnection = async (
  id: string,
  input: SSHConnectionInput,
): Promise<SSHConnection> => {
  const connection = await SSH.Update(id, input)
  upsertConnection(connection)
  return connection
}

export const connectSSH = async (id: string): Promise<void> => {
  upsertConnection(await SSH.Connect(id))
}

export const selectSSHConnection = (id: string | null): void => {
  sshState.selectedId = id
}

$effect.root(() => {
  $effect(() => {
    const workspaceId = workspacesState.selectedId
    untrack(() => {
      if (workspaceId !== sshState.workspaceId) void loadWorkspace(workspaceId)
    })
  })
})
