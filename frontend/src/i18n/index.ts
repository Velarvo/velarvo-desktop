import i18n from 'i18next'

import cs from './locales/cs.json'
import en from './locales/en.json'

export const defaultNS = 'translation'
export const resources = {
  cs: { translation: cs },
  en: { translation: en },
} as const

export const LANGUAGES = [
  { code: 'en', label: 'English', flag: 'EN' },
  { code: 'cs', label: 'Čeština', flag: 'CS' },
] as const

export type LanguageCode = (typeof LANGUAGES)[number]['code']

export const isLanguageCode = (
  value: string | null | undefined,
): value is LanguageCode =>
  LANGUAGES.some((language) => language.code === value)

const detectBrowserLanguage = (): LanguageCode => {
  if (typeof navigator !== 'undefined') {
    const primary = navigator.language.toLowerCase().split('-')[0]
    if (isLanguageCode(primary)) {
      return primary
    }
  }
  return 'en'
}

i18n.init({
  resources,
  defaultNS,
  lng: detectBrowserLanguage(),
  fallbackLng: 'en',
  supportedLngs: LANGUAGES.map((language) => language.code),
  interpolation: { escapeValue: false },
})

export default i18n
