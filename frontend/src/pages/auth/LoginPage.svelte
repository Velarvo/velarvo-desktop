<script lang="ts">
  import { ArrowRight } from 'lucide-svelte'
  import AuthFormCard from '@/components/auth/AuthFormCard.svelte'
  import AuthPageHeader from '@/components/auth/AuthPageHeader.svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import FormField from '@/components/ui/FormField.svelte'
  import PasswordField from '@/components/ui/PasswordField.svelte'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { login } from '@/lib/auth'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import { navigate } from '@/lib/router'

  let email = ''
  let password = ''
  let rememberMe = false
  let loading = false
  let errorMessage = ''

  const handleSubmit = async (event: SubmitEvent) => {
    event.preventDefault()
    errorMessage = ''

    if (!email.trim()) {
      errorMessage = t('validation.emailRequired')
      return
    }

    if (!password.trim()) {
      errorMessage = t('validation.passwordRequired')
      return
    }

    loading = true
    const resp = await login({ email, password })
    loading = false

    if (!resp.success) {
      errorMessage = getResponseMessage(resp, t('errors.UNKNOWN'))
    }
  }
</script>

<AuthLayout>
  <div class="w-full space-y-6">
    <AuthPageHeader
      title={$translate('auth.login.title')}
      subtitle={$translate('auth.login.subtitle')}
    />

    {#if errorMessage}
      <Alert variant="destructive" message={errorMessage} />
    {/if}

    <AuthFormCard>
      <form class="space-y-4" on:submit={handleSubmit}>
        <FormField
          id="email"
          type="email"
          label={$translate('auth.login.email')}
          bind:value={email}
        />

        <PasswordField
          id="password"
          label={$translate('auth.login.password')}
          bind:value={password}
        />

        <label
          class="flex cursor-pointer items-center gap-2.5 text-sm text-muted-foreground"
        >
          <input type="checkbox" bind:checked={rememberMe} class="h-4 w-4" />
          {$translate('auth.login.rememberMe')}
        </label>

        <Button
          type="submit"
          variant="primary"
          radius="md"
          fullWidth
          {loading}
          class="mt-1"
        >
          <span>
            {loading
              ? $translate('auth.login.submitting')
              : $translate('auth.login.submit')}
          </span>
          {#if !loading}
            <ArrowRight class="h-4 w-4" />
          {/if}
        </Button>
      </form>
    </AuthFormCard>

    <p class="text-center text-sm text-muted-foreground">
      {$translate('auth.login.noAccount')}
      <button
        type="button"
        class="ml-1 font-medium text-primary hover:text-primary/80"
        on:click={() => navigate('/register')}
      >
        {$translate('auth.login.createAccount')}
      </button>
    </p>
  </div>
</AuthLayout>
