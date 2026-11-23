import {
  DEFAULT_SECTION_KIND,
  isApiMethod,
  isSidebarItemStatus,
} from '@/lib/sidebar'
import type {
  SidebarItem,
  SidebarItemDraft,
  SidebarSection,
} from '@/types/sidebar'

const STORAGE_KEY = 'velarvo.sidebar.sections.v1'

const defaultSections: SidebarSection[] = []

const makeId = (prefix: string) => {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`
}

const normalizeItem = (item: unknown): SidebarItem | null => {
  if (!item || typeof item !== 'object') return null

  const candidate = item as Record<string, unknown>
  if (typeof candidate.id !== 'string') return null
  if (typeof candidate.name !== 'string') return null
  if (typeof candidate.type !== 'string') return null

  const method = isApiMethod(candidate.method) ? candidate.method : undefined
  const status = isSidebarItemStatus(candidate.status)
    ? candidate.status
    : undefined

  return {
    id: candidate.id,
    name: candidate.name,
    type: candidate.type,
    method,
    status,
  }
}

const normalizeSection = (section: unknown): SidebarSection | null => {
  if (!section || typeof section !== 'object') return null

  const candidate = section as Record<string, unknown>
  if (typeof candidate.id !== 'string') return null
  if (typeof candidate.label !== 'string') return null
  if (!Array.isArray(candidate.items)) return null

  const items = candidate.items
    .map((item) => normalizeItem(item))
    .filter((item): item is SidebarItem => item !== null)

  return {
    id: candidate.id,
    label: candidate.label,
    kind: DEFAULT_SECTION_KIND,
    items,
  }
}

const getInitialSections = () => {
  if (typeof window === 'undefined') return defaultSections

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return defaultSections

    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return defaultSections

    const sections = parsed
      .map((section) => normalizeSection(section))
      .filter((section): section is SidebarSection => section !== null)

    return sections.length > 0 ? sections : defaultSections
  } catch {
    return defaultSections
  }
}

class SidebarSections {
  sections = $state.raw<SidebarSection[]>(getInitialSections())

  #update(mutator: (sections: SidebarSection[]) => SidebarSection[]) {
    this.sections = mutator(this.sections)
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(this.sections))
    } catch {}
  }

  addSection(customLabel?: string): SidebarSection | null {
    let created: SidebarSection | null = null

    this.#update((sections) => {
      const label = (customLabel ?? '').trim()
      if (!label) return sections

      created = {
        id: makeId('section'),
        label,
        kind: DEFAULT_SECTION_KIND,
        items: [],
      }

      return [...sections, created]
    })

    return created
  }

  removeSection(sectionId: string) {
    this.#update((sections) =>
      sections.filter((section) => section.id !== sectionId),
    )
  }

  renameSection(sectionId: string, label: string) {
    const normalized = label.trim()
    if (!normalized) return

    this.#update((sections) =>
      sections.map((section) => {
        if (section.id !== sectionId) return section
        return {
          ...section,
          label: normalized,
        }
      }),
    )
  }

  addItem(sectionId: string, payload: SidebarItemDraft) {
    this.#update((sections) =>
      sections.map((section) => {
        if (section.id !== sectionId) return section

        const newItem: SidebarItem = {
          id: makeId('item'),
          name: payload.name,
          type: payload.type,
          status: payload.status,
          method: payload.method,
        }

        return {
          ...section,
          items: [...section.items, newItem],
        }
      }),
    )
  }

  removeItem(sectionId: string, itemId: string) {
    this.#update((sections) =>
      sections.map((section) => {
        if (section.id !== sectionId) return section
        return {
          ...section,
          items: section.items.filter((item) => item.id !== itemId),
        }
      }),
    )
  }

  renameItem(sectionId: string, itemId: string, name: string) {
    const normalized = name.trim()
    if (!normalized) return

    this.#update((sections) =>
      sections.map((section) => {
        if (section.id !== sectionId) return section

        return {
          ...section,
          items: section.items.map((item) => {
            if (item.id !== itemId) return item
            return {
              ...item,
              name: normalized,
            }
          }),
        }
      }),
    )
  }

  reorderItem(sectionId: string, itemId: string, toIndex: number) {
    this.#update((sections) =>
      sections.map((section) => {
        if (section.id !== sectionId) return section

        const sourceIndex = section.items.findIndex(
          (item) => item.id === itemId,
        )
        if (sourceIndex === -1) return section

        const boundedTargetIndex = Math.max(
          0,
          Math.min(Math.trunc(toIndex), section.items.length - 1),
        )

        if (boundedTargetIndex === sourceIndex) return section

        const nextItems = [...section.items]
        const [movedItem] = nextItems.splice(sourceIndex, 1)
        if (!movedItem) return section

        const insertIndex = Math.max(
          0,
          Math.min(boundedTargetIndex, nextItems.length),
        )
        nextItems.splice(insertIndex, 0, movedItem)

        return {
          ...section,
          items: nextItems,
        }
      }),
    )
  }

  moveItem(
    fromSectionId: string,
    itemId: string,
    toSectionId: string,
    toIndex: number,
  ) {
    if (fromSectionId === toSectionId) {
      this.reorderItem(fromSectionId, itemId, toIndex)
      return
    }

    this.#update((sections) => {
      const sourceSection = sections.find(
        (section) => section.id === fromSectionId,
      )
      const targetSection = sections.find(
        (section) => section.id === toSectionId,
      )

      if (!sourceSection || !targetSection) return sections

      const sourceIndex = sourceSection.items.findIndex(
        (item) => item.id === itemId,
      )
      if (sourceIndex === -1) return sections

      const movedItem = sourceSection.items[sourceIndex]
      if (!movedItem) return sections

      const boundedTargetIndex = Math.max(
        0,
        Math.min(Math.trunc(toIndex), targetSection.items.length),
      )

      return sections.map((section) => {
        if (section.id === fromSectionId) {
          return {
            ...section,
            items: section.items.filter((item) => item.id !== itemId),
          }
        }

        if (section.id === toSectionId) {
          const nextItems = [...section.items]
          nextItems.splice(boundedTargetIndex, 0, movedItem)

          return {
            ...section,
            items: nextItems,
          }
        }

        return section
      })
    })
  }
}

export const sidebarSections = new SidebarSections()
