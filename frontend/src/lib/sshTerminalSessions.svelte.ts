import { untrack } from 'svelte'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { Events } from '@wailsio/runtime'
import { Terminal as GoTerminal } from '@bindings/sshconn'
import { SshTerminalData, SshTerminalExit } from '@events'
import { describeError } from '@/lib/errors'
import {
  createPhaseTracker,
  ensureSSHConnected,
} from '@/lib/sshConnectProgress'
import {
  EditorTabKind,
  type LayoutNode,
  editor,
} from '@/lib/editorLayout.svelte'
import { logger } from '@/lib/logger'
import {
  emptyConnectProgress,
  SSHConnectPhase,
  SSHTerminalStatus,
  type SSHConnectProgress,
} from '@/types/ssh'

export interface TerminalEntry {
  paneId: string
  connectionId: string
  host: HTMLDivElement
  term: Terminal
  fitAddon: FitAddon
  sessionId: string | null
  view: TerminalView
  unlistenData: (() => void) | null
  unlistenExit: (() => void) | null
}

export interface TerminalView {
  status: SSHTerminalStatus
  error: string
  progress: SSHConnectProgress
}

const orphansClosed: Promise<void> = GoTerminal.CloseAll().catch((error) =>
  logger.error('Failed to close orphaned SSH terminals', error),
)

// eslint-disable-next-line svelte/prefer-svelte-reactivity
const registry = new Map<string, TerminalEntry>()

class TerminalsStore {
  statuses = $state.raw<Record<string, SSHTerminalStatus>>({})
}

export const terminals = new TerminalsStore()

const setStatus = (entry: TerminalEntry, status: SSHTerminalStatus) => {
  entry.view.status = status
  terminals.statuses = { ...terminals.statuses, [entry.paneId]: status }
}

const themeColor = (name: string, fallback: string): string => {
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim()
  return value || fallback
}

const rgbTuple = (name: string, fallback: string): string => {
  return themeColor(name, fallback).trim().split(/\s+/).join(', ')
}

const buildTerminalTheme = () => ({
  background: themeColor('--color-surface-canvas', '#07070b'),
  foreground: themeColor('--color-foreground', '#fafafa'),
  cursor: themeColor('--color-primary', '#12ce90'),
  cursorAccent: themeColor('--color-surface-canvas', '#07070b'),
  selectionBackground: `rgba(${rgbTuple('--color-primary-rgb', '18 206 144')}, 0.28)`,
  black: themeColor('--color-surface-base', '#0a0a0f'),
  red: '#ff6b6b',
  green: themeColor('--color-primary', '#12ce90'),
  yellow: themeColor('--color-warning', '#f4c23a'),
  blue: '#3b82f6',
  magenta: '#a566ff',
  cyan: '#22d3ee',
  white: '#d4d4d8',
  brightBlack: '#52525b',
  brightRed: '#ff8787',
  brightGreen: '#3ee6ab',
  brightYellow: '#ffd454',
  brightBlue: '#60a5fa',
  brightMagenta: '#c4a0ff',
  brightCyan: '#67e8f9',
  brightWhite: themeColor('--color-foreground', '#fafafa'),
})

const decode = (b64: string): Uint8Array =>
  Uint8Array.from(atob(b64), (char) => char.charCodeAt(0))

const createEntry = (paneId: string, connectionId: string): TerminalEntry => {
  const host = document.createElement('div')
  host.className = 'h-full w-full'

  const term = new Terminal({
    fontFamily:
      '"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace',
    fontSize: 13,
    lineHeight: 1.2,
    cursorBlink: true,
    allowProposedApi: true,
    theme: buildTerminalTheme(),
  })

  const fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(host)

  const view = $state<TerminalView>({
    status: SSHTerminalStatus.Connecting,
    error: '',
    progress: emptyConnectProgress(),
  })

  const entry: TerminalEntry = {
    paneId,
    connectionId,
    host,
    term,
    fitAddon,
    sessionId: null,
    view,
    unlistenData: null,
    unlistenExit: null,
  }
  setStatus(entry, SSHTerminalStatus.Connecting)

  term.onData((data) => {
    if (entry.sessionId && entry.view.status === SSHTerminalStatus.Open) {
      void GoTerminal.Write(entry.sessionId, data)
    }
  })

  return entry
}

export const acquireTerminal = (
  paneId: string,
  connectionId: string,
): TerminalEntry => {
  const existing = registry.get(paneId)
  if (existing) return existing

  const entry = createEntry(paneId, connectionId)
  registry.set(paneId, entry)
  void startSession(entry)
  return entry
}

const teardownStream = (entry: TerminalEntry): void => {
  entry.unlistenData?.()
  entry.unlistenData = null
  entry.unlistenExit?.()
  entry.unlistenExit = null
  if (entry.sessionId) {
    void GoTerminal.Close(entry.sessionId)
    entry.sessionId = null
  }
}

type StreamEvent =
  | { kind: 'data'; sessionId: string; chunk: string }
  | { kind: 'exit'; sessionId: string }

const listen = (entry: TerminalEntry): ((sessionId: string) => void) => {
  let backlog: StreamEvent[] | null = []

  const deliver = (event: StreamEvent) => {
    if (event.sessionId !== entry.sessionId) return
    if (event.kind === 'data') {
      entry.term.write(decode(event.chunk))
      return
    }
    setStatus(entry, SSHTerminalStatus.Closed)
    entry.term.write('\r\n\x1b[2msession closed\x1b[0m\r\n')
  }

  const handle = (event: StreamEvent) => {
    if (backlog) backlog.push(event)
    else deliver(event)
  }

  entry.unlistenData = Events.On(SshTerminalData, ({ data }) =>
    handle({ kind: 'data', sessionId: data.sessionId, chunk: data.data ?? '' }),
  )
  entry.unlistenExit = Events.On(SshTerminalExit, ({ data }) =>
    handle({ kind: 'exit', sessionId: data.sessionId }),
  )

  return (sessionId) => {
    entry.sessionId = sessionId
    const pending = backlog ?? []
    backlog = null
    pending.forEach(deliver)
  }
}

const startSession = async (entry: TerminalEntry): Promise<void> => {
  setStatus(entry, SSHTerminalStatus.Connecting)
  entry.view.error = ''

  const phases = createPhaseTracker(
    (progress) => (entry.view.progress = progress),
  )
  try {
    phases.begin(SSHConnectPhase.Channel)
    await ensureSSHConnected(entry.connectionId)
    phases.settle()

    phases.begin(SSHConnectPhase.Session)
    const attach = listen(entry)
    const cols = entry.term.cols || 80
    const rows = entry.term.rows || 24
    await orphansClosed
    const sessionId = await GoTerminal.Open(entry.connectionId, cols, rows)

    if (
      registry.get(entry.paneId) !== entry ||
      entry.view.status !== SSHTerminalStatus.Connecting
    ) {
      void GoTerminal.Close(sessionId)
      return
    }
    phases.settle()

    setStatus(entry, SSHTerminalStatus.Open)
    attach(sessionId)
  } catch (error) {
    teardownStream(entry)
    phases.fail()
    setStatus(entry, SSHTerminalStatus.Error)
    entry.view.error = describeError(error)
    logger.error('Failed to open SSH terminal', error)
  }
}

export const reconnectTerminal = async (paneId: string): Promise<void> => {
  const entry = registry.get(paneId)
  if (!entry) return
  teardownStream(entry)
  entry.term.clear()
  await startSession(entry)
}

export const fitTerminal = (entry: TerminalEntry): void => {
  if (entry.host.offsetWidth === 0 || entry.host.offsetHeight === 0) return
  try {
    entry.fitAddon.fit()
  } catch {
    return
  }
  if (entry.sessionId) {
    void GoTerminal.Resize(entry.sessionId, entry.term.cols, entry.term.rows)
  }
}

const disposeTerminal = (paneId: string): void => {
  const entry = registry.get(paneId)
  if (!entry) return
  registry.delete(paneId)
  const { [paneId]: _closed, ...remaining } = terminals.statuses
  terminals.statuses = remaining
  teardownStream(entry)
  try {
    entry.term.dispose()
  } catch (error) {
    logger.debug('Terminal dispose failed', error)
  }
  entry.host.remove()
}

const collectTerminalPaneIds = (node: LayoutNode, into: Set<string>): void => {
  if (node.type === 'group') {
    for (const tab of node.tabs) {
      if (tab.kind === EditorTabKind.SSHTerminal) into.add(tab.id)
    }
    return
  }
  node.children.forEach((child) => collectTerminalPaneIds(child, into))
}

$effect.root(() => {
  $effect(() => {
    const live = new Set<string>()
    collectTerminalPaneIds(editor.layout.root, live)
    untrack(() => {
      for (const paneId of [...registry.keys()]) {
        if (!live.has(paneId)) disposeTerminal(paneId)
      }
    })
  })
})
