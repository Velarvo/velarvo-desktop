<script lang="ts">
  import { ArrowRight, AlertCircle, Eye, EyeOff } from 'lucide-svelte'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { MASTER_PASSWORD_MIN_LENGTH } from '@/constants'
  import { navigate } from '@/lib/router'
  import { t } from '@/lib/i18n'

  let showPassword = false
  let showConfirmPassword = false
  let password = ''
  let confirmPassword = ''
  let errorMessage = ''

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

    navigate('/dashboard')
  }
</script>

<AuthLayout>
  <div class="w-full max-w-lg mx-auto space-y-6">
    <div class="space-y-1 text-center">
      <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-white">
        {t('auth.onboarding.title')}
      </h2>
      <p class="text-sm text-muted-foreground">
        {t('auth.onboarding.subtitle')}
      </p>
    </div>

    {#if errorMessage}
      <div
        class="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2"
      >
        <AlertCircle class="h-4 w-4" />
        <span>{errorMessage}</span>
      </div>
    {/if}

    <div
      class="rounded-xl border border-border/70 bg-card/70 backdrop-blur-sm p-5 sm:p-6 shadow-xl"
    >
      <form class="space-y-4" on:submit={handleCreateVault}>
        <div class="space-y-1.5">
          <label
            for="password"
            class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
            >{t('auth.onboarding.masterPassword')}</label
          >
          <div class="relative">
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              class="h-11 w-full rounded-md border border-border bg-card px-3 pr-10 text-sm font-mono text-white"
              placeholder={t('auth.onboarding.passwordPlaceholder')}
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
          <p class="text-xs text-muted-foreground">
            {t('auth.onboarding.validation.passwordMin')}
          </p>
        </div>

        <div class="space-y-1.5">
          <label
            for="confirmPassword"
            class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
            >{t('auth.onboarding.confirmPassword')}</label
          >
          <div class="relative">
            <input
              id="confirmPassword"
              type={showConfirmPassword ? 'text' : 'password'}
              bind:value={confirmPassword}
              class="h-11 w-full rounded-md border border-border bg-card px-3 pr-10 text-sm font-mono text-white"
              placeholder={t('auth.onboarding.confirmPasswordPlaceholder')}
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-white"
              on:click={() => (showConfirmPassword = !showConfirmPassword)}
            >
              {#if showConfirmPassword}<EyeOff class="h-4 w-4" />{:else}<Eye
                  class="h-4 w-4"
                />{/if}
            </button>
          </div>
        </div>

        <button
          type="submit"
          class="w-full h-11 rounded-md bg-primary text-primary-foreground font-semibold text-sm mt-1 inline-flex items-center justify-center gap-2"
        >
          <span>{t('auth.onboarding.createVault')}</span>
          <ArrowRight class="h-4 w-4" />
        </button>
      </form>
    </div>
  </div>
</AuthLayout>
