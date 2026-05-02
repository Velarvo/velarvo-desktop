import i18n from 'i18next'

import en from './locales/en.json'

export const defaultNS = 'translation'
export const resources = {
  en: { translation: en },
} as const

export const LANGUAGES = [{ code: 'en', label: 'English', flag: 'EN' }] as const

export type LanguageCode = (typeof LANGUAGES)[number]['code']

i18n.init({
  resources,
  defaultNS,
  lng: 'en',
  fallbackLng: 'en',
  supportedLngs: ['en'],
  interpolation: { escapeValue: false },
})

export default i18n
