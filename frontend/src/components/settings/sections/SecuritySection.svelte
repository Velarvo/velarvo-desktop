<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { LockKeyhole, ShieldCheck, TimerReset } from 'lucide-svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import InfoTile from '@/components/ui/InfoTile.svelte'
  import SectionHero from '@/components/ui/SectionHero.svelte'
  import StatusBadge from '@/components/ui/StatusBadge.svelte'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import { navigate } from '@/lib/router'
  import { lockVault, vaultState } from '@/lib/vault'

  const dispatch = createEventDispatcher<{ close: void; busy: boolean }>()

  let errorMessage = ''
  let isLocking = false

  $: dispatch('busy', isLocking)

  $: autoLockLabel = $vaultState?.autoLockSeconds
    ? `${Math.round($vaultState.autoLockSeconds / 60)} min`
    : $translate('common.off')

  const handleLockVault = async () => {
    isLocking = true
    errorMessage = ''

    const resp = await lockVault()
    if (!resp.success) {
      errorMessage = getResponseMessage(resp, t('errors.VAULT_LOCK_FAILED'))
      isLocking = false
      return
    }

    dispatch('close')
    navigate('/unlock')
    isLocking = false
  }
</script>

<section class="space-y-4">
  <header>
    <h3 class="text-xl font-semibold text-white">
      {$translate('settings.security.title')}
    </h3>
    <p class="mt-1 max-w-xl text-sm leading-6 text-muted-foreground">
      {$translate('settings.security.subtitle')}
    </p>
  </header>

  <SectionHero
    icon={ShieldCheck}
    title={$translate('settings.security.masterPasswordTitle')}
    description={$translate('settings.security.masterPasswordDescription')}
  >
    <div class="rounded-xl border border-white/8 bg-surface-security px-4 py-3">
      <p class="text-sm font-medium text-white">
        {$translate('settings.security.activeMethod')}
      </p>
      <p class="mt-1 text-xs leading-5 text-muted-foreground">
        {$translate('settings.security.activeMethodDescription')}
      </p>
    </div>
  </SectionHero>

  <div class="grid gap-3 sm:grid-cols-2">
    <InfoTile
      icon={LockKeyhole}
      tone="primary"
      title={$translate('settings.security.vault')}
      value={$vaultState?.isUnlocked
        ? $translate('common.unlocked')
        : $translate('common.locked')}
    >
      <StatusBadge
        slot="trailing"
        tone={$vaultState?.isUnlocked ? 'active' : 'muted'}
        label={$vaultState?.isUnlocked
          ? $translate('common.active')
          : $translate('common.locked')}
      />
    </InfoTile>

    <InfoTile
      icon={TimerReset}
      title={$translate('settings.security.autoLock')}
      value={autoLockLabel}
    />
  </div>

  <Button
    variant="secondary"
    fullWidth
    loading={isLocking}
    disabled={!$vaultState?.isUnlocked}
    on:click={handleLockVault}
  >
    {#if !isLocking}
      <LockKeyhole class="h-4 w-4" />
    {/if}
    {isLocking
      ? $translate('settings.security.locking')
      : $translate('settings.security.lockVault')}
  </Button>

  {#if errorMessage}
    <Alert variant="destructive" message={errorMessage} />
  {/if}
</section>
