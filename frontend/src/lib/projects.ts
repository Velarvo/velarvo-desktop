import { derived, get, writable } from 'svelte/store'
import {
  CreateProject,
  GetProjectIcon,
  ListProjects,
  SetProjectIcon,
} from '../../wailsjs/go/app/App'
import type { APIResponse } from '@/types/api'
import { getResponseMessage } from '@/lib/errors'
import { t } from '@/lib/i18n'
import { logger } from '@/lib/logger'

export interface Project {
  id: string
  name: string
  color: string
  sortOrder: number
  hasIcon: boolean
  createdAt: number
  updatedAt: number
  revision: string
}

export interface CreateProjectInput {
  name: string
  color?: string
  icon?: ProjectIconInput
}

export interface ProjectIconInput {
  mime: 'image/png' | 'image/jpeg' | 'image/webp'
  data: number[]
}

export interface ProjectIcon {
  projectId: string
  mime?: string
  dataBase64?: string
}

const SELECTED_PROJECT_KEY = 'velarvo:selected-project-id'

const readSelectedProjectId = (): string | null => {
  if (typeof window === 'undefined') {
    return null
  }

  return window.localStorage.getItem(SELECTED_PROJECT_KEY)
}

const persistSelectedProjectId = (projectId: string | null): void => {
  if (typeof window === 'undefined') {
    return
  }

  if (!projectId) {
    window.localStorage.removeItem(SELECTED_PROJECT_KEY)
    return
  }

  window.localStorage.setItem(SELECTED_PROJECT_KEY, projectId)
}

export const projects = writable<Project[]>([])
export const isLoadingProjects = writable(false)
export const projectsError = writable<string | null>(null)
export const projectIconUrls = writable<Record<string, string>>({})
export const selectedProjectId = writable<string | null>(
  readSelectedProjectId(),
)

export const selectedProject = derived(
  [projects, selectedProjectId],
  ([$projects, $selectedProjectId]) =>
    $projects.find((project) => project.id === $selectedProjectId) ?? null,
)

const normalizeProject = (project: unknown): Project => {
  const dto = (project ?? {}) as Record<string, unknown>
  return {
    id: String(dto.id ?? ''),
    name: String(dto.name ?? ''),
    color: String(dto.color ?? ''),
    sortOrder: Number(dto.sortOrder ?? 0),
    hasIcon: Boolean(dto.hasIcon),
    createdAt: Number(dto.createdAt ?? 0),
    updatedAt: Number(dto.updatedAt ?? 0),
    revision: String(dto.revision ?? ''),
  }
}

const applyProjectSelection = (
  nextProjects: Project[],
  preferredId?: string | null,
): void => {
  const currentSelectedId = preferredId ?? get(selectedProjectId)
  const selected =
    nextProjects.find((project) => project.id === currentSelectedId) ??
    nextProjects[0] ??
    null

  selectedProjectId.set(selected?.id ?? null)
  persistSelectedProjectId(selected?.id ?? null)
}

export const refreshProjects = async (): Promise<APIResponse<Project[]>> => {
  isLoadingProjects.set(true)
  projectsError.set(null)

  try {
    const resp = await ListProjects()
    if (!resp.success) {
      const message = getResponseMessage(resp, t('errors.PROJECT_LIST_FAILED'))
      projectsError.set(message)
      return resp as APIResponse<Project[]>
    }

    const nextProjects = ((resp.data ?? []) as Project[]).map(normalizeProject)
    projects.set(nextProjects)
    applyProjectSelection(nextProjects)
    void loadProjectIcons(nextProjects)
    return {
      ...(resp as APIResponse<Project[]>),
      data: nextProjects,
    }
  } catch (error) {
    const message = getResponseMessage(
      {
        success: false,
        code: 'CLIENT_ERROR',
        error: error instanceof Error ? error.message : String(error),
      },
      t('errors.PROJECT_LIST_FAILED'),
    )
    logger.error('Failed to refresh projects', error)
    projectsError.set(message)
    return {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
  } finally {
    isLoadingProjects.set(false)
  }
}

export const bootstrapProjects = async (): Promise<void> => {
  if (get(isLoadingProjects)) {
    return
  }

  await refreshProjects()
}

export const selectProject = (projectId: string): void => {
  selectedProjectId.set(projectId)
  persistSelectedProjectId(projectId)
}

export const createProject = async (
  input: CreateProjectInput,
): Promise<APIResponse<Project>> => {
  try {
    const resp = await CreateProject({
      name: input.name,
      color: input.color ?? '',
    })
    if (!resp.success || !resp.data) {
      return resp as APIResponse<Project>
    }

    const created = normalizeProject(resp.data)

    if (input.icon) {
      const iconResp = await setProjectIcon(created.id, input.icon)
      if (iconResp.success) {
        created.hasIcon = true
      } else {
        logger.error(
          'Failed to set project icon after project creation',
          iconResp,
        )
      }
    }

    const nextProjects = [...get(projects), created].sort((a, b) => {
      if (a.sortOrder !== b.sortOrder) {
        return a.sortOrder - b.sortOrder
      }

      return a.createdAt - b.createdAt
    })

    projects.set(nextProjects)
    applyProjectSelection(nextProjects, created.id)
    projectsError.set(null)

    return {
      ...(resp as APIResponse<Project>),
      data: created,
    }
  } catch (error) {
    logger.error('Failed to create project', error)
    return {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
  }
}

export const setProjectIcon = async (
  projectId: string,
  icon: ProjectIconInput,
): Promise<APIResponse<ProjectIcon>> => {
  try {
    const resp = await SetProjectIcon({
      projectId,
      mime: icon.mime,
      data: icon.data,
    })
    if (!resp.success || !resp.data) {
      return resp as APIResponse<ProjectIcon>
    }

    const nextIcon = resp.data as ProjectIcon
    if (nextIcon.mime && nextIcon.dataBase64) {
      projectIconUrls.update((current) => ({
        ...current,
        [projectId]: `data:${nextIcon.mime};base64,${nextIcon.dataBase64}`,
      }))
    }

    return {
      ...(resp as APIResponse<ProjectIcon>),
      data: nextIcon,
    }
  } catch (error) {
    logger.error('Failed to set project icon', error)
    return {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
  }
}

export const loadProjectIcon = async (projectId: string): Promise<void> => {
  try {
    const resp = await GetProjectIcon(projectId)
    if (!resp.success || !resp.data) {
      return
    }

    const icon = resp.data as ProjectIcon
    if (!icon.mime || !icon.dataBase64) {
      return
    }

    projectIconUrls.update((current) => ({
      ...current,
      [projectId]: `data:${icon.mime};base64,${icon.dataBase64}`,
    }))
  } catch (error) {
    logger.error('Failed to load project icon', error)
  }
}

export const loadProjectIcons = async (
  nextProjects: Project[],
): Promise<void> => {
  const currentIconUrls = get(projectIconUrls)
  const iconProjects = nextProjects.filter(
    (project) => project.hasIcon && !currentIconUrls[project.id],
  )

  await Promise.all(iconProjects.map((project) => loadProjectIcon(project.id)))
}

export const clearProjects = (): void => {
  projects.set([])
  projectIconUrls.set({})
  projectsError.set(null)
  isLoadingProjects.set(false)
  selectedProjectId.set(null)
  persistSelectedProjectId(null)
}
