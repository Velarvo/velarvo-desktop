import { derived, writable } from 'svelte/store'
import { GetAuthState, Login, Logout, Register } from '../../wailsjs/go/app/App'
import type { APIResponse } from '@/types/api'
import type { LoginCredentials, RegisterData } from '@/types/auth'
import { toClientErrorResponse } from '@/lib/errors'
import { logger } from '@/lib/logger'
import { clearProjects } from '@/lib/projects'

export interface User {
  userId: string
  email: string
  firstName: string
  lastName: string
}

export const user = writable<User | null>(null)
export const isLoadingAuth = writable(true)
export const isAuth = derived(user, ($user) => Boolean($user))

const toUser = (payload: Record<string, string>): User => ({
  userId: payload.userId ?? '',
  email: payload.email ?? '',
  firstName: payload.firstName ?? '',
  lastName: payload.lastName ?? '',
})

const applyUser = (payload?: Record<string, string>) => {
  user.set(payload ? toUser(payload) : null)
}

export const bootstrapAuth = async () => {
  isLoadingAuth.set(true)
  try {
    const resp = await GetAuthState()
    if (resp.success && resp.data) {
      applyUser(resp.data as Record<string, string>)
    } else {
      applyUser()
    }
  } catch (error) {
    logger.warn('Auth bootstrap failed', error)
    applyUser()
  } finally {
    isLoadingAuth.set(false)
  }
}

export const login = async (
  credentials: LoginCredentials,
): Promise<APIResponse> => {
  try {
    const resp = await Login(credentials.email, credentials.password)
    if (resp.success && resp.data) {
      applyUser(resp.data as Record<string, string>)
    }
    return resp as APIResponse
  } catch (error) {
    logger.error('Login failed', error)
    return toClientErrorResponse(error)
  }
}

export const register = async (data: RegisterData): Promise<APIResponse> => {
  try {
    return (await Register(
      data.email,
      data.firstName,
      data.lastName,
      data.password,
    )) as APIResponse
  } catch (error) {
    logger.error('Registration failed', error)
    return toClientErrorResponse(error)
  }
}

export const logout = async (): Promise<APIResponse> => {
  try {
    const resp = (await Logout()) as APIResponse
    applyUser()
    clearProjects()
    return resp
  } catch (error) {
    logger.warn('Logout failed', error)
    applyUser()
    clearProjects()
    return toClientErrorResponse(error)
  }
}
