<script lang="ts">
  import { getLanguage, LANGUAGES, setLanguage, t } from '@/lib/i18n.svelte'
  import { logger } from '@/lib/logger'
  import SegmentedControl from '@/components/ui/SegmentedControl.svelte'
  import SettingsRow from './SettingsRow.svelte'

  const choose = (code: (typeof LANGUAGES)[number]['code']) => {
    setLanguage(code).catch((error) =>
      logger.error('Failed to save language', error),
    )
  }
</script>

<SettingsRow
  title={t('settings.preferences.language')}
  description={t('settings.preferences.languageDescription')}
>
  <SegmentedControl
    label={t('settings.preferences.language')}
    value={getLanguage()}
    options={LANGUAGES.map((language) => ({
      value: language.code,
      label: language.label,
    }))}
    onChange={choose}
  />
</SettingsRow>
