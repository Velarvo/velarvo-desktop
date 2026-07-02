import { Circle, CircleCheck, CircleDot } from 'lucide-svelte'
import { isSSHConnected, type SSHConnection } from '@/types/ssh'

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

export const sshStatusLabel = (connection: SSHConnection): string => {
  if (isSSHConnected(connection)) return 'Connected'
  if (connection.hasPassword) return 'Ready'
  return 'Needs auth'
}

export const sshStatusIcon = (connection: SSHConnection): IconComponent => {
  if (isSSHConnected(connection)) return CircleCheck
  if (connection.hasPassword) return CircleDot
  return Circle
}

export const formatRelativeTime = (micros?: number): string => {
  if (!micros) return 'Never'
  const diffMs = Date.now() - micros / 1000
  if (diffMs < 0) return 'Just now'
  const minutes = Math.floor(diffMs / 60000)
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes} min`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} h`
  const days = Math.floor(hours / 24)
  return `${days} d`
}
