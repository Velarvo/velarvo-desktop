import type { APIResponse } from '@/types/api'
import { t } from '@/lib/i18n'

export const toClientErrorResponse = <T = unknown>(
  error: unknown,
  code = 'CLIENT_ERROR',
): APIResponse<T> => {
  return {
    success: false,
    code,
    error: error instanceof Error ? error.message : String(error ?? ''),
  }
}

export const getErrorMessage = (
  error: unknown,
  fallback = t('errors.UNKNOWN'),
): string => {
  if (error instanceof Error && error.message) {
    return error.message
  }

  if (typeof error === 'string' && error.trim()) {
    return error
  }

  if (error && typeof error === 'object') {
    const maybeMessage = Reflect.get(error, 'message')
    if (typeof maybeMessage === 'string' && maybeMessage.trim()) {
      return maybeMessage
    }

    try {
      const serialized = JSON.stringify(error)
      if (serialized && serialized !== '{}') {
        return serialized
      }
    } catch {
      return fallback
    }
  }

  return fallback
}

export const getResponseMessage = <T = unknown>(
  response: APIResponse<T> | null | undefined,
  fallback = t('errors.UNKNOWN'),
): string => {
  if (!response) {
    return fallback
  }

  const translationKey = response.code ? `errors.${response.code}` : ''
  if (translationKey) {
    const translated = t(translationKey)
    if (translated !== translationKey) {
      return translated
    }
  }

  if (response.message?.trim()) {
    return response.message
  }

  if (response.error?.trim()) {
    return response.error
  }

  return fallback
}

export const unwrapResponse = <T>(
  response: APIResponse<T>,
  fallback = t('errors.UNKNOWN'),
): T => {
  if (response.success && response.data !== undefined) {
    return response.data
  }

  throw new Error(getResponseMessage(response, fallback))
}
