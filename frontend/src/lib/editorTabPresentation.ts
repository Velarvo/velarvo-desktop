import { FolderTree, Plug, SquareTerminal, Workflow } from '@lucide/svelte'
import { EditorTabKind } from '@/lib/editorLayout.svelte'

type IconComponent = typeof SquareTerminal

export const TAB_KIND_ICON: Record<EditorTabKind, IconComponent> = {
  [EditorTabKind.Session]: Workflow,
  [EditorTabKind.SSHConnectionForm]: Plug,
  [EditorTabKind.SSHTerminal]: SquareTerminal,
  [EditorTabKind.SFTP]: FolderTree,
}

export const TAB_KIND_LABEL: Record<EditorTabKind, [string, string]> = {
  [EditorTabKind.Session]: ['editor.kindSession', 'Session'],
  [EditorTabKind.SSHConnectionForm]: ['editor.kindConnection', 'Connection'],
  [EditorTabKind.SSHTerminal]: ['ssh.terminal', 'Terminal'],
  [EditorTabKind.SFTP]: ['sftp.files', 'Files'],
}
