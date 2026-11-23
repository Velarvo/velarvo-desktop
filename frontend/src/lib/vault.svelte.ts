import { Service as Vault, type State } from '@bindings/vault'
import { clearProjects, refreshProjects } from '@/lib/projects.svelte'
import { describeError } from '@/lib/errors'
import { logger } from '@/lib/logger'

export type VaultState = State

class VaultStore {
  state = $state<VaultState | null>(null)
  loading = $state(true)
  error = $state<string | null>(null)
  isSetup = $derived(this.state?.isSetup ?? false)
  isUnlocked = $derived(this.state?.isUnlocked ?? false)
}

export const vault = new VaultStore()

const apply = (state: VaultState): void => {
  vault.state = state
  if (!state.isUnlocked) clearProjects()
}

export const bootstrapVault = async (): Promise<void> => {
  vault.loading = true
  vault.error = null
  try {
    apply(await Vault.State())
  } catch (error) {
    logger.error('Failed to read vault state', error)
    vault.state = null
    vault.error = describeError(error)
  } finally {
    vault.loading = false
  }
}

export const setupVault = async (masterPassword: string): Promise<void> => {
  apply(await Vault.Setup(masterPassword))
  await refreshProjects()
}

export const unlockVault = async (masterPassword: string): Promise<void> => {
  apply(await Vault.Unlock(masterPassword))
  await refreshProjects()
}

export const lockVault = async (): Promise<void> => {
  apply(await Vault.Lock())
}
