import {
  Service as Settings,
  type Language as SettingsLanguage,
} from '@bindings/settings'
import i18n, { isLanguageCode, LANGUAGES, type LanguageCode } from '@/i18n'
import { logger } from '@/lib/logger'

let language = $state<LanguageCode>(
  isLanguageCode(i18n.resolvedLanguage) ? i18n.resolvedLanguage : 'en',
)

i18n.on('languageChanged', (code) => {
  language = isLanguageCode(code) ? code : 'en'
})

export const getLanguage = (): LanguageCode => language

export const t = (key: string, defaultValue?: string): string =>
  i18n.t(key, { defaultValue, lng: language })

const applyLanguage = async (code: string) => {
  if (isLanguageCode(code) && i18n.resolvedLanguage !== code) {
    await i18n.changeLanguage(code)
  }
}

export const bootstrapLanguage = async () => {
  try {
    const settings = await Settings.Get()
    await applyLanguage(settings.language)
  } catch (error) {
    logger.error('Failed to load settings', error)
  }
}

export const setLanguage = async (code: LanguageCode) => {
  const settings = await Settings.SetLanguage(code as SettingsLanguage)
  await applyLanguage(settings.language)
}

export { LANGUAGES }
