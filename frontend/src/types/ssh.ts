export const SSH_DEFAULT_PORT = 22

export enum SSHKeyType {
  Ed25519 = 'ed25519',
  RSA4096 = 'rsa4096',
  ECDSA384 = 'ecdsa384',
}

export enum SSHTerminalStatus {
  Connecting = 'connecting',
  Open = 'open',
  Closed = 'closed',
  Error = 'error',
}

export enum SSHConnectPhase {
  Channel = 'channel',
  Session = 'session',
}

export const SSH_CONNECT_PHASES: readonly SSHConnectPhase[] = [
  SSHConnectPhase.Channel,
  SSHConnectPhase.Session,
]

export interface SSHConnectProgress {
  current: SSHConnectPhase | null
  timings: Partial<Record<SSHConnectPhase, number>>
  failedAt: SSHConnectPhase | null
}

export const emptyConnectProgress = (): SSHConnectProgress => ({
  current: null,
  timings: {},
  failedAt: null,
})

export const TERMINAL_STATUS_TONE: Record<SSHTerminalStatus, string> = {
  [SSHTerminalStatus.Connecting]: 'text-warning',
  [SSHTerminalStatus.Open]: 'text-primary',
  [SSHTerminalStatus.Closed]: 'text-muted-foreground',
  [SSHTerminalStatus.Error]: 'text-red-400',
}

export type {
  Connection as SSHConnection,
  ConnectionInput as SSHConnectionInput,
  ImportedKey as ImportSSHKeyInput,
  Key as SSHKey,
  NewKey as CreateSSHKeyInput,
} from '@bindings/sshconn'

export const isSSHConnected = (connection: { connected: boolean }): boolean =>
  connection.connected
