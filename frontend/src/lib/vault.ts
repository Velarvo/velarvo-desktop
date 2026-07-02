import { derived, writable } from 'svelte/store'
import {
  GetVaultState,
  LockVault,
  SetupVault,
  UnlockVault,
} from '../../wailsjs/go/app/App'
import type { APIResponse } from '@/types/api'
import { clearProjects, refreshProjects } from '@/lib/projects'
import { getResponseMessage } from '@/lib/errors'
import { t } from '@/lib/i18n'
import { logger } from '@/lib/logger'

export interface VaultState {
  isSetup: boolean
  isUnlocked: boolean
  autoLockSeconds: number
}

export const vaultState = writable<VaultState | null>(null)
export const isLoadingVault = writable(true)
export const vaultError = writable<string | null>(null)

export const isVaultSetup = derived(
  vaultState,
  ($vaultState) => $vaultState?.isSetup ?? false,
)

export const isVaultUnlocked = derived(
  vaultState,
  ($vaultState) => $vaultState?.isUnlocked ?? false,
)

const normalizeVaultState = (payload: unknown): VaultState => {
  const state = (payload ?? {}) as Record<string, unknown>
  return {
    isSetup: Boolean(state.isSetup),
    isUnlocked: Boolean(state.isUnlocked),
    autoLockSeconds: Number(state.autoLockSeconds ?? 0),
  }
}

const applyVaultState = (payload: unknown): VaultState => {
  const state = normalizeVaultState(payload)
  vaultState.set(state)

  if (!state.isUnlocked) {
    clearProjects()
  }

  return state
}

export const bootstrapVault = async (): Promise<APIResponse<VaultState>> => {
  isLoadingVault.set(true)
  vaultError.set(null)

  try {
    const resp = await GetVaultState()
    if (!resp.success || !resp.data) {
      vaultError.set(
        getResponseMessage(resp, t('errors.VAULT_READ_STATE_FAILED')),
      )
      vaultState.set(null)
      return resp as APIResponse<VaultState>
    }

    const state = applyVaultState(resp.data)
    return {
      ...(resp as APIResponse<VaultState>),
      data: state,
    }
  } catch (error) {
    logger.error('Vault bootstrap failed', error)
    const resp: APIResponse<VaultState> = {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
    vaultError.set(
      getResponseMessage(resp, t('errors.VAULT_READ_STATE_FAILED')),
    )
    vaultState.set(null)
    return resp
  } finally {
    isLoadingVault.set(false)
  }
}

export const setupVault = async (
  masterPassword: string,
): Promise<APIResponse<VaultState>> => {
  vaultError.set(null)

  try {
    const resp = await SetupVault({ masterPassword })
    if (!resp.success || !resp.data) {
      vaultError.set(getResponseMessage(resp, t('errors.VAULT_SETUP_FAILED')))
      return resp as APIResponse<VaultState>
    }

    const state = applyVaultState(resp.data)
    await refreshProjects()
    return {
      ...(resp as APIResponse<VaultState>),
      data: state,
    }
  } catch (error) {
    logger.error('Vault setup failed', error)
    const resp: APIResponse<VaultState> = {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
    vaultError.set(getResponseMessage(resp, t('errors.VAULT_SETUP_FAILED')))
    return resp
  }
}

export const unlockVault = async (
  masterPassword: string,
): Promise<APIResponse<VaultState>> => {
  vaultError.set(null)

  try {
    const resp = await UnlockVault({ masterPassword })
    if (!resp.success || !resp.data) {
      vaultError.set(getResponseMessage(resp, t('errors.VAULT_UNLOCK_FAILED')))
      return resp as APIResponse<VaultState>
    }

    const state = applyVaultState(resp.data)
    await refreshProjects()
    return {
      ...(resp as APIResponse<VaultState>),
      data: state,
    }
  } catch (error) {
    logger.error('Vault unlock failed', error)
    const resp: APIResponse<VaultState> = {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
    vaultError.set(getResponseMessage(resp, t('errors.VAULT_UNLOCK_FAILED')))
    return resp
  }
}

export const lockVault = async (): Promise<APIResponse<VaultState>> => {
  vaultError.set(null)

  try {
    const resp = await LockVault()
    if (!resp.success || !resp.data) {
      vaultError.set(getResponseMessage(resp, t('errors.VAULT_LOCK_FAILED')))
      return resp as APIResponse<VaultState>
    }

    const state = applyVaultState(resp.data)
    return {
      ...(resp as APIResponse<VaultState>),
      data: state,
    }
  } catch (error) {
    logger.error('Vault lock failed', error)
    const resp: APIResponse<VaultState> = {
      success: false,
      code: 'CLIENT_ERROR',
      error: error instanceof Error ? error.message : String(error),
    }
    vaultError.set(getResponseMessage(resp, t('errors.VAULT_LOCK_FAILED')))
    return resp
  }
}
