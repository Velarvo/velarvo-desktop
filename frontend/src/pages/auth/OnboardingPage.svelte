<script lang="ts">
  import { ArrowRight } from 'lucide-svelte'
  import AuthFormCard from '@/components/auth/AuthFormCard.svelte'
  import AuthPageHeader from '@/components/auth/AuthPageHeader.svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import PasswordField from '@/components/ui/PasswordField.svelte'
  import { MASTER_PASSWORD_MIN_LENGTH } from '@/constants'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import { navigate } from '@/lib/router'
  import { setupVault } from '@/lib/vault'

  let password = ''
  let confirmPassword = ''
  let errorMessage = ''
  let isSubmitting = false

  const handleCreateVault = async (event: SubmitEvent) => {
    event.preventDefault()
    errorMessage = ''

    if (password.length < MASTER_PASSWORD_MIN_LENGTH) {
      errorMessage = t('auth.onboarding.validation.passwordMin')
      return
    }

    if (!confirmPassword) {
      errorMessage = t('validation.confirmPasswordRequired')
      return
    }

    if (password !== confirmPassword) {
      errorMessage = t('validation.passwordsMismatch')
      return
    }

    isSubmitting = true
    const resp = await setupVault(password)
    try {
      if (!resp.success) {
        errorMessage = getResponseMessage(resp, t('errors.VAULT_SETUP_FAILED'))
        return
      }

      password = ''
      confirmPassword = ''
      navigate('/dashboard')
    } finally {
      isSubmitting = false
    }
  }
</script>

<AuthLayout>
  <div class="mx-auto w-full max-w-lg space-y-6">
    <AuthPageHeader
      title={$translate('auth.onboarding.title')}
      subtitle={$translate('auth.onboarding.subtitle')}
    />

    {#if errorMessage}
      <Alert variant="destructive" message={errorMessage} />
    {/if}

    <AuthFormCard>
      <form class="space-y-4" on:submit={handleCreateVault}>
        <PasswordField
          id="password"
          label={$translate('auth.onboarding.masterPassword')}
          placeholder={$translate('auth.onboarding.passwordPlaceholder')}
          hint={$translate('auth.onboarding.validation.passwordMin')}
          bind:value={password}
        />

        <PasswordField
          id="confirmPassword"
          label={$translate('auth.onboarding.confirmPassword')}
          placeholder={$translate('auth.onboarding.confirmPasswordPlaceholder')}
          bind:value={confirmPassword}
        />

        <Button
          type="submit"
          variant="primary"
          radius="md"
          fullWidth
          loading={isSubmitting}
          class="mt-1"
        >
          <span>
            {isSubmitting
              ? $translate('auth.onboarding.creatingVault')
              : $translate('auth.onboarding.createVault')}
          </span>
          {#if !isSubmitting}
            <ArrowRight class="h-4 w-4" />
          {/if}
        </Button>
      </form>
    </AuthFormCard>
  </div>
</AuthLayout>
