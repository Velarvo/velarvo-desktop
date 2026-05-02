<script lang="ts">
  import { Eye, EyeOff, ArrowRight, AlertCircle, Loader2 } from 'lucide-svelte'
  import AuthLayout from '@/layout/AuthLayout.svelte'
  import { navigate } from '@/lib/router'
  import { login } from '@/lib/auth'
  import { t } from '@/lib/i18n'

  let email = ''
  let password = ''
  let showPassword = false
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
      errorMessage = t(`errors.${resp.code}`, t('errors.UNKNOWN'))
      return
    }
  }
</script>

<AuthLayout>
  <div class="w-full space-y-6">
    <div class="space-y-1 text-center">
      <h2 class="text-2xl sm:text-3xl font-bold tracking-tight text-white">
        {t('auth.login.title')}
      </h2>
      <p class="text-sm text-muted-foreground">{t('auth.login.subtitle')}</p>
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
        <div class="space-y-1.5">
          <label
            for="email"
            class="text-xs font-medium text-muted-foreground uppercase tracking-wide"
            >{t('auth.login.email')}</label
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
            >{t('auth.login.password')}</label
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

        <label
          class="flex items-center gap-2.5 text-sm text-muted-foreground cursor-pointer"
        >
          <input type="checkbox" bind:checked={rememberMe} class="h-4 w-4" />
          {t('auth.login.rememberMe')}
        </label>

        <button
          type="submit"
          disabled={loading}
          class="w-full h-11 rounded-md bg-primary text-primary-foreground font-semibold text-sm mt-1 disabled:opacity-60 inline-flex items-center justify-center gap-2"
        >
          {#if loading}
            <Loader2 class="h-4 w-4 animate-spin" />
            <span>{t('auth.login.submitting')}</span>
          {:else}
            <span>{t('auth.login.submit')}</span>
            <ArrowRight class="h-4 w-4" />
          {/if}
        </button>
      </form>
    </div>

    <p class="text-center text-sm text-muted-foreground">
      {t('auth.login.noAccount')}
      <button
        type="button"
        class="ml-1 font-medium text-primary hover:text-primary/80"
        on:click={() => navigate('/register')}
      >
        {t('auth.login.createAccount')}
      </button>
    </p>
  </div>
</AuthLayout>
