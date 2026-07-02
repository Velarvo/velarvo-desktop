import { get, writable, type Writable } from 'svelte/store'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import {
  CloseSSHTerminal,
  OpenSSHTerminal,
  ResizeSSHTerminal,
  WriteSSHTerminal,
} from '../../wailsjs/go/app/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { unwrapResponse } from '@/lib/errors'
import { t } from '@/lib/i18n'
import { connectSSH, connectedSSHIds } from '@/lib/sshConnections'
import { EditorTabKind, editorState, type LayoutNode } from '@/lib/editorLayout'
import { logger } from '@/lib/logger'
import { SSHTerminalStatus } from '@/types/ssh'

export interface TerminalEntry {
  paneId: string
  connectionId: string
  host: HTMLDivElement
  term: Terminal
  fitAddon: FitAddon
  sessionId: string | null
  status: Writable<SSHTerminalStatus>
  error: Writable<string>
  unlistenData: (() => void) | null
  unlistenExit: (() => void) | null
  unsubStatus: (() => void) | null
}

const registry = new Map<string, TerminalEntry>()

export const terminalStatuses = writable<Record<string, SSHTerminalStatus>>({})

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

  const entry: TerminalEntry = {
    paneId,
    connectionId,
    host,
    term,
    fitAddon,
    sessionId: null,
    status: writable<SSHTerminalStatus>(SSHTerminalStatus.Connecting),
    error: writable(''),
    unlistenData: null,
    unlistenExit: null,
    unsubStatus: null,
  }

  entry.unsubStatus = entry.status.subscribe((value) => {
    terminalStatuses.update((map) => ({ ...map, [paneId]: value }))
  })

  term.onData((data) => {
    if (entry.sessionId && get(entry.status) === SSHTerminalStatus.Open) {
      void WriteSSHTerminal(entry.sessionId, data)
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

const ensureConnected = async (connectionId: string): Promise<void> => {
  if (get(connectedSSHIds).has(connectionId)) return
  await connectSSH(connectionId)
}

const teardownStream = (entry: TerminalEntry): void => {
  entry.unlistenData?.()
  entry.unlistenData = null
  entry.unlistenExit?.()
  entry.unlistenExit = null
  if (entry.sessionId) {
    void CloseSSHTerminal(entry.sessionId)
    entry.sessionId = null
  }
}

const startSession = async (entry: TerminalEntry): Promise<void> => {
  entry.status.set(SSHTerminalStatus.Connecting)
  entry.error.set('')
  try {
    await ensureConnected(entry.connectionId)

    const cols = entry.term.cols || 80
    const rows = entry.term.rows || 24
    const resp = await OpenSSHTerminal(entry.connectionId, cols, rows)
    const sessionId = unwrapResponse(
      resp,
      t('errors.SSH_CONNECT_FAILED'),
    ) as string

    entry.sessionId = sessionId
    entry.status.set(SSHTerminalStatus.Open)

    entry.unlistenData = EventsOn(
      `ssh:term:data:${sessionId}`,
      (payload: string) => entry.term.write(decode(payload)),
    )
    entry.unlistenExit = EventsOn(`ssh:term:exit:${sessionId}`, () => {
      entry.status.set(SSHTerminalStatus.Closed)
      entry.term.write('\r\n\x1b[2msession closed\x1b[0m\r\n')
    })
  } catch (error) {
    entry.status.set(SSHTerminalStatus.Error)
    entry.error.set(error instanceof Error ? error.message : String(error))
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
    void ResizeSSHTerminal(entry.sessionId, entry.term.cols, entry.term.rows)
  }
}

const disposeTerminal = (paneId: string): void => {
  const entry = registry.get(paneId)
  if (!entry) return
  registry.delete(paneId)
  entry.unsubStatus?.()
  terminalStatuses.update((map) => {
    const next = { ...map }
    delete next[paneId]
    return next
  })
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

editorState.subscribe((state) => {
  const live = new Set<string>()
  collectTerminalPaneIds(state.root, live)
  for (const paneId of [...registry.keys()]) {
    if (!live.has(paneId)) disposeTerminal(paneId)
  }
})
