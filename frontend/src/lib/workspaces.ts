import { derived, get, writable } from 'svelte/store'
import { CreateWorkspace, ListWorkspaces } from '../../wailsjs/go/app/App'
import type { APIResponse } from '@/types/api'
import { getResponseMessage } from '@/lib/errors'
import { t } from '@/lib/i18n'
import { logger } from '@/lib/logger'
import { selectedProjectId } from '@/lib/projects'

export interface Workspace {
  id: string
  projectId: string
  name: string
  color: string
  sortOrder: number
  createdAt: number
  updatedAt: number
  revision: string
}

export interface CreateWorkspaceInput {
  projectId: string
  name: string
  color?: string
}

const SELECTED_WORKSPACE_KEY_PREFIX = 'velarvo:selected-workspace-id:'

export const workspaces = writable<Workspace[]>([])
export const isLoadingWorkspaces = writable(false)
export const workspacesError = writable<string | null>(null)
export const selectedWorkspaceId = writable<string | null>(null)

export const selectedWorkspace = derived(
  [workspaces, selectedWorkspaceId],
  ([$workspaces, $selectedWorkspaceId]) =>
    $workspaces.find((workspace) => workspace.id === $selectedWorkspaceId) ??
    null,
)

let activeProjectId: string | null = null

const workspaceStorageKey = (projectId: string): string => {
  return `${SELECTED_WORKSPACE_KEY_PREFIX}${projectId}`
}

const readSelectedWorkspaceId = (projectId: string | null): string | null => {
  if (typeof window === 'undefined' || !projectId) {
    return null
  }

  return window.localStorage.getItem(workspaceStorageKey(projectId))
}

const persistSelectedWorkspaceId = (
  projectId: string | null,
  workspaceId: string | null,
): void => {
  if (typeof window === 'undefined' || !projectId) {
    return
  }

  if (!workspaceId) {
    window.localStorage.removeItem(workspaceStorageKey(projectId))
    return
  }

  window.localStorage.setItem(workspaceStorageKey(projectId), workspaceId)
}

const normalizeWorkspace = (workspace: unknown): Workspace => {
  const dto = (workspace ?? {}) as Record<string, unknown>
  return {
    id: String(dto.id ?? ''),
    projectId: String(dto.projectId ?? ''),
    name: String(dto.name ?? ''),
    color: String(dto.color ?? ''),
    sortOrder: Number(dto.sortOrder ?? 0),
    createdAt: Number(dto.createdAt ?? 0),
    updatedAt: Number(dto.updatedAt ?? 0),
    revision: String(dto.revision ?? ''),
  }
}

const applyWorkspaceSelection = (
  projectId: string,
  nextWorkspaces: Workspace[],
  preferredId?: string | null,
): void => {
  const currentSelectedId =
    preferredId ??
    readSelectedWorkspaceId(projectId) ??
    get(selectedWorkspaceId)
  const selected =
    nextWorkspaces.find((workspace) => workspace.id === currentSelectedId) ??
    nextWorkspaces[0] ??
    null

  selectedWorkspaceId.set(selected?.id ?? null)
  persistSelectedWorkspaceId(projectId, selected?.id ?? null)
}

export const refreshWorkspaces = async (
  projectId: string,
): Promise<APIResponse<Workspace[]>> => {
  if (!projectId) {
    workspaces.set([])
    selectedWorkspaceId.set(null)
    workspacesError.set(null)
    return { success: true, code: 'OK', data: [] }
  }

  isLoadingWorkspaces.set(true)
  workspacesError.set(null)

  try {
    const resp = await ListWorkspaces(projectId)
    if (!resp.success) {
      const message = getResponseMessage(
        resp,
        t('errors.WORKSPACE_LIST_FAILED'),
      )
      workspacesError.set(message)
      return resp as APIResponse<Workspace[]>
    }

    const nextWorkspaces = ((resp.data ?? []) as Workspace[]).map(
      normalizeWorkspace,
    )
    workspaces.set(nextWorkspaces)
    applyWorkspaceSelection(projectId, nextWorkspaces)
    return {
      ...(resp as APIResponse<Workspace[]>),
      data: nextWorkspaces,
    }
  } catch (error) {
    const message = getResponseMessage(
      {
        success: false,
        code: 'CLIENT_ERROR',
        error: error instanceof Error ? error.message : String(error),
      },
      t('errors.WORKSPACE_LIST_FAILED'),
    )
    logger.error('Failed to refresh workspaces', error)
    workspacesError.set(message)
    return {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
  } finally {
    isLoadingWorkspaces.set(false)
  }
}

export const selectWorkspace = (workspaceId: string): void => {
  selectedWorkspaceId.set(workspaceId)
  if (activeProjectId) {
    persistSelectedWorkspaceId(activeProjectId, workspaceId)
  }
}

export const createWorkspace = async (
  input: CreateWorkspaceInput,
): Promise<APIResponse<Workspace>> => {
  try {
    const resp = await CreateWorkspace({
      projectId: input.projectId,
      name: input.name,
      color: input.color ?? '',
    })
    if (!resp.success || !resp.data) {
      return resp as APIResponse<Workspace>
    }

    const created = normalizeWorkspace(resp.data)

    if (created.projectId === activeProjectId) {
      const nextWorkspaces = [...get(workspaces), created].sort((a, b) => {
        if (a.sortOrder !== b.sortOrder) {
          return a.sortOrder - b.sortOrder
        }

        return a.createdAt - b.createdAt
      })

      workspaces.set(nextWorkspaces)
      applyWorkspaceSelection(created.projectId, nextWorkspaces, created.id)
    }

    workspacesError.set(null)

    return {
      ...(resp as APIResponse<Workspace>),
      data: created,
    }
  } catch (error) {
    logger.error('Failed to create workspace', error)
    return {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
  }
}

export const clearWorkspaces = (): void => {
  workspaces.set([])
  workspacesError.set(null)
  isLoadingWorkspaces.set(false)
  selectedWorkspaceId.set(null)
  activeProjectId = null
}

selectedProjectId.subscribe((projectId) => {
  if (projectId === activeProjectId) {
    return
  }

  activeProjectId = projectId

  if (!projectId) {
    workspaces.set([])
    selectedWorkspaceId.set(null)
    workspacesError.set(null)
    return
  }

  void refreshWorkspaces(projectId).catch((error) => {
    logger.error('Failed to refresh workspaces after project change', error)
  })
})
