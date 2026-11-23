<script lang="ts">
  import { onMount } from 'svelte'
  import { ArrowRight } from '@lucide/svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import Field from '@/components/ui/Field.svelte'
  import Input from '@/components/ui/Input.svelte'
  import { MASTER_PASSWORD_MIN_LENGTH } from '@/constants'
  import VaultLayout from '@/layout/VaultLayout.svelte'
  import { describeError } from '@/lib/errors'
  import { t } from '@/lib/i18n.svelte'
  import { navigate } from '@/lib/router.svelte'
  import { setupVault } from '@/lib/vault.svelte'

  let password = $state('')
  let confirmPassword = $state('')
  let passwordError = $state('')
  let confirmError = $state('')
  let errorMessage = $state('')
  let isSubmitting = $state(false)
  let passwordInput = $state<HTMLInputElement | null>(null)
  let confirmInput = $state<HTMLInputElement | null>(null)

  onMount(() => passwordInput?.focus())

  const handleCreateVault = async (event: SubmitEvent) => {
    event.preventDefault()
    if (isSubmitting) return
    passwordError = confirmError = errorMessage = ''

    if (password.length < MASTER_PASSWORD_MIN_LENGTH) {
      passwordError = t('vault.onboarding.validation.passwordMin')
      passwordInput?.focus()
      return
    }

    if (!confirmPassword) {
      confirmError = t('validation.confirmPasswordRequired')
      confirmInput?.focus()
      return
    }

    if (password !== confirmPassword) {
      confirmError = t('validation.passwordsMismatch')
      confirmInput?.select()
      return
    }

    isSubmitting = true
    try {
      await setupVault(password)
      password = ''
      confirmPassword = ''
      navigate('/dashboard')
    } catch (error) {
      errorMessage = describeError(error)
    } finally {
      isSubmitting = false
    }
  }
</script>

<VaultLayout>
  <form onsubmit={handleCreateVault} novalidate>
    <h1 class="text-xl font-semibold tracking-tight text-white">
      {t('vault.onboarding.title')}
    </h1>
    <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">
      {t('vault.onboarding.subtitle')}
    </p>

    <div class="mt-6 space-y-4">
      <Field
        label={t('vault.onboarding.masterPassword')}
        for="password"
        error={passwordError}
        hint={t('vault.onboarding.validation.passwordMin')}
      >
        <Input
          id="password"
          type="password"
          autocomplete="new-password"
          placeholder={t('vault.onboarding.passwordPlaceholder')}
          invalid={Boolean(passwordError)}
          bind:ref={passwordInput}
          bind:value={password}
          oninput={() => (passwordError = '')}
        />
      </Field>

      <Field
        label={t('vault.onboarding.confirmPassword')}
        for="confirmPassword"
        error={confirmError}
      >
        <Input
          id="confirmPassword"
          type="password"
          autocomplete="new-password"
          placeholder={t('vault.onboarding.confirmPasswordPlaceholder')}
          invalid={Boolean(confirmError)}
          bind:ref={confirmInput}
          bind:value={confirmPassword}
          oninput={() => (confirmError = '')}
        />
      </Field>

      {#if errorMessage}
        <Alert variant="destructive" message={errorMessage} />
      {/if}

      <Button type="submit" size="lg" fullWidth loading={isSubmitting}>
        {isSubmitting
          ? t('vault.onboarding.creatingVault')
          : t('vault.onboarding.createVault')}
        {#if !isSubmitting}
          <ArrowRight class="h-4 w-4" />
        {/if}
      </Button>
    </div>
  </form>
</VaultLayout>
