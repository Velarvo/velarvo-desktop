import i18n, { LANGUAGES, type LanguageCode } from '@/i18n'

export const t = (key: string, defaultValue?: string) => {
  return i18n.t(key, { defaultValue })
}

export const setLanguage = (code: LanguageCode) => {
  void i18n.changeLanguage(code)
}

export { LANGUAGES }
