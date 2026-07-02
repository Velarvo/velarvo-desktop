export const SSH_DEFAULT_PORT = 22

export enum SSHConnectionStatus {
  Disconnected = 'disconnected',
  Connected = 'connected',
}

export enum SSHTerminalStatus {
  Connecting = 'connecting',
  Open = 'open',
  Closed = 'closed',
  Error = 'error',
}

export const TERMINAL_STATUS_INDICATOR: Record<SSHTerminalStatus, string> = {
  [SSHTerminalStatus.Connecting]: 'bg-warning animate-pulse',
  [SSHTerminalStatus.Open]:
    'bg-primary shadow-[0_0_6px_rgb(var(--color-primary-rgb)_/_0.7)]',
  [SSHTerminalStatus.Closed]: 'bg-muted-foreground',
  [SSHTerminalStatus.Error]: 'bg-destructive',
}

export interface SSHConnection {
  id: string
  workspaceId: string
  name: string
  host: string
  port: number
  username: string
  hasPassword: boolean
  os: string
  status: SSHConnectionStatus
  sortOrder: number
  lastUsedAt?: number
  createdAt: number
  updatedAt: number
  revision: string
}

export interface CreateSSHConnectionInput {
  name: string
  host: string
  port: number
  username: string
  password: string
}

export interface UpdateSSHConnectionInput extends CreateSSHConnectionInput {
  id: string
  clearPassword: boolean
}

export const isSSHConnected = (connection: SSHConnection): boolean =>
  connection.status === SSHConnectionStatus.Connected
