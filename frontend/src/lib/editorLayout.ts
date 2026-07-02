import { get, writable } from 'svelte/store'

export type Orientation = 'row' | 'column'
export type SplitDirection = 'right' | 'left' | 'down' | 'up'

export enum EditorTabKind {
  Session = 'session',
  SSHConnectionForm = 'ssh-form',
  SSHTerminal = 'ssh-terminal',
}

export enum EditorResourceKind {
  SSHConnection = 'ssh',
}

interface EditorTabDataByKind {
  [EditorTabKind.Session]: Record<string, unknown>
  [EditorTabKind.SSHConnectionForm]: { connectionId?: string }
  [EditorTabKind.SSHTerminal]: { connectionId: string }
}

type EditorTabContentByKind<K extends EditorTabKind> = {
  title: string
  kind: K
  data: EditorTabDataByKind[K]
}

export type EditorTabContent = {
  [K in EditorTabKind]: EditorTabContentByKind<K>
}[EditorTabKind]

export type EditorTab = {
  [K in EditorTabKind]: EditorTabContentByKind<K> & { id: string }
}[EditorTabKind]

const materializeTab = (id: string, content: EditorTabContent): EditorTab => {
  switch (content.kind) {
    case EditorTabKind.Session:
    case EditorTabKind.SSHConnectionForm:
    case EditorTabKind.SSHTerminal:
      return { id, ...content }
  }
}

export interface GroupNode {
  type: 'group'
  id: string
  tabs: EditorTab[]
  activeTabId: string | null
}

export interface SplitNode {
  type: 'split'
  id: string
  orientation: Orientation
  children: LayoutNode[]
  sizes: number[]
}

export type LayoutNode = GroupNode | SplitNode

export interface EditorState {
  root: LayoutNode
  activeGroupId: string
}

export const MIN_PANE_FRACTION = 0.08

let seq = 0
const nextId = (prefix: string): string => {
  seq += 1
  return `${prefix}-${seq.toString(36)}`
}

const createTab = (content: EditorTabContent): EditorTab =>
  materializeTab(nextId('tab'), content)

const createGroup = (tabs: EditorTab[] = []): GroupNode => ({
  type: 'group',
  id: nextId('grp'),
  tabs,
  activeTabId: tabs[tabs.length - 1]?.id ?? null,
})

const clone = <T>(value: T): T =>
  typeof structuredClone === 'function'
    ? structuredClone(value)
    : (JSON.parse(JSON.stringify(value)) as T)

const findGroup = (node: LayoutNode, id: string): GroupNode | null => {
  if (node.type === 'group') return node.id === id ? node : null
  for (const child of node.children) {
    const found = findGroup(child, id)
    if (found) return found
  }
  return null
}

const firstGroup = (node: LayoutNode): GroupNode => {
  if (node.type === 'group') return node
  return firstGroup(node.children[0])
}

const groupExists = (node: LayoutNode, id: string): boolean =>
  findGroup(node, id) !== null

const initialState = (): EditorState => {
  const group = createGroup()
  return { root: group, activeGroupId: group.id }
}

export const editorState = writable<EditorState>(initialState())

export const draggingTab = writable<{
  fromGroupId: string
  tabId: string
  title: string
} | null>(null)

export const draggingConnection = writable<{
  kind: EditorResourceKind
  id: string
  name: string
} | null>(null)

const update = (mutator: (draft: EditorState) => void) => {
  editorState.update((state) => {
    const draft = clone(state)
    mutator(draft)
    if (!groupExists(draft.root, draft.activeGroupId)) {
      draft.activeGroupId = firstGroup(draft.root).id
    }
    return draft
  })
}

export const setActiveGroup = (groupId: string) =>
  update((draft) => {
    if (groupExists(draft.root, groupId)) draft.activeGroupId = groupId
  })

export const activateTab = (groupId: string, tabId: string) =>
  update((draft) => {
    const group = findGroup(draft.root, groupId)
    if (!group) return
    if (group.tabs.some((tab) => tab.id === tabId)) {
      group.activeTabId = tabId
      draft.activeGroupId = groupId
    }
  })

export const openSession = (title: string) =>
  update((draft) => {
    const group =
      findGroup(draft.root, draft.activeGroupId) ?? firstGroup(draft.root)
    const tab = createTab({
      title,
      kind: EditorTabKind.Session,
      data: {},
    })
    group.tabs.push(tab)
    group.activeTabId = tab.id
    draft.activeGroupId = group.id
  })

export interface OpenContentTabOptions {
  groupId?: string
  isSame?: (tab: EditorTab) => boolean
  index?: number
}

export const openContentTab = (
  params: EditorTabContent,
  options: OpenContentTabOptions = {},
) => {
  const { groupId, isSame, index } = options
  const state = get(editorState)
  const targetGroupId =
    (groupId && groupExists(state.root, groupId) && groupId) ||
    (groupExists(state.root, state.activeGroupId) && state.activeGroupId) ||
    firstGroup(state.root).id

  if (isSame) {
    const group = findGroup(state.root, targetGroupId)
    const existing = group?.tabs.find(isSame)
    if (existing) {
      activateTab(targetGroupId, existing.id)
      return
    }
  }

  update((draft) => {
    const group = findGroup(draft.root, targetGroupId) ?? firstGroup(draft.root)
    const tab = createTab(params)
    if (index === undefined) {
      group.tabs.push(tab)
    } else {
      const clamped = Math.max(0, Math.min(index, group.tabs.length))
      group.tabs.splice(clamped, 0, tab)
    }
    group.activeTabId = tab.id
    draft.activeGroupId = group.id
  })
}

export const replaceTab = (tabId: string, params: EditorTabContent) =>
  update((draft) => {
    const visit = (node: LayoutNode): boolean => {
      if (node.type === 'group') {
        const index = node.tabs.findIndex((item) => item.id === tabId)
        if (index < 0) return false
        node.tabs[index] = materializeTab(tabId, params)
        node.activeTabId = tabId
        draft.activeGroupId = node.id
        return true
      }
      return node.children.some(visit)
    }
    visit(draft.root)
  })

export const updateTab = (
  tabId: string,
  patch: { title?: string; data?: Record<string, unknown> },
) =>
  update((draft) => {
    const visit = (node: LayoutNode): boolean => {
      if (node.type === 'group') {
        const tab = node.tabs.find((item) => item.id === tabId)
        if (!tab) return false
        if (patch.title !== undefined) tab.title = patch.title
        if (patch.data) Object.assign(tab.data, patch.data)
        return true
      }
      return node.children.some(visit)
    }
    visit(draft.root)
  })

export const closeTab = (groupId: string, tabId: string) =>
  update((draft) => {
    const group = findGroup(draft.root, groupId)
    if (!group) return

    const index = group.tabs.findIndex((tab) => tab.id === tabId)
    if (index < 0) return

    group.tabs.splice(index, 1)

    if (group.activeTabId === tabId) {
      const neighbour = group.tabs[index] ?? group.tabs[index - 1] ?? null
      group.activeTabId = neighbour?.id ?? null
    }

    if (group.tabs.length === 0) {
      const removed = removeGroup(draft.root, groupId)
      if (removed) draft.root = removed
    }
  })

export const reorderTab = (groupId: string, tabId: string, toIndex: number) =>
  update((draft) => {
    const group = findGroup(draft.root, groupId)
    if (!group) return
    const from = group.tabs.findIndex((tab) => tab.id === tabId)
    if (from < 0) return
    const [tab] = group.tabs.splice(from, 1)
    const clamped = Math.max(0, Math.min(toIndex, group.tabs.length))
    group.tabs.splice(clamped, 0, tab)
  })

export const moveTab = (
  fromGroupId: string,
  tabId: string,
  toGroupId: string,
  toIndex: number,
) =>
  update((draft) => {
    if (fromGroupId === toGroupId) {
      const group = findGroup(draft.root, fromGroupId)
      if (!group) return
      const from = group.tabs.findIndex((tab) => tab.id === tabId)
      if (from < 0) return
      const [tab] = group.tabs.splice(from, 1)
      const clamped = Math.max(0, Math.min(toIndex, group.tabs.length))
      group.tabs.splice(clamped, 0, tab)
      group.activeTabId = tab.id
      return
    }

    const source = findGroup(draft.root, fromGroupId)
    const target = findGroup(draft.root, toGroupId)
    if (!source || !target) return

    const from = source.tabs.findIndex((tab) => tab.id === tabId)
    if (from < 0) return

    const [tab] = source.tabs.splice(from, 1)
    if (source.activeTabId === tabId) {
      const neighbour = source.tabs[from] ?? source.tabs[from - 1] ?? null
      source.activeTabId = neighbour?.id ?? null
    }

    const clamped = Math.max(0, Math.min(toIndex, target.tabs.length))
    target.tabs.splice(clamped, 0, tab)
    target.activeTabId = tab.id
    draft.activeGroupId = target.id

    if (source.tabs.length === 0) {
      const removed = removeGroup(draft.root, fromGroupId)
      if (removed) draft.root = removed
    }
  })

export const splitGroup = (groupId: string, direction: SplitDirection) =>
  update((draft) => {
    const newGroup = createGroup()
    draft.root = splitInTree(draft.root, groupId, direction, newGroup)
    draft.activeGroupId = newGroup.id
  })

export const splitGroupWithTab = (
  targetGroupId: string,
  direction: SplitDirection,
  content: EditorTabContent,
): string => {
  let newGroupId = ''
  update((draft) => {
    const newGroup = createGroup([createTab(content)])
    draft.root = splitInTree(draft.root, targetGroupId, direction, newGroup)
    draft.activeGroupId = newGroup.id
    newGroupId = newGroup.id
  })
  return newGroupId
}

export const splitGroupMoveTab = (
  fromGroupId: string,
  tabId: string,
  targetGroupId: string,
  direction: SplitDirection,
) =>
  update((draft) => {
    const source = findGroup(draft.root, fromGroupId)
    if (!source) return

    const from = source.tabs.findIndex((tab) => tab.id === tabId)
    if (from < 0) return

    if (fromGroupId === targetGroupId && source.tabs.length === 1) return

    const [tab] = source.tabs.splice(from, 1)
    if (source.activeTabId === tabId) {
      const neighbour = source.tabs[from] ?? source.tabs[from - 1] ?? null
      source.activeTabId = neighbour?.id ?? null
    }

    const newGroup = createGroup([tab])
    draft.root = splitInTree(draft.root, targetGroupId, direction, newGroup)
    draft.activeGroupId = newGroup.id

    if (source.tabs.length === 0) {
      const removed = removeGroup(draft.root, fromGroupId)
      if (removed) draft.root = removed
    }
  })

export const closeGroup = (groupId: string) =>
  update((draft) => {
    if (draft.root.type === 'group' && draft.root.id === groupId) {
      draft.root.tabs = []
      draft.root.activeTabId = null
      return
    }
    const removed = removeGroup(draft.root, groupId)
    draft.root = removed ?? createGroup()
  })

export const resizeSplit = (splitId: string, sizes: number[]) =>
  update((draft) => {
    const split = findSplit(draft.root, splitId)
    if (split) split.sizes = sizes
  })

export const resetLayout = () => editorState.set(initialState())

const findSplit = (node: LayoutNode, id: string): SplitNode | null => {
  if (node.type === 'group') return null
  if (node.id === id) return node
  for (const child of node.children) {
    const found = findSplit(child, id)
    if (found) return found
  }
  return null
}

const orientationFor = (direction: SplitDirection): Orientation =>
  direction === 'left' || direction === 'right' ? 'row' : 'column'

const insertsBefore = (direction: SplitDirection): boolean =>
  direction === 'left' || direction === 'up'

const splitInTree = (
  node: LayoutNode,
  targetId: string,
  direction: SplitDirection,
  newGroup: GroupNode,
): LayoutNode => {
  const orientation = orientationFor(direction)
  const before = insertsBefore(direction)

  if (node.type === 'group') {
    if (node.id !== targetId) return node
    const children = before ? [newGroup, node] : [node, newGroup]
    return {
      type: 'split',
      id: nextId('spl'),
      orientation,
      children,
      sizes: [0.5, 0.5],
    }
  }

  const directIndex = node.children.findIndex(
    (child) => child.type === 'group' && child.id === targetId,
  )

  if (directIndex >= 0 && node.orientation === orientation) {
    const children = [...node.children]
    const sizes = [...node.sizes]
    const half = sizes[directIndex] / 2
    sizes[directIndex] = half
    const insertAt = before ? directIndex : directIndex + 1
    children.splice(insertAt, 0, newGroup)
    sizes.splice(insertAt, 0, half)
    return { ...node, children, sizes }
  }

  return {
    ...node,
    children: node.children.map((child) =>
      splitInTree(child, targetId, direction, newGroup),
    ),
  }
}

const removeGroup = (node: LayoutNode, targetId: string): LayoutNode | null => {
  if (node.type === 'group') {
    return node.id === targetId ? null : node
  }

  const children: LayoutNode[] = []
  const sizes: number[] = []
  node.children.forEach((child, index) => {
    const result = removeGroup(child, targetId)
    if (result !== null) {
      children.push(result)
      sizes.push(node.sizes[index])
    }
  })

  if (children.length === 0) return null
  if (children.length === 1) return children[0]

  const total = sizes.reduce((sum, value) => sum + value, 0)
  return { ...node, children, sizes: sizes.map((value) => value / total) }
}

export const countGroups = (node: LayoutNode): number =>
  node.type === 'group'
    ? 1
    : node.children.reduce((sum, child) => sum + countGroups(child), 0)
