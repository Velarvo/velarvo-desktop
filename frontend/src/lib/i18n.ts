import { derived, writable } from 'svelte/store'
import { GetSettings, SetUILanguage } from '../../wailsjs/go/app/App'
import i18n, { isLanguageCode, LANGUAGES, type LanguageCode } from '@/i18n'
import { logger } from '@/lib/logger'

export const t = (key: string, defaultValue?: string) => {
  return i18n.t(key, { defaultValue })
}

export const currentLanguage = writable<LanguageCode>(
  (i18n.resolvedLanguage as LanguageCode) || 'en',
)

export const translate = derived(currentLanguage, () => {
  return (key: string, defaultValue?: string) => {
    return i18n.t(key, { defaultValue })
  }
})

i18n.on('languageChanged', (code) => {
  currentLanguage.set((code as LanguageCode) || 'en')
})

export const bootstrapLanguage = async () => {
  try {
    const resp = await GetSettings()
    if (!resp.success || !resp.data) {
      logger.warn('Failed to load settings from backend', resp)
      return
    }

    const code = resp.data.language
    if (!isLanguageCode(code)) {
      return
    }

    if (i18n.resolvedLanguage !== code) {
      await i18n.changeLanguage(code)
    }
  } catch (error) {
    logger.error('Settings bootstrap failed', error)
  }
}

export const setLanguage = async (code: LanguageCode) => {
  const resp = await SetUILanguage(code)
  if (!resp.success || !resp.data) {
    logger.warn('Failed to persist language', resp)
    return
  }

  const persisted = resp.data.language
  if (isLanguageCode(persisted) && i18n.resolvedLanguage !== persisted) {
    await i18n.changeLanguage(persisted)
  }
}

export { LANGUAGES }
