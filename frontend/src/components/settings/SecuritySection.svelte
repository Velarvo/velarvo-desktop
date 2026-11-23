<script lang="ts">
  import { LockKeyhole } from '@lucide/svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import { describeError } from '@/lib/errors'
  import { t } from '@/lib/i18n.svelte'
  import { navigate } from '@/lib/router.svelte'
  import { lockVault, vault } from '@/lib/vault.svelte'
  import SettingsRow from './SettingsRow.svelte'

  interface Props {
    onBusy: (busy: boolean) => void
    onClose: () => void
  }

  let { onBusy, onClose }: Props = $props()

  let locking = $state(false)
  let errorMessage = $state('')

  let autoLock = $derived(
    vault.state?.autoLockSeconds
      ? `${Math.round(vault.state.autoLockSeconds / 60)} min`
      : t('common.off'),
  )

  const lock = async () => {
    locking = true
    onBusy(true)
    errorMessage = ''
    try {
      await lockVault()
      onClose()
      navigate('/unlock')
    } catch (error) {
      errorMessage = describeError(error)
    } finally {
      locking = false
      onBusy(false)
    }
  }
</script>

<SettingsRow
  title={t('settings.security.vault')}
  description={t('settings.security.vaultDescription')}
>
  <Button
    variant="secondary"
    size="sm"
    loading={locking}
    disabled={!vault.isUnlocked}
    onclick={lock}
  >
    {#if !locking}<LockKeyhole class="h-3.5 w-3.5" />{/if}
    {locking
      ? t('settings.security.locking')
      : t('settings.security.lockVault')}
  </Button>
</SettingsRow>

<SettingsRow
  title={t('settings.security.autoLock')}
  description={t('settings.security.autoLockDescription')}
>
  <span class="text-sm text-fg-muted tabular-nums">{autoLock}</span>
</SettingsRow>

<SettingsRow
  title={t('settings.security.unlockMethod')}
  description={t('settings.security.unlockMethodDescription')}
>
  <span class="text-sm text-fg-muted">
    {t('settings.security.masterPassword')}
  </span>
</SettingsRow>

{#if errorMessage}
  <div class="mt-4">
    <Alert variant="destructive" message={errorMessage} />
  </div>
{/if}
