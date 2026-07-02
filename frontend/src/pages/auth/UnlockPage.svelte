<script lang="ts">
  import { ArrowRight, KeyRound, ShieldCheck } from 'lucide-svelte'
  import AuthFormCard from '@/components/auth/AuthFormCard.svelte'
  import AuthPageHeader from '@/components/auth/AuthPageHeader.svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import FormField from '@/components/ui/FormField.svelte'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import { navigate } from '@/lib/router'
  import { unlockVault } from '@/lib/vault'

  let masterPassword = ''
  let errorMessage = ''
  let isSubmitting = false

  const handleUnlock = async (event: SubmitEvent) => {
    event.preventDefault()
    errorMessage = ''

    if (!masterPassword.trim()) {
      errorMessage = t('auth.unlock.validation.passwordRequired')
      return
    }

    isSubmitting = true
    const resp = await unlockVault(masterPassword)
    try {
      if (!resp.success) {
        errorMessage = getResponseMessage(resp, t('errors.VAULT_UNLOCK_FAILED'))
        return
      }

      masterPassword = ''
      navigate('/dashboard')
    } finally {
      isSubmitting = false
    }
  }
</script>

<AuthLayout>
  <div class="mx-auto w-full max-w-lg space-y-6">
    <AuthPageHeader
      icon={ShieldCheck}
      title={$translate('auth.unlock.title')}
      subtitle={$translate('auth.unlock.subtitle')}
    />

    {#if errorMessage}
      <Alert variant="destructive" message={errorMessage} />
    {/if}

    <AuthFormCard>
      <form class="space-y-4" on:submit={handleUnlock}>
        <FormField
          id="masterPassword"
          type="password"
          mono
          label={$translate('auth.unlock.masterPassword')}
          placeholder={$translate('auth.unlock.passwordPlaceholder')}
          bind:value={masterPassword}
        >
          <KeyRound slot="trailing" class="h-4 w-4 text-muted-foreground" />
        </FormField>

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
              ? $translate('auth.unlock.submitting')
              : $translate('auth.unlock.submit')}
          </span>
          {#if !isSubmitting}
            <ArrowRight class="h-4 w-4" />
          {/if}
        </Button>
      </form>
    </AuthFormCard>
  </div>
</AuthLayout>
