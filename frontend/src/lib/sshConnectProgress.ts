import { connectSSH, sshState } from '@/lib/sshConnections.svelte'
import {
  emptyConnectProgress,
  type SSHConnectPhase,
  type SSHConnectProgress,
} from '@/types/ssh'

export interface PhaseTracker {
  begin(phase: SSHConnectPhase): void
  settle(): void
  fail(): void
}

export const createPhaseTracker = (
  publishTo: (value: SSHConnectProgress) => void,
): PhaseTracker => {
  let progress: SSHConnectProgress = emptyConnectProgress()
  let mark = performance.now()

  const publish = () => publishTo({ ...progress })
  publish()

  return {
    begin(phase) {
      mark = performance.now()
      progress = { ...progress, current: phase }
      publish()
    },
    settle() {
      const phase = progress.current
      if (!phase) return
      progress = {
        ...progress,
        current: null,
        timings: {
          ...progress.timings,
          [phase]: Math.round(performance.now() - mark),
        },
      }
      publish()
    },
    fail() {
      progress = { ...progress, failedAt: progress.current, current: null }
      publish()
    },
  }
}

export const ensureSSHConnected = async (
  connectionId: string,
): Promise<void> => {
  if (sshState.connectedIds.has(connectionId)) return
  await connectSSH(connectionId)
}
