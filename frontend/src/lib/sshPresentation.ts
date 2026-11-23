import { Circle, CircleCheck, CircleDot } from '@lucide/svelte'
import { isSSHConnected, SSHKeyType, type SSHConnection } from '@/types/ssh'

const SSH_KEY_ALGORITHM_LABEL: Record<SSHKeyType, string> = {
  [SSHKeyType.Ed25519]: 'ssh-ed25519',
  [SSHKeyType.RSA4096]: 'ssh-rsa',
  [SSHKeyType.ECDSA384]: 'ecdsa-384',
}

const PRIVATE_KEY_HEADER = /^-----BEGIN (OPENSSH|RSA|EC|DSA) PRIVATE KEY-----/i

export const isPrivateKeyText = (text: string): boolean =>
  PRIVATE_KEY_HEADER.test(text.trim())

export const sshKeyAlgorithmLabel = (keyType: string): string =>
  SSH_KEY_ALGORITHM_LABEL[keyType as SSHKeyType] ?? keyType

type IconComponent = typeof Circle

export const matchesSSHQuery = (
  connection: SSHConnection,
  query: string,
): boolean => {
  if (!query) return true
  return [connection.name, connection.host, connection.username].some((value) =>
    value.toLowerCase().includes(query),
  )
}

const hasAuth = (connection: SSHConnection): boolean =>
  connection.hasPassword || Boolean(connection.keyId)

export const sshStatusLabel = (connection: SSHConnection): [string, string] => {
  if (isSSHConnected(connection)) return ['ssh.connected', 'Connected']
  if (hasAuth(connection)) return ['ssh.ready', 'Ready']
  return ['ssh.needsAuth', 'Needs auth']
}

export const sshStatusIcon = (connection: SSHConnection): IconComponent => {
  if (isSSHConnected(connection)) return CircleCheck
  if (hasAuth(connection)) return CircleDot
  return Circle
}

export const formatRelativeTime = (millis?: number): string => {
  if (!millis) return 'Never'
  const diffMs = Date.now() - millis
  if (diffMs < 0) return 'Just now'
  const minutes = Math.floor(diffMs / 60000)
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} h`
  const days = Math.floor(hours / 24)
  return `${days} d`
}
