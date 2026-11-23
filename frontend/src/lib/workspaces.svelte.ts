import { untrack } from 'svelte'
import { Service as Workspaces, type Workspace } from '@bindings/workspaces'
import { describeError } from '@/lib/errors'
import { logger } from '@/lib/logger'
import { projectsState } from '@/lib/projects.svelte'

export type { Workspace }

export interface CreateWorkspaceInput {
  projectId: string
  name: string
  color?: string
}

const selectionKey = (projectId: string) =>
  `velarvo:selected-workspace-id:${projectId}`

class WorkspacesStore {
  projectId = $state<string | null>(null)
  items = $state.raw<Workspace[]>([])
  loading = $state(false)
  error = $state<string | null>(null)
  selectedId = $state<string | null>(null)
  selected = $derived(
    this.items.find((workspace) => workspace.id === this.selectedId) ?? null,
  )
}

export const workspacesState = new WorkspacesStore()

const reconcileSelection = (projectId: string, preferredId?: string): void => {
  const items = workspacesState.items
  const wanted = preferredId ?? localStorage.getItem(selectionKey(projectId))
  const selected = items.find((w) => w.id === wanted) ?? items[0] ?? null
  selectWorkspace(selected?.id ?? null)
}

export const selectWorkspace = (workspaceId: string | null): void => {
  workspacesState.selectedId = workspaceId
  const projectId = workspacesState.projectId
  if (!projectId) return
  if (workspaceId) localStorage.setItem(selectionKey(projectId), workspaceId)
  else localStorage.removeItem(selectionKey(projectId))
}

const loadWorkspaces = async (projectId: string | null): Promise<void> => {
  workspacesState.projectId = projectId
  workspacesState.items = []
  workspacesState.selectedId = null
  workspacesState.error = null
  if (!projectId) return

  workspacesState.loading = true
  try {
    const items = (await Workspaces.List(projectId)) ?? []
    if (workspacesState.projectId !== projectId) return
    workspacesState.items = items
    reconcileSelection(projectId)
  } catch (error) {
    logger.error('Failed to load workspaces', error)
    workspacesState.error = describeError(error)
  } finally {
    workspacesState.loading = false
  }
}

export const createWorkspace = async (
  input: CreateWorkspaceInput,
): Promise<Workspace> => {
  const workspace = await Workspaces.Create(input.projectId, {
    name: input.name,
    color: input.color ?? '',
  })
  if (workspace.projectId === workspacesState.projectId) {
    workspacesState.items = [...workspacesState.items, workspace]
    reconcileSelection(workspace.projectId, workspace.id)
  }
  return workspace
}

$effect.root(() => {
  $effect(() => {
    const projectId = projectsState.selected?.id ?? null
    untrack(() => {
      if (projectId !== workspacesState.projectId)
        void loadWorkspaces(projectId)
    })
  })
})
