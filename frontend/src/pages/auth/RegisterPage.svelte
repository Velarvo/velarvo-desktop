<script lang="ts">
  import { Eye, EyeOff, ArrowRight, AlertCircle, Loader2 } from 'lucide-svelte'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { navigate } from '@/lib/router'
  import { register } from '@/lib/auth'
  import { t } from '@/lib/i18n'
  import {
    EMAIL_MAX_LENGTH,
    NAME_MAX_LENGTH,
    PASSWORD_MIN_LENGTH,
  } from '@/constants'

  let firstName = ''
  let lastName = ''
  let email = ''
  let password = ''
  let confirmPassword = ''
  let acceptTerms = false
  let showPassword = false
  let loading = false
  let errorMessage = ''

  const validate = () => {
    if (!firstName.trim()) return t('validation.firstNameRequired')
    if (!lastName.trim()) return t('validation.lastNameRequired')
    if (firstName.length > NAME_MAX_LENGTH || lastName.length > NAME_MAX_LENGTH)
      return t('validation.UNKNOWN')
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
      errorMessage = t(`errors.${resp.code}`, t('errors.UNKNOWN'))
      return
    }

    navigate('/login')
  }
</script>

<AuthLayout>
  <div class="w-full space-y-6">
    <div class="space-y-1 text-center">
      <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-white">
        {t('auth.register.title')}
      </h2>
      <p class="text-sm text-muted-foreground">{t('auth.register.subtitle')}</p>
    </div>

    {#if errorMessage}
      <div
        class="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2"
      >
        <AlertCircle class="h-4 w-4" />
        <span>{errorMessage}</span>
      </div>
    {/if}

    <div class="rounded-xl bg-card/60 backdrop-blur-sm p-5 sm:p-6 shadow-xl">
      <form class="space-y-4" on:submit={handleSubmit}>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div class="space-y-1.5">
            <label
              for="firstName"
              class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
              >{t('auth.register.firstName')}</label
            >
            <input
              id="firstName"
              bind:value={firstName}
              class="h-11 w-full rounded-md border border-border bg-card px-3 text-sm text-white"
            />
          </div>
          <div class="space-y-1.5">
            <label
              for="lastName"
              class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
              >{t('auth.register.lastName')}</label
            >
            <input
              id="lastName"
              bind:value={lastName}
              class="h-11 w-full rounded-md border border-border bg-card px-3 text-sm text-white"
            />
          </div>
        </div>

        <div class="space-y-1.5">
          <label
            for="email"
            class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
            >{t('auth.register.email')}</label
          >
          <input
            id="email"
            type="email"
            bind:value={email}
            class="h-11 w-full rounded-md border border-border bg-card px-3 text-sm text-white"
          />
        </div>

        <div class="space-y-1.5">
          <label
            for="password"
            class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
            >{t('auth.register.password')}</label
          >
          <div class="relative">
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              class="h-11 w-full rounded-md border border-border bg-card px-3 pr-10 text-sm font-mono text-white"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-white"
              on:click={() => (showPassword = !showPassword)}
            >
              {#if showPassword}<EyeOff class="h-4 w-4" />{:else}<Eye
                  class="h-4 w-4"
                />{/if}
            </button>
          </div>
        </div>

        <div class="space-y-1.5">
          <label
            for="confirmPassword"
            class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
            >{t('auth.register.confirmPassword')}</label
          >
          <input
            id="confirmPassword"
            type="password"
            bind:value={confirmPassword}
            class="h-11 w-full rounded-md border border-border bg-card px-3 text-sm font-mono text-white"
          />
        </div>

        <label
          class="flex items-start gap-2.5 text-sm text-muted-foreground leading-snug cursor-pointer"
        >
          <input
            type="checkbox"
            bind:checked={acceptTerms}
            class="mt-1 h-4 w-4"
          />
          <span>{t('auth.register.acceptTerms')}</span>
        </label>

        <button
          type="submit"
          disabled={loading}
          class="w-full h-11 rounded-md bg-primary text-primary-foreground font-semibold text-sm mt-1 disabled:opacity-60 inline-flex items-center justify-center gap-2"
        >
          {#if loading}
            <Loader2 class="h-4 w-4 animate-spin" />
            <span>{t('auth.register.submitting')}</span>
          {:else}
            <span>{t('auth.register.submit')}</span>
            <ArrowRight class="h-4 w-4" />
          {/if}
        </button>
      </form>
    </div>

    <p class="text-center text-sm text-muted-foreground">
      {t('auth.register.hasAccount')}
      <button
        type="button"
        class="ml-1 font-medium text-primary hover:text-primary/80"
        on:click={() => navigate('/login')}
      >
        {t('auth.register.signIn')}
      </button>
    </p>
  </div>
</AuthLayout>
