<script lang="ts">
  import { ArrowRight } from 'lucide-svelte'
  import AuthFormCard from '@/components/auth/AuthFormCard.svelte'
  import AuthPageHeader from '@/components/auth/AuthPageHeader.svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import FormField from '@/components/ui/FormField.svelte'
  import PasswordField from '@/components/ui/PasswordField.svelte'
  import {
    EMAIL_MAX_LENGTH,
    NAME_MAX_LENGTH,
    PASSWORD_MIN_LENGTH,
  } from '@/constants'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { register } from '@/lib/auth'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import { navigate } from '@/lib/router'

  let firstName = ''
  let lastName = ''
  let email = ''
  let password = ''
  let confirmPassword = ''
  let acceptTerms = false
  let loading = false
  let errorMessage = ''

  const validate = () => {
    if (!firstName.trim()) return t('validation.firstNameRequired')
    if (!lastName.trim()) return t('validation.lastNameRequired')
    if (firstName.length > NAME_MAX_LENGTH || lastName.length > NAME_MAX_LENGTH)
      return t('validation.unknown')
    if (!email.trim()) return t('validation.emailRequired')
    if (
      email.length > EMAIL_MAX_LENGTH ||
      !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
    )
      return t('validation.emailInvalid')
    if (password.length < PASSWORD_MIN_LENGTH)
      return t('validation.passwordMin')
    if (!confirmPassword.trim()) return t('validation.confirmPasswordRequired')
    if (password !== confirmPassword) return t('validation.passwordsMismatch')
    if (!acceptTerms) return t('validation.acceptTermsRequired')
    return ''
  }

  const handleSubmit = async (event: SubmitEvent) => {
    event.preventDefault()
    errorMessage = validate()
    if (errorMessage) return

    loading = true
    const resp = await register({ email, firstName, lastName, password })
    loading = false

    if (!resp.success) {
      errorMessage = getResponseMessage(resp, t('errors.UNKNOWN'))
      return
    }

    navigate('/login')
  }
</script>

<AuthLayout>
  <div class="w-full space-y-6">
    <AuthPageHeader
      title={$translate('auth.register.title')}
      subtitle={$translate('auth.register.subtitle')}
    />

    {#if errorMessage}
      <Alert variant="destructive" message={errorMessage} />
    {/if}

    <AuthFormCard>
      <form class="space-y-4" on:submit={handleSubmit}>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <FormField
            id="firstName"
            label={$translate('auth.register.firstName')}
            bind:value={firstName}
          />
          <FormField
            id="lastName"
            label={$translate('auth.register.lastName')}
            bind:value={lastName}
          />
        </div>

        <FormField
          id="email"
          type="email"
          label={$translate('auth.register.email')}
          bind:value={email}
        />

        <PasswordField
          id="password"
          label={$translate('auth.register.password')}
          bind:value={password}
        />

        <FormField
          id="confirmPassword"
          type="password"
          mono
          label={$translate('auth.register.confirmPassword')}
          bind:value={confirmPassword}
        />

        <label
          class="flex cursor-pointer items-start gap-2.5 text-sm leading-snug text-muted-foreground"
        >
          <input
            type="checkbox"
            bind:checked={acceptTerms}
            class="mt-1 h-4 w-4"
          />
          <span>{$translate('auth.register.acceptTerms')}</span>
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
              ? $translate('auth.register.submitting')
              : $translate('auth.register.submit')}
          </span>
          {#if !loading}
            <ArrowRight class="h-4 w-4" />
          {/if}
        </Button>
      </form>
    </AuthFormCard>

    <p class="text-center text-sm text-muted-foreground">
      {$translate('auth.register.hasAccount')}
      <button
        type="button"
        class="ml-1 font-medium text-primary hover:text-primary/80"
        on:click={() => navigate('/login')}
      >
        {$translate('auth.register.signIn')}
      </button>
    </p>
  </div>
</AuthLayout>
