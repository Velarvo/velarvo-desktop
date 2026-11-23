import { Service as Projects, type Project } from '@bindings/projects'
import { describeError } from '@/lib/errors'
import { logger } from '@/lib/logger'

export type { Project }

export interface ProjectIconInput {
  mime: 'image/png' | 'image/jpeg' | 'image/webp'
  data: string
}

export interface CreateProjectInput {
  name: string
  color?: string
  icon?: ProjectIconInput
}

const SELECTED_PROJECT_KEY = 'velarvo:selected-project-id'

const saveSelection = (projectId: string | null): void => {
  if (projectId) localStorage.setItem(SELECTED_PROJECT_KEY, projectId)
  else localStorage.removeItem(SELECTED_PROJECT_KEY)
}

class ProjectsStore {
  items = $state.raw<Project[]>([])
  loading = $state(false)
  error = $state<string | null>(null)
  iconUrls = $state.raw<Record<string, string>>({})
  selectedId = $state<string | null>(localStorage.getItem(SELECTED_PROJECT_KEY))
  selected = $derived(
    this.items.find((project) => project.id === this.selectedId) ?? null,
  )
}

export const projectsState = new ProjectsStore()

const reconcileSelection = (preferredId = projectsState.selectedId): void => {
  const items = projectsState.items
  const selected = items.find((p) => p.id === preferredId) ?? items[0] ?? null
  projectsState.selectedId = selected?.id ?? null
  saveSelection(projectsState.selectedId)
}

const setIconUrl = (projectId: string, url: string): void => {
  projectsState.iconUrls = { ...projectsState.iconUrls, [projectId]: url }
}

const loadIcons = async (projects: Project[]): Promise<void> => {
  const missing = projects.filter(
    (project) => project.hasIcon && !projectsState.iconUrls[project.id],
  )
  await Promise.all(
    missing.map(async (project) => {
      try {
        setIconUrl(project.id, await Projects.Icon(project.id))
      } catch (error) {
        logger.error('Failed to load project icon', error)
      }
    }),
  )
}

export const refreshProjects = async (): Promise<void> => {
  projectsState.loading = true
  projectsState.error = null
  try {
    projectsState.items = (await Projects.List()) ?? []
    reconcileSelection()
    void loadIcons(projectsState.items)
  } catch (error) {
    logger.error('Failed to load projects', error)
    projectsState.error = describeError(error)
  } finally {
    projectsState.loading = false
  }
}

export const bootstrapProjects = async (): Promise<void> => {
  if (!projectsState.loading) await refreshProjects()
}

export const selectProject = (projectId: string): void => {
  projectsState.selectedId = projectId
  saveSelection(projectId)
}

export const createProject = async (
  input: CreateProjectInput,
): Promise<Project> => {
  let project = await Projects.Create({
    name: input.name,
    color: input.color ?? '',
  })

  if (input.icon) {
    try {
      setIconUrl(
        project.id,
        await Projects.SetIcon(project.id, input.icon.mime, input.icon.data),
      )
      project = { ...project, hasIcon: true }
    } catch (error) {
      logger.error('Failed to set the icon of a new project', error)
    }
  }

  projectsState.items = [...projectsState.items, project]
  reconcileSelection(project.id)
  return project
}

export const clearProjects = (): void => {
  projectsState.items = []
  projectsState.iconUrls = {}
  projectsState.error = null
  projectsState.loading = false
  projectsState.selectedId = null
  saveSelection(null)
}
