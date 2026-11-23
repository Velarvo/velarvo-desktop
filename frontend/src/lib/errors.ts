import { CancelError } from '@wailsio/runtime'
import { t } from '@/lib/i18n.svelte'

export const describeError = (error: unknown): string => {
  const code = codeOf(error)
  if (code) {
    const key = `errors.${code}`
    const translated = t(key)
    if (translated !== key) return translated
  }
  if (error instanceof Error && error.message) return error.message
  if (typeof error === 'string' && error) return error
  return t('errors.UNKNOWN')
}

const codeOf = (error: unknown): string | undefined => {
  const cause = error instanceof Error ? error.cause : undefined
  if (cause && typeof cause === 'object' && 'code' in cause) {
    return String(cause.code)
  }
  return undefined
}

export const isCancelled = (error: unknown): boolean =>
  error instanceof CancelError
